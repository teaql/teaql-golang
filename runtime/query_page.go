package runtime

import (
	stdcontext "context"
	"fmt"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
)

// ExecuteTypedList is the generated-library bridge, not an application SDK
// hook. One request scope covers rows, explicitly requested count, facets and
// typed hydration. No invocation state is installed on UserContext.
func ExecuteTypedList[T any](context stdcontext.Context, service *RuntimeDataService, query *core.SelectQuery, plan *FacetPlan, hydrate func(core.Record) (T, error)) (result *core.SmartList[T], err error) {
	request, err := data_service.NewQueryRequest(query)
	if err != nil {
		return nil, err
	}
	context, operation := service.startListOperation(context, request.Query, request.Scope().Root())
	defer func() {
		if err != nil {
			operation.Failure(RuntimeErrorType(err))
		} else if result == nil {
			operation.Failure("operation_interrupted")
		} else {
			operation.Success(map[string]RuntimeAttributeValue{"teaql.result.cardinality": len(result.Data)})
		}
	}()
	if hydrate == nil {
		return nil, fmt.Errorf("typed list hydration must not be nil")
	}
	rows, err := service.FetchListWithFacetPlan(context, request.Query, plan)
	if err != nil {
		return nil, err
	}
	entities := make([]T, 0, len(rows.Data))
	for _, row := range rows.Data {
		entity, hydrationError := hydrate(row)
		if hydrationError != nil {
			return nil, hydrationError
		}
		entities = append(entities, entity)
	}
	result = core.NewSmartList(entities)
	result.TotalCount, result.Aggregations, result.Summary, result.Facets = rows.TotalCount, rows.Aggregations, rows.Summary, rows.Facets
	return result, nil
}

// FetchListWithFacetPlan never adds an implicit count. All list variants share
// the same invocation lifecycle; count is optional result metadata, not a
// separate generated terminal.
func (s *RuntimeDataService) FetchListWithFacetPlan(context stdcontext.Context, query *core.SelectQuery, plan *FacetPlan) (result *core.SmartList[core.Record], err error) {
	request, err := data_service.NewQueryRequest(query)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		query, plan, err = CaptureQueryPlan(request.Query, nil)
		if err != nil {
			return nil, err
		}
	} else {
		query = request.Query
	}
	if plan.countOperation != nil {
		return s.FetchPageWithFacetPlan(context, query, plan)
	}
	if !s.ownsListOperation(context, request.Scope().Root()) {
		var operation RuntimeTelemetryScope
		context, operation = s.startListOperation(context, query, request.Scope().Root())
		defer func() {
			if err != nil {
				operation.Failure(RuntimeErrorType(err))
			} else if result == nil {
				operation.Failure("operation_interrupted")
			} else {
				operation.Success(map[string]RuntimeAttributeValue{"teaql.result.cardinality": len(result.Data)})
			}
		}()
	}
	rows, err := s.FetchAllWithFacetPlan(context, query, plan)
	if err != nil {
		return nil, err
	}
	result = core.NewSmartList(rows)
	result.Facets, err = plan.Execute(context, s, query)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// FetchPageWithFacetPlan consumes one authorized invocation snapshot. The
// generated layer supplies native operation metadata and hydrates typed rows;
// exact-count ownership, cached-count reuse and facet execution live here.
func (s *RuntimeDataService) FetchPageWithFacetPlan(context stdcontext.Context, query *core.SelectQuery, plan *FacetPlan) (result *core.SmartList[core.Record], err error) {
	request, err := data_service.NewQueryRequest(query)
	if err != nil {
		return nil, err
	}
	if plan == nil {
		query, plan, err = CaptureQueryPlan(request.Query, nil)
		if err != nil {
			return nil, err
		}
	} else {
		query = request.Query
	}
	query = plan.WithQueryDiagnostics(query)
	if !s.ownsListOperation(context, request.Scope().Root()) {
		var operation RuntimeTelemetryScope
		context, operation = s.startListOperation(context, query, request.Scope().Root())
		defer func() {
			if err != nil {
				operation.Failure(RuntimeErrorType(err))
			} else if result == nil {
				operation.Failure("operation_interrupted")
			} else {
				operation.Success(map[string]RuntimeAttributeValue{"teaql.result.cardinality": len(result.Data)})
			}
		}()
	}
	alias := "__teaql_total"
	if plan.countOperation != nil && plan.countOperation.Alias != nil {
		alias = *plan.countOperation.Alias
	}
	rowsQuery := query.Clone()
	if plan.countOperation != nil {
		rowsQuery.Aggregates = nil
		for _, aggregate := range query.Aggregates {
			if aggregate != nil && !(aggregate.Function == core.AggCount && aggregate.Alias == alias) {
				rowsQuery.Aggregates = append(rowsQuery.Aggregates, aggregate)
			}
		}
	}
	if err := rowsQuery.PrepareForList(); err != nil {
		return nil, err
	}
	var rows []core.Record
	var total uint64
	count := func() error {
		var scope *core.QueryRequestScope
		var scopeErr error
		if plan.countOperation != nil {
			scope, scopeErr = request.Scope().Operation(*plan.countOperation)
		} else {
			scope, scopeErr = request.Scope().Helper(query.Entity, "count matching rows")
		}
		if scopeErr != nil {
			return scopeErr
		}
		countQuery, bindErr := query.ForExactCount(alias).WithQueryRequestScope(scope)
		if bindErr != nil {
			return bindErr
		}
		countRows, countErr := s.FetchAll(context, countQuery)
		if countErr != nil {
			return countErr
		}
		if len(countRows) != 1 {
			return fmt.Errorf("exact count returned %d rows", len(countRows))
		}
		var ok bool
		total, ok = countRows[0][alias].TryU64()
		if !ok {
			return fmt.Errorf("exact count did not return an unsigned integer")
		}
		return nil
	}
	if rowsQuery.IDSetPagination != nil {
		rows, err = s.FetchAllWithFacetPlan(context, rowsQuery, plan)
		if err != nil {
			return nil, err
		}
		userContext, ok := UserContextFrom(context)
		if ok {
			retained, accuracy := userContext.IDSetCount()
			if accuracy == "EXACT" {
				total = retained
			} else {
				err = count()
			}
		} else {
			err = count()
		}
	} else {
		err = count()
		if err == nil {
			rows, err = s.FetchAllWithFacetPlan(context, rowsQuery, plan)
		}
	}
	if err != nil {
		return nil, err
	}
	result = core.NewSmartList(rows).WithTotalCount(total)
	if plan.countOperation != nil && plan.countOperation.Alias != nil {
		result.WithAggregation(alias, core.ValU64(total))
	}
	result.Facets, err = plan.Execute(context, s, rowsQuery)
	if err != nil {
		return nil, err
	}
	return result, nil
}
