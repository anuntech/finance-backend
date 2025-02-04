package protocols

import (
	"io"
	"net/http"
	"net/url"
)

type HttpRequest struct {
	Body      io.ReadCloser
	Header    http.Header
	UrlParams url.Values
	UrlPath   string
}

type HttpResponse struct {
	Body       io.ReadCloser
	StatusCode int
}
