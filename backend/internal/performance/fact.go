package performance

import (
	"time"

	"mapp/internal/platform/money"
	"mapp/internal/platform/record"
)

type Origin string

const (
	OriginUpload    Origin = "upload"
	OriginSimulated Origin = "simulated"
)

type ImportStatus string

const (
	ImportCompleted ImportStatus = "completed"
	ImportFailed    ImportStatus = "failed"
	ImportUndone    ImportStatus = "undone"
)

type Outcome string

const (
	OutcomeMatched      Outcome = "matched"
	OutcomeHeld         Outcome = "held"
	OutcomeUnattributed Outcome = "unattributed"
	OutcomeRejected     Outcome = "rejected"
	OutcomeUnmatched    Outcome = "unmatched"
)

type Counts struct {
	Read         int `json:"read"`
	Matched      int `json:"matched"`
	Held         int `json:"held"`
	Unattributed int `json:"unattributed"`
	Rejected     int `json:"rejected"`
	Unmatched    int `json:"unmatched"`
}

// Reconciles is the invariant the database also carries as a CHECK (FR-022).
func (c Counts) Reconciles() bool {
	return c.Read == c.Matched+c.Held+c.Unattributed+c.Rejected+c.Unmatched
}

type Import struct {
	ID          string       `json:"id"`
	SourceID    string       `json:"sourceId"`
	FileName    string       `json:"fileName"`
	Origin      Origin       `json:"origin"`
	Status      ImportStatus `json:"status"`
	DataThrough *time.Time   `json:"dataThrough,omitempty"`
	Counts      Counts       `json:"counts"`
	record.Meta
}

// MatchedOn is the key a fact matched on, in the order they are tried (FR-018).
type MatchedOn string

const (
	MatchedNone          MatchedOn = ""
	MatchedActivityID    MatchedOn = "activityId"
	MatchedMediaCode     MatchedOn = "mediaCode"
	MatchedIntegrationID MatchedOn = "integrationId"
	MatchedExternalName  MatchedOn = "externalName"
)

// MatchOrder is the strict precedence: the first key present in a row decides the match (FR-018).
func MatchOrder() []MatchedOn {
	return []MatchedOn{MatchedActivityID, MatchedMediaCode, MatchedIntegrationID, MatchedExternalName}
}

type Fact struct {
	ID              string       `json:"id"`
	SourceID        string       `json:"sourceId"`
	ImportID        string       `json:"importId"`
	SourceReference string       `json:"sourceReference"`
	ActivityID      string       `json:"activityId"`
	MatchedOn       MatchedOn    `json:"matchedOn"`
	FactDate        time.Time    `json:"factDate"`
	Stage           string       `json:"stage"`
	WebMetric       WebMetric    `json:"webMetric"`
	Count           *int64       `json:"count,omitempty"`
	Amount          *money.Money `json:"amount,omitempty"`
	Held            bool         `json:"held"`
	ReplacedByID    string       `json:"replacedById"`
	record.Meta
}

// Counts reports whether this fact may appear in a figure: a held fact counts nowhere until its activity
// is approved (FR-020), and a replaced one has been superseded (FR-013).
func (f Fact) Counts() bool {
	return !f.Held && f.ActivityID != "" && f.ReplacedByID == ""
}

type ReviewStatus string

const (
	ReviewOpen      ReviewStatus = "open"
	ReviewResolved  ReviewStatus = "resolved"
	ReviewDismissed ReviewStatus = "dismissed"
)

// Interpreted is the row as the load read it through that day's mappings, kept so a correction and a
// re-check never restate it through later mappings (FR-007).
type Interpreted struct {
	Reference string    `json:"reference"`
	Date      string    `json:"date"`
	Currency  string    `json:"currency"`
	Keys      RowKeys   `json:"keys"`
	Readings  []Reading `json:"readings"`
	Corrected bool      `json:"corrected"`
}

// Row rebuilds what Classify takes, from the stored interpretation rather than from today's mappings.
func (i Interpreted) Row(raw map[string]string) Row {
	return Row{
		Raw:       raw,
		Reference: Cell{Column: "reference", Value: i.Reference},
		Date:      Cell{Column: "date", Value: i.Date},
		Currency:  Cell{Column: "currency", Value: i.Currency},
		Keys:      i.Keys,
		Readings:  i.Readings,
	}
}

type ReviewRow struct {
	ID         string       `json:"id"`
	ImportID   string       `json:"importId"`
	SourceID   string       `json:"sourceId"`
	Raw        []byte       `json:"raw"`
	Parsed     Interpreted  `json:"parsed"`
	Reason     string       `json:"reason"`
	Faults     []Fault      `json:"faults"`
	Candidates []string     `json:"candidates,omitempty"`
	Status     ReviewStatus `json:"status"`
	Comment    string       `json:"comment"`
	record.Meta
}
