package main

import (
	"strings"
)

/*
Idea here is to merge views which combine horizontally (on the same line)

------   ------   ------------
|    | + |    | = |          |
|    | + |    | = |          |
------   ------   ------------

*/

type ViewLines interface {
	Lines() []string
}

type HasWidth interface {
	Width() int
}

type CrossMerge []interface {
	ViewLines
	HasWidth
}

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
