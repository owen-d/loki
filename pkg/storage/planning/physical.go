package planning

import (
	"fmt"
	"strings"
)

type PhysicalPlan interface {
	// Schema returns the schema of the physical plan
	Schema() (Schema, error)
	// Execute performs the actual data processing and returns a sequence of RecordBatches
	Execute() (Sequence, error)
	// Children returns the child physical plans
	Children() []PhysicalPlan
}

// PhysicalExpr represents a physical expression that can be evaluated on a RecordBatch
type PhysicalExpr interface {
	// Evaluate performs the expression evaluation on the input RecordBatch
	// and returns a Column representing the result
	Evaluate(input RecordBatch) (Column, error)
}

// ColumnExpression represents a physical expression that evaluates to a specific column in a RecordBatch
type ColumnExpression struct {
	index int
}

// NewColumnExpression creates a new ColumnExpression with the given index
func NewColumnExpression(index int) *ColumnExpression {
	return &ColumnExpression{index: index}
}

// Evaluate returns the column at the specified index from the input RecordBatch
func (c *ColumnExpression) Evaluate(input RecordBatch) (Column, error) {
	_, col, err := input.GetFieldIndex(c.index)
	return col, err
}

// String returns a string representation of the ColumnExpression
func (c *ColumnExpression) String() string {
	return fmt.Sprintf("#%d", c.index)
}

// PhysicalLiteralValue represents a physical expression for a literal value
type PhysicalLiteralValue struct {
	value any
	dtype DataTypeSignal
}

// Evaluate creates a LiteralColumn with the expression's value and the input RecordBatch's row count
func (l *PhysicalLiteralValue) Evaluate(input RecordBatch) (Column, error) {
	return &LiteralColumn{
		value: l.value,
		n:     input.n,
		dtype: l.dtype,
	}, nil
}

// Create constructors to match all the logical literals
// newLiteralValueExpr creates a new LiteralValueExpr
func newLiteralValueExpr(value any, dtype DataTypeSignal) *PhysicalLiteralValue {
	return &PhysicalLiteralValue{
		value: value,
		dtype: dtype,
	}
}

// NewLiteralStringExpr creates a new LiteralValueExpr with a string value
func NewLiteralStringExpr(value string) *PhysicalLiteralValue {
	return newLiteralValueExpr(value, String)
}

// NewLiteralBytesExpr creates a new LiteralValueExpr with a bytes value
func NewLiteralBytesExpr(value []byte) *PhysicalLiteralValue {
	return newLiteralValueExpr(value, Bytes)
}

// NewLiteralInt64Expr creates a new LiteralValueExpr with an int64 value
func NewLiteralInt64Expr(value int64) *PhysicalLiteralValue {
	return newLiteralValueExpr(value, I64)
}

// NewLiteralFloat64Expr creates a new LiteralValueExpr with a float64 value
func NewLiteralFloat64Expr(value float64) *PhysicalLiteralValue {
	return newLiteralValueExpr(value, F64)
}

// String returns a string representation of the LiteralValueExpr
func (l *PhysicalLiteralValue) String() string {
	return fmt.Sprintf("%v", l.value)
}

// PhysicalBinaryExpr represents a binary expression in the physical plan
type PhysicalBinaryExpr struct {
	Left  PhysicalExpr
	Right PhysicalExpr
	Op    string
}

// NewBinaryExpr creates a new BinaryExpr
func NewBinaryExpr(left, right PhysicalExpr, op string) *PhysicalBinaryExpr {
	return &PhysicalBinaryExpr{
		Left:  left,
		Right: right,
		Op:    op,
	}
}

// Evaluate performs the binary operation on the input RecordBatch
func (b *PhysicalBinaryExpr) Evaluate(input RecordBatch) (Column, error) {
	leftCol, err := b.Left.Evaluate(input)
	if err != nil {
		return nil, fmt.Errorf("error evaluating left expression: %w", err)
	}

	rightCol, err := b.Right.Evaluate(input)
	if err != nil {
		return nil, fmt.Errorf("error evaluating right expression: %w", err)
	}

	if leftCol.N() != rightCol.N() {
		return nil, fmt.Errorf("mismatched column sizes: left=%d, right=%d", leftCol.N(), rightCol.N())
	}

	if leftCol.Type() != rightCol.Type() {
		return nil, fmt.Errorf("mismatched column types: left=%v, right=%v", leftCol.Type(), rightCol.Type())
	}

	return b.evaluateBinaryOp(leftCol, rightCol)
}

