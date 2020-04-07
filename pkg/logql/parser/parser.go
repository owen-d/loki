package parser

import (
	"fmt"
	"strings"
)

// Parser is the interface any parser must impl.
type Parser interface {
	// Type is used as an annotation to facilitate human readable errors.
	Type() string
	// Parsing an input string yields the parsed value,
	// the remaining input, and any error. Thus they can be sequenced,
	// so that for the input `"ab"`, a parser of "a" will return ("a", "b", nil)
	// and the successive parser "b" will return ("b", "", nil).
	Parse(string) (interface{}, string, error)
}

// ParseError keeps track of where an error occurred in an input string
type ParseError struct {
	pos int
	err error
}

// String implements Stringer for convenience
func (e ParseError) String() string {
	return e.Error()
}

// Error implementation
func (e ParseError) Error() string {
	return fmt.Sprintf("Parse error at %d: %s", e.pos, e.err.Error())
}

// IsParseError returns true if the err is a ast parsing error.
func IsParseError(err error) bool {
	_, ok := err.(ParseError)
	return ok
}

// RunParser executes a parser against a string, returning the result or an error
func RunParser(parser Parser, input string) (value interface{}, err error) {
	defer func() {
		r := recover()
		if r != nil {
			var ok bool
			if err, ok = r.(error); ok {
				if IsParseError(err) {
					return
				}
				err = ParseError{
					pos: 0,
					err: err,
				}
			}
		}
	}()

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

/*
FunctorParser is a type useful for applying mapping functions against the result of a parser (the FMap constructor). It allows taking the successful result of a parser and applying a mapping function to it, resulting in a parser that yields a different type or value. This is commonly useful for transforming string results into their internal go structs.
*/
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

/* FMap is a constructor for a FunctorParser. It will apply a mapping function to the successful
result of another parser, returning a new parser of the mapped result's type. For instance,
given some arbitrary `IntMapper` which maps an input into an integer, we can construct an
`IncrementMapper` which will parse an integer and increment it's parsed value:
var IncrementMapper = FMap(func(x interface{})interface{}{
	return x.(int)+1
}, "IncrementedInt", IntParser)
It takes the form `fmap :: (a -> b) -> f a -> f b`
*/
func FMap(fn func(interface{}) interface{}, p Parser, mappedType string) FunctorParser {
	return FunctorParser{
		mappedType: mappedType,
		embedded:   p,
		fn:         fn,
	}
}

// Const returns a function which will always return a specified value, disregarding its input.
// It takes the form `const :: a -> b -> a`
func Const(x interface{}) func(interface{}) interface{} {
	return func(_ interface{}) interface{} {
		return x
	}
}

// ConstParser is a parser which will always return a specified value
// without consuming additional input.
type ConstParser struct {
	t   string
	val interface{}
}

func (p ConstParser) Type() string { return p.t }
func (p ConstParser) Parse(s string) (interface{}, string, error) {
	return p.val, s, nil
}

// MonadParser is a type useful for creating parsers from the successful results of other parsers.
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

/*
Bind constructs a MonadParser with a function which maps the successful result
 of one parser into another parser. For instance, given a Parser which returns an alphanumeric
character, `AlphaNumParser`, you may want to return another parser which only parses
that same character:
var TwiceParser(IntParser, `Twice`, func(x interface{}) Parser {
	return StringParser{x.(string)}
})
it takes the form: `bind :: m a -> (a -> m b) -> m b`
*/
func Bind(p Parser, mappedType string, fn func(interface{}) Parser) MonadParser {
	return MonadParser{
		embedded:   p,
		mappedType: mappedType,
		fn:         fn,
	}
}

// AlternativeParser is a type which attempts one parser, then attempts an alternative
// parser if the first fails.
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

// Option constructs an alternative parser from two parsers.
// (<|>) :: f a -> f a -> f a
func Option(p1, p2 Parser) AlternativeParser {
	return AlternativeParser{
		p1: p1,
		p2: p2,
	}
}

// ErrP constructs an ErrParser
func ErrP(err error) ErrParser { return ErrParser{err} }

// ErrParser is a parser which always returns an error
type ErrParser struct{ error }

func (p ErrParser) Type() string { return "error" }
func (p ErrParser) Parse(s string) (interface{}, string, error) {
	return nil, s, p.error
}

// StringP constructs a string parser
func StringP(s string) StringParser { return StringParser{s} }

// StringParser is a parser which matches a specific string
type StringParser struct{ match string }

func (p StringParser) Type() string { return p.match }

func (p StringParser) Parse(s string) (interface{}, string, error) {
	ln := len(p.match)
	if len(s) >= ln && s[:ln] == p.match {
		return p.match, s[ln:], nil
	}

	return nil, s, fmt.Errorf("Expecting (%s)", p.match)
}

// ManyP constructs a ManyParser from a parser
func ManyP(p Parser) ManyParser { return ManyParser{p} }

// ManyParser is a parser which matches zero or more occurences of a parser.
type ManyParser struct {
	p Parser
}

func (p ManyParser) Type() string {
	return fmt.Sprintf("[%T]", p.p.Type())
}

func (p ManyParser) Parse(s string) (interface{}, string, error) {
	rem := s
	var all []interface{}
	for len(rem) > 0 {
		res, nextRem, err := p.p.Parse(rem)
		if err != nil {
			return all, rem, nil
		}
		all = append(all, res)
		rem = nextRem

	}

	return all, rem, nil

}

// SomeParser is a parser which matches one or more occurences of a parser.
type SomeParser struct {
	p Parser
}

func (p SomeParser) Type() string {
	return fmt.Sprintf("[%T]", p.p.Type())
}

func (p SomeParser) Parse(s string) (interface{}, string, error) {
	rem := s
	var all []interface{}
	for len(rem) > 0 {
		res, nextRem, err := p.p.Parse(rem)
		if err != nil {
			// require at least one successful parse
			if len(all) == 0 {
				return nil, rem, err
			}
			return all, rem, nil
		}
		all = append(all, res)
		rem = nextRem

	}

	return all, rem, nil

}
