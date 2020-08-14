package storage

import (
	"sync"
	"time"

	"github.com/cortexproject/cortex/pkg/chunk"
	"github.com/grafana/loki/pkg/logql"
	"github.com/prometheus/common/model"
	"github.com/prometheus/prometheus/pkg/labels"
)

// - [ ] map[matcher]map[seriesID]bool // detect if series passes matchers or unknown
// {job="foo"} |= "bar"
// - [ ] map[chunkid]map[filter]BoundsDoesNotExist // keep track of the parts within a chunk that don't contain data for a certain matcher.

type MatchersCache interface {
	PassesMatchers(model.Fingerprint, ...labels.Matcher) (passes bool, ok bool)
	SetMatchers(model.Fingerprint, bool, ...labels.Matcher) error
}

type NoopMatchersCache struct{}

func (NoopMatchersCache) PassesMatchers(_ model.Fingerprint, _ ...labels.Matcher) (passes bool, ok bool) {
	return false, false
}

func (NoopMatchersCache) SetMatchers(_ model.Fingerprint, _ bool, x ...labels.Matcher) error {
	return nil
}

type MemoryMatchersCache struct {
	sync.RWMutex
	matchers map[string]*seriesCache
}

type seriesCache struct {
	sync.RWMutex
	m map[model.Fingerprint]bool
}

func (c *seriesCache) Get(fp model.Fingerprint)

type FilterCache interface {
	FailureBounds(chunk.Chunk, ...logl.FilterExpr) (start time.Time, end time.Time, ok bool)
	SetFailureBounds(chk chunk.Chunk, lBound, rBound time.Time, _ ...logql.FilterExpr) error
}

type NoopFilterCache struct{}

func (NoopFilterCache) FailureBounds(_ chunk.Chunk, _ ...logl.FilterExpr) (start time.Time, end time.Time, ok bool) {
	return start, end, false
}
func (NoopFilterCache) SetFailureBounds(chk chunk.Chunk, lBound, rBound time.Time, _ ...logql.FilterExpr) error {
	return nil
}

type SeriesCache interface {
	MatcherCache
	FilterCache
}

type NoopSeriesCache struct {
	NoopMatchersCache
	NoopFilterCache
}
