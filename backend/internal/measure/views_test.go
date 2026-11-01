package measure_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/history"
	"mapp/internal/history/historytest"
	"mapp/internal/measure"
	"mapp/internal/measure/measuretest"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/errs"
	"mapp/internal/platform/fake"
	"mapp/internal/platform/ids"
)

var (
	priyaSaves  = actor.Actor{ID: "U-1002", Name: "Priya Raman"}
	danielSaves = actor.Actor{ID: "U-1003", Name: "Daniel Okoro"}
)

func viewService(t *testing.T, reports ...measure.Report) (*measure.Service, *history.Service) {
	t.Helper()
	db := fake.NewDB()
	generator := ids.NewGenerator(fake.NewCounters(db))
	now := func() time.Time { return time.Date(2026, 10, 30, 9, 0, 0, 0, time.UTC) }
	log := history.NewService(db, historytest.NewStore(db), generator, now)
	registry := measure.NewRegistry()
	for _, r := range reports {
		if err := registry.Register(r); err != nil {
			t.Fatalf("Register(%s): %v", r.Key(), err)
		}
	}
	s := measure.NewService(history.Tools{Tx: db, IDs: generator, Log: log, Now: now},
		measure.Deps{Views: measuretest.NewStore(db)}, registry)
	if err := log.Register(s.ViewKind()); err != nil {
		t.Fatalf("Register: %v", err)
	}
	return s, log
}

func create(t *testing.T, s *measure.Service, a actor.Actor, report string, in measure.ViewCreate) measure.ReportView {
	t.Helper()
	v, err := s.CreateView(t.Context(), a, report, in)
	if err != nil {
		t.Fatalf("CreateView(%+v): %v", in, err)
	}
	return v
}

// problemOf is the message a person reads: the first field's, else the error's; "" without an error.
func problemOf(err error) string {
	var e *errs.Error
	switch {
	case err == nil:
		return ""
	case !errors.As(err, &e):
		return err.Error()
	case len(e.Fields) > 0:
		return e.Fields[0].Message
	}
	return e.Message
}

func manyValues() string {
	var b strings.Builder
	for i := range 700 {
		fmt.Fprintf(&b, "&attr=SA-1022:SO-%d", 1000+i)
	}
	return b.String()
}

func TestViews_NameAndQuery_FR024_FR027(t *testing.T) {
	t.Parallel()
	s, _ := viewService(t)
	saved := create(t, s, priyaSaves, measure.KeyPlanBudget,
		measure.ViewCreate{Name: "  Europe  ", Query: "owner=U-1003&year=2026&owner=U-1002"})
	if saved.Name != "Europe" || saved.Query != "year=2026&owner=U-1002&owner=U-1003" || saved.Version != 1 {
		t.Errorf("saved %+v: want the name trimmed and the canonical query", saved)
	}

	for _, c := range []struct {
		name, report, view, query, want string
		by                              actor.Actor
	}{
		{"no name", measure.KeyPlanBudget, "  ", "", "Enter a name for the view.", priyaSaves},
		{"too long", measure.KeyPlanBudget, strings.Repeat("x", 61), "", "Use 60 characters or fewer.", priyaSaves},
		{"taken, ignoring case", measure.KeyPlanBudget, "EUROPE", "", "You already have a view called 'Europe'.", priyaSaves},
		{"a line break", measure.KeyPlanBudget, "Euro\npe", "", "Names can't contain tabs, line breaks or text-direction characters.", priyaSaves},
		{"a value the report does not take", measure.KeyPlanBudget, "Yearly", "period=year", "The view can't be saved — reload the page and try again.", priyaSaves},
		{"too long an address", measure.KeyPlanBudget, "Many", manyValues(), "The view can't be saved — reload the page and try again.", priyaSaves},
		{"the same name on another report", measure.KeyApprovals, "Europe", "", "", priyaSaves},
		{"the same name for another person", measure.KeyPlanBudget, "Europe", "", "", danielSaves},
	} {
		_, err := s.CreateView(t.Context(), c.by, c.report, measure.ViewCreate{Name: c.view, Query: c.query})
		if got := problemOf(err); got != c.want {
			t.Errorf("%s: error %v, want %q", c.name, err, c.want)
		}
	}
	if _, err := s.CreateView(t.Context(), priyaSaves, "salesPipeline", measure.ViewCreate{Name: "X"}); problemOf(err) != "There is no such report." {
		t.Errorf("an unknown report: %v, want not found", err)
	}
}

