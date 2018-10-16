// +build !linux,!windows

package rjsocks

// FindAllAdapters returns all the Network Adapters Info
func FindAllAdapters() ([]NwAdapterInfo, error) {
	return nil, nil
}

func refreshIP(adapter string) {

}
