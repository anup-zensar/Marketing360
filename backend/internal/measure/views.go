package measure

import (
	"context"
	"fmt"
	"net/url"
	"slices"
	"strings"
	"unicode/utf8"

	"mapp/internal/history"
	"mapp/internal/platform/actor"
	"mapp/internal/platform/errs"
	"mapp/internal/platform/ids"
	"mapp/internal/platform/record"
)

const (
	msgEnterViewName = "Enter a name for the view."
	msgViewNameLong  = "Use 60 characters or fewer."
	msgViewNameTaken = "You already have a view called '%s'."
	msgViewQuery     = "The view can't be saved — reload the page and try again."
	msgViewGone      = "This view no longer exists — it may have been deleted in another tab."
	msgViewChanged   = "This view was changed in another tab — reload it to see the latest"
)

const (
	maxViewName  = 60
	maxViewQuery = 4000
)

func (s *Service) ListViews(ctx context.Context, a actor.Actor, report string) (ReportViewPage, error) {
	if err := knownReport(report); err != nil {
		return ReportViewPage{}, err
	}
	page := ReportViewPage{Items: []ReportView{}}
	err := s.tx.Read(ctx, func(ctx context.Context) error {
		own, err := s.deps.Views.ListViews(ctx, a.ID, report)
		if err != nil {
			return err
		}
		slices.SortFunc(own, func(x, y StoredView) int {
			return strings.Compare(strings.ToLower(x.Name)+"\x00"+x.ID, strings.ToLower(y.Name)+"\x00"+y.ID)
		})
		for _, v := range own {
			page.Items = append(page.Items, v.view())
		}
		page.Total = len(page.Items)
		return nil
	})
	return page, err
}

func (s *Service) CreateView(ctx context.Context, a actor.Actor, report string, in ViewCreate) (ReportView, error) {
	if err := knownReport(report); err != nil {
		return ReportView{}, err
	}
	var created ReportView
	err := s.tx.Write(ctx, func(ctx context.Context) error {
		own, err := s.deps.Views.ListViews(ctx, a.ID, report)
		if err != nil {
			return err
		}
		var problems errs.Problems
		name := checkViewName(&problems, in.Name, own, "")
		query := checkViewQuery(&problems, in.Query, s.paramsOf(report))
		if err := problems.Err(); err != nil {
			return err
		}
		id, err := s.ids.Next(ctx, ids.PrefixReportView)
		if err != nil {
			return err
		}
		v := StoredView{ID: id, PersonID: a.ID, Report: report, Name: name, IsDefault: in.IsDefault, Query: query,
			Meta: record.NewMeta(a, s.now())}
		if err := s.saveView(ctx, a, v, own); err != nil {
			return err
		}
		created = v.view()
		return s.logView(ctx, a, v, history.ActionCreated, nil)
	})
	return created, err
}

func (s *Service) UpdateView(ctx context.Context, a actor.Actor, ref ViewRef, in ViewUpdate) (ReportView, error) {
	var updated ReportView
	err := s.tx.Write(ctx, func(ctx context.Context) error {
		before, err := s.ownView(ctx, a, ref)
		if err != nil {
			return err
		}
		if in.Version != before.Version {
			conflict := errs.VersionConflict("view", before.UpdatedBy.Name, before.UpdatedAt)
			conflict.Message = msgViewChanged
			return conflict
		}
		own, err := s.deps.Views.ListViews(ctx, a.ID, before.Report)
		if err != nil {
			return err
		}
		after, err := changedView(before, in, own, s.paramsOf(before.Report))
		if err != nil || after.logView() == before.logView() {
			updated = before.view()
			return err
		}
		after.Touch(a, s.now())
		if err := s.saveView(ctx, a, after, own); err != nil {
			return err
		}
		updated = after.view()
		return s.logChange(ctx, a, before, after)
	})
	return updated, err
}

