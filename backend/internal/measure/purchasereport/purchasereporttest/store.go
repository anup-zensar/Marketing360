// Package purchasereporttest holds the in-memory fakes of the Purchase requests report's ports and a fixture with a
// request in every status, for the report's unit tests and contract tests.
package purchasereporttest

import (
	"context"
	"slices"

	"mapp/internal/measure"
	"mapp/internal/measure/purchasereport"
	"mapp/internal/platform/actor"
)

type Scoper struct {
	Result measure.Scope
}

func (s Scoper) Scope(context.Context, actor.Actor, measure.Filter, []string) (measure.Scope, error) {
	return s.Result, nil
}

type Purchases struct {
	Requests []purchasereport.Request
	Asked    *[]string
	Off      bool
}

func (p Purchases) OnLineItems(_ context.Context, lineItemIDs []string) (purchasereport.Requests, error) {
	if p.Asked != nil {
		*p.Asked = append(*p.Asked, lineItemIDs...)
	}
	found := []purchasereport.Request{}
	for _, r := range p.Requests {
		if slices.ContainsFunc(r.Lines, func(l purchasereport.Line) bool { return slices.Contains(lineItemIDs, l.LineItemID) }) {
			found = append(found, r)
		}
	}
	return purchasereport.Requests{Items: found, SwitchedOn: !p.Off}, nil
}
