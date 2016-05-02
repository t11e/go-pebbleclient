package pebbleclient

import (
	"errors"
	"fmt"
	"net/http"
)

var NotFound = errors.New("Not found")

type ClientRequestError struct {
	Response *http.Response
}

func (err *ClientRequestError) Error() string {
	return fmt.Sprintf("Request to %s failed with status %d: %s",
		err.Response.Request.URL, err.Response.StatusCode, err.Response.Status)
}
