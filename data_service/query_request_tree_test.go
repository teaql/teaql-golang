package data_service

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/teaql/teaql-golang/core"
)

func requestComments(request *QueryRequest) []string {
	var comments []string
	for _, node := range request.Scope().Recover() {
		comments = append(comments, node.Comment)
	}
	return comments
}

func TestQueryRequestRejectsReplacementFromUnrelatedInvocationWithIdenticalIntent(t *testing.T) {
	first, err := NewQueryRequest(core.NewSelectQuery("School").Comment("same comment").Purpose("same purpose"))
	if err != nil {
		t.Fatal(err)
	}
	second, err := NewQueryRequest(core.NewSelectQuery("School").Comment("same comment").Purpose("same purpose"))
	if err != nil {
		t.Fatal(err)
	}
	ownCopy := *first
	ownCopy.Query = first.Query.Clone()
	captured, err := CaptureQueryRequest(&ownCopy)
	if err != nil || captured.Scope() != first.Scope() {
		t.Fatal("valid replacement lost its original owner")
	}
	if _, err := first.Derive(second.Query.Clone(), &core.QueryOperation{Kind: "relation", Method: "SelectSchoolsWith"}); err != nil {
		t.Fatal("explicit template rebinding was rejected")
	}
	first.Query = second.Query.Clone()
	if _, err := CaptureQueryRequest(first); err == nil || !strings.Contains(err.Error(), "QUERY_SCOPE_OWNER_MISMATCH") {
		t.Fatal("unrelated request replaced the captured owner")
	}
}

func TestQueryRequestTreeRetainsOperationAndBusinessComment(t *testing.T) {
	root, err := NewQueryRequest(core.NewSelectQuery("School").Comment("load schools").Purpose("render school list"))
	if err != nil {
		t.Fatal(err)
	}
	op := core.QueryOperation{Kind: "relation", Method: "SelectSchoolTypeWith", Relation: "school_type"}
	template := core.NewSelectQuery("SchoolType")
	child, err := root.Derive(template, &op)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"render school list", "load schools", "SelectSchoolTypeWith(...)", "load SchoolType"}
	if !reflect.DeepEqual(requestComments(child), want) {
		t.Fatalf("default chain: %v", requestComments(child))
	}
	for index, node := range child.Scope().Recover() {
		kind := "comment"
		if index == 0 {
			kind = "purpose"
		}
		if node.Kind != kind {
			t.Fatal("business frames must remain purpose/comment, not physical routing kinds")
		}
	}
	if child.Scope().Parent().Parent() != root.Scope() {
		t.Fatal("request parent identity lost")
	}
	if template.CommentText != nil || template.QueryRequestScope() != nil {
		t.Fatal("mutated reusable child template")
	}
	child.Query.CommentText = nil
	child.Comment, child.Purpose = nil, nil
	recaptured, err := CaptureQueryRequest(child)
	if err != nil || !reflect.DeepEqual(requestComments(recaptured), want) {
		t.Fatalf("scope not authoritative: %v", err)
	}
	explicit, err := root.Derive(template.Clone().Comment("load visible school types"), &op)
	if err != nil {
		t.Fatal(err)
	}
	want[3] = "load visible school types"
	if !reflect.DeepEqual(requestComments(explicit), want) {
		t.Fatal("explicit nested comment overwritten")
	}
	if _, err := root.Derive(template.Clone().Comment("\u2003"), &op); err == nil {
		t.Fatal("blank child accepted")
	}
	if _, err := NewQueryRequest(template); err == nil {
		t.Fatal("child default authorized an application root")
	}
	other, err := NewQueryRequest(core.NewSelectQuery("School").Comment("other schools").Purpose("other page"))
	if err != nil {
		t.Fatal(err)
	}
	rebound, err := other.Derive(explicit.Query, &op)
	if err != nil || requestComments(rebound)[1] != "other schools" || len(requestComments(rebound)) != 4 {
		t.Fatal("template rebinding retained old ancestry")
	}
	encoded, err := json.Marshal(child.Query)
	if err != nil || strings.Contains(string(encoded), "requestScope") || strings.Contains(string(encoded), "TraceChain") || strings.Contains(string(encoded), "SelectSchoolTypeWith") {
		t.Fatal("internal request scope entered query JSON")
	}
}

