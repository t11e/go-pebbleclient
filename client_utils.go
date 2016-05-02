package pebbleclient

import (
	"encoding/json"
	"io"
	"net/http"
)

func performJSONRequest(fn func() (*http.Response, error), result interface{}) error {
	resp, err := fn()
	if err != nil {
		return err
	}
	return unmarshalJSONResponse(resp, result)
}

func unmarshalJSONResponse(resp *http.Response, dest interface{}) error {
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return &ClientRequestError{resp}
	}
	return unmarshalJSONStream(resp.Body, &dest)
}

func unmarshalJSONStream(reader io.Reader, dest interface{}) error {
	decoder := json.NewDecoder(reader)
	err := decoder.Decode(&dest)
	return err
}
