package main

import (
	"strings"

	"github.com/muesli/reflow/wordwrap"
	"github.com/muesli/termenv"
)

var profile = termenv.ColorProfile()

type Draw interface {
	Draw(n int) string // draw n width
	Advance()          // newline
}

type Content struct {
	s     string
	color termenv.Color
	wrap  int // should words be wrapped to newlines based on length?

	// stateful drawing internals
	i     int
	lines []string
}

func NewContent(s string) Content {
	return Content{
		s: s,
	}
}

func (c Content) Color(color termenv.Color) Content {
	c.color = color
	return c.Reset()
}

func (c Content) Wrap(n int) Content {
	c.wrap = n
	return c.Reset()
}

func (c Content) Reset() Content {
	c.lines = nil
	return c
}

func (c *Content) Draw(n int) string {
	// initialize internals if not done already
	if len(c.lines) == 0 && len(c.s) > 0 {
		if c.wrap > 0 {
			c.lines = strings.Split(wordwrap.String(c.s, c.wrap), "\n")
		} else {
			c.lines = strings.Split(c.s, "\n")
		}
	}

	var str string
	if c.i < len(c.lines) {
		str = ExactWidth(c.lines[c.i], n)
		if c.color != nil {
			str = termenv.String(str).Foreground(profile.Convert(c.color)).String()
		}
	} else {
		str = RPad(str, n)
	}

	return str
}

func (c *Content) Seek(i int) {
	if i < 0 {
		i = 0
	}
	c.i = i
}

func (c *Content) Retreat(n int) { c.Seek(c.i - n) }
func (c *Content) Advance()      { c.i++ }
