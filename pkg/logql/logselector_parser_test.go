package logql

import (
	"testing"

	"github.com/grafana/loki/pkg/logql/parser"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/stretchr/testify/require"
)

func TestLogSelectorParser(t *testing.T) {
	for _, tc := range []struct {
		in  string
		exp interface{}
		err string
	}{
		{
			in:  `{foo="bar"}`,
			exp: &matchersExpr{matchers: []*labels.Matcher{mustNewMatcher(labels.MatchEqual, "foo", "bar")}},
		},
		{
			in:  `{ foo = "bar" }`,
			exp: &matchersExpr{matchers: []*labels.Matcher{mustNewMatcher(labels.MatchEqual, "foo", "bar")}},
		},
		{
			in:  `{ foo != "bar" }`,
			exp: &matchersExpr{matchers: []*labels.Matcher{mustNewMatcher(labels.MatchNotEqual, "foo", "bar")}},
		},
		{
			in:  `{ foo =~ "bar" }`,
			exp: &matchersExpr{matchers: []*labels.Matcher{mustNewMatcher(labels.MatchRegexp, "foo", "bar")}},
		},
		{
			in:  `{ foo !~ "bar" }`,
			exp: &matchersExpr{matchers: []*labels.Matcher{mustNewMatcher(labels.MatchNotRegexp, "foo", "bar")}},
		},
		{
			in: `{ foo = "bar", bar != "baz" }`,
			exp: &matchersExpr{matchers: []*labels.Matcher{
				mustNewMatcher(labels.MatchEqual, "foo", "bar"),
				mustNewMatcher(labels.MatchNotEqual, "bar", "baz"),
			}},
		},
		{
			in: `{foo="bar"} |= "baz"`,
			exp: &filterExpr{
				left:  &matchersExpr{matchers: []*labels.Matcher{mustNewMatcher(labels.MatchEqual, "foo", "bar")}},
				ty:    labels.MatchEqual,
				match: "baz",
			},
		},
		{
			in: `{foo="bar"} |= "baz" |~ "blip" != "flip" !~ "flap"`,
			exp: &filterExpr{
				left: &filterExpr{
					left: &filterExpr{
						left: &filterExpr{
							left:  &matchersExpr{matchers: []*labels.Matcher{mustNewMatcher(labels.MatchEqual, "foo", "bar")}},
							ty:    labels.MatchEqual,
							match: "baz",
						},
						ty:    labels.MatchRegexp,
						match: "blip",
					},
					ty:    labels.MatchNotEqual,
					match: "flip",
				},
				ty:    labels.MatchNotRegexp,
				match: "flap",
			},
		},
		{
			in:  `{foo="bar"} |~`,
			err: `Parse error at 11: Unterminated input:  |~`,
		},
	} {
		t.Run(tc.in, func(t *testing.T) {
			out, err := parser.RunParser(LogSelectorParser, tc.in)
			if tc.err != "" {
				require.NotNil(t, err)
				require.Equal(t, tc.err, err.Error())
			}
			require.Equal(t, tc.exp, out)
		})
	}

}