// evaluateBinaryOp performs the actual binary operation
func (b *PhysicalBinaryExpr) evaluateBinaryOp(left, right Column) (Column, error) {
	// Implementation depends on the specific binary operations supported
	// This is a placeholder for demonstration
	switch b.Op {
	case "+", "-", "*", "/":
		return b.arithmeticOp(left, right)
	case "=", "!=", "<", "<=", ">", ">=":
		return b.comparisonOp(left, right)
	default:
		return nil, fmt.Errorf("unsupported binary operation: %s", b.Op)
	}
}

// arithmeticOp handles arithmetic operations
func (b *PhysicalBinaryExpr) arithmeticOp(left, right Column) (Column, error) {
	// Placeholder implementation
	return nil, fmt.Errorf("arithmetic operations not implemented")
}

// comparisonOp handles comparison operations
// comparisonOp handles comparison operations
func (b *PhysicalBinaryExpr) comparisonOp(left, right Column) (Column, error) {
	result := make([]any, left.N())
	for i := 0; i < left.N(); i++ {
		leftVal, err := left.At(i)
		if err != nil {
			return nil, fmt.Errorf("error accessing left column at index %d: %w", i, err)
		}
		rightVal, err := right.At(i)
		if err != nil {
			return nil, fmt.Errorf("error accessing right column at index %d: %w", i, err)
		}

		switch b.Op {
		case "=":
			result[i] = leftVal == rightVal
		case "!=":
			result[i] = leftVal != rightVal
		case "<":
			result[i] = compare(leftVal, rightVal) < 0
		case "<=":
			result[i] = compare(leftVal, rightVal) <= 0
		case ">":
			result[i] = compare(leftVal, rightVal) > 0
		case ">=":
			result[i] = compare(leftVal, rightVal) >= 0
		default:
			return nil, fmt.Errorf("unsupported comparison operation: %s", b.Op)
		}
	}
	return &SliceColumn{items: result, dtype: Bytes}, nil
}

// compare is a helper function to compare two values
func compare(a, b any) int {
	switch x := a.(type) {
	case int64:
		return int(x - b.(int64))
	case float64:
		y := b.(float64)
		if x < y {
			return -1
		} else if x > y {
			return 1
		}
		return 0
	case string:
		return strings.Compare(x, b.(string))
	default:
		panic(fmt.Sprintf("unsupported type for comparison: %T", a))
	}
}

// String returns a string representation of the BinaryExpr
func (b *PhysicalBinaryExpr) String() string {
	return fmt.Sprintf("(%v %s %v)", b.Left, b.Op, b.Right)
}

// ScanExec represents a physical execution plan for scanning data from a data source
type ScanExec struct {
	ds         DataSource
	projection []string
}

// NewScanExec creates a new ScanExec instance
func NewScanExec(ds DataSource, projection []string) *ScanExec {
	return &ScanExec{
		ds:         ds,
		projection: projection,
	}
}

// Schema returns the schema of the ScanExec
// Schema returns the schema of the ScanExec
func (s *ScanExec) Schema() (Schema, error) {
	return s.ds.Schema().SelectNames(s.projection)
}

// Children returns an empty slice as ScanExec is a leaf node
func (s *ScanExec) Children() []PhysicalPlan {
	return []PhysicalPlan{}
}

// Execute performs the scan operation and returns a Sequence of RecordBatches
func (s *ScanExec) Execute() (Sequence, error) {
	return s.ds.Scan(s.projection)
}

// String returns a string representation of the ScanExec
func (s *ScanExec) String() string {
	schema, err := s.Schema()
	if err != nil {
		return fmt.Sprintf("ScanExec: schema=Error(%v), projection=%v", err, s.projection)
	}
	return fmt.Sprintf("ScanExec: schema=%v, projection=%v", schema, s.projection)
}

// ProjectionExec represents a physical execution plan for projection operations
type ProjectionExec struct {
	input  PhysicalPlan
	schema Schema
	expr   []PhysicalExpr
}

