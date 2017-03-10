package pebbleclient_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	pebble "github.com/t11e/go-pebbleclient"
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
		values := pebble.Params(map[string]interface{}{
			"myKey": testCase.In,
		}).ToValues()
		assert.Equal(t, []string{testCase.Expect}, values["myKey"])
		assert.Len(t, values, 1)
	}
}

func TestUID_Class(t *testing.T) {
	for idx, test := range []struct {
		in          string
		expected    string
		expectedErr string
	}{
		{
			in:          "",
			expectedErr: "invalid uid",
		},
		{
			in:       "a:b$1",
			expected: "a",
		},
		{
			in:       "post.example:this.is.my.path$1234",
			expected: "post.example",
		},
		{
			in:          "a:",
			expectedErr: "invalid uid",
		},
		{
			in:          "a:$",
			expectedErr: "invalid uid",
		},
		{
			in:          ":b$1",
			expectedErr: "invalid uid",
		},
		{
			in:          "a:b$c",
			expectedErr: "invalid uid",
		},
	} {
		t.Run(fmt.Sprintf("[%d] %s", idx, test.in), func(t *testing.T) {
			actual, err := pebble.UID(test.in).Class()
			if test.expectedErr != "" {
				assert.EqualError(t, err, test.expectedErr)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestUID_Path(t *testing.T) {
	for idx, test := range []struct {
		in          string
		expected    string
		expectedErr string
	}{
		{
			in:          "",
			expectedErr: "invalid uid",
		},
		{
			in:       "a:b$1",
			expected: "b",
		},
		{
			in:       "post.example:this.is.my.path$1234",
			expected: "this.is.my.path",
		},
		{
			in:          ":b",
			expectedErr: "invalid uid",
		},
		{
			in:          ":b$",
			expectedErr: "invalid uid",
		},
		{
			in:          "a:$1",
			expectedErr: "invalid uid",
		},
		{
			in:          "a:b$c",
			expectedErr: "invalid uid",
		},
	} {
		t.Run(fmt.Sprintf("[%d] %s", idx, test.in), func(t *testing.T) {
			actual, err := pebble.UID(test.in).Path()
			if test.expectedErr != "" {
				assert.EqualError(t, err, test.expectedErr)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.expected, actual)
		})
	}
}

func TestUID_NUID(t *testing.T) {
	for idx, test := range []struct {
		in          string
		expected    int
		expectedErr string
	}{
		{
			in:          "",
			expectedErr: "invalid uid",
		},
		{
			in:       "a:b$1",
			expected: 1,
		},
		{
			in:       "post.example:this.is.my.path$1234",
			expected: 1234,
		},
		{
			in:          "$b",
			expectedErr: "invalid uid",
		},
		{
			in:          ":$1",
			expectedErr: "invalid uid",
		},
		{
			in:          "a:b$",
			expectedErr: "invalid uid",
		},
		{
			in:          "a:b$c",
			expectedErr: "invalid uid",
		},
	} {
		t.Run(fmt.Sprintf("[%d] %s", idx, test.in), func(t *testing.T) {
			actual, err := pebble.UID(test.in).NUID()
			if test.expectedErr != "" {
				assert.EqualError(t, err, test.expectedErr)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, test.expected, actual)
		})
	}
}
