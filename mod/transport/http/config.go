package http

import (
	"net/http"
)

func (c *Config) getNormalizedPath() string {
	if c.Path == "" {
		return "/"
	}
	if c.Path[0] != '/' {
		return "/" + c.Path
	}
	return c.Path
}

func (c *Config) applyHeader(header http.Header) {
	for _, httpHeader := range c.Header {
		for _, httpHeaderValue := range httpHeader.Value {
			header.Set(httpHeader.Name, httpHeaderValue)
		}
	}
}
