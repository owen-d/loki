package logql

import (
	"time"

	"github.com/grafana/loki/pkg/logql/parser"
)

// unexported
var (
	rangeVectorOps = []string{OpTypeCountOverTime, OpTypeRate}

	// <aggr-op>([parameter,] <vector expression>) [without|by (<label list>)]
	// 	parser.OneOfStrings([]string{OpTypeSum, OpTypeAvg, OpTypeMax, OpTypeMin, OpTypeCount, OpTypeStddev, OpTypeStdvar}...),
	// binOps         = []string{OpTypeOr, OpTypeAnd, OpTypeUnless, OpTypeAdd, OpTypeSub, OpTypeMul, OpTypeDiv, OpTypeMod, OpTypePow}

	// rate({job="mysql"} |= "error" != "timeout" [5m])
	// rate({job="mysql"}[5m] |= "error" != "timeout")

)

// operation parens(labels filtersP logRange)
// operation parens(labels logRange filtersP)

// var rangeOpOne = parser.BindWith2(
// 	parser.OneOfStrings(rangeVectorOps...),
// 	parser.Parens(
// 		parser.BindWith3(
// 			parser.Labels,
// 			FiltersP,
// 			rangeParser,
// 		),
// 	),
// )

var rangeParser = parser.Brackets(
	parser.BindWith2(
		parser.IntParser,
		durationParser,
		func(i, dur interface{}) interface{} {
			return time.Duration(i.(int)) * dur.(time.Duration)
		},
	),
)

var durationParser = parser.OneOf(
	parser.FMap(
		parser.Const(time.Millisecond),
		parser.StringP("ms"),
		"milliseconds",
	),
	parser.FMap(
		parser.Const(time.Second),
		parser.StringP("s"),
		"seconds",
	),
	parser.FMap(
		parser.Const(time.Minute),
		parser.StringP("m"),
		"minutes",
	),
	parser.FMap(
		parser.Const(time.Hour),
		parser.StringP("h"),
		"hours",
	),
)

type VectorAgg struct {
	params    int
	operation string
}
