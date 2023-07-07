package spandex

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_bucketFor(t *testing.T) {
	for _, tc := range []struct {
		dist uint64
		exp  int
	}{
		{
			dist: 0,
			exp:  0,
		},
		{
			dist: 1,
			exp:  0,
		},
		{
			dist: 15,
			exp:  0,
		}, {
			dist: 16,
			exp:  1,
		}, {
			dist: 255,
			exp:  1,
		}, {
			dist: 256,
			exp:  2,
		}, {
			dist: 257,
			exp:  2,
		},
	} {
		t.Run(fmt.Sprint(tc.dist), func(t *testing.T) {
			require.Equal(t, tc.exp, bucketFor(tc.dist))
		})
	}
}
