package main

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
)

type Grid struct {
	hSpacing,
	vSpacing,
	width,
	unitWidth,
	height,
	unitHeight int
	rows [][]Viewport
}

type Viewable interface {
	View() string
}

type Content string

func (c Content) View() string { return string(c) }

func NewGrid(vSpacing, hSpacing, height, width, maxColums int, views ...Viewable) Grid {

	ln := len(views)
	cols := min(maxColums, ln)
	rows := ln / cols
	if ln%cols > 0 {
		rows++
	}

	unitHeight := (height - (rows-1)*vSpacing) / rows
	unitWidth := (width - (cols-1)*hSpacing) / cols

	grid := Grid{
		vSpacing:   vSpacing,
		hSpacing:   hSpacing,
		height:     height,
		unitHeight: unitHeight,
		width:      width,
		unitWidth:  unitWidth,
		rows:       make([][]Viewport, rows),
	}

	for i, v := range views {
		row := i / cols
		vp := Viewport{
			Model: viewport.Model{
				Height: unitHeight,
				Width:  unitWidth,
			},
		}
		vp.SetContent(v.View())
		grid.rows[row] = append(grid.rows[row], vp)
	}

	return grid
}

func (g Grid) View() string {
	mergers := make([]CrossMerge, 0, len(g.rows))

	for _, row := range g.rows {
		merger := make(CrossMerge, 0, len(row))
		for _, v := range row {
			merger = append(merger, v)
		}
		mergers = append(mergers, merger.Intersperse(MergableSep{
			Height: g.unitHeight,
			Sep:    " ",
		}))
	}

	if len(g.rows) < 2 {
		return mergers[0].View()
	}

	// build vertical separator
	var sb strings.Builder
	unit := strings.Repeat(" ", g.width)
	for i := 0; i < g.vSpacing; i++ {
		sb.WriteString(unit + "\n")
	}
	vSep := sb.String()

	var result strings.Builder
	for i, m := range mergers {
		result.WriteString(m.View())
		if i < len(mergers)-1 {
			result.WriteString(vSep)
		}

	}
	return result.String()

}

func min(x, y int) int {
	res := x
	if y < x {
		res = y
	}
	return res
}

func max(x, y int) int {
	res := x
	if y > x {
		res = y
	}
	return res
}
