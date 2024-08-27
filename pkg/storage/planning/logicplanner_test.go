package planning

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// compiler checks
var (
	_ LogicalPlan = &Scan{}
	_ LogicalPlan = &Projection{}
	_ LogicalPlan = &Selection{}
)

func TestLogicalChain(t *testing.T) {

	scan := func() (*Scan, error) {
		return NewScan(
			"<name>",
			&StreamDataSource{},
			[]string{"timestamp", "line", "labels"},
		)
	}

	p := NewLogicalPlanner(
		// scan
		NewExecCtx(scan),
	).Project(
		// capture 2 fields
		[]LogicalExpr{
			ColumnRef("timestamp"),
			ColumnRef("line"),
		},
	).Select(
		// filter by line equality
		&LogicalBinaryExpr{
			Name:  "line_comparison",
			Op:    "=",
			Left:  ColumnRef("line"),
			Right: LiteralBytes([]byte("foo")),
		},
	)

	exp := `
Filter: line_comparison (=)
  #line
  foo
`
	out, err := p.inner.Run()
	require.NoError(t, err)

	require.Equal(t, exp, "\n"+out.String()+"\n")
}
