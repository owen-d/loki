package planning

import (
	"fmt"

	"github.com/grafana/loki/v3/pkg/logproto"
)

// Temporary impl. Will eventually be refactored into an iterator pattern
type Sequence struct {
	Batches []RecordBatch
}

type DataSource interface {
	Schema() Schema
	Scan(projection []string) (Sequence, error)
}

type StreamDataSource struct {
	stream *logproto.Stream
}

// implement DataSource
func (s *StreamDataSource) Schema() Schema {
	return NewSchema(
		NewFieldInfo("timestamp", I64),
		NewFieldInfo("line", Bytes),
		NewFieldInfo("labels", String),
	)
}

// CreateSequenceFromStream creates a Sequence from a logproto.Stream
// Below is the structure of `logproto.Stream`
//
//		type Stream struct {
//			Labels  string  `protobuf:"bytes,1,opt,name=labels,proto3" json:"labels"`
//			Entries []Entry `protobuf:"bytes,2,rep,name=entries,proto3,customtype=EntryAdapter" json:"entries"`
//			Hash    uint64  `protobuf:"varint,3,opt,name=hash,proto3" json:"-"`
//		}
//
//	 Below is the structure of `logproto.Entry`
//
//	type Entry struct {
//		Timestamp          time.Time     `protobuf:"bytes,1,opt,name=timestamp,proto3,stdtime" json:"ts"`
//		Line               string        `protobuf:"bytes,2,opt,name=line,proto3" json:"line"`
//		StructuredMetadata LabelsAdapter `protobuf:"bytes,3,opt,name=structuredMetadata,proto3" json:"structuredMetadata,omitempty"`
//		Parsed             LabelsAdapter `protobuf:"bytes,4,opt,name=parsed,proto3" json:"parsed,omitempty"`
//	}

// Scan creates a Sequence from the StreamDataSource
func (s *StreamDataSource) Scan(projection []string) (Sequence, error) {
	if s.stream == nil {
		return Sequence{}, fmt.Errorf("nil stream")
	}

	n := len(s.stream.Entries)
	if n == 0 {
		return Sequence{}, nil
	}

	// Create columns based on projection
	schema := s.Schema()
	columns := make([]Column, 0, len(projection))

	for _, field := range projection {
		switch field {
		case "timestamp":
			timestampCol := make([]any, n)
			for i, entry := range s.stream.Entries {
				timestampCol[i] = entry.Timestamp.UnixNano()
			}
			columns = append(columns, &SliceColumn{items: timestampCol, dtype: I64})
		case "line":
			lineCol := make([]any, n)
			for i, entry := range s.stream.Entries {
				lineCol[i] = []byte(entry.Line)
			}
			columns = append(columns, &SliceColumn{items: lineCol, dtype: Bytes})
		case "labels":
			columns = append(columns, &LiteralColumn{value: s.stream.Labels, n: n, dtype: String})
		default:
			return Sequence{}, fmt.Errorf("unknown field: %s", field)
		}
	}

	// Create RecordBatch
	batch, err := NewRecordBatch(schema, columns, n)
	if err != nil {
		return Sequence{}, fmt.Errorf("failed to create RecordBatch: %w", err)
	}

	return Sequence{Batches: []RecordBatch{batch}}, nil
}
