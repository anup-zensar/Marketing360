package widgetdata_test

import (
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"

	"mapp/internal/platform/apitest"
	"mapp/internal/widgetdata"
)

const (
	dataSources = "/api/v1/data-sources"
	query       = "/api/v1/data-sources/query"
	priyaEmail  = "priya.raman@shell.com"
	carlosEmail = "carlos.mendes@shell.com"
)

type sourceAnswer struct {
	Items []struct {
		Key    string `json:"key"`
		Fields []struct {
			Key          string   `json:"key"`
			Groupable    bool     `json:"groupable"`
			Aggregations []string `json:"aggregations"`
		} `json:"fields"`
	} `json:"items"`
}

func TestListDataSources_SourcesOfThePerson_FR016(t *testing.T) {
	t.Parallel()
	server := apitest.New(t)
	tests := []struct {
		email string
		want  []string
	}{
		{priyaEmail, []string{"plans", "campaigns", "initiatives", "activities", "lineItems", "plannedByMonth",
			"financialTransactions", "approvals", "purchaseRequests", "DS-1001"}},
		{carlosEmail, []string{"plans", "campaigns", "initiatives", "activities", "lineItems", "plannedByMonth",
			"financialTransactions", "approvals", "purchaseRequests", "globalBudgets"}},
	}
	for _, tt := range tests {
		t.Run(tt.email, func(t *testing.T) {
			t.Parallel()
			r := server.SignIn(t, tt.email).Do(t, http.MethodGet, dataSources, nil)
			if r.Status != http.StatusOK {
				t.Fatalf("GET %s = %d %s", dataSources, r.Status, r.Body)
			}
			var answer sourceAnswer
			r.Decode(t, &answer)
			keys := []string{}
			for _, s := range answer.Items {
				keys = append(keys, s.Key)
			}
			if diff := cmp.Diff(tt.want, keys); diff != "" {
				t.Errorf("sources (-want +got):\n%s", diff)
			}
		})
	}
}

func TestQueryWidgetData_AnswersTheWidget_FR016(t *testing.T) {
	t.Parallel()
	c := apitest.New(t).SignIn(t, priyaEmail)
	r := sumByType("activities")
	r.Query.Filters = []widgetdata.Filter{{Field: "status", Operator: "ne", Values: []string{"Declined"}}}
	r.Filters = &widgetdata.Filters{PlanIDs: []string{}, TeamIDs: []string{}}

	answer := c.Do(t, http.MethodPost, query, r)

	if answer.Status != http.StatusOK {
		t.Fatalf("POST %s = %d %s", query, answer.Status, answer.Body)
	}
	var data widgetdata.Data
	answer.Decode(t, &data)
	if len(data.Rows) == 0 || data.GroupCount != len(data.Rows) || data.Totals[0].Money == nil {
		t.Errorf("answer = %+v, want activity types with their planned cost", data)
	}
}

func TestQueryWidgetData_Refusals_FR019(t *testing.T) {
	t.Parallel()
	c := apitest.New(t).SignIn(t, priyaEmail)
	invalid := sumByType("activities")
	invalid.Query.Measures[0].Field = ptr("activity")
	tests := []struct {
		name   string
		r      widgetdata.Request
		status int
		code   string
	}{
		{"every problem of the query at once", invalid, http.StatusBadRequest, "validation_failed"},
		{"a source the person may not read", sumByType("globalBudgets"), http.StatusForbidden, "forbidden"},
		{"a dataset the person may not see", sumByType("DS-9999"), http.StatusNotFound, "not_found"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			answer := c.Do(t, http.MethodPost, query, tt.r)
			var envelope struct {
				Error struct {
					Code string `json:"code"`
				} `json:"error"`
			}
			answer.Decode(t, &envelope)
			if answer.Status != tt.status || envelope.Error.Code != tt.code {
				t.Errorf("POST %s = %d %s, want %d %s", query, answer.Status, answer.Body, tt.status, tt.code)
			}
		})
	}
}
