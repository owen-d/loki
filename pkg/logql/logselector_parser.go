package logql

import (
	"github.com/grafana/loki/pkg/logql/parser"
	"github.com/prometheus/prometheus/pkg/labels"
)

var (
	LogSelectorParser = parser.BindWith2(
		Labels,
		FiltersP,
		func(ls, fs interface{}) interface{} {
			var expr LogSelectorExpr = &matchersExpr{ls.([]*labels.Matcher)}
			for _, f := range fs.([]Filter) {
				expr = NewFilterExpr(expr, f.ty, f.match)
			}
			return expr
		},
	)
)
