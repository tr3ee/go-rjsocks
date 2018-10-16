package rjsocks

import (
	"bytes"
	"encoding/csv"
	"io"
	"net"
	"os/exec"
	"strings"

	"golang.org/x/net/html/charset"
)

func FindAllAdapters() ([]NwAdapterInfo, error) {
	output, err := exec.Command("getmac", "/V", "/NH", "/FO", "csv").CombinedOutput()
	if err != nil {
		return nil, err
	}
	utf8reader, err := charset.NewReaderLabel("gb2312", bytes.NewReader(output))
	if err != nil {
		return nil, err
	}
	r := csv.NewReader(utf8reader)
	var infos []NwAdapterInfo
	for {
		record, err := r.Read()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return nil, err
			}
		}
		if len(record) != 4 {
			continue
		}

		mac, err := net.ParseMAC(record[2])
		if err != nil {
			//TODO log the error
			continue
		}
		deviceName := record[3]
		lBrace := strings.IndexRune(deviceName, '{')
		rBrace := strings.IndexRune(deviceName, '}')
		if (lBrace+1 >= rBrace) && (lBrace == -1 || rBrace == -1) {
			//panic(fmt.Sprintf("Error on FindAllAdapters(): deviceName(%q) leftBrace(%d) rightBrace(%d)", deviceName, lBrace, rBrace))
			continue
		}
		deviceName = deviceName[lBrace+1 : rBrace]
		infos = append(infos, NwAdapterInfo{
			AdapterName: record[0],
			DeviceDesc:  record[1],
			Mac:         mac,
			DeviceName:  deviceName,
		})
	}
	return infos, nil
}
