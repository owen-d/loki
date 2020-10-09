package main

import (
	"strings"

	"github.com/mattn/go-runewidth"
)

/*
Idea here is to merge views which combine horizontally (on the same line)

------   ------   ------------
|    | + |    | = |          |
|    | + |    | = |          |
------   ------   ------------

*/

type CrossMergable interface {
	Lines() []string
	Width() int
}

type MergableSep struct {
	Height int
	Sep    string
}

func (s MergableSep) Lines() []string {
	lines := make([]string, 0, s.Height)
	for i := 0; i < s.Height; i++ {
		lines = append(lines, s.Sep)
	}
	return lines
}

func (s MergableSep) Width() int {
	return runewidth.StringWidth(s.Sep)
}

type CrossMerge []CrossMergable

func (c CrossMerge) Width() (res int) {
	for _, x := range c {
		res += x.Width()
	}
	return

}

func (c CrossMerge) Lines() (lines []string) {
	var maxLines int
	subGrps := make([]struct {
		lines []string
		width int
	}, len(c))

	for i, x := range c {
		subGrps[i].lines = x.Lines()
		subGrps[i].width = x.Width()

		if lineCt := len(subGrps[i].lines); lineCt > maxLines {
			maxLines = lineCt
		}
	}

	for i := 0; i < maxLines; i++ {
		var line string
		for _, x := range subGrps {
			var msg string
			if i < len(x.lines) {
				msg = x.lines[i]
			}
			line += ExactWidth(msg, x.width)

		}

		lines = append(lines, line)
	}

	return lines

}

func (c CrossMerge) View() string {
	return strings.Join(c.Lines(), "\n")
}
