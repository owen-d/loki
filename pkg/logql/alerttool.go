package logql

import (
	"fmt"

	"github.com/prometheus/prometheus/pkg/labels"

	"github.com/grafana/loki/pkg/logql/log"
)

type rule struct {
	test   func(SampleExpr) (bool, error)
	mapper func(SampleExpr) (SampleExpr, error)
}

func MetaAlert(expr SampleExpr) (Expr, error) {
	rules := []rule{
		onlyMatchersAlert,
		queryAbesenceAlert,
	}
	for _, rule := range rules {
		shouldMap, err := rule.test(expr)
		if err != nil {
			return nil, err
		}
		if shouldMap {
			return rule.mapper(expr)
		}
	}

	return nil, nil
}

var queryAbesenceAlert = rule{
	test: func(expr SampleExpr) (bool, error) {
		return true, nil
	},
	mapper: func(expr SampleExpr) (SampleExpr, error) {
		return QueryAbsence(expr, astMetadata{})
	},
}

// QueryAbsence maps one SampleExpr into another which will act as a heuristic for whether the first expression
// could fire as an alert. It's used to create "meta-alerts" which can detect when log based alerting
// drifts out of sync with the code it monitors.
func QueryAbsence(expr SampleExpr, hints astMetadata) (SampleExpr, error) {
	switch e := expr.(type) {
	case *literalExpr:
		return nil, nil
	case *binOpExpr:
		lhs, err := QueryAbsence(e.SampleExpr, hints)
		if err != nil {
			return nil, err
		}
		rhs, err := QueryAbsence(e.RHS, hints)
		if err != nil {
			return nil, err
		}

		if lhs == nil {
			return rhs, nil
		}
		if rhs == nil {
			return lhs, nil
		}

		// OR both sides when building a validation for an alert.
		// Note: we could alternatively be more stringent and AND them together
		// to ensure both sides have data present.
		// That however, wouldn't work great for queries like
		// `rate({foo="bar", mode="panic"}[5m]) / rate({foo="bar"}[5m])`
		return &binOpExpr{
			SampleExpr: lhs.(SampleExpr),
			RHS:        rhs.(SampleExpr),
			op:         OpTypeOr,
		}, nil

	case *vectorAggregationExpr:
		// add any groupings as extraLabels
		if e.grouping != nil && !e.grouping.without {
			hints.extraLabels = append(hints.extraLabels, e.grouping.groups...)
		}
		return QueryAbsence(e.left, hints)
	case *labelReplaceExpr:
		return QueryAbsence(e.left, hints)
	case *rangeAggregationExpr:
		if e.operation == OpRangeTypeAbsent {
			// Absent queries already satisfy the absence heuristic
			return nil, nil
		}
		// add any groupings as extraLabels
		if e.grouping != nil && !e.grouping.without {
			hints.extraLabels = append(hints.extraLabels, e.grouping.groups...)
		}
		return AbsenceLogRange(e.left, hints)
	default:
		return nil, fmt.Errorf("QueryAbsence unexpected type %T", expr)
	}
}

func AbsenceLogRange(expr *logRange, hints astMetadata) (SampleExpr, error) {
	if expr.unwrap != nil {
		hints.extraLabels = append(hints.extraLabels, expr.unwrap.identifier)
	}

	selector, err := AbsenceLogSelector(expr.left, hints)
	if err != nil || selector == nil {
		return nil, err
	}

	return newRangeAggregationExpr(
		newLogRange(
			selector,
			expr.interval,
			nil,
		),
		OpRangeTypeAbsent,
		nil,
		nil,
	), nil
}

func AbsenceLogSelector(expr LogSelectorExpr, hints astMetadata) (LogSelectorExpr, error) {
	switch e := expr.(type) {
	case *literalExpr:
		return nil, nil
	case *matchersExpr:
		if len(hints.extraLabels) > 0 {
			return newPipelineExpr(e, nonEmptyRegexpStages(hints.extraLabels...)), nil
		}
		return e, nil
	case *pipelineExpr:
		var stages MultiStageExpr
		for i, p := range e.pipeline {
			switch s := p.(type) {
			case *lineFilterExpr:

				// if a parser follows this filter, include it
				if hasParser(e.pipeline[i+1:]) {
					stages = append(stages, s)
				}
				continue
			case *labelFilterExpr:

				// special handling for binops -- we want to preserve their (AND|OR)
				// groupings but otherwise make them use non-empty regexp matchers.
				if binop, ok := s.LabelFilterer.(*log.BinaryLabelFilter); ok {
					stages = append(
						stages,
						&labelFilterExpr{
							LabelFilterer: mapAbsenceBinFilter(binop),
						},
					)
					continue
				}

				// Any label filters should be ammended
				// to only require the tested label _exists_.
				// We don't want to test the exact condition which
				// would cause the alert to fire. Instead, we
				// want to ensure the labels that will be tested
				// are _present_.
				extraStages := nonEmptyRegexpStages(s.RequiredLabelNames()...)
				stages = append(stages, extraStages...)
			default:
				stages = append(stages, s)
			}
		}

		// add any extra labels that may be referenced by upstream aggregations
		stages = append(stages, nonEmptyRegexpStages(hints.extraLabels...)...)

		if len(stages) == 0 {
			return e.left, nil
		}

		return newPipelineExpr(e.left, stages), nil

	default:
		return nil, fmt.Errorf("AbsenceLogSelector unexpected type %T", expr)
	}

}

