package main

import (
	"encoding/json"
	"net/http"
)

type releaseInfo struct {
	TagName string `json:"tag_name"`
}

var currentVersion = `v2.2.0b`

func CheckUpdate() (string, error) {

	latest := `https://api.github.com/repos/tr3ee/go-rjsocks/releases/latest`
	resp, err := http.Get(latest)
	if err != nil {
		return "", err
	}
	info := releaseInfo{}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", err
	}
	if info.TagName != currentVersion {
		return info.TagName, nil
	}
	return "", nil
}
