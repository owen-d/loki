package logql

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_Validity(t *testing.T) {
	for _, tc := range []struct {
		query, validity string
		err             error
	}{
		{
			// when it's just a set of matchers, do nothing. The alert & meta alert would be equivalent.
			query: `rate({foo="bar"}[5m])`,
		},
		{
			// when it's just a set of matchers but is a binop, ensure the underlying data exists.
			query:    `rate({foo="bar"}[5m]) > 1`,
			validity: `absent_over_time({foo="bar"}[5m])`,
		},
		// {
		// 	// when a grouping exists, ensure the label is present.
		// 	// Note: this could be improved to detect when indexed labels are used vs
		// 	// when it's not known.
		// 	query:    `sum by (cluster) (rate({foo="bar"}[1m]))`,
		// 	validity: `absent_over_time({foo="bar"} | cluster=~".+" [1m])`,
		// },
		{
			// respect OR'd label filters
			query:    `rate({foo="bar"} | logfmt | bazz="a" or buzz="b" [5m])`,
			validity: `absent_over_time({foo="bar"} | logfmt | bazz=~".+" or buzz=~".+" [5m])`,
		},
		{
			query: "1 + 1",
		},
		{
			// strip out line filters -- we just want to know if the underlying streams exist, not if the alert is firing.
			// future optimization: this is reducible
			query:    `sum(rate({foo="bar"} |= "error" [5m])) / sum(rate({foo="bar"}[5m]))`,
			validity: `absent_over_time({foo="bar"} [5m]) or absent_over_time({foo="bar"} [5m])`,
		},
		{
			// for binops, ensure at least one leg exists as these are usually comparisons between a subset & a full set.
			query:    `sum(rate({foo="bar", level="error"}[5m])) / sum(rate({foo="bar"}[5m])) > 0.1`,
			validity: `absent_over_time({foo="bar", level="error"}[5m]) or absent_over_time({foo="bar"} [5m])`,
		},
		{
			// In this case the meta-alert and the actual alert are identical.
			// Once there are no more `foo="bar"` streams, the alert itself would fire.
			query: `absent_over_time({foo="bar"}[5m])`,
		},
		// {
		// 	query:    `sum by (org_id) (count_over_time({job="loki-prod/query-frontend"} |= "metrics.go:84" | logfmt | duration > 10s [5m])) > 10 `,
		// 	validity: `{job="loki-prod/query-frontend"} |= "metrics.go:84" | logfmt | duration=~".+" | org_id =~".+"`,
		// },
	} {
		t.Run(tc.query, func(t *testing.T) {
			expr, err := ParseExpr(tc.query)
			require.Nil(t, err)
			sampleExpr, ok := expr.(SampleExpr)
			require.Equal(t, true, ok)

			absence, err := MetaAlert(sampleExpr)
			if tc.err != nil {
				require.Equal(t, tc.err, err)
				return
			}
			require.Nil(t, err)

			// No validity test is needed for this query
			if tc.validity == "" {
				require.Nil(t, absence)
				return
			}

			expected, err := ParseExpr(tc.validity)
			require.Nil(t, err)
			require.Equal(t, expected.String(), absence.String())

		})
	}
}
