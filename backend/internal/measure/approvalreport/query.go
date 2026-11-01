package approvalreport

import "net/url"

// Query is the report's own part of the address; an unknown value is kept, and answers with empty figures (041-C10).
type Query struct {
	StatusType RecordType
}

func Params() []string {
	return []string{"statusType", "recordType", "approver", "outcome", "q", "sort", "order", "page", "pageSize"}
}

func QueryOf(own url.Values) Query {
	return Query{StatusType: RecordType(own.Get("statusType"))}
}
