package planning

type LogicalPlanner[T LogicalPlan] struct {
	inner ExecCtx[T]
}

func NewLogicalPlanner[T LogicalPlan](ctx ExecCtx[T]) LogicalPlanner[T] {
	return LogicalPlanner[T]{
		inner: ctx,
	}
}

func (p LogicalPlanner[T]) Project(exprs []LogicalExpr) LogicalPlanner[*Projection] {
	inner := Bind(p.inner, func(plan T) (*Projection, error) {
		return NewProjection(plan, exprs)
	})
	return NewLogicalPlanner(inner)
}

func (p LogicalPlanner[T]) Select(expr LogicalExpr) LogicalPlanner[*Selection] {
	inner := Bind(p.inner, func(plan T) (*Selection, error) {
		return NewSelection(plan, expr)
	})
	return NewLogicalPlanner(inner)
}

func example() (LogicalPlan, error) {
	scan := func() (*Scan, error) {
		return NewScan("foo", &StreamDataSource{}, []string{"timestamp", "line", "labels"})
	}

	p := NewLogicalPlanner(
		// scan
		NewExecCtx(scan),
	).Project(
		// capture 2 fields
		[]LogicalExpr{
			&ColumnReference{"timestamp"},
			&ColumnReference{"line"},
		},
	).Select(
		// filter by line equality
		&BinaryExpr{
			Name: "foo",
			Op:   "=",
			Left: &ColumnReference{"line"},
			Right: &LiteralValue{
				Value: "foo",
				Type:  String,
			},
		},
	)

	return p.inner.Run()
}
