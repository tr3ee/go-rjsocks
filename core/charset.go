package rjsocks

import (
	"bytes"
	"io/ioutil"

	"golang.org/x/net/html/charset"
)

var (
	sysDefaultCharset = "gb2312"
)

func toUTF8(data []byte) ([]byte, error) {
	r, err := charset.NewReaderLabel(sysDefaultCharset, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return buf, nil
}
