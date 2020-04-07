package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLibParser(t *testing.T) {
	for _, tc := range []struct {
		parser Parser
		in     string
		err    string
		out    interface{}
	}{
		{
			parser: CharSetParser,
			in:     "8",
			out:    "8",
		},
		{
			parser: Parens(Equals),
			in:     "(=)",
			out:    "=",
		},
		{
			parser: Characters,
			in:     "abc",
			out:    "abc",
		},
	} {
		t.Run(tc.in, func(t *testing.T) {
			out, err := RunParser(tc.parser, tc.in)
			if tc.err != "" {
				require.NotNil(t, err)
				require.Equal(t, tc.err, err.Error())
			} else {
				require.Nil(t, err)
				require.Equal(t, tc.out, out)
			}
		})
	}
}

func TestIntParser(t *testing.T) {
	for _, tc := range []struct {
		in  string
		err string
		out int
	}{
		{
			in:  "8",
			out: 8,
		},
		{
			in:  "0123",
			out: 123,
		},
		{
			in:  "abc",
			err: `Parse error at 0: Expecting Alternative<0, 1, 2, 3, 4, 5, 6, 7, 8, 9>`,
		},
	} {
		t.Run(tc.in, func(t *testing.T) {
			out, err := RunParser(IntParser, tc.in)
			if tc.err != "" {
				require.NotNil(t, err)
				require.Equal(t, tc.err, err.Error())
			} else {
				require.Nil(t, err)
				require.Equal(t, tc.out, out)
			}
		})
	}
}
