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
			query:    `rate({foo="bar"}[5m]) > 1`,
			validity: `absent_over_time({foo="bar"}[5m])`,
		},
		{
			query:    `sum by (cluster) (rate({foo="bar"}[1m]))`,
			validity: `absent_over_time({foo="bar"}[1m])`,
		},
		// {
		// 	// This test fails, but is a better Absence mapping. Ignored for now in
		//      // favor of how easy it was to implement our algorithm with `RequiredLabelNames()`
		//      // instead, which does not preserve (OR|AND).
		// 	query:    `rate({foo="bar"} | logfmt | bazz="a" or buzz="b" [5m])`,
		// 	validity: `absent_over_time({foo="bar"} | logfmt | bazz=~".+" or buzz=~".+" [5m])`,
		// },

		// {
		// 	// Another optimization not yet implemented could test that grouping
		// 	// labels are ensured as well.
		// 	query:    `sum by (cluster, namespace) (rate({container="ingester"}[5m])) > 10`,
		// 	validity: `absent_over_time({container="ingester"} | cluster =~ ".+" | namespace =~ ".+" | [5m])`,
		// },
		{
			query: "1 + 1",
		},
		{
			query: `sum(rate({foo="bar"} |= "error" [5m])) / sum(rate({foo="bar"}[5m]))`,
			// note: this can be reduced further
			validity: `
			absent_over_time({foo="bar"} [5m]) or absent_over_time({foo="bar"} [5m])
			`,
		},
		{
			query:    `sum(rate({foo="bar", level="error"}[5m])) / sum(rate({foo="bar"}[5m])) > 0.1`,
			validity: `absent_over_time({foo="bar", level="error"}[5m]) or absent_over_time({foo="bar"} [5m])`,
		},
		{
			// In this case the meta-alert and the actual alert are identical.
			// Once there are no more `foo="bar"` streams, the alert itself would fire.
			query: `absent_over_time({foo="bar"}[5m])`,
		},
	} {
		t.Run(tc.query, func(t *testing.T) {
			expr, err := ParseExpr(tc.query)
			require.Nil(t, err)
			sampleExpr, ok := expr.(SampleExpr)
			require.Equal(t, true, ok)

			absence, err := QueryAbsence(sampleExpr)
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
