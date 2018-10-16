package rjsocks

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"

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
	EventIdle = Event(iota)
	EventStart
	EventRespIdentity
	EventRespMd5Chall
	EventSuccess
	EventFailure
	EventKeepAlive
	EventError
)

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
	handle    *Handle
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
		handle:   handle,
	}
	return srv, nil
}

func (s *Service) nextEvent() (Event, error) {
	var pkt gopacket.Packet
	for {
		packet, err := s.pktSrc.NextPacket()
		if err != nil {
			return 0, err
		}

		_eth := packet.Layer(layers.LayerTypeEthernet)
		if _eth == nil {
			continue
		}
		eth := _eth.(*layers.Ethernet)
		if bytes.Compare(eth.DstMAC, s.nwinfo.Mac) != 0 {
			continue
		}

		eap := packet.Layer(layers.LayerTypeEAP)
		if eap != nil {
			pkt = packet
			break
		}
	}
	s.lastPkt = pkt
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

func (s *Service) handleEvent(e Event) error {
	switch e {
	case EventStart:
		if err := s.handle.SendStartPkt(); err != nil {
			return err
		}
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
	case EventFailure:
		// TODO
	case EventError:
		// TODO
	}
	s.lastEvent = e
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
	if err := s.handle.SendStartPkt(); err != nil {
		panic(err)
	}
	for {
		e := s.NextEvent()
		if e == EventIdle {
			fmt.Println(hex.Dump(s.lastPkt.Data()))
		}
		switch e {
		case EventStart:
			log.Println("Start")
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
