package v1

import (
	"context"
	"sort"
	"sync"

	"github.com/efficientgo/core/errors"
	"github.com/grafana/dskit/concurrency"
	"github.com/prometheus/common/model"
)

type request struct {
	ctx      context.Context
	fp       model.Fingerprint
	chks     ChunkRefs
	searches [][]byte
	response chan output
}

type SeriesChunks struct {
	fp   model.Fingerprint
	chks ChunkRefs
}

type Rpc struct {
	ctx      context.Context
	inputs   []SeriesChunks
	searches [][]byte

	results  chan output
	removals map[model.Fingerprint]ChunkRefs
}

func newRpc(ctx context.Context, inputs []SeriesChunks, searches [][]byte) *Rpc {
	return &Rpc{
		ctx:      ctx,
		inputs:   inputs,
		searches: searches,
		results:  make(chan output),
		removals: make(map[model.Fingerprint]ChunkRefs),
	}
}

func (r *Rpc) Slice(min, max int) PeekingIterator[request] {
	reqItr := NewMapIter[SeriesChunks, request](
		NewSliceIter[SeriesChunks](r.inputs[min:max]),
		func(sc SeriesChunks) request {
			return request{
				ctx:      r.ctx,
				fp:       sc.fp,
				chks:     sc.chks,
				searches: r.searches,
				response: r.results,
			}
		},
	)

	return NewPeekingIter[request](
		// we need to cancel the request iterator when the rpc is canceled.
		NewCancelableIter[request](r.ctx, reqItr),
	)
}

// output represents a chunk that failed to pass all searches
// and must be downloaded
type output struct {
	fp model.Fingerprint
	// chunks which can be removed
	ignore ChunkRefs
}

// Fuse combines multiple requests into a single loop iteration
// over the data set and returns the corresponding outputs
// TODO(owen-d): better async control
func (bq *BlockQuerier) Fuse(inputs []PeekingIterator[request]) *FusedQuerier {
	return NewFusedQuerier(bq, inputs)
}

// BlockIter is an interface for iterating over a block
// It gives access to the underlying linked iterators individually for performance,
// but calling `Next()` on the top level iterator will re-syncronize the underlying iterators
type BlockIter interface {
	// main iter
	SeekIter[model.Fingerprint, *SeriesWithBloom]
	// access to underlying specific iters for more performant access
	Series() SeekIter[model.Fingerprint, *SeriesWithOffset]
	Blooms() SeekIter[BloomOffset, *Bloom]
	// bounds
	FingerprintBounds() FingerprintBounds
}

type FusedQuerier struct {
	block  BlockIter
	inputs Iterator[[]request]
}

func NewFusedQuerier(b BlockIter, inputs []PeekingIterator[request]) *FusedQuerier {
	heap := NewHeapIterator[request](
		func(a, b request) bool {
			return a.fp < b.fp
		},
		inputs...,
	)

	merging := NewDedupingIter[request, []request](
		func(a request, b []request) bool {
			return a.fp == b[0].fp
		},
		func(a request) []request { return []request{a} },
		func(a request, b []request) []request {
			return append(b, a)
		},
		NewPeekingIter[request](heap),
	)
	return &FusedQuerier{
		block:  b,
		inputs: merging,
	}
}

