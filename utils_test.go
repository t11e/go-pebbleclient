package pebbleclient

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

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

type valueTestCase struct {
	In     interface{}
	Expect string
}

func Test_Values(t *testing.T) {
	for _, testCase := range []valueTestCase{
		valueTestCase{nil, ""},
		valueTestCase{"s", "s"},
		valueTestCase{int(1), "1"},
		valueTestCase{int32(1), "1"},
		valueTestCase{int64(1), "1"},
		valueTestCase{float32(3.14), "3.14"},
		valueTestCase{float64(3.14), "3.14"},
		valueTestCase{true, "true"},
		valueTestCase{false, "false"},
	} {
		vs := Values(map[string]interface{}{
			"myKey": testCase.In,
		})
		assert.Equal(t, []string{testCase.Expect}, vs["myKey"])
		assert.Len(t, vs, 1)
	}
}
