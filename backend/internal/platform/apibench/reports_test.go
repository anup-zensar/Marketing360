package apibench_test

import (
	"net/http"
	"testing"

	"mapp/internal/platform/apibench"
)

// BenchmarkReportsFrame measures 039's and 040's SC on the demo data and on the volume: a report's frame and Plan
// budget within 2 seconds, for every plan and for one market, with and without filters, and the shared export.
func BenchmarkReportsFrame(b *testing.B) {
	for _, data := range []apibench.Data{apibench.Demo, apibench.Volume} {
		b.Run(string(data), func(b *testing.B) {
			server := apibench.New(b, data)
			clients := signIn(b, server)
			for _, p := range reportPages() {
				b.Run(p.name, func(b *testing.B) { clients[p.who].Measure(b, http.MethodGet, p.path, nil) })
			}
		})
	}
}

func reportPages() []page {
	return []page{
		{"catalogue", alex, "/api/v1/reports"},
		{"frame every plan", alex, "/api/v1/reports/frame?report=planBudget&year=2026"},
		{"frame one market", priya, "/api/v1/reports/frame?report=planBudget&year=2026"},
		{"frame filtered", alex, "/api/v1/reports/frame?report=planBudget&year=2026&planType=operational&activityStatus=approved"},
		{"saved views", priya, "/api/v1/reports/plan-budget/views"},
		{"export xlsx", alex, "/api/v1/reports/plan-budget/export?year=2026&format=xlsx"},
		{"plan budget", alex, "/api/v1/reports/plan-budget?year=2026"},
		{"plan budget ranked", alex, "/api/v1/reports/plan-budget/ranked?year=2026&show=activity&group=owner"},
	}
}
