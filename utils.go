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

// BuildValues is a convenience function to generate url.Values from a map.
// Conversion to string is done like so:
//
// - If value is nil, use empty string.
//
// - If value is a string, use that.
//
// - If value implements fmt.Stringer, use that.
//
// - Otherwise, use fmt.Sprintf("%v", v).
//
func Values(m map[string]interface{}) url.Values {
	values := url.Values{}
	for k, v := range m {
		if v == nil {
			values.Set(k, "")
		} else if s, ok := v.(string); ok {
			values.Set(k, s)
		} else if s, ok := v.(fmt.Stringer); ok {
			values.Set(k, s.String())
		} else {
			values.Set(k, fmt.Sprintf("%v", v))
		}
	}
	return values
}

func URIEscape(path string) string {
	escaped := (&url.URL{Path: path}).EscapedPath()
	escaped = strings.Replace(escaped, "/", "%2F", -1)
	return escaped
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
