package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// RoundTripFunc .
type RoundTripFunc func(req *http.Request) *http.Response

// RoundTrip .
func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

// NewTestClient returns *http.Client with Transport replaced to avoid making real calls
func NewTestClient(fn RoundTripFunc) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(fn),
	}
}

func TestMyIP(t *testing.T) {
	tests := []struct {
		code     int
		text     string
		ip       string
		hasError bool
	}{
		{code: 200, text: "{\"ip\":\"1.2.3.4\"}", ip: "1.2.3.4", hasError: false},
		{code: 403, text: "", ip: "", hasError: true},
		{code: 200, text: "abcd", ip: "", hasError: true},
	}

	for row, test := range tests {
		client := NewTestClient(func(req *http.Request) *http.Response {
			assert.Equal(t, req.URL.String(), MyIPUrl, "ip url should match, row %d", row)
			return &http.Response{
				StatusCode: test.code,
				Body:       ioutil.NopCloser(bytes.NewBufferString(test.text)),
				Header:     make(http.Header),
			}
		})
		api := &IPApi{Client: client}

		ip, err := api.MyIP()
		if test.hasError {
			assert.Error(t, err, "row %d", row)
		} else {
			assert.NoError(t, err, "row %d", row)
		}
		assert.Equal(t, test.ip, ip, "ip should equal, row %d", row)
	}
}
