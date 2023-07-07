package spandex

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_KeyCmp(t *testing.T) {
	for _, tc := range []struct {
		desc string
		a, b *key
		exp  Cmp
	}{
		{
			desc: "nils",
			exp:  Eq,
		},
		{
			desc: "nil cmp",
			b: func() *key {
				x := key(1)
				return &x
			}(),
			exp: Gt,
		},
		{
			desc: "cmp nil",
			a: func() *key {
				x := key(1)
				return &x
			}(),
			exp: Lt,
		},
		{
			desc: "lt",
			a: func() *key {
				x := key(0)
				return &x
			}(),
			b: func() *key {
				x := key(1)
				return &x
			}(),
			exp: Lt,
		},
		{
			desc: "gt",
			a: func() *key {
				x := key(1)
				return &x
			}(),
			b: func() *key {
				x := key(0)
				return &x
			}(),
			exp: Gt,
		},
		{
			desc: "Eq",
			a: func() *key {
				x := key(1)
				return &x
			}(),
			b: func() *key {
				x := key(1)
				return &x
			}(),
			exp: Eq,
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			require.Equal(t, tc.exp, tc.a.Cmp(tc.b), "expected %s to be %v when compared to %s", tc.a, tc.exp, tc.b)
		})
	}
}
