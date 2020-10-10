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
	totals               tea.WindowSizeMsg
	ready                bool
	focusPane            Pane
	separator            MergableSep
	params, labels, logs Viewport
	help                 HelpPane
}

func (v *viewports) focused() *viewport.Model {
	switch v.focusPane {
	case Labels:
		return &v.labels.Model
	case Logs:
		return &v.logs.Model
	default:
		return &v.params.Model
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
			v.Size(v.totals)
		case "p":
			v.focusPane = v.focusPane.Prev()
			v.Size(v.totals)
		}
	}

	focused := v.focused()
	updated, cmd := viewport.Update(msg, *focused)
	*focused = updated
	cmds = append(cmds, cmd)

	return tea.Batch(cmds...)
}

func (v *viewports) selected() (main *Viewport, secondaries []*Viewport) {
	switch v.focusPane {
	case Labels:
		return &v.labels, []*Viewport{&v.params, &v.logs}
	case Logs:
		return &v.logs, []*Viewport{&v.params, &v.labels}
	// Params is the default
	default:
		return &v.params, []*Viewport{&v.labels, &v.logs}
	}
}

// Size sets pane sizes (primary & secondaries) based on the golden ratio.
func (v *viewports) Size(msg tea.WindowSizeMsg) {
	v.totals = msg
	width := msg.Width - v.separator.Width()*2
	if !v.ready {
		v.ready = true
	}

	v.help.Height = 4
	v.help.Width = v.totals.Width

	withoutHeaders := msg.Height - 3
	withoutHelp := withoutHeaders - v.help.Height

	height := withoutHelp

	v.separator.Height = height
	v.params.Model.Height = height
	v.labels.Model.Height = height
	v.logs.Model.Height = height

	primary := int(float64(width) / GoldenRatio)
	secondary := (width - primary) / 2
	main, secondaries := v.selected()
	main.Model.Width = primary
	for _, s := range secondaries {
		s.Model.Width = secondary
	}
}

func (v *viewports) header() string {
	pane := v.focusPane
	width := v.totals.Width
	var start int

	switch pane {
	case Labels:
		start = v.params.Width() + v.separator.Width()
	case Logs:
		start = v.params.Width()*2 + v.separator.Width()*2 // all non-primary panes have the same size
	}

	headerTopFrame := "╭─────────────╮"
	headerBotFrame := "╰─────────────╯"
	headerTop := ExactWidth(LPad(headerTopFrame, start+runewidth.StringWidth(headerTopFrame)), width)
	headerBot := ExactWidth(LPad(headerBotFrame, start+runewidth.StringWidth(headerBotFrame)), width)

	lConnector := "│"
	if start > 0 {
		lConnector = "┤"
	}
	headerMid := lConnector + CenterTo(pane.String(), runewidth.StringWidth(headerTopFrame)-2) + "├"
	headerMid = LPadWith(headerMid, '─', start+runewidth.StringWidth(headerMid))
	headerMid = RPadWith(headerMid, '─', width)

	return strings.Join([]string{headerTop, headerMid, headerBot}, "\n")
}

func (v *viewports) View() string {
	if !v.ready {
		return "\n  Initializing..."
	}

	merger := CrossMerge{
		v.params,
		v.separator,
		v.labels,
		v.separator,
		v.logs,
	}

	return strings.Join(
		[]string{
			v.header(),
			merger.View(),
			v.help.View(),
		},
		"\n",
	)
}