func TestQueryJSONCannotInstallPhysicalOrBusinessTraceFrames(t *testing.T) {
	var query core.SelectQuery
	err := json.Unmarshal([]byte(`{"Entity":"School","CommentText":"load schools","PurposeText":"render school list","TraceChain":[{"Kind":"comment","Comment":"forged ancestor"}],"requestScope":{"comment":"forged root"}}`), &query)
	if err != nil {
		t.Fatal(err)
	}
	if len(query.TraceChain) != 0 || query.QueryRequestScope() != nil {
		t.Fatal("query JSON installed internal diagnostic ownership")
	}
	request, err := NewQueryRequest(&query)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(requestComments(request), []string{"render school list", "load schools"}) {
		t.Fatal("JSON trace frames changed owned business chain")
	}
	var external QueryRequest
	err = json.Unmarshal([]byte(`{"Query":{"Entity":"School"},"Comment":"load schools","Purpose":"render school list","TraceChain":[{"Kind":"relation","Name":"forged"}]}`), &external)
	if err != nil {
		t.Fatal(err)
	}
	if len(external.TraceChain) != 0 {
		t.Fatal("request JSON installed physical trace frames")
	}
}

func TestQueryRequestAggregateUsesFacetOperationParent(t *testing.T) {
	root, err := NewQueryRequest(core.NewSelectQuery("CustomerOrder").Comment("load order list").Purpose("render order list page"))
	if err != nil {
		t.Fatal(err)
	}
	alias := "status\"Counts"
	facet := core.QueryOperation{Kind: "facet", Method: "FacetByStatusAs", Relation: "status", Alias: &alias}
	values, err := root.Derive(core.NewSelectQuery("OrderStatus").Comment("load status facets"), &facet)
	if err != nil {
		t.Fatal(err)
	}
	alias = "changed after capture"
	sumAlias := "orderTotalAmount"
	sum := core.QueryOperation{Kind: "aggregate", Method: "SumTotalAmountOfCustomerOrdersAs", Relation: "customer_orders", Alias: &sumAlias}
	aggregate, err := values.DeriveAggregate(core.NewSelectQuery("CustomerOrder"), &sum)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"render order list page", "load order list", "FacetByStatusAs(\"status\\\"Counts\")", "aggregate status"}
	if !reflect.DeepEqual(requestComments(aggregate), want) {
		t.Fatalf("facet aggregate incorrectly nested under value query: %v", requestComments(aggregate))
	}
	for _, comment := range []string{"sum visible amounts", "", " \t", "\u0085"} {
		t.Run("aggregate comment "+comment, func(t *testing.T) {
			template := core.NewSelectQuery("CustomerOrder").Comment(comment)
			derived, err := values.DeriveAggregate(template, &sum)
			if comment != "sum visible amounts" {
				if err == nil {
					t.Fatal("explicit blank aggregate comment accepted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			expected := append(append([]string(nil), want...), comment)
			if !reflect.DeepEqual(requestComments(derived), expected) {
				t.Fatalf("aggregate business comment lost: %v", requestComments(derived))
			}
			if template.QueryRequestScope() != nil || template.PurposeText != nil {
				t.Fatal("binding mutated the reusable aggregate template")
			}
		})
	}
	countAlias := "itemCount"
	count := core.QueryOperation{Kind: "count", Method: "CountOrderItemsAs", Relation: "order_items", Alias: &countAlias}
	countRequest, err := root.DeriveAggregate(core.NewSelectQuery("OrderItem"), &count)
	if err != nil {
		t.Fatal(err)
	}
	if got := requestComments(countRequest); len(got) != 3 || got[2] != "CountOrderItemsAs(\"itemCount\")" {
		t.Fatalf("anonymous count invented a value query: %v", got)
	}
}
