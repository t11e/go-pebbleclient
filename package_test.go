package pebbleclient

import (
	"fmt"
	"net/url"
)

func hostFromUrl(anUrl string) string {
	u, err := url.Parse(anUrl)
	if err != nil {
		panic(fmt.Sprintf("Invalid URL %q", anUrl))
	}
	return u.Host
}
