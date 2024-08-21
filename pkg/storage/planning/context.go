package planning

type ExecCtx[O any] struct {
	fn func() (O, error)
}

func NewExecCtx[O any](fn func() (O, error)) ExecCtx[O] {
	return ExecCtx[O]{
		fn: fn,
	}
}

func (e ExecCtx[O]) Run() (O, error) {
	return e.fn()
}

func Pure[O any](value O) ExecCtx[O] {
	return NewExecCtx(func() (O, error) {
		return value, nil
	})
}

// Map applies a pure function to the result of a computation within a context.
// It allows transformation of the result.
func Map[O, O1 any](
	c ExecCtx[O],
	fn func(O) O1,
) ExecCtx[O1] {
	return ExecCtx[O1]{
		fn: func() (O1, error) {
			result, err := c.Run()
			if err != nil {
				var zero O1
				return zero, err
			}
			return fn(result), nil
		},
	}
}

// Apply (Applicative Functor) applies a function wrapped in an ExecCtx to a value wrapped in another ExecCtx
func Apply[
	A, O1 any,
	O ~func(A) O1,
](
	left ExecCtx[O],
	right ExecCtx[A],
) ExecCtx[O1] {
	return ExecCtx[O1]{
		fn: func() (O1, error) {
			f, err := left.Run()
			if err != nil {
				var zero O1
				return zero, err
			}

			a, err := right.Run()
			if err != nil {
				var zero O1
				return zero, err
			}

			return f(a), nil
		},
	}
}

// Monadic Bind over a generic Context.
// in short, allows us to sequence actions which may fail without explicitly handling failure.
func Bind[O, O1 any](
	c ExecCtx[O],
	fn func(O) (O1, error),
) ExecCtx[O1] {
	return ExecCtx[O1]{
		fn: func() (zero O1, err error) {
			result, err := c.Run()
			if err != nil {
				return zero, err
			}
			return fn(result)
		},
	}
}

// Bind2 chains two functions in sequence, handling errors at each step
func Bind2[O, O1, O2 any](
	c ExecCtx[O],
	fn1 func(O) (O1, error),
	fn2 func(O1) (O2, error),
) ExecCtx[O2] {
	return ExecCtx[O2]{
		fn: func() (zero O2, err error) {
			result, err := c.Run()
			if err != nil {
				return zero, err
			}
			intermediate, err := fn1(result)
			if err != nil {
				return zero, err
			}
			return fn2(intermediate)
		},
	}
}

// Bind3 chains three functions in sequence, handling errors at each step
func Bind3[O, O1, O2, O3 any](
	c ExecCtx[O],
	fn1 func(O) (O1, error),
	fn2 func(O1) (O2, error),
	fn3 func(O2) (O3, error),
) ExecCtx[O3] {
	return ExecCtx[O3]{
		fn: func() (zero O3, err error) {
			result, err := c.Run()
			if err != nil {
				return zero, err
			}
			intermediate1, err := fn1(result)
			if err != nil {
				return zero, err
			}
			intermediate2, err := fn2(intermediate1)
			if err != nil {
				return zero, err
			}
			return fn3(intermediate2)
		},
	}
}
