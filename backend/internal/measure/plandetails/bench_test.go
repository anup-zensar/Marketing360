package plandetails

import (
	"fmt"
	"testing"

	"mapp/internal/allwork"
	"mapp/internal/measure"
	"mapp/internal/platform/actor"
)

// benchYear is the 028 volume (SC-001): 50 plans of 4 campaigns, 5,000 activities with 4 line items each, and 30
// active activity attributes of every kind, half of them filled in on each activity.
func benchYear(b *testing.B) (*year, []allwork.Attribute) {
	b.Helper()
	kinds := []string{allwork.TypeSingleChoice, allwork.TypeMultipleChoice, allwork.TypeMultipleChoiceSplit,
		allwork.TypeYesNo, "text"}
	var attributes []allwork.Attribute
	for i := range 30 {
		attributes = append(attributes, attribute(fmt.Sprintf("SA-%d", 100+i), fmt.Sprintf("Attribute %d", i),
			kinds[i%len(kinds)], "O1", "O2", "O3", "O4", "O5", "O6", "O7", "O8", "O9", "O10", "O11", "O12"))
	}
	y := &year{t: b, unmatched: map[string]bool{}}
	for p := range 50 {
		plan := fmt.Sprintf("P-%d", p)
		y.plan(plan, "BU-LUB")
		for c := range 4 {
			campaign := fmt.Sprintf("C-%d-%d", p, c)
			y.campaign(campaign, plan)
			for a := range 25 {
				y.benchActivity(fmt.Sprintf("A-%d-%d-%d", p, c, a), campaign, attributes)
			}
		}
	}
	return y, attributes
}

func (y *year) benchActivity(id, parent string, attributes []allwork.Attribute) {
	y.activity(id, parent, "1234.56", func(n *allwork.Node) {
		n.Attrs = map[string]allwork.AttrValue{}
		for i, a := range attributes {
			if i%2 == len(id)%2 {
				continue
			}
			option := a.Options[(len(id)+i)%len(a.Options)].ID
			switch a.Type {
			case allwork.TypeYesNo:
				yes := i%3 == 0
				n.Attrs[a.ID] = allwork.AttrValue{Yes: &yes}
			default:
				n.Attrs[a.ID] = allwork.AttrValue{OptionIDs: []string{option}}
			}
		}
	})
	for l := range 4 {
		y.nodes = append(y.nodes, allwork.Node{Level: allwork.LevelLineItem, ID: fmt.Sprintf("L-%s-%d", id, l),
			ParentID: id, PlanID: y.nodes[len(y.nodes)-1].PlanID})
	}
}

// BenchmarkPlanDetails_SC001 is each level's report at the 028 volume; the year read is measured by 036 and 028.
func BenchmarkPlanDetails_SC001(b *testing.B) {
	y, attributes := benchYear(b)
	s := NewService(fakeTx{}, &fakeScoped{matching: y.matching(), attributes: attributes})
	for _, show := range []string{measure.ShowActivity, measure.ShowCampaign, measure.ShowPlan} {
		b.Run(show, func(b *testing.B) {
			for b.Loop() {
				if _, err := s.Report(b.Context(), actor.Actor{ID: "U-1002"}, measure.Filter{Show: show}); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
