package rjsocks

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
)

/*
 Event represents the event currently obtained from the auth-server.
 the event table defined below.
*/
type Event int

/*
 all Event defined here.
*/
const (
	EventError = Event(iota)
	EventIdle
	EventRespIdentity
	EventRespMd5Chall
	EventSuccess
	EventFailure
)

const idleTimeout = 29 * time.Second

/*
 Service defines all the data required in the authentication process.
 You can get an instance by calling NewService().
*/
type Service struct {
	username  []byte
	password  []byte
	nwinfo    *NwAdapterInfo
	lastPkt   gopacket.Packet
	lastEvent Event
	pktSrc    *gopacket.PacketSource
	pktChan   chan gopacket.Packet
	handle    *Handle
	ads       string
	echoNo    uint32
	echoKey   uint32
}

func NewService(user, pass string, nwAdapterinfo *NwAdapterInfo) (*Service, error) {
	handle, err := NewHandle(fmt.Sprintf(`\Device\NPF_{%s}`, nwAdapterinfo.DeviceName), nwAdapterinfo.Mac)
	if err != nil {
		return nil, err
	}
	srv := &Service{
		username: []byte(user),
		password: []byte(pass),
		nwinfo:   nwAdapterinfo,
		pktSrc:   gopacket.NewPacketSource(handle.PcapHandle, layers.LayerTypeEthernet),
		pktChan:  make(chan gopacket.Packet, 1024),
		handle:   handle,
	}
	return srv, nil
}

func (s *Service) prePacket() {
	for packet := range s.pktSrc.Packets() {
		// am I the target?
		_eth := packet.Layer(layers.LayerTypeEthernet)
		if _eth == nil {
			continue
		}
		eth := _eth.(*layers.Ethernet)
		if bytes.Compare(eth.DstMAC, s.nwinfo.Mac) != 0 {
			continue
		}
		// try to decode as EAP layer.
		eap := packet.Layer(layers.LayerTypeEAP)
		if eap != nil {
			s.pktChan <- packet
		}
	}
}

func (s *Service) nextEvent() (Event, error) {
	var pkt gopacket.Packet
	select {
	case pkt = <-s.pktChan:
		s.lastPkt = pkt
	case <-time.After(idleTimeout):
		return EventIdle, nil
	}
	eap := pkt.Layer(layers.LayerTypeEAP).(*layers.EAP)
	switch eap.Code {
	case layers.EAPCodeRequest:
		switch eap.Type {
		case layers.EAPTypeIdentity:
			return EventRespIdentity, nil
		case layers.EAPTypeOTP:
			return EventRespMd5Chall, nil
		}
	case layers.EAPCodeSuccess:
		return EventSuccess, nil
	case layers.EAPCodeFailure:
		return EventFailure, nil
	}
	return EventIdle, nil
}

func (s *Service) NextEvent() Event {
	event, err := s.nextEvent()
	if err != nil {
		// TODO
		event = EventError
	}
	return event
}

func (s *Service) parseExdata(data []byte) ([]byte, uint32, error) {
	r := bufio.NewReader(bytes.NewReader(data))
	ads, err := parseMTLV(r)
	if err != nil {
		return nil, 0, err
	}
	utf8Ads, err := toUTF8(ads.Buffer)
	if err == nil {
		ads.Buffer = utf8Ads
	}
	_, err = r.Discard(0x7B + 6)
	if err != nil {
		return nil, 0, err
	}
	buf := make([]byte, 4)
	_, err = r.Read(buf)
	if err != nil {
		return nil, 0, err
	}
	symEncode(buf)
	key := binary.BigEndian.Uint32(buf)
	return ads.Buffer, key, nil
}

func (s *Service) handleEvent(e Event) error {
	defer func() { s.lastEvent = e }()
	switch e {
	case EventRespIdentity:
		eth := s.lastPkt.Layer(layers.LayerTypeEthernet).(*layers.Ethernet)
		eap := s.lastPkt.Layer(layers.LayerTypeEAP).(*layers.EAP)
		s.handle.SetDstMacAddr(eth.SrcMAC)
		if err := s.handle.SendResponseIdentity(eap.Id, s.username); err != nil {
			return err
		}
	case EventRespMd5Chall:
		eap := s.lastPkt.Layer(layers.LayerTypeEAP).(*layers.EAP)
		if eap.TypeData[0] == '\x10' && len(eap.TypeData) >= 17 {
			seed := eap.TypeData[1:17]
			if err := s.handle.SendResponseMD5Chall(eap.Id, seed, s.username, s.password); err != nil {
				return err
			}
		}
	case EventSuccess:
		// TODO
		eap := s.lastPkt.Layer(layers.LayerTypeEAP).(*layers.EAP)
		if len(eap.Contents) > 4 {
			ads, key, err := s.parseExdata(eap.Contents[4:])
			if err != nil {
				return err
			}
			s.ads = string(ads)
			s.echoKey = key
			s.echoNo = uint32(0x102b)
			log.Printf("adv: %s\nechoNo: %x\nechoKey: %x\n", s.ads, s.echoNo, s.echoKey)
			/*
				TODO: keep it alive on EventKeepAlive
					for {
						if err := s.handle.SendEchoPkt(echoNo, key); err != nil {
							panic(err)
						}
						echoNo++
						time.Sleep(30 * time.Second)
					}
			*/
		} else {
			return fmt.Errorf("packet corrupted: no enough data to parse on SUCCESS")
		}
	case EventIdle:
		if s.lastEvent == EventSuccess || s.lastEvent == EventIdle {
			if err := s.handle.SendEchoPkt(s.echoNo, s.echoKey); err != nil {
				return err
			}
			s.echoNo++
		} else {
			if err := s.handle.SendStartPkt(); err != nil {
				return err
			}
		}
	case EventFailure:
		// TODO
	case EventError:
		// TODO
	}
	return nil
}

func (s *Service) HandleEvent(e Event) {
	err := s.handleEvent(e)
	if err != nil {
		// TODO
		log.Println(err)
	}
}

func (s *Service) Run() {
	go s.prePacket()
	s.HandleEvent(EventIdle)
	for {
		e := s.NextEvent()
		switch e {
		case EventRespIdentity:
			log.Println("RequestIdentity")
		case EventRespMd5Chall:
			log.Println("RequestMd5Chall")
		case EventSuccess:
			log.Println("Success")
		case EventFailure:
			log.Println("Failure")
		case EventIdle:
			log.Println("Idle")
		default:
			log.Printf("unknown(%d)\n", e)
		}
		s.HandleEvent(e)
	}
}
