package infra

import (
	"net/http"
	"net/url"
	"strings"
)

var DefaultHttpClient = http.DefaultClient

func InitHttpClientWithProxy(proxyURLString string) error {
	transport := &http.Transport{}
	if len(proxyURLString) > 0 && strings.HasPrefix(proxyURLString, "http") {
		if proxyURL, err := url.Parse(proxyURLString); err == nil {
			transport.Proxy = http.ProxyURL(proxyURL)
		} else {
			return err
		}
	}

	httpClient := &http.Client{
		Transport: transport,
	}

	DefaultHttpClient = httpClient
	return nil
}
