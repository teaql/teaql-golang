package sqlite

import (
	"github.com/teaql/teaql-golang/core"
	ds "github.com/teaql/teaql-golang/data_service"
	"testing"
)

func TestStreamRejectsUnsupportedWorkBeforeSQL(t *testing.T) {
	for _, shape := range []string{"relation", "relation-aggregate", "object-group", "enhancement", "nil-consumer"} {
		t.Run(shape, func(t *testing.T) {
			e := nativeBatchEnvironment(t)
			query := core.NewSelectQuery("Customer").Limit(2).Comment("stream shape guard").Purpose("reject work that cannot be executed")
			child := core.NewSelectQuery("Customer").Limit(1)
			switch shape {
			case "relation":
				query.Relations = []*core.RelationLoad{core.NewRelationLoadWithQuery("child", child)}
			case "relation-aggregate":
				query.RelationAggregates = []*core.RelationAggregate{core.NewRelationAggregate("child", "n", child, true)}
			case "object-group":
				query.ObjectGroupBys = []*core.ObjectGroupBy{core.NewObjectGroupBy("child", "id", child)}
			case "enhancement":
				query.ChildEnhancements = []*core.SelectQuery{child}
			}
			request, err := ds.NewQueryRequest(query)
			if err != nil {
				t.Fatal(err)
			}
			calls := 0
			consumer := func(*ds.StreamChunk) error { calls++; return nil }
			if shape == "nil-consumer" {
				consumer = nil
			}
			if err := e.executor.QueryStream(e.ctx, request, 1, consumer); err == nil {
				t.Fatal("unsupported stream request was accepted")
			}
			if calls != 0 || len(e.logs.entries) != 0 {
				t.Fatal("invalid stream performed SQL or invoked its consumer")
			}
		})
	}
}
