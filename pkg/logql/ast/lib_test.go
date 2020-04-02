package ast

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
			parser: CharSet,
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
		// {
		// 	parser: nil,
		// 	in:     "",
		// 	err:    "",
		// 	out:    nil,
		// },
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
