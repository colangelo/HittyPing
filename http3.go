package main

import (
	"crypto/tls"
	"net/http"
	"time"

	"github.com/quic-go/quic-go/http3"
)

func newHTTP3Client(timeout time.Duration, insecure bool) *http.Client {
	return &http.Client{
		Timeout: timeout,
		Transport: &http3.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: insecure,
			},
		},
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}
