package rjsocks

import (
	"fmt"
	"testing"
)

func TestFindAllAdapters(t *testing.T) {
	adapters, err := FindAllAdapters()
	if err != nil {
		t.Fatal(err)
	}
	for _, adapter := range adapters {
		fmt.Printf("%+v\n", adapter)
	}
}
