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
	params, labels, logs Viewport
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
		v.totals = msg
		v.Size()
	case tea.KeyMsg:
		switch msg.String() {
		case "n":
			v.focusPane = v.focusPane.Next()
			v.Size()
		case "p":
			v.focusPane = v.focusPane.Prev()
			v.Size()
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
func (v *viewports) Size() {
	msg := v.totals
	height := msg.Height - v.separatorsHeight() // reserver separator space
	if !v.ready {
		v.params.Model.Width = msg.Width
		v.labels.Model.Width = msg.Width
		v.logs.Model.Width = msg.Width
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
	headerTop := "╭─────────────"
	headerBot := "╰─────────────"
	if focus {
		headerTop += "┬"
		headerBot += "┴"
	} else {
		headerTop += "─"
		headerBot += "─"
	}

	headerMid := "│" + CenterTo(pane.String(), runewidth.StringWidth(headerTop)-2) + "│"
	headerTop = headerTop + strings.Repeat("─", width-runewidth.StringWidth(headerTop)-1) + "╮"
	headerBot = headerBot + strings.Repeat("─", width-runewidth.StringWidth(headerBot)-1) + "╯"
	headerMid = headerMid + strings.Repeat(" ", width-runewidth.StringWidth(headerMid)-1) + "│"

	if !focus {
		return strings.Join([]string{headerTop, headerBot}, "\n")
	}

	return strings.Join([]string{headerTop, headerMid, headerBot}, "\n")

}

func (v *viewports) headers() []string {
	return []string{
		paneHeader(Params, v.focusPane == Params, v.params.Width()),
		paneHeader(Labels, v.focusPane == Labels, v.labels.Width()),
		paneHeader(Logs, v.focusPane == Logs, v.logs.Width()),
	}
}

func (v *viewports) View() string {
	if !v.ready {
		return "\n  Initializing..."
	}

	strs := intersperse(v.headers(), []string{
		viewport.View(v.params.Model),
		viewport.View(v.labels.Model),
		viewport.View(v.logs.Model),
	})

	return strings.Join(strs, "\n")
}

// height of combined primary header & the two others
// TODO: derive this
func (*viewports) separatorsHeight() int { return 3 + 2 + 2 }