// DeleteView removes the view for good; the answer is the view as it was, so Undo can save it again.
func (s *Service) DeleteView(ctx context.Context, a actor.Actor, ref ViewRef) (ReportView, error) {
	var deleted ReportView
	err := s.tx.Write(ctx, func(ctx context.Context) error {
		v, err := s.ownView(ctx, a, ref)
		if err != nil {
			return err
		}
		if err := s.logView(ctx, a, v, history.ActionDeleted, nil); err != nil {
			return err
		}
		deleted = v.view()
		return s.deps.Views.DeleteView(ctx, v.ID)
	})
	return deleted, err
}

type ViewRef struct {
	Report string
	ID     string
}

func (s *Service) ownView(ctx context.Context, a actor.Actor, ref ViewRef) (StoredView, error) {
	if err := knownReport(ref.Report); err != nil {
		return StoredView{}, err
	}
	v, found, err := s.deps.Views.GetView(ctx, ref.ID)
	if err != nil {
		return StoredView{}, err
	}
	if !found || v.PersonID != a.ID || v.Report != ref.Report {
		return StoredView{}, errs.NotFound(msgViewGone)
	}
	return v, nil
}

// saveView keeps one default view per person and report: starring v unstars the other one first.
func (s *Service) saveView(ctx context.Context, a actor.Actor, v StoredView, own []StoredView) error {
	for _, other := range own {
		if !v.IsDefault || other.ID == v.ID || !other.IsDefault {
			continue
		}
		unstarred := other
		unstarred.IsDefault = false
		unstarred.Touch(a, s.now())
		if err := s.deps.Views.PutView(ctx, unstarred); err != nil {
			return err
		}
		if err := s.logChange(ctx, a, other, unstarred); err != nil {
			return err
		}
	}
	return s.deps.Views.PutView(ctx, v)
}

func changedView(v StoredView, in ViewUpdate, own []StoredView, keep []string) (StoredView, error) {
	var problems errs.Problems
	if in.Name != nil {
		v.Name = checkViewName(&problems, *in.Name, own, v.ID)
	}
	if in.Query != nil {
		v.Query = checkViewQuery(&problems, *in.Query, keep)
	}
	if in.IsDefault != nil {
		v.IsDefault = *in.IsDefault
	}
	return v, problems.Err()
}

func checkViewName(problems *errs.Problems, name string, own []StoredView, exceptID string) string {
	name = strings.TrimSpace(name)
	notSingleLine := record.CheckSingleLine(name)
	switch {
	case name == "":
		problems.Add("name", msgEnterViewName)
	case utf8.RuneCountInString(name) > maxViewName:
		problems.Add("name", msgViewNameLong)
	case notSingleLine != "":
		problems.Add("name", notSingleLine)
	default:
		for _, v := range own {
			if v.ID != exceptID && strings.EqualFold(v.Name, name) {
				problems.Add("name", fmt.Sprintf(msgViewNameTaken, v.Name))
				break
			}
		}
	}
	return name
}

// checkViewQuery keeps the address in the API's canonical form, then the report's own parameters it keeps, so a
// saved view and the page it was saved from always give the same report.
func checkViewQuery(problems *errs.Problems, query string, keep []string) string {
	values, err := url.ParseQuery(query)
	if err != nil {
		problems.Add("query", msgViewQuery)
		return ""
	}
	f, err := ParseFilter(values)
	canonical := withOwn(Canonical(f), f.Own, keep)
	if err != nil || utf8.RuneCountInString(canonical) > maxViewQuery {
		problems.Add("query", msgViewQuery)
	}
	return canonical
}

func withOwn(canonical string, own url.Values, keep []string) string {
	parts := []string{}
	if canonical != "" {
		parts = append(parts, canonical)
	}
	for _, name := range slices.Sorted(slices.Values(keep)) {
		for _, v := range slices.Sorted(slices.Values(own[name])) {
			parts = append(parts, url.QueryEscape(name)+"="+url.QueryEscape(v))
		}
	}
	return strings.Join(parts, "&")
}

func (s *Service) paramsOf(report string) []string {
	r, found := s.registry.Get(report)
	if p, parameterised := r.(Parameterised); found && parameterised {
		return p.Params()
	}
	return nil
}

func knownReport(report string) error {
	if !slices.Contains(Keys(), report) {
		return errs.NotFound(msgNoReport)
	}
	return nil
}
