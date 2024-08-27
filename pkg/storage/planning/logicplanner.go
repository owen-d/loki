package planning

// LogicalPlanner. Ergonomic way to sequence operations.
// Also known by the name `DataFrame` elsewhere (r, python, etc)
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
