package main

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mattn/go-runewidth"
)

type Pane int

const (
	Params Pane = iota
	Labels
	Logs

	MinPane = Params
	MaxPane = Logs

	GoldenRatio = 1.618
)

func (p Pane) String() string {
	switch p {
	case Labels:
		return "labels"
	case Logs:
		return "logs"
	default:
		return "params"
	}
}

func (p Pane) Next() Pane {
	n := p + 1
	if n > MaxPane {
		n = MinPane
	}
	return n
}

func (p Pane) Prev() Pane {
	n := p - 1
	if n < MinPane {
		n = MaxPane
	}
	return n
}

type Model struct {
	views viewports
}

type viewports struct {
	ready                bool
	focusPane            Pane
	params, labels, logs viewport.Model
}

func (v *viewports) focused() *viewport.Model {
	switch v.focusPane {
	case Labels:
		return &v.labels
	case Logs:
		return &v.logs
	default:
		return &v.params
	}
}

func (v *viewports) Update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.Size(msg)
	case tea.KeyMsg:
		switch msg.String() {
		case "n":
			v.focusPane = v.focusPane.Next()
		case "p":
			v.focusPane = v.focusPane.Prev()
		}
	}

	focused := v.focused()
	updated, cmd := viewport.Update(msg, *focused)
	*focused = updated
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

func (v *viewports) selected() (main *viewport.Model, secondaries []*viewport.Model) {
	switch v.focusPane {
	case Labels:
		return &v.labels, []*viewport.Model{&v.params, &v.logs}
	case Logs:
		return &v.logs, []*viewport.Model{&v.params, &v.labels}
	// Params is the default
	default:
		return &v.params, []*viewport.Model{&v.labels, &v.logs}
	}
}

// Size sets pane sizes (primary & secondaries) based on the golden ratio.
func (v *viewports) Size(msg tea.WindowSizeMsg) {
	height := msg.Height - v.separatorsHeight() // reserver separator space
	if !v.ready {
		v.params.Width = msg.Width
		v.labels.Width = msg.Width
		v.logs.Width = msg.Width
		v.ready = true
	}

	primary := int(float64(height) / GoldenRatio)
	secondary := (height - primary) / 2
	main, secondaries := v.selected()
	main.Height = primary
	for _, s := range secondaries {
		s.Height = secondary
	}
}

func paneHeader(pane Pane, focus bool, width int) string {
	headerTop := "╭─────────────╮"
	headerBot := "╰─────────────╯"

	if !focus {
		return strings.Join([]string{headerTop, headerBot}, "\n")
	}

	headerMid := "│" + padTo(pane.String(), runewidth.StringWidth(headerTop)-2) + "├"
	headerMid += strings.Repeat("-", width-runewidth.StringWidth(headerMid))
	return strings.Join([]string{headerTop, headerMid, headerBot}, "\n")

}

func (v *viewports) headers() []string {
	return []string{
		paneHeader(Params, v.focusPane == Params, v.params.Width),
		paneHeader(Labels, v.focusPane == Labels, v.labels.Width),
		paneHeader(Logs, v.focusPane == Logs, v.logs.Width),
	}
}

func intersperse(xs, ys []string) (res []string) {
	for i := 0; i < len(xs) && i < len(ys); i++ {
		res = append(res, xs[i])
		res = append(res, ys[i])
	}
	return res
}

func (v *viewports) View() string {
	if !v.ready {
		return "\n  Initializing..."
	}

	strs := intersperse(v.headers(), []string{
		viewport.View(v.params),
		viewport.View(v.labels),
		viewport.View(v.logs),
	})

	return strings.Join(strs, "\n")
}

func (*viewports) separatorsHeight() int { return 3 + 2 + 2 } // height of combined primary header & the two others

func padTo(msg string, ln int) string {
	rem := ln - len(msg)
	if rem < 1 {
		return msg
	}

	div, mod := rem/2, rem%2
	lpad := strings.Repeat(" ", div)
	rpad := lpad

	// on odd, prefer rpad
	if mod != 0 {
		rpad += " "
	}

	return lpad + msg + rpad
}
