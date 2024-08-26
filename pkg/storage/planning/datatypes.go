package planning

import (
	"fmt"
)

// Most of this is experimental+incomplete. It's just enough to build the library structure around it,
// but I expect to extend/reimpl parts as we mature.

// DataTypeSignal is a poor man's "sum type" to signal the format of a referenced field.
type DataTypeSignal uint

// Initial data types supported
const (
	F64 DataTypeSignal = iota
	I64
	Bytes
	String
)

func (d DataTypeSignal) Valid() bool {
	switch d {
	case F64, I64, Bytes, String:
		return true
	default:
		return false
	}
}

// FieldInfo represents information about a field in a data structure
// Name is the identifier of the field
// DType indicates the data type of the field
type FieldInfo struct {
	Name  string
	DType DataTypeSignal
}

// function to generate fieldinfo from (name, dtype)
func NewFieldInfo(name string, dtype DataTypeSignal) FieldInfo {
	return FieldInfo{
		Name:  name,
		DType: dtype,
	}
}

// Schema represents the structure of data
type Schema struct {
	Fields []FieldInfo
}

// NewSchema creates a new Schema from a list of FieldInfo
func NewSchema(fields ...FieldInfo) Schema {
	return Schema{
		Fields: fields,
	}
}

// GetFieldIndex retrieves the FieldInfo for a given index
func (s Schema) GetFieldIndex(i int) (field FieldInfo, err error) {
	if i < 0 || i >= len(s.Fields) {
		return FieldInfo{}, fmt.Errorf("index out of range")
	}
	return s.Fields[i], nil
}

// GetFieldByName returns the first field matching a given name
func (s Schema) GetFieldByName(name string) (FieldInfo, error) {
	for _, field := range s.Fields {
		if field.Name == name {
			return field, nil
		}
	}
	return FieldInfo{}, fmt.Errorf("field %s not found", name)
}

// Select returns a sub-schema if all the fields exist
func (s Schema) Select(fields []FieldInfo) (Schema, error) {
	selectedFields := make([]FieldInfo, 0, len(fields))
	for _, field := range fields {
		existingField, err := s.GetFieldByName(field.Name)
		if err != nil {
			return Schema{}, fmt.Errorf("field '%s' not found in schema", field.Name)
		}
		if existingField.DType != field.DType {
			return Schema{}, fmt.Errorf("field '%s' type mismatch: expected %v, got %v", field.Name, existingField.DType, field.DType)
		}
		selectedFields = append(selectedFields, existingField)
	}
	return NewSchema(selectedFields...), nil
}

// SelectNames is like Select but does not enforce datatype.
// Helpful for initially populating schemas from `Scan`s
func (s Schema) SelectNames(names []string) (Schema, error) {
	selectedFields := make([]FieldInfo, 0, len(names))
	for _, name := range names {
		field, err := s.GetFieldByName(name)
		if err != nil {
			return Schema{}, fmt.Errorf("field '%s' not found in schema", name)
		}
		selectedFields = append(selectedFields, field)
	}
	return NewSchema(selectedFields...), nil
}

// Column represents a column of data in a RecordBatch
type Column interface {
	// N returns the number of elements in the column
	N() int
	// Type returns the data type of the column
	Type() DataTypeSignal
	// At retrieves the element at the specified index
	At(int) (any, error)
}

// RecordBatch represents a collection of columns with a common schema
type RecordBatch struct {
	schema Schema
	fields []Column // The actual data columns
	n      int      // The number of rows in the RecordBatch
}

// GetFieldIndex retrieves the FieldInfo and Column for a given index
func (r RecordBatch) GetFieldIndex(i int) (field FieldInfo, col Column, err error) {
	field, err = r.schema.GetFieldIndex(i)
	if err != nil {
		return FieldInfo{}, nil, err
	}
	if i >= len(r.fields) {
		return field, nil, fmt.Errorf("column data not available for field %s", field.Name)
	}
	return field, r.fields[i], nil
}

// GetFieldByName returns the first field matching a given name
func (r RecordBatch) GetFieldByName(name string) (FieldInfo, Column, error) {
	field, err := r.schema.GetFieldByName(name)
	if err != nil {
		return FieldInfo{}, nil, err
	}
	for i, f := range r.schema.Fields {
		if f.Name == name {
			if i < len(r.fields) {
				return field, r.fields[i], nil
			}
			return field, nil, fmt.Errorf("column data not available for field %s", name)
		}
	}
	return FieldInfo{}, nil, fmt.Errorf("field %s not found", name)
}

// NewRecordBatch creates a new RecordBatch with the given schema, fields, and number of records
func NewRecordBatch(schema Schema, fields []Column, n int) (RecordBatch, error) {
	if len(schema.Fields) != len(fields) {
		return RecordBatch{}, fmt.Errorf("mismatch between schema and field count")
	}
	return RecordBatch{
		schema: schema,
		fields: fields,
		n:      n,
	}, nil
}
