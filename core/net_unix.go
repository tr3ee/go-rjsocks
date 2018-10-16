// +build unix,cgo

package rjsocks

import (
	"net"
)

func FindAllAdapters() ([]NwAdapterInfo, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	var infos []NwAdapterInfo
	for _, itf := range interfaces {
		infos = append(infos, NwAdapterInfo{
			AdapterName: itf.Name,
			DeviceDesc:  itf.Name,
			Mac:         itf.HardwareAddr,
			DeviceName:  itf.Name,
		})
	}
	return infos, nil
}
