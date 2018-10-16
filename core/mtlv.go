package rjsocks

import (
	"encoding/binary"
	"fmt"
	"io"
)

const (
	mtlvMagic = 0x1311
)

// Magic Type Length struct
type mtl struct {
	Magic  uint32
	Type   uint8
	Length uint8
}

// Magic Type Length Value struct
type mtlv struct {
	Header mtl
	Buffer []byte
}

func parseMTLV(r io.Reader) (*mtlv, error) {
	var header mtl
	if err := binary.Read(r, binary.BigEndian, &header); err != nil {
		return nil, err
	}
	if header.Magic != mtlvMagic {
		return nil, fmt.Errorf("magic number should be %x, but got %x", mtlvMagic, header.Magic)
	}
	switch {
	case header.Type == 1:
		header.Length--
	case header.Type > 0x50:
		header.Length -= 2
	}
	buf := make([]byte, int(header.Length))
	_, err := io.ReadFull(r, buf)
	if err != nil {
		return nil, fmt.Errorf("length corrupted: %s", err)
	}
	return &mtlv{
		Header: header,
		Buffer: buf,
	}, nil
}

// func parseMTLVs(r io.Reader, n int) ([]*mtlv, error) {
// 	var ret []*mtlv
// 	for i := 0; i < n; i++ {
// 		mtlv, err := parseMTLV(r)
// 		if err != nil {
// 			return nil, err
// 		}
// 		ret = append(ret, mtlv)
// 	}
// 	return ret, nil
// }
