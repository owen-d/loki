package ast

import (
	"errors"
	"fmt"
	"strings"
)

type Parser interface {
	Type() string
	Parse(string) (interface{}, string, error)
}

type ParseError struct {
	pos int
	err error
}

func (e ParseError) String() string {
	return e.Error()
}

func (e ParseError) Error() string {
	return fmt.Sprintf("Parse error at %d: %s", e.pos, e.err.Error())
}

func RunParser(parser Parser, input string) (interface{}, error) {
	value, rem, err := parser.Parse(input)
	if err != nil {
		return nil, ParseError{
			pos: len(input) - len(rem),
			err: err,
		}
	}

	if rem != "" {
		return nil, ParseError{
			pos: len(input) - len(rem),
			err: fmt.Errorf("Unterminated input: %s", rem),
		}
	}

	return value, nil
}

type FunctorParser struct {
	mappedType string
	embedded   Parser
	fn         func(interface{}) interface{}
}

func (p FunctorParser) Type() string { return p.mappedType }
func (p FunctorParser) Parse(s string) (interface{}, string, error) {
	out, rem, err := p.embedded.Parse(s)
	if err != nil {
		return out, rem, err
	}

	return p.fn(out), rem, err
}

// (a -> b) -> f a -> f b
func FMap(fn func(interface{}) interface{}, p Parser, mappedType string) Parser {
	return FunctorParser{
		mappedType: mappedType,
		embedded:   p,
		fn:         fn,
	}
}

// const :: a -> b -> a
func Const(x interface{}) func(interface{}) interface{} {
	return func(_ interface{}) interface{} {
		return x
	}
}

type ConstParser struct {
	t   string
	val interface{}
}

func (p ConstParser) Type() string { return p.t }
func (p ConstParser) Parse(s string) (interface{}, string, error) {
	return p.val, s, nil
}

// a -> m a
func Unit(x interface{}) ConstParser {
	return ConstParser{
		t:   fmt.Sprintf("%T", x),
		val: x,
	}
}

type MonadParser struct {
	embedded   Parser
	mappedType string
	fn         func(interface{}) Parser
}

func (p MonadParser) Type() string { return p.mappedType }
func (p MonadParser) Parse(s string) (interface{}, string, error) {
	v, rem, err := p.embedded.Parse(s)
	if err != nil {
		return v, rem, err
	}

	next := p.fn(v)
	return next.Parse(rem)
}

// m a -> (a -> m b) -> m b
func Bind(p Parser, mappedType string, fn func(interface{}) Parser) MonadParser {
	return MonadParser{
		embedded:   p,
		mappedType: mappedType,
		fn:         fn,
	}
}

type AlternativeParser struct {
	p1 Parser
	p2 Parser
}

// subtypes can recursively check legs of alternative parsers, mainly
// for nicer error messages
func (p AlternativeParser) subTypes() []string {
	var lhs, rhs []string

	switch left := p.p1.(type) {
	case AlternativeParser:
		lhs = append(lhs, left.subTypes()...)
	default:
		lhs = append(lhs, left.Type())
	}

	switch right := p.p2.(type) {
	case AlternativeParser:
		rhs = append(rhs, right.subTypes()...)
	default:
		rhs = append(rhs, right.Type())
	}

	return append(lhs, rhs...)
}

func (p AlternativeParser) Type() string {
	return fmt.Sprintf("Alternative<%s>", strings.Join(p.subTypes(), ", "))
}

func (p AlternativeParser) Parse(s string) (interface{}, string, error) {
	if v, rem, err := p.p1.Parse(s); err == nil {
		return v, rem, err
	}

	if v, rem, err := p.p2.Parse(s); err == nil {
		return v, rem, err
	}

	return nil, s, fmt.Errorf("Expecting %s", p.Type())
}

// (<|>) :: f a -> f a -> f a
func Option(p1, p2 Parser) AlternativeParser {
	return AlternativeParser{
		p1: p1,
		p2: p2,
	}
}

type ErrParser struct{ error }

func (p ErrParser) Type() string { return "error" }
func (p ErrParser) Parse(s string) (interface{}, string, error) {
	return nil, s, p.error
}

func Satisfy(predicate func(interface{}) bool, failureErr error, p Parser) MonadParser {
	return Bind(
		p,
		p.Type(),
		func(v interface{}) Parser {
			if predicate(v) {
				return Unit(v)
			}
			return ErrParser{failureErr}
		},
	)
}

type StringParser struct{ match string }

func (StringParser) Type() string { return "string" }

func (p StringParser) Parse(s string) (interface{}, string, error) {
	ln := len(p.match)
	if len(s) >= ln && s[:ln] == p.match {
		return p.match, s[ln:], nil
	}

	return nil, s, fmt.Errorf("Expecting (%s)", p.match)
}

func OneOf(xs ...Parser) Parser {
	if len(xs) == 0 {
		return ErrParser{errors.New("No available options")}
	}

	res := xs[0]
	for _, x := range xs[1:] {
		res = Option(res, x)
	}
	return res

}

func OneOfStrings(xs ...string) Parser {
	parsers := make([]Parser, 0, len(xs))
	for _, x := range xs {
		parsers = append(parsers, StringParser{x})
	}
	return OneOf(parsers...)

}

// Zero or more
type ManyParser struct {
	p Parser
}

func (p ManyParser) Type() string {
	return fmt.Sprintf("[%T]", p.p.Type())
}

func (p ManyParser) Parse(s string) (res interface{}, rem string, err error) {
	rem = s
	var all []interface{}
	for len(rem) > 0 {
		res, rem, err = p.p.Parse(rem)
		if err != nil {
			return nil, rem, err
		}

		all = append(all, res)

	}

	return all, rem, nil

}