// NewProjectionExec creates a new ProjectionExec instance
func NewProjectionExec(input PhysicalPlan, schema Schema, expr []PhysicalExpr) *ProjectionExec {
	return &ProjectionExec{
		input:  input,
		schema: schema,
		expr:   expr,
	}
}

// Schema returns the schema of the ProjectionExec
func (p *ProjectionExec) Schema() (Schema, error) {
	return p.schema, nil
}

// Children returns the child physical plans
func (p *ProjectionExec) Children() []PhysicalPlan {
	return []PhysicalPlan{p.input}
}

// Execute performs the projection operation and returns a Sequence of RecordBatches
func (p *ProjectionExec) Execute() (Sequence, error) {
	inputSeq, err := p.input.Execute()
	if err != nil {
		return Sequence{}, err
	}

	var projectedBatches []RecordBatch
	for _, batch := range inputSeq.Batches {
		columns := make([]Column, len(p.expr))
		for i, expr := range p.expr {
			col, err := expr.Evaluate(batch)
			if err != nil {
				return Sequence{}, err
			}
			columns[i] = col
		}
		projectedBatch, err := NewRecordBatch(p.schema, columns, batch.n)
		if err != nil {
			return Sequence{}, err
		}
		projectedBatches = append(projectedBatches, projectedBatch)
	}

	return Sequence{Batches: projectedBatches}, nil
}

// String returns a string representation of the ProjectionExec
func (p *ProjectionExec) String() string {
	return fmt.Sprintf("ProjectionExec: expr=%v", p.expr)
}

// SelectionExec represents a physical execution plan for selection (filter) operations
type SelectionExec struct {
	input PhysicalPlan
	expr  PhysicalExpr
}

// NewSelectionExec creates a new SelectionExec instance
func NewSelectionExec(input PhysicalPlan, expr PhysicalExpr) *SelectionExec {
	return &SelectionExec{
		input: input,
		expr:  expr,
	}
}

// Schema returns the schema of the SelectionExec
func (s *SelectionExec) Schema() (Schema, error) {
	return s.input.Schema()
}

// Children returns the child physical plans
func (s *SelectionExec) Children() []PhysicalPlan {
	return []PhysicalPlan{s.input}
}

// Execute performs the selection operation and returns a Sequence of RecordBatches
func (s *SelectionExec) Execute() (Sequence, error) {
	inputSeq, err := s.input.Execute()
	if err != nil {
		return Sequence{}, err
	}

	var outputBatches []RecordBatch
	for _, batch := range inputSeq.Batches {
		result, err := s.expr.Evaluate(batch)
		if err != nil {
			return Sequence{}, err
		}

		boolColumn, ok := result.(*SliceColumn)
		if !ok || boolColumn.Type() != Bytes {
			return Sequence{}, fmt.Errorf("selection expression must evaluate to a boolean column")
		}

		filteredBatch, err := s.filterBatch(batch, boolColumn)
		if err != nil {
			return Sequence{}, err
		}

		outputBatches = append(outputBatches, filteredBatch)
	}

	return Sequence{Batches: outputBatches}, nil
}

// filterBatch applies the selection filter to a single RecordBatch
func (s *SelectionExec) filterBatch(input RecordBatch, filter *SliceColumn) (RecordBatch, error) {
	filteredColumns := make([]Column, len(input.fields))
	for i, col := range input.fields {
		filteredCol, err := s.filterColumn(col, filter)
		if err != nil {
			return RecordBatch{}, err
		}
		filteredColumns[i] = filteredCol
	}

	return NewRecordBatch(input.schema, filteredColumns, filteredColumns[0].N())
}

// filterColumn applies the selection filter to a single Column
func (s *SelectionExec) filterColumn(col Column, filter *SliceColumn) (Column, error) {
	var filteredItems []any
	for i := 0; i < col.N(); i++ {
		keep, err := filter.At(i)
		if err != nil {
			return nil, err
		}
		if keep.(bool) {
			value, err := col.At(i)
			if err != nil {
				return nil, err
			}
			filteredItems = append(filteredItems, value)
		}
	}
	return &SliceColumn{items: filteredItems, dtype: col.Type()}, nil
}

// String returns a string representation of the SelectionExec
func (s *SelectionExec) String() string {
	return fmt.Sprintf("SelectionExec: %s", s.expr)
}
