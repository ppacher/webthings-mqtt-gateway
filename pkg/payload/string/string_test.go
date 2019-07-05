package string

import (
	"testing"

	"github.com/own-home/central/pkg/payload"
	"github.com/stretchr/testify/assert"
)

func Test_StringHandler(t *testing.T) {
	cases := []struct {
		m   payload.HandlerSpec
		i   string
		o   string
		err bool
	}{
		{
			payload.HandlerSpec{
				"regex": "\\d",
				"group": 0,
				"index": 0,
			},
			"foo 1 bar",
			"1",
			false,
		},
		{
			payload.HandlerSpec{
				"regex": "[a-z]+ (\\d)",
				"group": 0,
				"index": 1,
			},
			"foo 1 bar",
			"1",
			false,
		},
		{
			payload.HandlerSpec{
				"regex": "\\d [a-z]+",
				"index": 0,
			},
			"foo 1 bar",
			"1 bar",
			false,
		},
		{
			payload.HandlerSpec{
				"regex": "\\d",
			},
			"foo 1 bar",
			"1",
			false,
		},
		{
			payload.HandlerSpec{
				"regex": "\\d [a-z]+",
				"index": 1,
			},
			"foo 1 bar",
			"",
			true,
		},
		{
			payload.HandlerSpec{
				"regex": "\\d",
				"group": 0,
				"index": 2,
			},
			"foo 1 bar",
			"",
			true,
		},
		{
			payload.HandlerSpec{
				"regex": "\\d",
				"group": 1,
				"index": 0,
			},
			"foo 1 bar",
			"",
			true,
		},
	}

	for _, c := range cases {
		c.m["type"] = "string"

		res, err := c.m.Parse([]byte(c.i))
		if c.err {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
			assert.Equal(t, c.o, res)
		}
	}
}
