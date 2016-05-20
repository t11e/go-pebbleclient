package pebbleclient

import "testing"

func Test_escapedPath(t *testing.T) {
	for idx, scenario := range []struct {
		in       string
		expected string
	}{
		{"", ""},
		{"foo bar", "foo%20bar"},
		{"foo/bar", "foo%2Fbar"},
		{"!@#$%^&*():;?.,", "%21@%23$%25%5E&%2A%28%29:;%3F.,"},
	} {
		actual := escapedPath(scenario.in)
		if scenario.expected != actual {
			t.Errorf("Scenario %d\nexpected: %s\nactual: %s\n", idx, scenario.expected, actual)
		}
	}
}
