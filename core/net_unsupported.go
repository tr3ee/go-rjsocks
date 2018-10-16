// +build !linux,!windows

package rjsocks

func FindAllAdapters() ([]NwAdapterInfo, error) {
	return nil, nil
}
