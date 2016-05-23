package pebbleclient

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

type formatTestCase struct {
}

func Test_formatPath(t *testing.T) {
	for _, testCase := range []struct {
		path   string
		params url.Values
		expect string
	}{
		{
			path:   "/",
			params: url.Values{"x": []string{"y"}},
			expect: "/",
		},
		{
			path:   "/foo/bar",
			params: url.Values{"x": []string{"y"}},
			expect: "/foo/bar",
		},
		{
			path:   "/foo/b:x",
			params: url.Values{"x": []string{"y"}},
			expect: "/foo/b:x",
		},
		{
			path:   "/foo/:x",
			params: url.Values{"x": []string{"y"}},
			expect: "/foo/y",
		},
		{
			path:   "/foo/:x",
			params: url.Values{"x": []string{"a b"}},
			expect: `/foo/a b`,
		},
		{
			path:   "/foo/:x",
			params: url.Values{"x": []string{"a/b"}},
			expect: `/foo/a/b`,
		},
		{
			path:   "/foo/:x",
			params: url.Values{"x": []string{"a", "b"}},
			expect: `/foo/a,b`,
		},
	} {
		result, err := formatPath(testCase.path, testCase.params)
		assert.NoError(t, err)
		assert.Equal(t, testCase.expect, result)
	}
}

func Test_formatPath_missingKey(t *testing.T) {
	_, err := formatPath("/foo/:bar", url.Values{})
	assert.Error(t, err)
	assert.IsType(t, &MissingParameter{}, err)
}

func Test_formatPath_removesFromParams(t *testing.T) {
	p := url.Values{"x": []string{"y"}}
	_, err := formatPath("/foo/:x", p)
	assert.NoError(t, err)
	assert.Len(t, p, 0)
}

func Test_URIEscape(t *testing.T) {
	for idx, scenario := range []struct {
		in       string
		expected string
	}{
		{"", ""},
		{"foo bar", "foo%20bar"},
		{"foo/bar", "foo%2Fbar"},
		{"!@#$%^&*():;?.,", "%21@%23$%25%5E&%2A%28%29:;%3F.,"},
	} {
		actual := URIEscape(scenario.in)
		if scenario.expected != actual {
			t.Errorf("Scenario %d\nexpected: %s\nactual: %s\n", idx, scenario.expected, actual)
		}
	}
}
