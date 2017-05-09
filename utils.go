package pebbleclient

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

func hostFromRequest(req *http.Request) (string, bool) {
	var host string
	if hosts, ok := req.Header["X-Forwarded-Host"]; ok && len(hosts) > 0 {
		host = hosts[len(hosts)-1]
	} else {
		host = req.Host
	}
	if host != "" {
		return host, true
	}
	return "", false
}

type MissingParameter struct {
	Key string
}

func (err *MissingParameter) Error() string {
	return fmt.Sprintf("The parameter %q is referenced from the path, but is not specified", err.Key)
}

func formatPath(path string, values url.Values) (string, error) {
	if !(strings.ContainsRune(path, '/') && strings.ContainsRune(path, ':')) {
		return path, nil
	}
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if len(part) > 1 && part[0] == ':' {
			key := part[1:]
			if value, ok := values[key]; ok && len(value) > 0 {
				parts[i] = strings.Join(value, ",")
				values.Del(key)
			} else {
				return "", &MissingParameter{key}
			}
		}
	}
	return strings.Join(parts, "/"), nil
}

func URIEscape(path string) string {
	escaped := (&url.URL{Path: path}).EscapedPath()
	escaped = strings.Replace(escaped, "/", "%2F", -1)
	return escaped
}

// isRetriableStatus returns true if the HTTP status code indicates that the request
// can be retried.
func isRetriableStatus(statusCode int) bool {
	switch statusCode {
	case http.StatusGatewayTimeout, http.StatusBadGateway, http.StatusServiceUnavailable:
		return true
	}
	return false
}

func isNonSuccessStatus(statusCode int) bool {
	return statusCode < 200 || statusCode > 299
}

func doesStatusCodeYieldBody(statusCode int) bool {
	switch statusCode {
	case http.StatusNoContent, http.StatusResetContent:
		return false
	default:
		return true
	}
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

	b, err := ioutil.ReadAll(body)
	if err != nil {
		return errors.Wrap(err, "Could not read entire response")
	}

	if err := json.Unmarshal(b, out); err != nil {
		return errors.Wrap(err, "Could not decode response JSON")
	}

	return nil
}
