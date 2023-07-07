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

func Test_Owned(t *testing.T) {
	type comparisons []struct {
		k   key
		exp bool
	}
	for _, tc := range []struct {
		desc string
		ks   keyspace
		cmps comparisons
	}{
		{
			desc: "full keyspace owns everything",
			ks: keyspace{
				From: 0,
			},
			cmps: comparisons{
				{0, true},
				{1, true},
			},
		},
		{
			desc: "partial keyspace",
			ks: keyspace{
				From: 5,
				Through: func() *key {
					x := key(10)
					return &x
				}(),
			},
			cmps: comparisons{
				{0, false},
				{5, true},
				{6, true},
				{10, false},
			},
		},
		{
			desc: "intersection keyspace",
			ks: func() keyspace {
				ten := key(10)
				fifteen := key(15)
				left := newKeySpace(0, &ten)
				right := newKeySpace(5, &fifteen)
				return left.Intersect(right)
			}(),
			cmps: comparisons{
				{0, false},
				{5, true},
				{9, true},
				{10, false},
			},
		},
		{
			desc: "union keyspace",
			ks: func() keyspace {
				ten := key(10)
				fifteen := key(15)
				left := newKeySpace(1, &ten)
				right := newKeySpace(5, &fifteen)
				return left.Union(right)
			}(),
			cmps: comparisons{
				{0, false},
				{1, true},
				{5, true},
				{10, true},
				{14, true},
				{15, false},
			},
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			for _, exp := range tc.cmps {
				require.Equal(t, exp.exp, tc.ks.Owned(exp.k))
			}
		})
	}
}

func Test_Center(t *testing.T) {
	for _, tc := range []struct {
		desc string
		ks   keyspace
		exp  key
	}{
		{
			desc: "full keyspace",
			ks:   keyspace{},
			exp:  key(1 << 63),
		},
		{
			desc: "splits keyspace with nil through",
			ks: keyspace{
				From: key(1 << 63),
			},
			exp: key(uint64(3) << 62),
		},
		{
			desc: "splits bounded keyspace",
			ks: keyspace{
				From: key(10),
				Through: func() *key {
					x := key(20)
					return &x
				}(),
			},
			exp: key(15),
		},
		{
			desc: "integer division",
			ks: keyspace{
				From: key(10),
				Through: func() *key {
					x := key(21)
					return &x
				}(),
			},
			exp: key(16),
		},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			require.Equal(t, tc.exp, tc.ks.Center())
		})
	}
}
