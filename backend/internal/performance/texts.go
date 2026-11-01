package performance

import "strconv"

const (
	activityKind    = "activity"
	factKind        = "performanceFact"
	mediaCodeKind   = "mediaCode"
	aliasKind       = "externalNameAlias"
	integrationKind = "integrationId"

	reviewKind         = "performanceReviewRow"
	sourceKind         = "performanceSource"
	sourceKindSingular = "data source"
	sourceKindPlural   = "data sources"

	msgSourceNotFound   = "This data source doesn't exist or you can't see it."
	mappingKindSingular = "column mapping"
	msgMappingNotFound  = "This column mapping doesn't exist or you can't see it."
	msgHoldsFacts       = "Not deleted — %s holds performance data. Deactivate it instead: its data stays and it stops accepting new loads."
	msgSourceInactive   = "Not loaded — %s is deactivated and accepts no further data. Reactivate it first."
	msgStageSupplied    = "Not saved — %s already supplies %s for this model. Deactivate that source first, or choose another stage."
	msgRevenueSupplied  = "Not saved — %s already supplies revenue for this model. Deactivate that source first."
)

const (
	msgNameLength  = "Give the data source a name of 1 to 120 characters."
	msgOwnerNeeded = "Choose an owner for this data source."
	msgNameHeld    = "Not saved — %s is already the external name or alias of activity %s. A name can belong to only one activity."
)

const (
	deactivatedVerb = "deactivated"
	reactivatedVerb = "reactivated"
	undoneVerb      = "undone"
	dismissedVerb   = "dismissed"
	reopenedVerb    = "reopened"
	resolvedVerb    = "resolved"
	supersededVerb  = "superseded by a later load"

	importKind         = "performanceImport"
	importKindSingular = "upload"
	importKindPlural   = "uploads"

	msgPersonalData        = "Not loaded - the file carries personal data in %s. Remove those columns and upload it again."
	msgImportNotFound      = "This upload doesn't exist or you can't see it."
	msgNotUndoable         = "Not undone - %s is not a completed upload."
	msgFileUnreadable      = "This file can't be read. Save it as a .csv file in UTF-8 and try again."
	msgNoPull              = "Not pulled - %s does not deliver by simulated feed. Change its delivery route first."
	msgActivityNotFound    = "This activity doesn't exist or you can't see it."
	msgAliasNotFound       = "This alias doesn't exist or you can't see it."
	msgIntegrationNotFound = "This activity has no integration ID for that data source."
	msgIntegrationExists   = "Not saved - %s already addresses this activity as %s. Remove that integration ID first."
	msgIntegrationValue    = "Give the integration ID a value of 1 to 120 characters."
	msgReviewRowNotFound   = "This row doesn't exist or you can't see it."
	msgCommentNeeded       = "Say why this row is being dismissed."
	msgReviewNotChangeable = "Not changed - this row is %s."
	msgUnknownColumn       = "This row carries no column named %s."
)

func codeText(code int64) string {
	return strconv.FormatInt(code, 10)
}

const (
	fieldColumnName = "columnName"
	fieldMeaning    = "meaning"
	fieldKeyType    = "keyType"
	fieldStage      = "stage"
	fieldWebMetric  = "webMetric"
	fieldModelID    = "modelId"

	msgColumnLength   = "Give the column a name of 1 to 120 characters."
	msgMeaningUnknown = "Choose what this column means."
	msgStageNeeded    = "A stage column names its stage, and only a stage column does."
	msgKeyTypeNeeded  = "A key column names which key it carries, and only a key column does."
	msgMetricNeeded   = "A web-metric column names its metric, and only a web-metric column does."
	msgMetricUnknown  = "A web metric is sessions, visits or engagement touches."
	msgModelNeeded    = "Only a stage or revenue column is supplied by a model, and it must name one."
)

const (
	reasonNoKey     = "unattributed:noKey"
	reasonRejected  = "rejected:"
	reasonUnknown   = "unmatched:unknown"
	reasonAmbiguous = "unmatched:ambiguous"
	reasonDisagree  = "unmatched:disagree"
)

const (
	faultEmptyColumn       = "emptyColumn"
	faultInvalidDate       = "invalidDate"
	faultFutureDate        = "futureDate"
	faultNegativeCount     = "negativeCount"
	faultFractionalCount   = "fractionalCount"
	faultInvalidCount      = "invalidCount"
	faultInvalidAmount     = "invalidAmount"
	faultUnknownCurrency   = "unconfiguredCurrency"
	faultMissingReference  = "missingReference"
	faultReferenceTooLong  = "referenceTooLong"
	faultRepeatedReference = "repeatedReference"
	faultNoMeasure         = "noMeasure"
	faultDuplicateMeasure  = "duplicateMeasure"

	msgReferenceNeeded   = "Every row needs the source's own reference, so a later load can replace it."
	msgDateNeeded        = "Every row needs the date its data is for."
	msgDateInvalid       = "Write the date as YYYY-MM-DD."
	msgDateFuture        = "A row cannot be dated in the future."
	msgCountInvalid      = "A count is a whole number."
	msgCountNegative     = "A count cannot be below zero."
	msgCountFractional   = "A count is a whole number."
	msgAmountInvalid     = "Write the amount as a number, such as 1234.50."
	msgCurrencyUnknown   = "Name a currency this organisation has configured."
	msgReferenceTooLong  = "A source reference is at most 200 characters."
	msgReferenceRepeated = "Another row in this file already carries this reference."
	msgMeasureNeeded     = "A row must carry a figure in at least one of the source's measure columns."
	msgMeasureTwice      = "Two filled columns carry the same measure; a row answers each measure once."
)
