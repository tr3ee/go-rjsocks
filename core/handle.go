package rjsocks

import (
	"bytes"
	"crypto/md5"
	"encoding/binary"
	"net"

	"github.com/google/gopacket/layers"

	"github.com/google/gopacket"
	"github.com/google/gopacket/pcap"
)

var (
	DefaultSnaplen int32 = 1024
	// MultiCastAddr        = net.HardwareAddr{0x01, 0x80, 0xc2, 0x00, 0x00, 0x03}
	MultiCastAddr = net.HardwareAddr{0x01, 0xD0, 0xF8, 0x00, 0x00, 0x03}
)

type Handle struct {
	PcapHandle             *pcap.Handle
	srcMacAddr, dstMacAddr net.HardwareAddr
	buffer                 gopacket.SerializeBuffer
	options                gopacket.SerializeOptions
}

func NewHandle(dev *pcap.Interface, srcMacAddr net.HardwareAddr) (*Handle, error) {
	handler, err := pcap.OpenLive(dev.Name, DefaultSnaplen, false, pcap.BlockForever)
	if err != nil {
		return nil, err
	}
	h := &Handle{
		PcapHandle: handler,
		srcMacAddr: srcMacAddr,
		dstMacAddr: MultiCastAddr,
		buffer:     gopacket.NewSerializeBuffer(),
		options:    gopacket.SerializeOptions{FixLengths: false, ComputeChecksums: true},
	}
	return h, nil
}

// Close cleans up the pcap Handle.
func (h *Handle) Close() {
	h.PcapHandle.Close()
}

func (h *Handle) send(l ...gopacket.SerializableLayer) error {
	if err := gopacket.SerializeLayers(h.buffer, h.options, l...); err != nil {
		return err
	}
	return h.PcapHandle.WritePacketData(h.buffer.Bytes())
}

func (h *Handle) SetDstMacAddr(addr net.HardwareAddr) {
	if bytes.Compare(h.dstMacAddr, MultiCastAddr) == 0 {
		h.dstMacAddr = addr
	}
}

func (h *Handle) SendStartPkt() error {
	eth := layers.Ethernet{
		SrcMAC:       h.srcMacAddr,
		DstMAC:       h.dstMacAddr,
		EthernetType: layers.EthernetTypeEAPOL,
	}
	eapol := layers.EAPOL{
		Version: 0x01,
		Type:    layers.EAPOLTypeStart,
	}
	if err := h.send(&eth, &eapol, &fillLayer); err != nil {
		return err
	}
	return nil
}

func (h *Handle) SendResponseIdentity(id uint8, identity []byte) error {
	eth := layers.Ethernet{
		SrcMAC:       h.srcMacAddr,
		DstMAC:       h.dstMacAddr,
		EthernetType: layers.EthernetTypeEAPOL,
	}
	eapol := layers.EAPOL{
		Version: 0x01,
		Type:    layers.EAPOLTypeEAP,
		Length:  uint16(0x10),
	}
	eap := layers.EAP{
		Code:     layers.EAPCodeResponse,
		Id:       id,
		Type:     layers.EAPTypeIdentity,
		TypeData: identity,
		Length:   uint16(0x10),
	}
	if err := h.send(&eth, &eapol, &eap, &fillLayer); err != nil {
		return err
	}
	return nil
}

func (h *Handle) SendResponseMD5Chall(id uint8, salt, user, pass []byte) error {
	plain := []byte{id}
	plain = append(plain, pass...)
	plain = append(plain, salt[:0x10]...)
	cipher := md5.Sum(plain)
	data := append([]byte{uint8(len(cipher))}, cipher[:]...)
	data = append(data, user...)
	eth := layers.Ethernet{
		SrcMAC:       h.srcMacAddr,
		DstMAC:       h.dstMacAddr,
		EthernetType: layers.EthernetTypeEAPOL,
	}
	eapol := layers.EAPOL{
		Version: 0x01,
		Type:    layers.EAPOLTypeEAP,
		Length:  uint16(5 + len(data)),
	}
	eap := layers.EAP{
		Code:     layers.EAPCodeResponse,
		Id:       id,
		Type:     layers.EAPTypeOTP,
		TypeData: data,
		Length:   eapol.Length,
	}
	if err := h.send(&eth, &eapol, &eap, &fillLayer); err != nil {
		return err
	}
	return nil
}

func (h *Handle) SendLogoffPkt() error {
	eth := layers.Ethernet{
		SrcMAC:       h.srcMacAddr,
		DstMAC:       h.dstMacAddr,
		EthernetType: layers.EthernetTypeEAPOL,
	}
	eapol := layers.EAPOL{
		Version: 0x01,
		Type:    layers.EAPOLTypeLogOff,
	}
	if err := h.send(&eth, &eapol); err != nil {
		return err
	}
	return nil
}

var echoPacket = []byte{0xFF, 0xFF, 0x37, 0x77, 0x7F, 0x9F, 0xFF, 0xFF, 0xD9, 0x13, 0xFF, 0xFF, 0x37, 0x77, 0x7F, 0x9F, 0xFF, 0xFF, 0xF7, 0x2B, 0xFF, 0xFF, 0x37, 0x77, 0x7F, 0x3F, 0xFF}

func (h *Handle) SendEchoPkt(echoNo, echoKey uint32) error {
	if echoNo != 0x0000102B {
		dd1, dd2 := echoNo+echoKey, echoNo
		buf1, buf2 := echoPacket[6:10], echoPacket[16:20]
		binary.BigEndian.PutUint32(buf1, dd1)
		binary.BigEndian.PutUint32(buf2, dd2)
		Symmetric(buf1)
		Symmetric(buf2)
	}
	eth := layers.Ethernet{
		SrcMAC:       h.srcMacAddr,
		DstMAC:       h.dstMacAddr,
		EthernetType: layers.EthernetTypeEAPOL,
	}
	eapol := layers.EAPOL{
		Version: 0x01,
		Type:    0xbf,
		Length:  uint16(len(echoPacket)),
	}
	echo := gopacket.Payload(echoPacket)
	if err := h.send(&eth, &eapol, &echo); err != nil {
		return err
	}
	return nil
}
