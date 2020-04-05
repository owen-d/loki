package ast

import (
	"fmt"
	"testing"

	"github.com/grafana/loki/pkg/logql/parser"
	"github.com/prometheus/prometheus/pkg/labels"
	"github.com/stretchr/testify/require"
)

func TestFilterParser(t *testing.T) {
	for _, tc := range []struct {
		parser parser.Parser
		in     string
		err    string
		out    interface{}
	}{
		{
			parser: FilterType,
			in:     `|=`,
			out:    labels.MatchEqual,
		},
		{
			parser: FilterType,
			in:     `|~`,
			out:    labels.MatchRegexp,
		},
		{
			parser: FilterType,
			in:     `!=`,
			out:    labels.MatchNotEqual,
		},
		{
			parser: FilterType,
			in:     `!~`,
			out:    labels.MatchNotRegexp,
		},
		{
			parser: FilterP,
			in:     `|= "baz"`,
			out:    Filter{labels.MatchEqual, "baz"},
		},
		{
			parser: FilterP,
			in:     `|= "baz`,
			err:    `Parse error at 7: Expecting (")`,
		},
		{
			parser: FilterP,
			in:     `|~ "baz"`,
			out:    Filter{labels.MatchRegexp, "baz"},
		},
		{
			parser: FilterP,
			in:     `!= "baz"`,
			out:    Filter{labels.MatchNotEqual, "baz"},
		},
		{
			parser: FilterP,
			in:     `!~ "baz"`,
			out:    Filter{labels.MatchNotRegexp, "baz"},
		},
		{
			parser: FiltersP,
			in:     `|= "foo" != "bar" |~ "baz" !~ "buzz"`,
			out: []Filter{
				{labels.MatchEqual, "foo"},
				{labels.MatchNotEqual, "bar"},
				{labels.MatchRegexp, "baz"},
				{labels.MatchNotRegexp, "buzz"},
			},
		},
		{
			parser: FiltersP,
			in:     `|= "foo" != "bar" |~ "baz" !~ "buzz`,
			err:    `Parse error at 27: Unterminated input: !~ "buzz`,
		},
	} {
		t.Run(fmt.Sprintf("%s-%s", tc.in, tc.parser.Type()), func(t *testing.T) {
			out, err := parser.RunParser(tc.parser, tc.in)
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
