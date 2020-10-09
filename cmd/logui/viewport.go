package main

import (
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
)

type Viewport struct {
	viewport.Model
}

func (v Viewport) Width() int {
	return v.Model.Width
}

func (v Viewport) Lines() []string {
	return strings.Split(viewport.View(v.Model), "\n")
}
