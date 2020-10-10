package main

import (
	"fmt"
	"strings"
)

type Intent struct {
	Primary string
	Aliases []string
	Msg     string
}

func (i Intent) View() string {
	strs := []string{i.Primary}
	if len(i.Aliases) > 0 {
		strs = append(strs, fmt.Sprintf("(or %s)", strings.Join(i.Aliases, ",")))
	}
	strs = append(strs, []string{"->", i.Msg}...)
	return strings.Join(strs, " ")
}

type HelpPane struct {
	Height, Width int
	intents       []Intent
}

func (h HelpPane) View() string {
	vs := make([]Viewable, 0, len(h.intents))
	for _, intent := range h.intents {
		vs = append(vs, intent)
	}

	topBorder := strings.Repeat("─", h.Width)
	return topBorder + NewGrid(
		0,
		4,
		h.Height,
		h.Width,
		7,
		vs...,
	).View()
}

var DefaultHelp = HelpPane{
	Height: 0,
	Width:  0,
	intents: []Intent{
		{
			Primary: "n",
			Msg:     "next pane",
		},
		{
			Primary: "p",
			Msg:     "previous pane",
		},
		{
			Primary: "q",
			Aliases: []string{"C-c"},
			Msg:     "quit",
		},
		{
			Primary: "h",
			Aliases: []string{"←"},
			Msg:     "move left",
		},
		{
			Primary: "j",
			Aliases: []string{"↓"},
			Msg:     "move down",
		},
		{
			Primary: "k",
			Aliases: []string{"↑"},
			Msg:     "move up",
		},
		{
			Primary: "l",
			Aliases: []string{"→"},
			Msg:     "move right",
		},
	},
}
