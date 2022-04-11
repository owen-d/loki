package tsdb

import (
	"context"

	"github.com/grafana/loki/pkg/logproto"
	"github.com/grafana/loki/pkg/storage/chunk"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/model/labels"
)

// IndexAdapter is wraps a tsdb.Index and exposes a stores.Index interface implementation
type IndexAdapter struct {
	idx Index
}

func NewIndexAdapter(idx Index) *IndexAdapter {
	return &IndexAdapter{
		idx: idx,
	}
}

func (i *IndexAdapter) GetChunkRefs(ctx context.Context, userID string, from, through model.Time, matchers ...*labels.Matcher) ([]logproto.ChunkRef, error) {
	return i.idx.GetChunkRefs(ctx, userID, from, through, nil, nil, matchers...)
}

func (i *IndexAdapter) GetSeries(ctx context.Context, userID string, from, through model.Time, matchers ...*labels.Matcher) ([]labels.Labels, error) {
	// TODO(owen-d): is it worth having a superset struct for tsdb.Series
	// given it requires an extra linear cycle here to convert?
	xs, err := i.idx.Series(ctx, userID, from, through, nil, nil, matchers...)
	if err != nil {
		return nil, err
	}

	mapped := make([]labels.Labels, 0, len(xs))
	for _, x := range xs {
		mapped = append(mapped, x.Labels)
	}
	return mapped, nil
}

func (i *IndexAdapter) LabelValuesForMetricName(ctx context.Context, userID string, from, through model.Time, metricName string, labelName string, matchers ...*labels.Matcher) ([]string, error) {
	// Here, we ignore the metricName as it's a relic of the old index and is unused in Loki's TSDB.
	// It was always hardcoded to "logs" prior.
	return i.idx.LabelValues(ctx, userID, from, through, labelName, matchers...)
}

func (i *IndexAdapter) LabelNamesForMetricName(ctx context.Context, userID string, from, through model.Time, metricName string) ([]string, error) {
	// Here, we ignore the metricName as it's a relic of the old index and is unused in Loki's TSDB.
	// It was always hardcoded to "logs" prior.
	return i.idx.LabelNames(ctx, userID, from, through)
}

// SetChunkFilterer sets a chunk filter to be used when retrieving chunks.
// This is only used for GetSeries implementation.
// Todo we might want to pass it as a parameter to GetSeries instead.
func (i *IndexAdapter) SetChunkFilterer(chunkFilter chunk.RequestChunkFilterer) {
	// TODO(owen-d): determine how best to implement this
	panic("unimplemented")
}
