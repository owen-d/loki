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

	v.separator.Height = msg.Height
	v.params.Model.Height = msg.Height
	v.labels.Model.Height = msg.Height
	v.logs.Model.Height = msg.Height

	primary := int(float64(width) / GoldenRatio)
	secondary := (width - primary) / 2
	main, secondaries := v.selected()
	main.Model.Width = primary
	for _, s := range secondaries {
		s.Model.Width = secondary
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
	merger := CrossMerge{
		v.params,
		v.separator,
		v.labels,
		v.separator,
		v.logs,
	}

	return merger.View()
}

// func (v *viewports) View() string {
// 	if !v.ready {
// 		return "\n  Initializing..."
// 	}

// 	strs := intersperse(v.headers(), []string{
// 		viewport.View(v.params.Model),
// 		viewport.View(v.labels.Model),
// 		viewport.View(v.logs.Model),
// 	})

// 	return strings.Join(strs, "\n")
// }

func (v *viewports) separatorsWidth() int { return runewidth.StringWidth(v.separator.Sep) }
