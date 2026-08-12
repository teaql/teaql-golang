package runtime

import (
	"context"
	"fmt"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
)

type RuntimeDataService struct {
	metadata MetadataStore
	executor data_service.DataServiceExecutor
}

func NewRuntimeDataService(metadata MetadataStore, executor data_service.DataServiceExecutor) *RuntimeDataService {
	return &RuntimeDataService{
		metadata: metadata,
		executor: executor,
	}
}

func (s *RuntimeDataService) FetchAll(ctx context.Context, query *core.SelectQuery) ([]core.Record, error) {
	rows, err := s.fetchRows(ctx, query)
	if err != nil {
		return nil, err
	}
	if err := s.enhanceRelations(ctx, rows, query); err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *RuntimeDataService) fetchRows(ctx context.Context, query *core.SelectQuery) ([]core.Record, error) {
	qExec, ok := s.executor.(data_service.QueryExecutor)
	if !ok {
		return nil, fmt.Errorf("executor does not support Query")
	}

	req := &data_service.QueryRequest{
		Query:      query,
		TraceChain: query.TraceChain,
		Comment:    query.CommentText,
	}

	res, err := qExec.Query(ctx, req)
	if err != nil {
		return nil, &DataServiceError{Type: "Executor", ExecutorError: err}
	}

	return res.Rows, nil
}

func (s *RuntimeDataService) enhanceRelations(ctx context.Context, parents []core.Record, query *core.SelectQuery) error {
	if len(parents) == 0 || len(query.Relations) == 0 {
		return nil
	}
	parentDescriptor := s.metadata.Entity(query.Entity)
	if parentDescriptor == nil {
		return fmt.Errorf("unknown entity %s", query.Entity)
	}
	for _, load := range query.Relations {
		relation := parentDescriptor.RelationByName(load.Name)
		if relation == nil {
			return fmt.Errorf("missing relation %s.%s", query.Entity, load.Name)
		}
		ids := make([]core.Value, 0, len(parents))
		for _, parent := range parents {
			if id, ok := parent[relation.LocKey]; ok {
				ids = append(ids, id)
			}
		}
		childQuery := cloneSelectQuery(load.Query, relation.TargetEntity)
		ensureProjection(childQuery, relation.ForKey)
		childQuery.AndFilter(core.ExprInList(relation.ForKey, ids))
		if childQuery.Slice != nil {
			childQuery.PartitionByField(relation.ForKey)
		}
		children, err := s.fetchRows(ctx, childQuery)
		if err != nil {
			return err
		}
		for _, child := range children {
			delete(child, "__teaql_partition_rank")
		}
		if err := s.enhanceRelations(ctx, children, childQuery); err != nil {
			return err
		}
		attachRelationRows(parents, children, load.Name, relation)
	}
	return nil
}

func cloneSelectQuery(source *core.SelectQuery, entity string) *core.SelectQuery {
	if source == nil {
		return core.NewSelectQuery(entity)
	}
	clone := *source
	clone.Entity = entity
	clone.Projection = append([]string(nil), source.Projection...)
	clone.OrderBy = append([]*core.OrderBy(nil), source.OrderBy...)
	clone.Relations = append([]*core.RelationLoad(nil), source.Relations...)
	if source.Slice != nil {
		slice := *source.Slice
		clone.Slice = &slice
	}
	return &clone
}

func ensureProjection(query *core.SelectQuery, field string) {
	// An empty projection means SELECT all entity properties. Appending only the
	// relation foreign key would accidentally turn it into a narrow projection
	// and discard child id and business fields during relation loading.
	if len(query.Projection) == 0 {
		return
	}
	for _, selected := range query.Projection {
		if selected == field {
			return
		}
	}
	query.Projection = append(query.Projection, field)
}

func relationKey(value core.Value) string {
	return fmt.Sprintf("%T:%v", value.V, value.V)
}

func attachRelationRows(parents, children []core.Record, name string, relation *core.RelationDescriptor) {
	buckets := make(map[string][]core.Record)
	for _, child := range children {
		if foreignKey, ok := child[relation.ForKey]; ok {
			key := relationKey(foreignKey)
			buckets[key] = append(buckets[key], child)
		}
	}
	for _, parent := range parents {
		localKey, ok := parent[relation.LocKey]
		if !ok {
			continue
		}
		related := buckets[relationKey(localKey)]
		if relation.IsMany {
			parent[name] = core.Value{V: related}
		} else if len(related) > 0 {
			parent[name] = core.Value{V: related[0]}
		} else {
			parent[name] = core.ValNull()
		}
	}
}

// TODO: fetch entities etc
