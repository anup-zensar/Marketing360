package apibench_test

import (
	"net/http"
	"testing"

	"mapp/internal/platform/apibench"
)

// BenchmarkApprovalsReport measures feature 041's SC-006: the Approvals report of 2026 with and without its
// grid parameters, whose statements must stay the same whatever the number of rows.
func BenchmarkApprovalsReport(b *testing.B) {
	for _, data := range []apibench.Data{apibench.Demo, apibench.Volume} {
		b.Run(string(data), func(b *testing.B) {
			server := apibench.New(b, data)
			clients := signIn(b, server)
			for _, p := range approvalsPages() {
				b.Run(p.name, func(b *testing.B) { clients[p.who].Measure(b, http.MethodGet, p.path, nil) })
			}
		})
	}
}

func approvalsPages() []page {
	const report = "/api/v1/reports/approvals?year=2026"
	return []page{
		{"report", alex, report},
		{"report one market", priya, report},
		{"purchase requests", alex, report + "&statusType=purchaseRequest"},
		{"grid searched and sorted", alex, report + "&q=launch&sort=amount&order=desc&pageSize=100"},
		{"grid filtered", alex, report + "&outcome=waiting&recordType=activity"},
		{"export grid", alex, "/api/v1/reports/approvals/export?year=2026&widget=approvals&format=csv"},
	}
}
