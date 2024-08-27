package planning

import (
	"fmt"
	"strings"
)

var (
	_ LogicalPlan = &Selection{}
	_ LogicalPlan = &Projection{}
)

// LogicalPlan represents a logical query plan node in a query execution tree.
// It defines methods to retrieve the schema and children of the plan node.
type LogicalPlan interface {
	Schema() Schema
	Children() []LogicalPlan
}

// Here is an example of a logical plan formatted using this method.
// Projection: #id, #first_name, #last_name, #state, #salary
//
//	Filter: #state = 'CO'
//	  Scan: employee.csv; projection=None
type LogicalExpr interface {
	FieldInfo(input LogicalPlan) (FieldInfo, error)
	Format(indent int) string
}

// CompatibleSchema checks if the given expressions are compatible with the logical plan's schema
// and returns a new schema based on the expressions.
func CompatibleSchema(plan LogicalPlan, exprs ...LogicalExpr) (Schema, error) {
	// Initialize a slice to store the field information
	fields := make([]FieldInfo, 0, len(exprs))
	// Use a map to track unique fields and detect conflicts
	fieldMap := make(map[FieldInfo]struct{})

	// Iterate through each expression
	for _, expr := range exprs {
		// Get the field information for the expression
		field, err := expr.FieldInfo(plan)
		if err != nil {
			return Schema{}, fmt.Errorf("incompatible expression: %v", err)
		}

		// Check for conflicts
		if _, exists := fieldMap[field]; !exists {
			fields = append(fields, field)
			// Add the field to our tracking structures
			fieldMap[field] = struct{}{}
		}
	}

	// Create and return the new schema
	return plan.Schema().Select(fields)
}

// Here are examples of logical exprs
//
// Literal Value	"hello", 12.34
// Column Reference	user_id, first_name, last_name
// Math Expression	salary * state_tax
// Comparison Expression	x >= y
// Boolean Expression	birthday = today() AND age >= 21
// Aggregate Expression	MIN(salary), MAX(salary), SUM(salary), AVG(salary), COUNT(*)
// Scalar Function	CONCAT(first_name, " ", last_name)
// ColumnReference Logical Expression
type ColumnReference struct {
	Name string
}

// shorthand to create a column reference
func ColumnRef(name string) *ColumnReference {
	return &ColumnReference{Name: name}
}

// FieldInfo returns the field information for the ColumnReference
// It retrieves the field from the input LogicalPlan's schema based on the column name
func (c *ColumnReference) FieldInfo(input LogicalPlan) (FieldInfo, error) {
	return input.Schema().GetFieldByName(c.Name)
}

func (c *ColumnReference) Format(indent int) string {
	return fmt.Sprintf("%s#%s", strings.Repeat(" ", indent), c.Name)
}

// LiteralValue Logical Expression
type LiteralValue struct {
	Value any
	Type  DataTypeSignal
}

// Literal is shorthand for creating a LiteralValue
func Literal(value any, dtype DataTypeSignal) *LiteralValue {
	return &LiteralValue{
		Value: value,
		Type:  dtype,
	}
}

// Helper functions for common literal types
func LiteralString(value any) *LiteralValue {
	return &LiteralValue{
		Value: value,
		Type:  String,
	}
}

func LiteralInt64(value int64) *LiteralValue {
	return &LiteralValue{
		Value: value,
		Type:  I64,
	}
}

func LiteralFloat64(value float64) *LiteralValue {
	return &LiteralValue{
		Value: value,
		Type:  F64,
	}
}

func (l *LiteralValue) FieldInfo(input LogicalPlan) (FieldInfo, error) {
	return NewFieldInfo(fmt.Sprintf("%v", l.Value), l.Type), nil
}

func (l *LiteralValue) Format(indent int) string {
	return fmt.Sprintf("%s%v", strings.Repeat(" ", indent), l.Value)
}

// BinaryExpr Logical Expression for mathematical and comparison operations
type BinaryExpr struct {
	Name  string
	Op    string
	Left  LogicalExpr
	Right LogicalExpr
}

func (b *BinaryExpr) FieldInfo(input LogicalPlan) (FieldInfo, error) {
	leftField, err := b.Left.FieldInfo(input)
	if err != nil {
		return FieldInfo{}, err
	}
	rightField, err := b.Right.FieldInfo(input)
	if err != nil {
		return FieldInfo{}, err
	}
	// Determine the resulting type based on the operation and operand types
	resultType, err := determineResultType(b.Op, leftField.DType, rightField.DType)
	if err != nil {
		return FieldInfo{}, fmt.Errorf("failed to determine result type: %w", err)
	}

	return NewFieldInfo(b.Name, resultType), nil
}

