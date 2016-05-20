package pebbleclient

import (
	"fmt"
	"net/http"
)

type RequestError struct {
	Options     *RequestOptions
	Req         *http.Request
	Resp        *http.Response
	PartialBody []byte

	client *HTTPClient
}

func (err *RequestError) Error() string {
	return fmt.Sprintf("Request to %s [%s] failed with status %d: %s",
		err.client.options.ServiceName, err.Req.URL, err.Resp.StatusCode, err.Resp.Status)
}
