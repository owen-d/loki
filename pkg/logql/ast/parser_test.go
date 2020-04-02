package ast

import (
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNums(t *testing.T) {
	nums := OneOfStrings("0", "1", "2", "3", "4", "5", "6", "7", "8", "9")

	out, rem, err := nums.Parse("abc")
	require.Nil(t, out)
	require.Equal(t, "abc", rem)
	require.NotNil(t, err)

	out, rem, err = nums.Parse("5")
	require.Nil(t, err)
	require.Equal(t, "", rem)
	require.Equal(t, out, "5")
}

func TestMany(t *testing.T) {
	nums := OneOfStrings("0", "1", "2", "3", "4", "5", "6", "7", "8", "9")
	p := ManyParser{nums}

	out, rem, err := p.Parse("567")
	require.Nil(t, err)
	require.Equal(t, "", rem)
	require.Equal(t, []interface{}{"5", "6", "7"}, out)
}

func TestJoin(t *testing.T) {
	nums := OneOfStrings("0", "1", "2", "3", "4", "5", "6", "7", "8", "9")
	many := ManyParser{nums}

	joined := FMap(func(in interface{}) interface{} {
		var str string
		xs := in.([]interface{})
		if len(xs) == 0 {
			return 0
		}

		for _, x := range xs {
			str = str + x.(string)
		}

		n, _ := strconv.Atoi(str)
		return n

	}, many, "int")
	require.Equal(t, "int", joined.Type())

	out, rem, err := joined.Parse("567")
	require.Nil(t, err)
	require.Equal(t, "", rem)
	require.Equal(t, 567, out)

	p := Satisfy(
		func(in interface{}) bool {
			v := in.(int)
			return v%2 == 1
		},
		errors.New("expected odd number"),
		joined,
	)

	out, err = RunParser(p, "3")
	require.Nil(t, err)
	require.Equal(t, 3, out)

	_, err = RunParser(p, "34")
	require.Equal(t, err.Error(), "Parse error at 2: expected odd number")

}

func TestParse(t *testing.T) {
	for _, tc := range []struct {
		parser Parser
		in     string
		out    interface{}
		rest   string
		err    string
	}{
		{
			parser: ManyParser{StringParser{"a"}},
			in:     "aa",
			out:    []interface{}{"a", "a"},
			rest:   "",
			err:    "",
		},
		{
			parser: ManyParser{StringParser{"a"}},
			in:     "aab",
			out:    []interface{}{"a", "a"},
			rest:   "b",
		},
		{
			parser: ManyParser{StringParser{"a"}},
			in:     "bba",
			out:    []interface{}(nil),
			rest:   "bba",
		},
		{
			parser: SomeParser{StringParser{"a"}},
			in:     "aa",
			out:    []interface{}{"a", "a"},
			rest:   "",
			err:    "",
		},
		{
			parser: SomeParser{StringParser{"a"}},
			in:     "aab",
			out:    []interface{}{"a", "a"},
			rest:   "b",
		},
		{
			parser: SomeParser{StringParser{"a"}},
			in:     "bba",
			rest:   "bba",
			err:    "Expecting (a)",
		},
	} {

		t.Run(tc.in, func(t *testing.T) {
			out, rest, err := tc.parser.Parse(tc.in)
			if tc.err != "" {
				require.NotNil(t, err)
				require.Equal(t, tc.err, err.Error())
			} else {
				require.Nil(t, err)
				require.Equal(t, tc.out, out)
				require.Equal(t, tc.rest, rest)
			}
		})
	}
}

func TestParser(t *testing.T) {
	for _, tc := range []struct {
		parser Parser
		in     string
		err    string
		out    interface{}
	}{
		{
			parser: StringParser{"a"},
			in:     "a",
			out:    "a",
		},
		{
			parser: StringParser{"a"},
			in:     "b",
			err:    "Parse error at 0: Expecting (a)",
		},
	} {
		t.Run(tc.in, func(t *testing.T) {
			out, err := RunParser(tc.parser, tc.in)
			if tc.err != "" {
				require.NotNil(t, err)
				require.Equal(t, tc.err, err.Error())
			} else {
				require.Nil(t, err)
				require.Equal(t, tc.out, out)
			}
		})
	}
}
