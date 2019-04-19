package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// MyIPUrl ...
const MyIPUrl = "https://api.myip.com"

// ErrInvalidRespResult Error when response result is invalid
var ErrInvalidRespResult = fmt.Errorf("invalid response result")

// IPApi .
type IPApi struct {
	Client *http.Client
}

// NewIPApi return IPApi object configed with timeout
func NewIPApi(timeout time.Duration) *IPApi {
	return &IPApi{
		Client: &http.Client{
			Timeout: timeout,
		},
	}
}

// MyIP return public ip address of current machine
func (ia *IPApi) MyIP() (ip string, err error) {
	resp, err := ia.Client.Get(MyIPUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("status code: %d", resp.StatusCode)
	}

	infos := make(map[string]string)
	err = json.Unmarshal(body, &infos)
	if err != nil {
		return "", err
	}

	ip, ok := infos["ip"]
	if !ok {
		return "", ErrInvalidRespResult
	}
	return ip, nil
}
