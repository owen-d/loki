package v1

import (
	"bytes"
	"context"
	"sync"
	"testing"

	"github.com/grafana/dskit/concurrency"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"

	"github.com/grafana/loki/pkg/chunkenc"
)

func seedBlockQuerier(t *testing.T, numSeries int, startFp, endFp model.Fingerprint) (*BlockQuerier, []SeriesWithBloom) {
	// references for linking in memory reader+writer
	indexBuf := bytes.NewBuffer(nil)
	bloomsBuf := bytes.NewBuffer(nil)
	writer := NewMemoryBlockWriter(indexBuf, bloomsBuf)
	reader := NewByteReader(indexBuf, bloomsBuf)
	numKeysPerSeries := 10000
	data, _ := mkBasicSeriesWithBlooms(numSeries, numKeysPerSeries, startFp, endFp, 0, 10000)

	builder, err := NewBlockBuilder(
		BlockOptions{
			schema: Schema{
				version:  DefaultSchemaVersion,
				encoding: chunkenc.EncSnappy,
			},
			SeriesPageSize: 100,
			BloomPageSize:  10 << 10,
		},
		writer,
	)
	require.Nil(t, err)
	itr := NewSliceIter[SeriesWithBloom](data)
	require.Nil(t, builder.BuildFrom(itr))
	block := NewBlock(reader)
	querier := NewBlockQuerier(block)
	return querier, data
}

func TestFusedQuerier(t *testing.T) {
	numSeries := 100
	querier, data := seedBlockQuerier(t, 100, 0, 0xffff)

	nReqs := 10
	var inputs [][]request
	for i := 0; i < nReqs; i++ {
		ch := make(chan output)
		var reqs []request
		// find 2 series for each
		for j := 0; j < 2; j++ {
			idx := numSeries/nReqs*i + j
			reqs = append(reqs, request{
				fp:       data[idx].Series.Fingerprint,
				chks:     data[idx].Series.Chunks,
				response: ch,
			})
		}
		inputs = append(inputs, reqs)
	}

	var itrs []PeekingIterator[request]
	for _, reqs := range inputs {
		itrs = append(itrs, NewPeekingIter[request](NewSliceIter[request](reqs)))
	}

	resps := make([][]output, nReqs)
	var g sync.WaitGroup
	g.Add(1)
	go func() {
		require.Nil(t, concurrency.ForEachJob(
			context.Background(),
			len(resps),
			len(resps),
			func(_ context.Context, i int) error {
				for v := range inputs[i][0].response {
					resps[i] = append(resps[i], v)
				}
				return nil
			},
		))
		g.Done()
	}()

	fused := querier.Fuse(itrs)

	require.Nil(t, fused.Run())
	for _, input := range inputs {
		close(input[0].response)
	}
	g.Wait()

	for i, input := range inputs {
		for j, req := range input {
			resp := resps[i][j]
			require.Equal(
				t,
				output{
					fp:     req.fp,
					ignore: nil,
				},
				resp,
			)
		}
	}
}

func TestMultiFuse(t *testing.T) {
	// a few nonoverlapping blocks
	minFp, maxFp := 0, 4000
	nBlocks := 4
	nReqs := 10
	fpPerReq := 100
	var fpRanges []FingerprintBounds
	for i := 0; i < nBlocks; i++ {
		fpRanges = append(fpRanges, FingerprintBounds{
			Min: model.Fingerprint(minFp + i*(maxFp-minFp)/nBlocks),
			Max: model.Fingerprint(minFp + (i+1)*(maxFp-minFp)/nBlocks),
		})
	}

	qs := Map(
		fpRanges,
		func(x FingerprintBounds) *BlockQuerier {
			querier, _ := seedBlockQuerier(t, 100, x.Min, x.Max)
			return querier
		},
	)

	reqs := make([]*Rpc, nReqs)
	for i := 0; i < nReqs; i++ {
		var inputs []SeriesChunks
		for j := 0; j < fpPerReq; j++ {
			inputs = append(inputs, SeriesChunks{
				fp: model.Fingerprint(i*(maxFp-minFp)/nBlocks + j),
			})
		}

		reqs[i] = newRpc(context.Background(), inputs, nil)
	}

	require.Nil(t, runRPCs(reqs, qs))
}

