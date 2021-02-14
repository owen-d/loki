package logql

import (
	"fmt"

	"github.com/prometheus/prometheus/pkg/labels"

	"github.com/grafana/loki/pkg/logql/log"
)

func MetaAlert(expr SampleExpr) (Expr, error) {
	rules := []func(SampleExpr) (SampleExpr, error){
		OnlyMatchersAlert,
		QueryAbsence,
		func(mapped SampleExpr) (SampleExpr, error) {
			// no need for a meta alert which is identical to the alert condition itself
			if expr.String() == mapped.String() {
				return nil, nil
			}
			return mapped, nil
		},
	}

	next := expr
	var err error
	for _, rule := range rules {
		next, err = rule(next)
		if err != nil || next == nil {
			return nil, err
		}
	}

	return next, nil
}

// QueryAbsence maps one SampleExpr into another which will act as a heuristic for whether the first expression
// could fire as an alert. It's used to create "meta-alerts" which can detect when log based alerting
// drifts out of sync with the code it monitors.
func QueryAbsence(expr SampleExpr) (SampleExpr, error) {
	switch e := expr.(type) {
	case *literalExpr:
		return nil, nil
	case *binOpExpr:
		lhs, err := QueryAbsence(e.SampleExpr)
		if err != nil {
			return nil, err
		}
		rhs, err := QueryAbsence(e.RHS)
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
		// does not affect, descend
		return QueryAbsence(e.left)
	case *labelReplaceExpr:
		return QueryAbsence(e.left)
	case *rangeAggregationExpr:
		if e.operation == OpRangeTypeAbsent {
			// Absent queries already satisfy the absence heuristic
			return nil, nil
		}
		return AbsenceLogRange(e.left)
	default:
		return nil, fmt.Errorf("QueryAbsence unexpected type %T", expr)
	}
}

func AbsenceLogRange(expr *logRange) (SampleExpr, error) {
	selector, err := AbsenceLogSelector(expr.left)
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

func AbsenceLogSelector(expr LogSelectorExpr) (LogSelectorExpr, error) {
	switch e := expr.(type) {
	case *literalExpr:
		return nil, nil
	case *matchersExpr:
		return e, nil
	case *pipelineExpr:
		var stages MultiStageExpr
		for _, p := range e.pipeline {
			switch s := p.(type) {
			case *lineFilterExpr:
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
				for _, l := range s.RequiredLabelNames() {
					stages = append(
						stages,
						&labelFilterExpr{
							LabelFilterer: nonEmptyRegexpLabelFilter(l),
						},
					)
				}

			default:
				stages = append(stages, s)
			}

		}

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

// Only returns non nil when this is not a binOpExpr and there is nothing more complex than
// a set of matchers. This is because the generated alert would be the inverse of the firing state
// and thus not helpful.
// For instance, the alert `rate({foo="bar"}[5m])` would likely create a dead mans switch like
// `absent_over_time({foo="bar"}[5m])`, which would mean either the alert or the switch would
// always be firing: not very helpful.
// In contrast, consider `rate({foo="bar"}[5m]) > 1`. The generated switch,
// `absent_over_time({foo="bar"}[5m])`, makes a lot of sense to keep as there is a threshold
// in the original alert.
func OnlyMatchersAlert(expr SampleExpr) (SampleExpr, error) {
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
		return nil, err
	}

	_, isBinOp := expr.(*binOpExpr)

	if onlyMatchers.(bool) && !isBinOp {
		return nil, nil
	}

	return expr, nil

}
