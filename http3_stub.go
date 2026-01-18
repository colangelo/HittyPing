//go:build !http3

package main

import (
	"net/http"
	"time"
)

var http3Available = false

func newHTTP3Client(timeout time.Duration, insecure bool) *http.Client {
	// This function should not be called when http3Available is false
	// The main function checks http3Available before calling this
	return nil
}
