package main

import (
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
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
	if !v.ready {
		v.params.Width = msg.Width
		v.labels.Width = msg.Width
		v.logs.Width = msg.Width
		v.ready = true
	}

	primary := int(float64(msg.Height) / GoldenRatio)
	secondary := (msg.Height - primary) / 3
	main, secondaries := v.selected()
	main.Height = primary
	for _, s := range secondaries {
		s.Height = secondary
	}
}
