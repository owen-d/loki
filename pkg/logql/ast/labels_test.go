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
		{
			parser: Labels,
			in:     `{foo="bar"}`,
			out:    labels.MustNewMatcher(labels.MatchEqual, "foo", "bar"),
		},
		{
			parser: Labels,
			in:     `{="bar"}`,
			err:    `Parse error at 1: Expecting Alternative<a, b, c, d, e, d, f, g, h, i, j, k, l, m, n, o, p, q, r, s, t, u, v, w, x, y, z, A, B, C, D, E, F, G, H, I, J, K, L, M, N, O, P, Q, R, S, T, U, V, W, X, Y, Z, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, /, ., *, +, ?, !>`,
		},
		{
			parser: Labels,
			in:     `{foo="a}`,
			err:    `Parse error at 7: Expecting (")`,
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
