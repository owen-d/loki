package logql

import (
	"fmt"

	"github.com/grafana/loki/pkg/logql/log"
	"github.com/prometheus/prometheus/pkg/labels"
)

// QueryAbsence maps one SampleExpr into another which will act as a heuristic for whether the first expression
// could fire as an alert. It's used to create "meta-alerts" which can detect when log based alerting
// drifts out of sync with the code it monitors.
func QueryAbsence(expr SampleExpr) (Expr, error) {
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

func AbsenceLogRange(expr *logRange) (Expr, error) {
	selector, err := AbsenceLogSelector(expr.left)
	if err != nil {
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
							LabelFilterer: log.NewStringLabelFilter(
								labels.MustNewMatcher(
									labels.MatchRegexp,
									l, ".+",
								),
							),
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