func TestViews_PersonalWithOneDefault_FR025_FR026(t *testing.T) {
	t.Parallel()
	s, _ := viewService(t)
	ctx := t.Context()
	europe := create(t, s, priyaSaves, measure.KeyPlanBudget, measure.ViewCreate{Name: "europe", IsDefault: true})
	create(t, s, priyaSaves, measure.KeyPlanBudget, measure.ViewCreate{Name: "Asia", IsDefault: true})
	create(t, s, priyaSaves, measure.KeyApprovals, measure.ViewCreate{Name: "Mine", IsDefault: true})

	page, err := s.ListViews(ctx, priyaSaves, measure.KeyPlanBudget)
	if err != nil {
		t.Fatalf("ListViews: %v", err)
	}
	var got []string
	for _, v := range page.Items {
		got = append(got, fmt.Sprintf("%s %t v%d", v.Name, v.IsDefault, v.Version))
	}
	if diff := cmp.Diff([]string{"Asia true v1", "europe false v2"}, got); diff != "" || page.Total != 2 {
		t.Errorf("Priya's Plan budget views (-want +got):\n%s", diff)
	}
	if daniels, _ := s.ListViews(ctx, danielSaves, measure.KeyPlanBudget); daniels.Total != 0 {
		t.Errorf("Daniel sees %d of Priya's views", daniels.Total)
	}
	ref := measure.ViewRef{Report: measure.KeyPlanBudget, ID: europe.ID}
	gone := "This view no longer exists — it may have been deleted in another tab."
	if _, err := s.DeleteView(ctx, danielSaves, ref); problemOf(err) != gone {
		t.Errorf("Daniel deletes Priya's view: %v, want not found", err)
	}
	if _, err := s.DeleteView(ctx, priyaSaves, measure.ViewRef{Report: measure.KeyApprovals, ID: europe.ID}); problemOf(err) != gone {
		t.Error("a view was deleted through another report's address")
	}
}

func TestViews_VersionAndLog_FR028(t *testing.T) {
	t.Parallel()
	s, log := viewService(t)
	ctx := t.Context()
	europe := create(t, s, priyaSaves, measure.KeyPlanBudget, measure.ViewCreate{Name: "Europe", IsDefault: true})
	asia := create(t, s, priyaSaves, measure.KeyPlanBudget, measure.ViewCreate{Name: "Asia", IsDefault: true})
	asiaRef := measure.ViewRef{Report: measure.KeyPlanBudget, ID: asia.ID}
	renamed := "Asia Pacific"

	updated, err := s.UpdateView(ctx, priyaSaves, asiaRef, measure.ViewUpdate{Version: 1, Name: &renamed})
	if err != nil || updated.Name != renamed || updated.Version != 2 {
		t.Fatalf("rename = %+v, %v", updated, err)
	}
	same, err := s.UpdateView(ctx, priyaSaves, asiaRef, measure.ViewUpdate{Version: 2, Name: &renamed})
	if err != nil || same.Version != 2 {
		t.Errorf("an update that changes nothing = %+v, %v: want version 2 kept", same, err)
	}
	_, err = s.UpdateView(ctx, priyaSaves, asiaRef, measure.ViewUpdate{Version: 1, Name: &renamed})
	var conflict *errs.Error
	if !errors.As(err, &conflict) || conflict.Code != errs.CodeVersionConflict ||
		conflict.Message != "This view was changed in another tab — reload it to see the latest" {
		t.Errorf("a stale version: %v, want the version conflict", err)
	}
	deleted, err := s.DeleteView(ctx, priyaSaves, measure.ViewRef{Report: measure.KeyPlanBudget, ID: europe.ID})
	if err != nil || deleted.Name != "Europe" || deleted.IsDefault {
		t.Errorf("delete = %+v, %v: want the view as it was, for Undo", deleted, err)
	}

	entries, err := log.Entries(ctx)
	if err != nil {
		t.Fatalf("Entries: %v", err)
	}
	var got []string
	for _, e := range entries {
		got = append(got, fmt.Sprintf("%s %s %s %d", e.RecordType, e.Action, e.RecordID, len(e.Changes)))
	}
	want := []string{
		"reportView created " + europe.ID + " 0", "reportView changed " + europe.ID + " 1",
		"reportView created " + asia.ID + " 0", "reportView changed " + asia.ID + " 1",
		"reportView deleted " + europe.ID + " 0",
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("log entries (-want +got):\n%s", diff)
	}
}

type sortedReport struct{}

func (sortedReport) Key() string     { return measure.KeyApprovals }
func (sortedReport) Available() bool { return true }
func (sortedReport) Params() []string {
	return []string{"status", "sort"}
}

func (sortedReport) Sheets(context.Context, actor.Actor, measure.Filter) ([]measure.WidgetSheet, error) {
	return nil, nil
}

func TestViews_KeepTheReportsOwnParameters_039_B4(t *testing.T) {
	t.Parallel()
	s, _ := viewService(t, sortedReport{})
	query := "status=open&page=3&sort=name&status=declined&owner=U-1002&junk=1"

	approvals := create(t, s, priyaSaves, measure.KeyApprovals, measure.ViewCreate{Name: "Open", Query: query})
	budget := create(t, s, priyaSaves, measure.KeyPlanBudget, measure.ViewCreate{Name: "Open", Query: query})

	if want := "owner=U-1002&sort=name&status=declined&status=open"; approvals.Query != want {
		t.Errorf("Approvals view query = %q, want %q: its own parameters kept, page and unknown ones dropped", approvals.Query, want)
	}
	if budget.Query != "owner=U-1002" {
		t.Errorf("Plan budget view query = %q: a report without parameters of its own keeps only the filter", budget.Query)
	}
}
