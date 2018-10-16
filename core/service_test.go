package rjsocks

import (
	"fmt"
	"testing"
)

func TestService(t *testing.T) {
	infos, err := FindAllAdapters()
	if err != nil {
		t.Fatal(err)
	}
	for _, info := range infos {
		if info.AdapterName == "以太网" {
			fmt.Printf("%+v\n", info)
			service, err := NewService("usr", "pwd", &info)
			if err != nil {
				t.Fatal(err)
			}
			service.Run()
			break
		}
	}
}
