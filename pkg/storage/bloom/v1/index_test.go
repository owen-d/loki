package v1

import (
	"testing"

	"github.com/grafana/loki/pkg/chunkenc"
	"github.com/grafana/loki/pkg/util/encoding"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
)

// does not include a real bloom offset
func mkBasicSeries(n int, fromFp, throughFp model.Fingerprint, fromTs, throughTs model.Time) []SeriesWithOffset {
	var seriesList []SeriesWithOffset
	for i := 0; i < n; i++ {
		var series SeriesWithOffset
		step := (throughFp - fromFp) / (model.Fingerprint(n))
		series.Fingerprint = fromFp + model.Fingerprint(i)*step
		timeDelta := fromTs + (throughTs-fromTs)/model.Time(n)*model.Time(i)
		series.Chunks = []ChunkRef{
			{
				Start:    fromTs + timeDelta*model.Time(i),
				End:      fromTs + timeDelta*model.Time(i),
				Checksum: uint32(i),
			},
		}
		seriesList = append(seriesList, series)
	}
	return seriesList
}

func mkBasicSeriesPage(series []SeriesWithOffset) SeriesPage {
	pg := SeriesPage{
		Series: series,
	}

	var initialized bool
	var header SeriesHeader

	for _, s := range series {
		if !initialized {
			header.FromFp = s.Fingerprint
			header.FromTs = s.Chunks[0].Start
			initialized = true
		} else {
			if s.Fingerprint < header.FromFp {
				header.FromFp = s.Fingerprint
			}
			if s.Chunks[0].Start < header.FromTs {
				header.FromTs = s.Chunks[0].Start
			}
		}
		header.NumSeries++
		header.ThroughFp = s.Fingerprint
		for _, chk := range s.Chunks {
			if chk.End > header.ThroughTs {
				header.ThroughTs = chk.End
			}
		}
	}

	pg.Header = header
	return pg
}

func mkBasicSeriesPages(pages, series int, fromFp, throughFp model.Fingerprint, fromTs, throughTs model.Time) (res []SeriesPage, resultingSeries []Series) {
	for i := 0; i < pages; i++ {
		fpStep := (throughFp - fromFp) / model.Fingerprint(pages)
		series := mkBasicSeries(series/pages, fromFp+fpStep*model.Fingerprint(i), fromFp+fpStep*model.Fingerprint(i+1), fromTs, throughTs)
		res = append(res, mkBasicSeriesPage(series))
		for _, s := range series {
			resultingSeries = append(resultingSeries, s.Series)
		}
	}
	return
}

func TestBloomOffsetEncoding(t *testing.T) {
	src := BloomOffset{PageOffset: 1, ByteOffset: 2}
	enc := &encoding.Encbuf{}
	src.Encode(enc, BloomOffset{})

	var dst BloomOffset
	dec := encoding.DecWith(enc.Get())
	require.Nil(t, dst.Decode(&dec, BloomOffset{}))

	require.Equal(t, src, dst)
}

func TestSeriesEncoding(t *testing.T) {
	src := SeriesWithOffset{
		Series: Series{
			Fingerprint: model.Fingerprint(1),
			Chunks: []ChunkRef{
				{
					Start:    1,
					End:      2,
					Checksum: 3,
				},
				{
					Start:    4,
					End:      5,
					Checksum: 6,
				},
			},
		},
		Offset: BloomOffset{PageOffset: 2, ByteOffset: 3},
	}

	enc := &encoding.Encbuf{}
	src.Encode(enc, 0, BloomOffset{})

	dec := encoding.DecWith(enc.Get())
	var dst SeriesWithOffset
	fp, offset, err := dst.Decode(&dec, 0, BloomOffset{})
	require.Nil(t, err)
	require.Equal(t, src.Fingerprint, fp)
	require.Equal(t, src.Offset, offset)
	require.Equal(t, src, dst)
}

func TestSeriesPageEncoding(t *testing.T) {
	series := []SeriesWithOffset{
		{
			Series: Series{
				Chunks: []ChunkRef{
					{
						Start:    1,
						End:      2,
						Checksum: 3,
					},
					{
						Start:    4,
						End:      5,
						Checksum: 6,
					},
				},
				Fingerprint: model.Fingerprint(1),
			},
			Offset: BloomOffset{PageOffset: 2, ByteOffset: 3},
		},
		{
			Series: Series{
				Fingerprint: model.Fingerprint(2),
				Chunks: []ChunkRef{
					{
						Start:    7,
						End:      8,
						Checksum: 9,
					},
					{
						Start:    10,
						End:      11,
						Checksum: 12,
					},
				},
			},
			Offset: BloomOffset{PageOffset: 2, ByteOffset: 3},
		},
	}

	src := SeriesPage{
		Header: SeriesHeader{
			NumSeries: len(series),
			FromFp:    series[0].Fingerprint,
			ThroughFp: series[len(series)-1].Fingerprint,
			FromTs:    1,
			ThroughTs: 11,
		},
		Series: series,
	}

	enc := &encoding.Encbuf{}
	decompressedLen, err := src.Encode(enc, chunkenc.GetWriterPool(chunkenc.EncGZIP), Crc32HashPool.Get())
	require.Nil(t, err)

	var dst SeriesPage
	dec := encoding.DecWith(enc.Get())
	require.Nil(t, dst.Decode(&dec, chunkenc.GetReaderPool(chunkenc.EncGZIP), decompressedLen))
}