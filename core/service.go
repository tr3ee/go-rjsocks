package rjsocks

import (
	"encoding/binary"
	"sync"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

type SrvStat int

const (
	SrvStatStart = SrvStat(iota)
	SrvStatRespIdentity
	SrvStatRespMd5Chall
	SrvStatSuccess
	SrvStatFailure
	SrvStatKeepAlive
	SrvStatError
)

func (s SrvStat) String() string {
	switch s {
	case SrvStatStart:
		return "请求认证..."
	case SrvStatRespIdentity:
		return "开始认证..."
	case SrvStatRespMd5Chall:
		return "认证中..."
	case SrvStatKeepAlive:
		return "保持认证状态"
	case SrvStatSuccess:
		return "认证成功"
	case SrvStatFailure:
		return "认证失败"
	case SrvStatError:
		return "内部错误"
	}
	return "未知错误"
}

type Service struct {
	State           SrvStat
	user, pass      []byte
	device, adapter string
	handle          *Handle
	echoNo, echoKey uint32
	advertising     []rune
	chanPkt         chan gopacket.Packet
	threadLock      sync.Mutex
	crontab         *Crontab
	isClosed        bool
	isStopped       bool
}

func NewService(usr, pass, dev, adap string) (*Service, error) {
	ifc, err := SelectNetworkDev(dev)
	if err != nil {
		return nil, err
	}
	macAddr, err := SelectNetworkAdapter(adap)
	if err != nil {
		return nil, err
	}
	hnd, err := NewHandle(ifc, macAddr)
	if err != nil {
		return nil, err
	}
	return &Service{
		user:    []byte(usr),
		pass:    []byte(pass),
		device:  dev,
		adapter: adap,
		handle:  hnd,
		State:   SrvStatFailure,
		// chanPkt: make(chan gopacket.Packet, 1024),
		crontab: NewCrontab(),
	}, nil
}

func (s *Service) packets() (<-chan gopacket.Packet, error) {
	if s.chanPkt != nil {
		return s.chanPkt, nil
	}
	s.chanPkt = make(chan gopacket.Packet, 1000)
	go func() {
		defer close(s.chanPkt)
		src := gopacket.NewPacketSource(s.handle.PcapHandle, layers.LayerTypeEthernet)
		in := src.Packets()
		for packet := range in {
			if s.isClosed {
				break
			}
			pkt := packet.Layer(layers.LayerTypeEAP)
			if pkt == nil {
				continue
			}
			s.chanPkt <- packet
		}
	}()
	return s.chanPkt, nil
}

func (s *Service) Run() error {
	s.threadLock.Lock()
	defer s.threadLock.Unlock()
	go s.crontab.Run()
	s.crontab.ForceRegister("Monitor", NewCronItem(func() { s.handle.SendStartPkt() }, 40*time.Second))
	in, err := s.packets()
	if err != nil {
		return err
	}
	failcount := int64(0)
	s.handle.SendStartPkt()
	for packet := range in {
		if s.isClosed {
			break
		}
		if s.isStopped {
			s.crontab.UpdateLastAccess("Echo", time.Now())
			s.crontab.UpdateLastAccess("Monitor", time.Now())
			continue
		}
		eap := packet.Layer(layers.LayerTypeEAP).(*layers.EAP)
		switch eap.Code {
		case layers.EAPCodeRequest:
			switch eap.Type {
			case layers.EAPTypeIdentity:
				s.updateStat(SrvStatRespIdentity)
				eth := packet.Layer(layers.LayerTypeEthernet).(*layers.Ethernet)
				s.handle.SetDstMacAddr(eth.SrcMAC)
				if err := s.handle.SendResponseIdentity(eap.Id, s.user); err != nil {
					return err
				}
			case layers.EAPTypeOTP:
				s.updateStat(SrvStatRespMd5Chall)
				if len(eap.TypeData) >= 17 {
					seed := eap.TypeData[1:17]
					if err := s.handle.SendResponseMD5Chall(eap.Id, seed, s.user, s.pass); err != nil {
						return err
					}
				}
			}
		case layers.EAPCodeSuccess:
			s.updateStat(SrvStatSuccess)
			if len(eap.Contents) > 10 {
				pos := int(eap.Contents[9] + 0x8B)
				if len(eap.Contents) >= pos+4 {
					key := eap.Contents[pos : pos+4]
					Symmetric(key)
					s.echoKey = binary.BigEndian.Uint32(key)
					s.echoNo = uint32(0x102b)
					reNewIP(s.adapter)
					s.crontab.ForceRegister("Echo", NewCronItem(func() { s.updateStat(SrvStatKeepAlive); s.handle.SendEchoPkt(s.echoNo, s.echoKey); s.echoNo++ }, 30*time.Second))
				}
			}
		case layers.EAPCodeFailure:
			// 平方退避 0 1 4 9 16 25...
			s.updateStat(SrvStatFailure)
			interval := time.Duration(failcount*failcount) * time.Second
			failcount++
			s.crontab.Delete("Echo")
			time.Sleep(interval)
			if err := s.handle.SendStartPkt(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Service) updateStat(stat SrvStat) {
	s.crontab.UpdateLastAccess("Monitor", time.Now())
	s.State = stat
}

func (s *Service) Continue() {
	s.isStopped = false
	s.handle.SendStartPkt()
}

func (s *Service) Stop() {
	s.isStopped = true
	s.handle.SendLogoffPkt()
}

func (s *Service) Close() {
	s.handle.SendLogoffPkt()
	s.handle.Close()
	s.crontab.Close()
	s.isClosed = true
}