func TestPartitionFingerprintRange(t *testing.T) {
	queries := []SeriesChunks{
		{
			fp: 0,
		},
		{
			fp: 1,
		},
		{
			fp: 2,
		},
		{
			fp: 3,
		},
	}

	consumers := []FingerprintBounds{
		{0, 2},
		{1, 3},
		{5, 7},
	}

	partitions := partitionFingerprintRange(newRpc(context.Background(), queries, nil), consumers)
	expected := [][]model.Fingerprint{
		{
			0, 1, 2,
		},
		{
			1, 2, 3,
		},
		nil,
	}

	var got [][]model.Fingerprint
	for _, p := range partitions {
		if p == nil {
			got = append(got, nil)
			continue
		}

		got = append(got, Collect[model.Fingerprint](
			NewMapIter[request, model.Fingerprint](
				p,
				func(x request) model.Fingerprint {
					return x.fp
				},
			),
		))
	}

	require.Equal(t, expected, got)

}

func setupBlockForBenchmark(b *testing.B) (*BlockQuerier, [][]request) {
	indexBuf := bytes.NewBuffer(nil)
	bloomsBuf := bytes.NewBuffer(nil)
	writer := NewMemoryBlockWriter(indexBuf, bloomsBuf)
	reader := NewByteReader(indexBuf, bloomsBuf)
	numSeries := 10000
	numKeysPerSeries := 100
	data, _ := mkBasicSeriesWithBlooms(numSeries, numKeysPerSeries, 0, 0xffffff, 0, 10000)

	builder, err := NewBlockBuilder(
		BlockOptions{
			schema: Schema{
				version:  DefaultSchemaVersion,
				encoding: chunkenc.EncSnappy,
			},
			SeriesPageSize: 256 << 10, // 256k
			BloomPageSize:  1 << 20,   // 1MB
		},
		writer,
	)
	require.Nil(b, err)
	itr := NewSliceIter[SeriesWithBloom](data)
	require.Nil(b, builder.BuildFrom(itr))
	block := NewBlock(reader)
	querier := NewBlockQuerier(block)

	numRequestChains := 100
	seriesPerRequest := 100
	var requestChains [][]request
	for i := 0; i < numRequestChains; i++ {
		var reqs []request
		// ensure they use the same channel
		ch := make(chan output)
		// evenly spread out the series queried within a single request chain
		// to mimic series distribution across keyspace
		for j := 0; j < seriesPerRequest; j++ {
			// add the chain index (i) for a little jitter
			idx := numSeries*j/seriesPerRequest + i
			if idx >= numSeries {
				idx = numSeries - 1
			}
			reqs = append(reqs, request{
				fp:       data[idx].Series.Fingerprint,
				chks:     data[idx].Series.Chunks,
				response: ch,
			})
		}
		requestChains = append(requestChains, reqs)
	}

	return querier, requestChains
}

func BenchmarkBlockQuerying(b *testing.B) {
	b.StopTimer()
	querier, requestChains := setupBlockForBenchmark(b)
	// benchmark
	b.StartTimer()

	b.Run("single-pass", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			for _, chain := range requestChains {
				for _, req := range chain {
					_, _ = querier.CheckChunksForSeries(req.fp, req.chks, nil)
				}
			}
		}

	})
	b.Run("fused", func(b *testing.B) {
		// spin up some goroutines to consume the responses so they don't block
		go func() {
			require.Nil(b, concurrency.ForEachJob(
				context.Background(),
				len(requestChains), len(requestChains),
				func(_ context.Context, idx int) error {
					for range requestChains[idx][0].response {
					}
					return nil
				},
			))
		}()

		var itrs []PeekingIterator[request]

		for i := 0; i < b.N; i++ {
			itrs = itrs[:0]
			for _, reqs := range requestChains {
				itrs = append(itrs, NewPeekingIter[request](NewSliceIter[request](reqs)))
			}
			fused := querier.Fuse(itrs)
			_ = fused.Run()
		}
	})

}
