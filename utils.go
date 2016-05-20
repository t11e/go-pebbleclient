package pebbleclient

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

func escapedPath(path string) string {
	escaped := (&url.URL{Path: path}).EscapedPath()
	escaped = strings.Replace(escaped, "/", "%2F", -1)
	return escaped
}

func isNonSuccessStatus(statusCode int) bool {
	return statusCode < 200 || statusCode >= 300
}

func doesStatusCodeYieldBody(statusCode int) bool {
	return statusCode != 204
}

func decodeResponseAsJSON(resp *http.Response, body io.Reader, out interface{}) error {
	if resp.ContentLength == 0 {
		// We treat this is as a non-error
		return nil
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		return errors.New("Expected response to be JSON, received bytes")
	}

	if mediaType, _, err := mime.ParseMediaType(contentType); err != nil {
		return errors.Wrap(err, "Invalid content type")
	} else if mediaType != "application/json" {
		return fmt.Errorf("Expected response to be JSON, got %q", mediaType)
	}

	if err := json.NewDecoder(body).Decode(out); err != nil {
		return errors.Wrap(err, "Could not decode error response JSON")
	}
	return nil
}