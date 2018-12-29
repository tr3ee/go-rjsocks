// +build unix linux

package rjsocks

import (
	"net"
	"os/exec"
)

// FindAllAdapters returns all the Network Adapters Info
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

func refreshIP(adapter string) {
	cmd := exec.Command("dhclient", adapter)
	cmd.Run()
}
