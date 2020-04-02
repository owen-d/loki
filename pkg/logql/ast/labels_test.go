package ast

import (
	"testing"

	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/stretchr/testify/require"
)

func TestLabelsParser(t *testing.T) {
	for _, tc := range []struct {
		parser Parser
		in     string
		err    string
		out    interface{}
	}{
		{
			parser: MatchRegexp,
			in:     "=~",
			out:    labels.MatchRegexp,
		},
		{
			parser: MatchType,
			in:     "=",
			out:    labels.MatchEqual,
		},
		{
			parser: MatchType,
			in:     "!=",
			out:    labels.MatchNotEqual,
		},
		{
			parser: MatchType,
			in:     "=~",
			out:    labels.MatchRegexp,
		},
		{
			parser: MatchType,
			in:     "!~",
			out:    labels.MatchNotRegexp,
		},
		{
			parser: MatchType,
			in:     "~~",
			out:    labels.MatchNotRegexp,
			err:    "Parse error at 0: Expecting Alternative<labels.MatchRegexp, labels.MatchNotRegexp, labels.MatchNotEqual, labels.MatchEquals>",
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
