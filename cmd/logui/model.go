package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/grafana/loki/pkg/logproto"
	"github.com/prometheus/prometheus/pkg/labels"
)

type Pane int

const (
	ParamsPane Pane = iota
	LabelsPane
	LogsPane

	MinPane = ParamsPane
	MaxPane = LogsPane

	GoldenRatio = 1.618
)

func (p Pane) String() string {
	switch p {
	case LabelsPane:
		return "labels"
	case LogsPane:
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
	views  viewports
	params Params
}

// Hilarious we don't have type for this that's not bound to the ast.
// Mimic 2/3 of a label matcher :)
type Filter struct {
	Type  labels.MatchType
	Match string
}

func (f Filter) String() (res string) {
	switch f.Type {
	case labels.MatchEqual:
		res = "|="
	case labels.MatchRegexp:
		res = "|~"
	case labels.MatchNotEqual:
		res = "!="
	case labels.MatchNotRegexp:
		res = "!~"
	}

	return res + fmt.Sprintf(`"%s"`, f.Match)

}

type Params struct {
	Matchers     []labels.Matcher
	Filters      []Filter
	Since, Until time.Duration
	Direction    logproto.Direction
	Limit        int

	// internals

}

func (p Params) Query() string {
	mStrs := make([]string, 0, len(p.Matchers))
	for _, m := range p.Matchers {
		mStrs = append(mStrs, m.String())
	}

	var fStr strings.Builder
	for _, f := range p.Filters {
		fStr.WriteString(f.String())
	}
	return fmt.Sprintf("{%s}%s", strings.Join(mStrs, ","), fStr.String())
}
