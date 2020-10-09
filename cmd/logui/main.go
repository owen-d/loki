package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {

	// Set PAGER_LOG to a path to log to a file. For example:
	//
	//     export PAGER_LOG=debug.log
	//
	// This becomes handy when debugging stuff since you can't debug to stdout
	// because the UI is occupying it!
	path := os.Getenv("PAGER_LOG")
	if path != "" {
		f, err := tea.LogToFile(path, "pager")
		if err != nil {
			fmt.Printf("Could not open file %s: %v", path, err)
			os.Exit(1)
		}
		defer f.Close()
	}

	p := tea.NewProgram(initialize(), update, view)

	// Use the full size of the terminal in its "alternate screen buffer"
	p.EnterAltScreen()
	defer p.ExitAltScreen()

	// We also turn on mouse support so we can track the mouse wheel
	p.EnableMouseCellMotion()
	defer p.DisableMouseCellMotion()

	if err := p.Start(); err != nil {
		fmt.Println("could not run program:", err)
		os.Exit(1)
	}
}

func initialize() func() (tea.Model, tea.Cmd) {
	return func() (tea.Model, tea.Cmd) {
		var m Model

		var garbage string
		for i := 0; i < 200; i++ {
			garbage += fmt.Sprintf("%d - lorem ipsum\n", i)
		}
		m.views.params.SetContent(garbage)
		m.views.labels.SetContent(garbage)
		m.views.logs.SetContent(garbage)
		return m, nil
	}
}

func update(msg tea.Msg, mdl tea.Model) (tea.Model, tea.Cmd) {
	m, _ := mdl.(Model)

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Ctrl+c exits
		if msg.Type == tea.KeyCtrlC || msg.String() == "q" {
			return m, tea.Quit
		}
	}

	if cmd := m.views.Update(msg); cmd != nil {
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func view(mdl tea.Model) string {
	m, _ := mdl.(Model)
	return m.views.View()
}

// type model struct {
// 	client   client.Client
// 	content  string
// 	ready    bool
// 	viewport viewport.Model
// }

// func initialize() func() (tea.Model, tea.Cmd) {
// 	return func() (tea.Model, tea.Cmd) {
// 		m := model{
// 			client: &client.DefaultClient{
// 				Address: "http://localhost:3100",
// 				OrgID:   "fake",
// 			},
// 		}

// 		return m, checkServer(m.client)
// 	}
// }

// type errMsg error

// func checkServer(c client.Client) func() tea.Msg {
// 	return func() tea.Msg {

// 		resp, err := c.QueryRange(
// 			`{filename!~".*(install|wifi).log"}`,
// 			1000,
// 			time.Now().Add(-time.Hour),
// 			time.Now(),
// 			logproto.BACKWARD,
// 			0,
// 			0,
// 			true,
// 		)

// 		if err != nil {
// 			return errMsg(err)
// 		}
// 		return resp
// 	}
// }

// func update(msg tea.Msg, mdl tea.Model) (tea.Model, tea.Cmd) {
// 	m, _ := mdl.(model)

// 	var (
// 		cmd  tea.Cmd
// 		cmds []tea.Cmd
// 	)

// 	switch msg := msg.(type) {
// 	case tea.KeyMsg:
// 		// Ctrl+c exits
// 		if msg.Type == tea.KeyCtrlC || msg.String() == "q" {
// 			return m, tea.Quit
// 		}

// 	case tea.WindowSizeMsg:
// 		verticalMargins := headerHeight + footerHeight

// 		if !m.ready {
// 			// Since this program is using the full size of the viewport we need
// 			// to wait until we've received the window dimensions before we
// 			// can initialize the viewport. The initial dimensions come in
// 			// quickly, though asynchronously, which is why we wait for them
// 			// here.
// 			m.viewport = viewport.Model{Width: msg.Width, Height: msg.Height - verticalMargins}
// 			m.viewport.YPosition = headerHeight
// 			m.viewport.HighPerformanceRendering = useHighPerformanceRenderer
// 			m.viewport.SetContent(m.content)
// 			m.ready = true
// 		} else {
// 			m.viewport.Width = msg.Width
// 			m.viewport.Height = msg.Height - verticalMargins
// 		}

// 		if useHighPerformanceRenderer {
// 			// Render (or re-render) the whole viewport. Necessary both to
// 			// initialize the viewport and when the window is resized.
// 			//
// 			// This is needed for high-performance rendering only.
// 			cmds = append(cmds, viewport.Sync(m.viewport))
// 		}

// 	case *loghttp.QueryResponse:
// 		wrapped := wordwrap.NewWriter(m.viewport.Width * 9 / 10)

// 		for _, stream := range msg.Data.Result.(loghttp.Streams) {
// 			fmt.Fprintf(wrapped, "\n%s", stream.Labels)

// 			for _, entry := range stream.Entries {
// 				fmt.Fprintf(
// 					wrapped,
// 					"\n%s",
// 					indent.String(fmt.Sprintf("%v: %s", entry.Timestamp, entry.Line), 4),
// 				)
// 			}
// 		}

// 		m.content = wrapped.String()
// 		m.viewport.SetContent(m.content)

// 	case errMsg:
// 		m.content = msg.Error()
// 		m.viewport.SetContent(m.content)

// 	}

// 	// Because we're using the viewport's default update function (with pager-
// 	// style navigation) it's important that the viewport's update function:
// 	//
// 	// * Recieves messages from the Bubble Tea runtime
// 	// * Returns commands to the Bubble Tea runtime
// 	//
// 	m.viewport, cmd = viewport.Update(msg, m.viewport)
// 	if useHighPerformanceRenderer {
// 		cmds = append(cmds, cmd)
// 	}

// 	return m, tea.Batch(cmds...)
// }

// func view(mdl tea.Model) string {
// 	m, _ := mdl.(model)

// 	if !m.ready {
// 		return "\n  Initalizing..."
// 	}

// 	headerTop := "╭───────────╮"
// 	headerMid := "│   LogUI   ├"
// 	headerBot := "╰───────────╯"
// 	headerMid += strings.Repeat("─", m.viewport.Width-runewidth.StringWidth(headerMid))
// 	header := fmt.Sprintf("%s\n%s\n%s", headerTop, headerMid, headerBot)

// 	footerTop := "╭──────╮"
// 	footerMid := fmt.Sprintf("┤ %3.f%% │", m.viewport.ScrollPercent()*100)
// 	footerBot := "╰──────╯"
// 	gapSize := m.viewport.Width - runewidth.StringWidth(footerMid)
// 	footerTop = strings.Repeat(" ", gapSize) + footerTop
// 	footerMid = strings.Repeat("─", gapSize) + footerMid
// 	footerBot = strings.Repeat(" ", gapSize) + footerBot
// 	footer := fmt.Sprintf("%s\n%s\n%s", footerTop, footerMid, footerBot)

// 	return fmt.Sprintf("%s\n%s\n%s", header, viewport.View(m.viewport), footer)
// }
