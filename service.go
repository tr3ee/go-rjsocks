package rjsocks

import (
	"encoding/binary"
	"fmt"
	"log"
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
	echoChan        chan bool
	echoNo, echoKey uint32
	ctrlServeChan   chan bool
	cond            *sync.Cond
	advertise       []rune
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
		user:          []byte(usr),
		pass:          []byte(pass),
		device:        dev,
		adapter:       adap,
		handle:        hnd,
		echoChan:      make(chan bool, 2),
		ctrlServeChan: make(chan bool, 4),
		cond:          sync.NewCond(&sync.Mutex{}),
	}, nil
}

func (s *Service) Serve() error {
	failcount := uint(0)
	src := gopacket.NewPacketSource(s.handle.PcapHandle, layers.LayerTypeEthernet)
	in := src.Packets()
	// 定时任务：若40s没有任何操作，则重新发送start包
	go s.crontabStart(40 * time.Second)
	go s.KeepAlive()

	s.Log(SrvStatStart, "发送初始start包...")
	s.handle.SendStartPkt()
	stop := false
	for packet := range in {
		//非阻塞判断是否有控制信号传入
		select {
		case b := <-s.ctrlServeChan:
			// 收到下线信号,则发送Logoff通知
			if !b {
				s.handle.SendLogoffPkt()
				s.Log(SrvStatFailure, "手动禁用网络连接...")
				stop = true
			} else {
				s.handle.SendLogoffPkt()
				time.Sleep(1 * time.Second)
				s.handle.SendStartPkt()
				stop = false
				continue
			}
		default:
		}
		if stop {
			continue
		}

		pkt := packet.Layer(layers.LayerTypeEAP)
		if pkt == nil {
			continue
		}
		eap := pkt.(*layers.EAP)
		switch eap.Code {
		case layers.EAPCodeRequest:
			switch eap.Type {
			case layers.EAPTypeIdentity:
				eth := packet.Layer(layers.LayerTypeEthernet).(*layers.Ethernet)
				s.handle.SetDstMacAddr(eth.SrcMAC)
				s.handle.SendResponseIdentity(eap.Id, s.user)
				s.Log(SrvStatRespIdentity, "向认证端响应Identity包...")
			case layers.EAPTypeOTP:
				if len(eap.TypeData) >= 17 {
					seed := eap.TypeData[1:17]
					s.handle.SendResponseMD5Chall(eap.Id, seed, s.user, s.pass)
					s.Log(SrvStatRespMd5Chall, fmt.Sprintf("响应Md5-Challenge包，seed=%v", seed))
				}
			}
		case layers.EAPCodeSuccess:
			//TODO
			s.Log(SrvStatSuccess, "认证成功")
			if len(eap.TypeData) > 0x15e+4 {
				key := eap.TypeData[0x15e : 0x15e+4]
				Symmetric(key)
				echoKey := binary.BigEndian.Uint32(key)
				echoNo := uint32(0x102b)
				if err := reNewIP(s.adapter); err != nil {
					panic(err)
					s.Log(SrvStatError, err.Error())
				}
				s.StartEchoing(echoNo, echoKey)
			}

		case layers.EAPCodeFailure:
			// 平方退避 0 1 4 9 16 25.。。
			interval := time.Duration(failcount*failcount) * time.Second
			s.Log(SrvStatFailure, "认证失败，等待"+interval.String()+"后重试")
			failcount++
			s.StopEchoing()
			time.Sleep(interval)
			s.handle.SendStartPkt()
		}
	}
	return nil
}

func (s *Service) Log(stat SrvStat, msg string) {
	log.Println(msg)
	s.State = stat
	s.cond.Broadcast()

}

func (s *Service) ContinueServe() {
	s.ctrlServeChan <- true
}

func (s *Service) StopServe() {
	s.ctrlServeChan <- false
}

func (s *Service) WaitStat() {
	s.cond.L.Lock()
	defer s.cond.L.Unlock()
	s.cond.Wait()
}

func (s *Service) crontabStart(interval time.Duration) {
	c := make(chan bool, 1)
	go func() {
		for {
			s.WaitStat()
			c <- true
		}
	}()
	for {
		select {
		case <-time.After(interval):
			if err := s.handle.SendStartPkt(); err != nil {
				log.Println(err)
			}
			s.Log(SrvStatStart, "等待时间过长无状态改变，重新认证")
		case <-c:
			//代表在interval时间内有操作
		}
	}
}

func (s *Service) KeepAlive() {
	var n, k uint32
	enable := false
	for {
		select {
		case <-time.After(29 * time.Second):
			if enable {
				s.Log(SrvStatKeepAlive, fmt.Sprintf("发送心跳包以维持连接... (No=%d,Key=%d)", n, k))
				if err := s.handle.SendEchoPkt(n, k); err != nil {
					s.Log(SrvStatError, err.Error())
				}
				n++
			}
		case b := <-s.echoChan:
			if b {
				enable = true
				n, k = s.echoNo, s.echoKey
			} else {
				enable = false
			}
		}
	}
}

func (s *Service) StartEchoing(n, k uint32) {
	s.echoChan <- true
	s.echoNo, s.echoKey = n, k
}

func (s *Service) StopEchoing() {
	s.echoChan <- false
}

func (s *Service) Close() {
	s.handle.SendLogoffPkt()
	s.handle.Close()
}
