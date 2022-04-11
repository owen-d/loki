package tsdb

import (
	"context"
	"testing"

	"github.com/go-kit/log"
	"github.com/prometheus/prometheus/model/labels"
	"github.com/stretchr/testify/require"

	"github.com/grafana/loki/pkg/storage/stores/tsdb/index"
)

func TestSingleIdx(t *testing.T) {
	cases := []LoadableSeries{
		{
			Labels: mustParseLabels(`{foo="bar"}`),
			Chunks: []index.ChunkMeta{
				{
					MinTime:  0,
					MaxTime:  3,
					Checksum: 0,
				},
				{
					MinTime:  1,
					MaxTime:  4,
					Checksum: 1,
				},
				{
					MinTime:  2,
					MaxTime:  5,
					Checksum: 2,
				},
			},
		},
		{
			Labels: mustParseLabels(`{foo="bar", bazz="buzz"}`),
			Chunks: []index.ChunkMeta{
				{
					MinTime:  1,
					MaxTime:  10,
					Checksum: 3,
				},
			},
		},
		{
			Labels: mustParseLabels(`{foo="bard", bazz="bozz", bonk="borb"}`),
			Chunks: []index.ChunkMeta{
				{
					MinTime:  1,
					MaxTime:  7,
					Checksum: 4,
				},
			},
		},
	}

	for _, variant := range []struct {
		desc string
		fn   func() *TSDBIndex
	}{
		{
			desc: "file",
			fn: func() *TSDBIndex {
				return BuildIndex(t, t.TempDir(), "fake", cases)
			},
		},
		{
			desc: "head",
			fn: func() *TSDBIndex {
				head := NewHead("fake", NewHeadMetrics(nil), log.NewNopLogger())
				for _, x := range cases {
					head.Append(x.Labels, x.Chunks)
				}
				reader, err := head.Index()
				require.Nil(t, err)
				return NewTSDBIndex(reader)
			},
		},
	} {
		t.Run(variant.desc, func(t *testing.T) {
			idx := variant.fn()
			t.Run("GetChunkRefs", func(t *testing.T) {
				refs, err := idx.GetChunkRefs(context.Background(), "fake", 1, 5, nil, nil, labels.MustNewMatcher(labels.MatchEqual, "foo", "bar"))
				require.Nil(t, err)

				expected := []ChunkRef{
					{
						UserID:      "fake",
						Fingerprint: mustParseLabels(`{foo="bar"}`).Hash(),
						From:        0,
						Through:     3,
						Checksum:    0,
					},
					{
						UserID:      "fake",
						Fingerprint: mustParseLabels(`{foo="bar"}`).Hash(),
						From:        1,
						Through:     4,
						Checksum:    1,
					},
					{
						UserID:      "fake",
						Fingerprint: mustParseLabels(`{foo="bar"}`).Hash(),
						From:        2,
						Through:     5,
						Checksum:    2,
					},
					{
						UserID:      "fake",
						Fingerprint: mustParseLabels(`{foo="bar", bazz="buzz"}`).Hash(),
						From:        1,
						Through:     10,
						Checksum:    3,
					},
				}
				require.Equal(t, expected, refs)
			})

			t.Run("GetChunkRefsSharded", func(t *testing.T) {
				shard := index.ShardAnnotation{
					Shard: 1,
					Of:    2,
				}
				shardedRefs, err := idx.GetChunkRefs(context.Background(), "fake", 1, 5, nil, &shard, labels.MustNewMatcher(labels.MatchEqual, "foo", "bar"))

				require.Nil(t, err)

				require.Equal(t, []ChunkRef{{
					UserID:      "fake",
					Fingerprint: mustParseLabels(`{foo="bar", bazz="buzz"}`).Hash(),
					From:        1,
					Through:     10,
					Checksum:    3,
				}}, shardedRefs)

			})

			t.Run("Series", func(t *testing.T) {
				xs, err := idx.Series(context.Background(), "fake", 8, 9, nil, nil, labels.MustNewMatcher(labels.MatchEqual, "foo", "bar"))
				require.Nil(t, err)

				expected := []Series{
					{
						Labels:      mustParseLabels(`{foo="bar", bazz="buzz"}`),
						Fingerprint: mustParseLabels(`{foo="bar", bazz="buzz"}`).Hash(),
					},
				}
				require.Equal(t, expected, xs)
			})

			t.Run("SeriesSharded", func(t *testing.T) {
				shard := index.ShardAnnotation{
					Shard: 0,
					Of:    2,
				}

				xs, err := idx.Series(context.Background(), "fake", 0, 10, nil, &shard, labels.MustNewMatcher(labels.MatchEqual, "foo", "bar"))
				require.Nil(t, err)

				expected := []Series{
					{
						Labels:      mustParseLabels(`{foo="bar"}`),
						Fingerprint: mustParseLabels(`{foo="bar"}`).Hash(),
					},
				}
				require.Equal(t, expected, xs)
			})

			t.Run("LabelNames", func(t *testing.T) {
				// request data at the end of the tsdb range, but it should return all labels present
				ls, err := idx.LabelNames(context.Background(), "fake", 9, 10)
				require.Nil(t, err)
				require.Equal(t, []string{"bazz", "bonk", "foo"}, ls)
			})

			t.Run("LabelNamesWithMatchers", func(t *testing.T) {
				// request data at the end of the tsdb range, but it should return all labels present
				ls, err := idx.LabelNames(context.Background(), "fake", 9, 10, labels.MustNewMatcher(labels.MatchEqual, "bazz", "buzz"))
				require.Nil(t, err)
				require.Equal(t, []string{"bazz", "foo"}, ls)
			})

			t.Run("LabelValues", func(t *testing.T) {
				vs, err := idx.LabelValues(context.Background(), "fake", 9, 10, "foo")
				require.Nil(t, err)
				require.Equal(t, []string{"bar", "bard"}, vs)
			})

			t.Run("LabelValuesWithMatchers", func(t *testing.T) {
				vs, err := idx.LabelValues(context.Background(), "fake", 9, 10, "foo", labels.MustNewMatcher(labels.MatchEqual, "bazz", "buzz"))
				require.Nil(t, err)
				require.Equal(t, []string{"bar"}, vs)
			})

		})
	}

}
