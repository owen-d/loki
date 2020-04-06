package parser

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCombinators(t *testing.T) {
	for _, tc := range []struct {
		parser Parser
		in     string
		err    string
		out    interface{}
	}{
		{
			parser: First(StringP("a"), StringP("b")),
			in:     "a",
			err:    "Parse error at 1: Expecting (b)",
		},
		{
			parser: ManyP(First(StringP("a"), StringP("b"))),
			in:     "aba",
			err:    "Parse error at 2: Unterminated input: a",
		},
		{
			parser: ManyP(First(StringP("a"), StringP("b"))),
			in:     "a",
			err:    "Parse error at 0: Unterminated input: a",
		},
		{
			parser: ManyP(First(StringP("a"), StringP("b"))),
			in:     "abab",
			out:    []interface{}{"a", "a"},
		},
	} {
		t.Run(fmt.Sprintf("%s-%s", tc.in, tc.parser.Type()), func(t *testing.T) {
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
