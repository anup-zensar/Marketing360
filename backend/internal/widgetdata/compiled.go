package widgetdata

const (
	aggCount         = "count"
	aggCountDistinct = "countDistinct"
	aggSum           = "sum"
	aggAvg           = "avg"
	aggMin           = "min"
	aggMax           = "max"
)

type grouping struct {
	index  int
	field  Field
	bucket string
}

// measure.index is -1 for a count of records.
type measure struct {
	agg   string
	index int
	field Field
	label string
}

type condition struct {
	index  int
	field  Field
	op     string
	values []Value
}

type compiled struct {
	typ        string
	source     Source
	dimension  *grouping
	series     *grouping
	measures   []measure
	conditions []condition
	sort       Sort
	limit      int
	groupRest  bool
	seriesCap  int
}

func (m measure) kind() FieldKind {
	if m.agg == aggCount || m.agg == aggCountDistinct {
		return KindNumber
	}
	return m.field.Kind
}
