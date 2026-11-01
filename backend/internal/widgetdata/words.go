package widgetdata

const (
	msgNoAccess          = "You don't have access to custom dashboards."
	msgNoSource          = "This data source doesn't exist."
	msgNoDatasets        = "This dataset doesn't exist."
	msgChooseType        = "Choose a widget type."
	msgTextNote          = "A text note has no data."
	msgNoField           = "This field isn't in the data source."
	msgNotGroupable      = "This field can't be grouped."
	msgChooseBucket      = "Choose day, week, month, quarter or year."
	msgBucketOnlyDates   = "Only dates are grouped by day, week, month, quarter or year."
	msgChooseGroup       = "Choose what to group by."
	msgNoGroup           = "%s shows no groups."
	msgChooseBreakdown   = "Choose what to break down by."
	msgNoBreakdown       = "%s has no breakdown."
	msgSameBreakdown     = "Break down by a different field or date group."
	msgOneMeasure        = "%s shows one measure."
	msgOneWithBreakdown  = "With a breakdown, %s shows one measure."
	msgMeasureRange      = "%s shows %d to %d measures."
	msgChooseAggregation = "Choose how to measure."
	msgChooseField       = "Choose a field to measure."
	msgSumNeeds          = "Sum needs a number or money field."
	msgAvgNeeds          = "Average needs a number or money field."
	msgMinNeeds          = "Minimum needs a number, money or date field."
	msgMaxNeeds          = "Maximum needs a number, money or date field."
	msgScatterNeeds      = "A scatter chart measures numbers or money."
	msgTreemapNeeds      = "A treemap shows a sum or a count."
	msgLabelLength       = "Use at most 40 characters."
	msgTooManyFilters    = "Use at most 10 filters."
	msgChooseCondition   = "Choose a condition."
	msgConditionKind     = "This condition doesn't apply to this field."
	msgOneValue          = "Enter one value."
	msgSomeValues        = "Enter 1 to 50 values."
	msgTwoValues         = "Enter two values."
	msgNoValues          = "This condition takes no values."
	msgEnterNumber       = "Enter a number with up to 4 decimals."
	msgEnterMoney        = "Enter an amount in USD with up to 2 decimals."
	msgEnterDate         = "Enter a date as YYYY-MM-DD."
	msgTextLength        = "Use 1 to 200 characters."
	msgContainsLength    = "Use 1 to 100 characters."
	msgBetweenOrder      = "The first value is after the second."
	msgSortBy            = "Sort by the group or a measure."
	msgSortMeasure       = "Choose one of the measures."
	msgSortDirection     = "Choose ascending or descending."
	msgLimit             = "Show 1 to 500 groups."
	msgDateOrder         = "The start date is after the end date."
	msgTooManyPlans      = "Choose at most 20 plans."
	msgTooManyTeams      = "Choose at most 20 teams."
	msgComparisonKpi     = "Only a key figure compares."
	msgComparisonKind    = "Compare with a target or the previous period."
	msgEnterTarget       = "Enter the target."
	msgPeriodNeedsRange  = "Choose a date range to compare with the previous period."
	msgPeriodNeedsDate   = "This data has no date to compare periods by."
	msgTooLarge          = "The values are too large to add up."
	noteNoDate           = "Not filtered by date"
	noteNoPlan           = "Not filtered by plan"
	noteNoTeam           = "Not filtered by team"
	noteSeveralTeams     = "A plan with several teams counts in each"
	noteShowing          = "Showing %d of %d"
	noteUnknownCurrency  = "Amounts in a currency no longer in Settings are left out"
	financeNotePrefix    = "Financial figures"
	financeNote          = financeNotePrefix + " cover %d of %d plans"
)

var typeNames = map[string]string{ //nolint:gochecknoglobals // the widget types' names in messages
	typeBar: "A bar chart", typeLine: "A line chart", typeArea: "An area chart", typeCombo: "A bar and line chart",
	typePie: "A pie chart", typeDonut: "A donut chart", typeScatter: "A scatter chart", typeFunnel: "A funnel",
	typeGauge: "A gauge", typeHeatmap: "A heatmap", typeTreemap: "A treemap", typeRadar: "A radar chart",
	typeKpi: "A key figure", typeTable: "A table",
}

var aggregationWords = map[string]string{ //nolint:gochecknoglobals // default measure labels
	aggCount: "Count of", aggCountDistinct: "Count distinct of", aggSum: "Sum of", aggAvg: "Average of",
	aggMin: "Minimum of", aggMax: "Maximum of",
}
