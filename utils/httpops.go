package utils

import (
	"net/http"
	"strings"
	"time"
)

func RunHTTPGet(url string) (*http.Response, error) {
	client := &http.Client{
		Timeout: 60 * time.Second,
	}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func isGzipResponse(r *http.Response) bool {
	// Check Content-Encoding
	for _, s := range r.Header["Content-Encoding"] {
		if s == "gzip" {
			return true
		}
	}
	// If Content-Encoding is not set, check Content-Type
	for _, s := range r.Header["Content-Type"] {
		if strings.Contains(s, "gzip") {
			return true
		}
	}
	return false
}

func ReadNBytesFromResponse(r *http.Response, n int) ([]byte, error) {
	buf := make([]byte, n)
	nRead, err := r.Body.Read(buf)
	if err != nil {
		return nil, err
	}
	return buf[:nRead], nil
}