func (b *BinaryExpr) Format(indent int) string {
	indentation := strings.Repeat(" ", indent)
	return fmt.Sprintf("%s%s (%s)\n%s\n%s",
		indentation, b.Name, b.Op,
		b.Left.Format(indent+2),
		b.Right.Format(indent+2))
}

// Helper function to determine the result type of a binary operation
func determineResultType(op string, left, right DataTypeSignal) (DataTypeSignal, error) {
	if !left.Valid() || !right.Valid() {
		return 0, fmt.Errorf("invalid data type")
	}

	if left != right {
		return 0, fmt.Errorf("incompatible types: %v and %v", left, right)
	}

	switch op {
	case "+", "-", "*", "/":
		if left == String {
			return 0, fmt.Errorf("arithmetic operations not supported for string type")
		}
		return left, nil
	case "=", "!=", "<", "<=", ">", ">=":
		return left, nil
	default:
		return 0, fmt.Errorf("unknown operation: %s", op)
	}
}

// Scan logical plan
type Scan struct {
	Path       string
	DataSource DataSource
	Projection []string
	schema     Schema
}

// NewScan creates a new Scan logical plan
func NewScan(path string, dataSource DataSource, projection []string) (*Scan, error) {
	s := &Scan{
		Path:       path,
		DataSource: dataSource,
		Projection: projection,
	}
	schema, err := s.deriveSchema()
	if err != nil {
		return nil, fmt.Errorf("failed to derive schema: %w", err)
	}
	s.schema = schema
	return s, nil
}

// Schema returns the schema of the Scan
func (s *Scan) Schema() Schema {
	return s.schema
}

// Children returns nil as Scan has no child plans
func (s *Scan) Children() []LogicalPlan {
	return nil
}

// deriveSchema derives the schema based on the projection
func (s *Scan) deriveSchema() (Schema, error) {
	schema := s.DataSource.Schema()
	if len(s.Projection) == 0 {
		return schema, nil
	}

	return schema.SelectNames(s.Projection)
}

// String returns a string representation of the Scan
func (s *Scan) String() string {
	if len(s.Projection) == 0 {
		return fmt.Sprintf("Scan: %s; projection=None", s.Path)
	}
	return fmt.Sprintf("Scan: %s; projection=%v", s.Path, s.Projection)
}

// Projection logical plan
type Projection struct {
	input  LogicalPlan
	expr   []LogicalExpr
	schema Schema
}

// NewProjection creates a new Projection logical plan
func NewProjection(input LogicalPlan, expr []LogicalExpr) (*Projection, error) {
	schema, err := CompatibleSchema(input, expr...)
	if err != nil {
		return nil, err
	}
	return &Projection{
		input:  input,
		expr:   expr,
		schema: schema,
	}, nil
}

// Schema returns the schema of the Projection
func (p *Projection) Schema() Schema {
	fields := make([]FieldInfo, len(p.expr))
	for i, e := range p.expr {
		field, _ := e.FieldInfo(p.input)
		fields[i] = field
	}
	return NewSchema(fields...)
}

// Children returns the child plans of the Projection
func (p *Projection) Children() []LogicalPlan {
	return []LogicalPlan{p.input}
}

// String returns a string representation of the Projection
func (p *Projection) String() string {
	exprStrings := make([]string, len(p.expr))
	for i, e := range p.expr {
		exprStrings[i] = e.Format(0)
	}
	return fmt.Sprintf("Projection: %s", strings.Join(exprStrings, ", "))
}

// Selection (also known as Filter) logical plan
type Selection struct {
	input  LogicalPlan
	expr   LogicalExpr
	schema Schema
}

// NewSelection creates a new Selection logical plan
func NewSelection(input LogicalPlan, expr LogicalExpr) (*Selection, error) {
	schema, err := CompatibleSchema(input, expr)
	if err != nil {
		return nil, err
	}

	return &Selection{
		input:  input,
		expr:   expr,
		schema: schema,
	}, nil
}

// Schema returns the schema of the Selection
func (s *Selection) Schema() Schema {
	// selection does not change the schema of the input
	return s.input.Schema()
}

// Children returns the child plans of the Selection
func (s *Selection) Children() []LogicalPlan {
	return []LogicalPlan{s.input}
}

// String returns a string representation of the Selection
func (s *Selection) String() string {
	return fmt.Sprintf("Filter: %s", s.expr.Format(0))
}
