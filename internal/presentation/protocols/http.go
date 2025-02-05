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
	Req       *http.Request
}

type HttpResponse struct {
	Body       io.ReadCloser
	StatusCode int
}
