package planning

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBind(t *testing.T) {
	a := Pure(2)

	b := Bind[int, int](
		a,
		func(i int) (int, error) {
			return i + 1, nil
		},
	)

	v, err := b.Run()
	require.NoError(t, err)
	require.Equal(t, 3, v)
}

func TestBindShortCircuitsError(t *testing.T) {
	a := Pure(0)
	var calls int
	f := func(i int) (int, error) {
		calls++
		return i + 1, nil
	}
	errF := func(i int) (int, error) {
		return 0, errors.New("buzz")
	}

	_, err := Bind(
		Bind(
			Bind(a, f),
			errF,
		),
		f,
	).Run()

	require.Error(t, err)
	require.Equal(t, calls, 1)
}