func mapAbsenceBinFilter(expr *log.BinaryLabelFilter) *log.BinaryLabelFilter {
	res := &log.BinaryLabelFilter{
		And: expr.And,
	}

	if lhs, ok := expr.Left.(*log.BinaryLabelFilter); ok {
		res.Left = mapAbsenceBinFilter(lhs)
	} else {
		// Otherwise it's not a binop and it can't possibly be a noop
		// because a binop isn't created when one leg is a noop, so it must have one label.
		res.Left = nonEmptyRegexpLabelFilter(expr.Left.RequiredLabelNames()[0])
	}

	if rhs, ok := expr.Right.(*log.BinaryLabelFilter); ok {
		res.Right = mapAbsenceBinFilter(rhs)
	} else {
		// Otherwise it's not a binop and it can't possibly be a noop
		// because a binop isn't created when one leg is a noop, so it must have one label.
		res.Right = nonEmptyRegexpLabelFilter(expr.Right.RequiredLabelNames()[0])
	}

	return res
}

func nonEmptyRegexpLabelFilter(lName string) *log.StringLabelFilter {
	return log.NewStringLabelFilter(
		labels.MustNewMatcher(
			labels.MatchRegexp,
			lName, ".+",
		),
	)
}

func nonEmptyRegexpStages(ls ...string) MultiStageExpr {
	var stages MultiStageExpr
	for _, l := range ls {
		stages = append(
			stages,
			&labelFilterExpr{
				LabelFilterer: nonEmptyRegexpLabelFilter(l),
			},
		)
	}
	return stages
}

func hasParser(xs MultiStageExpr) (hasParser bool) {
	for _, s := range xs {
		if _, ok := s.(*labelParserExpr); ok {
			hasParser = true
		}
	}
	return hasParser
}

// Only returns non nil when this is not a binOpExpr and there is nothing more complex than
// a set of matchers. This is because the generated alert would be the inverse of the firing state
// and thus not helpful.
// For instance, the alert `rate({foo="bar"}[5m])` would likely create a dead mans switch like
// `absent_over_time({foo="bar"}[5m])`, which would mean either the alert or the switch would
// always be firing: not very helpful.
// In contrast, consider `rate({foo="bar"}[5m]) > 1`. The generated switch,
// `absent_over_time({foo="bar"}[5m])`, makes a lot of sense to keep as there is a threshold
// in the original alert.
var onlyMatchersAlert = rule{
	test: func(expr SampleExpr) (bool, error) {

		switch e := expr.(type) {
		case *rangeAggregationExpr:
			// not only matchers, we'll need to check for aggregated labels
			if e.grouping != nil && !e.grouping.without && len(e.grouping.groups) > 0 {
				return false, nil
			}
		case *vectorAggregationExpr:
			// not only matchers, we'll need to check for aggregated labels
			if e.grouping != nil && !e.grouping.without && len(e.grouping.groups) > 0 {
				return false, nil
			}

		default:
			// We're only interested in range/vec aggs
			return false, nil
		}

		onlyMatchers, err := expr.Fold(func(accum interface{}, expr interface{}) (interface{}, error) {
			acc := accum.(bool)
			switch e := expr.(type) {
			case *logRange:
				_, isMatchers := e.left.(*matchersExpr)
				// only return true if all log ranges are only
				// simple matchersExprs
				return acc && isMatchers, nil
			}
			return acc, nil
		}, true)

		if err != nil {
			return false, err
		}

		return onlyMatchers.(bool), nil
	},

	mapper: func(expr SampleExpr) (SampleExpr, error) {

		if _, isBinOp := expr.(*binOpExpr); !isBinOp {
			return nil, nil
		}

		return expr, nil

	},
}

type astMetadata struct {
	matchers    []*labels.Matcher
	groupings   []*grouping
	extraLabels []string
}

func astMeta(expr SampleExpr) (*astMetadata, error) {
	md, err := expr.Fold(func(accum interface{}, expr interface{}) (interface{}, error) {
		md := accum.(*astMetadata)

		switch e := expr.(type) {
		case *matchersExpr:
			md.matchers = append(md.matchers, e.matchers...)
		case *vectorAggregationExpr:
			md.groupings = append(md.groupings, e.grouping)
		case *rangeAggregationExpr:
			md.groupings = append(md.groupings, e.grouping)
		}

		return md, nil

	}, &astMetadata{})

	if err != nil {
		return nil, err
	}
	return md.(*astMetadata), nil

}
