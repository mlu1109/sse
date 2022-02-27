package sseclt

import (
	"fmt"
	"io"
	"net/http"
)

type Request struct {
	url     string
	headers http.Header
	body    io.Reader
	method  string
}

func NewRequest(url string) *Request {
	headers := http.Header{
		"Cache-Control": {"no-cache"},
		"Accept":        {"text/event-stream"},
		"Connection":    {"keep-alive"},
	}
	return &Request{
		url:     url,
		headers: headers,
		method:  http.MethodGet,
	}
}

func (r *Request) AddHeader(key, value string) *Request {
	r.headers.Add(key, value)
	return r
}

func (r *Request) WithHeaders(headers http.Header) *Request {
	r.headers = headers
	return r
}

func (r *Request) WithBody(body io.Reader) *Request {
	r.body = body
	return r
}

func (r *Request) WithMethod(method string) *Request {
	r.method = method
	return r
}

func (r *Request) Do(client ...http.Client) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, r.url, r.body)
	if err != nil {
		return nil, err
	}
	if len(client) == 0 {
		return http.DefaultClient.Do(req)
	} else if len(client) == 1 {
		return client[0].Do(req)
	} else {
		return nil, fmt.Errorf("several clients provided")
	}
}
