package pebbleclient

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type paramTestCase struct {
	In     interface{}
	Expect string
}

func Test_Params(t *testing.T) {
	for _, testCase := range []paramTestCase{
		paramTestCase{nil, ""},
		paramTestCase{"s", "s"},
		paramTestCase{int(1), "1"},
		paramTestCase{int32(1), "1"},
		paramTestCase{int64(1), "1"},
		paramTestCase{float32(3.14), "3.14"},
		paramTestCase{float64(3.14), "3.14"},
		paramTestCase{true, "true"},
		paramTestCase{false, "false"},
	} {
		values := Params(map[string]interface{}{
			"myKey": testCase.In,
		}).ToValues()
		assert.Equal(t, []string{testCase.Expect}, values["myKey"])
		assert.Len(t, values, 1)
	}
}
