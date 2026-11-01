// Package useraccess is the User access report (feature 045): everyone with roles, teams and last sign-in.
package useraccess

import (
	"time"

	"mapp/internal/platform/authz"
	"mapp/internal/platform/date"
)

// The Status filter (FR-006).
const (
	StatusActive   = "active"
	StatusInactive = "inactive"
	StatusAll      = "all"
)

// The bands of the Last sign-in widget (FR-007).
const (
	BandWithin30Days = "within30Days"
	BandDays31to90   = "days31to90"
	BandOver90Days   = "over90Days"
	BandNever        = "never"
)

// Report is the contract's UserAccessReport: the cards and bands over every matching person, one page of rows.
type Report struct {
	Total         int       `json:"total"`
	Page          int       `json:"page"`
	PageSize      int       `json:"pageSize"`
	Sort          string    `json:"sort"`
	Cards         Cards     `json:"cards"`
	Bands         []Band    `json:"bands"`
	RecordedSince date.Date `json:"recordedSince"`
	Rows          []Row     `json:"rows"`
}

type Cards struct {
	People               int `json:"people"`
	SignedInWithin30Days int `json:"signedInWithin30Days"`
	NeverSignedIn        int `json:"neverSignedIn"`
}

type Band struct {
	Key   string `json:"key"`
	Label string `json:"label"`
	Count int    `json:"count"`
}

type Row struct {
	Person         authz.PersonRef        `json:"person"`
	Email          string                 `json:"email"`
	Roles          []authz.RoleAssignment `json:"roles"`
	Teams          []authz.TeamRef        `json:"teams"`
	LastSignedInAt *time.Time             `json:"lastSignedInAt"`
}

// Query is the report's own parameters: search, teams, status, sort and page (FR-006).
type Query struct {
	Search   string
	Teams    []string
	Status   string
	Sort     string
	Page     int
	PageSize int
}
