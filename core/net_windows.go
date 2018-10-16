package rjsocks

import (
	"bytes"
	"encoding/csv"
	"io"
	"net"
	"os/exec"
	"strings"
	"syscall"

	"golang.org/x/net/html/charset"
)

/*
FindAllAdapters returns all the Network Adapters Info:
	It can be listed for users to choose
	it uses 'getmac.exe' on windows and decode the result from gb2312 to utf-8
	ATTENTION: the index of the return value is not fixed.
*/
func FindAllAdapters() ([]NwAdapterInfo, error) {
	output, err := exec.Command("getmac", "/V", "/NH", "/FO", "csv").CombinedOutput()
	if err != nil {
		return nil, err
	}
	utf8reader, err := charset.NewReaderLabel(sysDefaultCharset, bytes.NewReader(output))
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

func refreshIP(adapter string) {
	cmd := exec.Command("ipconfig", "/renew", adapter)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	cmd.Run()
}
