// Package approvalreporttest holds the in-memory fakes of the Approvals report's ports and a fixture with a
// round in every outcome, for the report's unit tests and the contract tests.
package approvalreporttest

import (
	"context"
	"slices"

	"mapp/internal/measure"
	"mapp/internal/measure/approvalreport"
	"mapp/internal/platform/actor"
)

type Scoper struct {
	Result measure.Scope
}

func (s Scoper) Scope(context.Context, actor.Actor, measure.Filter, []string) (measure.Scope, error) {
	return s.Result, nil
}

type Entry struct {
	approvalreport.PurchaseRequest
	LineItemIDs []string
}

type Purchases struct {
	Entries []Entry
}

func (p Purchases) OnLineItems(_ context.Context, lineItemIDs []string) ([]approvalreport.PurchaseRequest, error) {
	found := []approvalreport.PurchaseRequest{}
	for _, entry := range p.Entries {
		if slices.ContainsFunc(entry.LineItemIDs, func(id string) bool { return slices.Contains(lineItemIDs, id) }) {
			found = append(found, entry.PurchaseRequest)
		}
	}
	return found, nil
}

type People map[string]string

func (p People) Names(_ context.Context, personIDs []string) (map[string]string, error) {
	names := map[string]string{}
	for _, id := range personIDs {
		if name, known := p[id]; known {
			names[id] = name
		}
	}
	return names, nil
}
