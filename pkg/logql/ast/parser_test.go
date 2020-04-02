package ast

import (
	"errors"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNums(t *testing.T) {
	nums := OneOf("0", "1", "2", "3", "4", "5", "6", "7", "8", "9")

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
	nums := OneOf("0", "1", "2", "3", "4", "5", "6", "7", "8", "9")
	p := ManyParser{nums}

	out, rem, err := p.Parse("567")
	require.Nil(t, err)
	require.Equal(t, "", rem)
	require.Equal(t, []interface{}{"5", "6", "7"}, out)
}

func TestJoin(t *testing.T) {
	nums := OneOf("0", "1", "2", "3", "4", "5", "6", "7", "8", "9")
	require.Equal(t, "string", nums.Type())
	many := ManyParser{nums}
	require.Equal(t, "[string]", many.Type())

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
