package pebbleclient

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type formatTestCase struct {
}

func Test_FormatPath(t *testing.T) {
	for _, testCase := range []struct {
		path   string
		params Params
		expect string
	}{
		{
			path:   "/",
			params: Params{"x": "y"},
			expect: "/",
		},
		{
			path:   "/foo/bar",
			params: Params{"x": "y"},
			expect: "/foo/bar",
		},
		{
			path:   "/foo/:notexists",
			params: Params{},
			expect: "/foo/",
		},
		{
			path:   "/foo/b:x",
			params: Params{"x": "y"},
			expect: "/foo/b:x",
		},
		{
			path:   "/foo/:x",
			params: Params{"x": "y"},
			expect: "/foo/y",
		},
		{
			path:   "/foo/:x",
			params: Params{"x": "a b"},
			expect: `/foo/a%20b`,
		},
		{
			path:   "/foo/:x",
			params: Params{"x": "a/b"},
			expect: `/foo/a%2Fb`,
		},
	} {
		assert.Equal(t, testCase.expect, FormatPath(testCase.path, testCase.params))
	}
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