func (fq *FusedQuerier) Run() error {
	for fq.inputs.Next() {
		// find all queries for the next relevant fingerprint
		nextBatch := fq.inputs.At()

		fp := nextBatch[0].fp

		// advance the series iterator to the next fingerprint
		if err := fq.block.Seek(fp); err != nil {
			return errors.Wrap(err, "seeking to fingerprint")
		}

		seriesItr := fq.block.Series()
		if !seriesItr.Next() {
			// no more series, we're done since we're iterating desired fingerprints in order
			return nil
		}

		series := seriesItr.At()
		if series.Fingerprint != fp {
			// fingerprint not found, can't remove chunks
			for _, input := range nextBatch {
				input.response <- output{
					fp:     fp,
					ignore: nil,
				}
			}
		}

		// Now that we've found the series, we need to unpack the bloom
		bloomItr := fq.block.Blooms()
		if err := bloomItr.Seek(series.Offset); err != nil {
			return errors.Wrapf(err, "seeking to bloom for series: %v", series.Fingerprint)
		}

		if !bloomItr.Next() {
			// fingerprint not found, can't remove chunks
			for _, input := range nextBatch {
				input.response <- output{
					fp:     fp,
					ignore: nil,
				}
			}
			continue
		}

		bloom := bloomItr.At()

		// test every input against this chunk
	inputLoop:
		for _, input := range nextBatch {
			_, inBlooms := input.chks.Compare(series.Chunks, true)

			// First, see if the search passes the series level bloom before checking for chunks individually
			for _, search := range input.searches {
				if !bloom.Test(search) {
					// the entire series bloom didn't pass one of the searches,
					// so we can skip checking chunks individually.
					// We still return all chunks that are not included in the bloom
					// as they may still have the data
					input.response <- output{
						fp:     fp,
						ignore: nil,
					}
					continue inputLoop
				}
			}

			out := output{
				fp:     fp,
				ignore: nil, // TODO(owen-d): pool
			}
		chunkLoop:
			for _, chk := range inBlooms {
				for _, search := range input.searches {
					// TODO(owen-d): meld chunk + search into a single byte slice from the block schema
					var combined = search

					if !bloom.ScalableBloomFilter.Test(combined) {
						out.ignore = append(out.ignore, chk)
						continue chunkLoop
					}
				}
			}

			input.response <- out
		}

	}

	return nil
}

type FingerprintBounds struct {
	Min, Max model.Fingerprint
}

// Cmp returns the fingerprint's position relative to the bounds
func (b FingerprintBounds) Cmp(fp model.Fingerprint) BoundsCheck {
	if fp < b.Min {
		return Before
	} else if fp > b.Max {
		return After
	}
	return Overlap
}

func partitionFingerprintRange(rpc *Rpc, consumers []FingerprintBounds) (res []PeekingIterator[request]) {
	for _, cons := range consumers {
		min := sort.Search(len(rpc.inputs), func(i int) bool {
			return cons.Cmp(rpc.inputs[i].fp) > Before
		})

		max := sort.Search(len(rpc.inputs), func(i int) bool {
			return cons.Cmp(rpc.inputs[i].fp) == After
		})

		// All fingerprints fall outside of the consumer's range
		if min == len(rpc.inputs) || max == 0 {
			// TODO(owen-d): better way to express that we don't need this block
			res = append(res, nil)
			continue
		}

		res = append(res, rpc.Slice(min, max))
	}
	return res
}

/*
Given a set of overlapping blocks and overlapping queries,
We send the relevant queries to the relevant blocks. In the case of overlaps,
the same query can be sent to multiple blocks.

blocks
-------                      queries
|      |-                    -------
|      | |-                  |      |-
|      | | | 							   |      | |-
|      | | |                 |      | | |
-------  | |                 -------  | |

	 |			 | |                  |			  | |
	 --------  |                  --------  |
	  |        |                   |        |
		---------                   	---------
*/
func fuseBlocks(rpcs []*Rpc, blocks []*BlockQuerier) (res []*FusedQuerier) {
	bounds := Map(blocks, func(bq *BlockQuerier) FingerprintBounds {
		return bq.FingerprintBounds()
	})

	computations := make([][]PeekingIterator[request], len(bounds))

	// group queries by request & block
	for _, rpc := range rpcs {
		partitions := partitionFingerprintRange(rpc, bounds)
		for i, partition := range partitions {
			if partition != nil {
				computations[i] = append(computations[i], partition)
			}
		}
	}

	// Assign any relevant computations to their blocks & fuse them for single pass iteration
	for i, comp := range computations {
		if len(comp) > 0 {
			res = append(res, blocks[i].Fuse(comp))
		}
	}

	return
}

// TODO(owen-d): allow early return of individual RPCs.
// Currently, we wait for all RPCs to finish
// before returning any results
func runRPCs(rpcs []*Rpc, blocks []*BlockQuerier) error {
	fusedBlocks := fuseBlocks(rpcs, blocks)

	// another goroutine to collect results
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		_ = concurrency.ForEachJob(
			context.Background(),
			len(rpcs),
			len(rpcs),
			func(_ context.Context, i int) error {
				rpc := rpcs[i]
				for x := range rpc.results {
					rpc.removals[x.fp] = rpc.removals[x.fp].Union(x.ignore)
				}
				return nil
			},
		)
		wg.Done()
	}()

	for _, b := range fusedBlocks {
		if err := b.Run(); err != nil {
			return err
		}
	}

	for _, rpc := range rpcs {
		close(rpc.results)
	}
	wg.Wait()
	return nil

}
