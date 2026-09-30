package sql

import (
	stdcontext "context"
	"fmt"
	"time"

	"github.com/teaql/teaql-golang/core"
)

const BusinessIDSchemaSQL = "CREATE TABLE IF NOT EXISTS teaql_business_id_space (scope_key VARCHAR(512) NOT NULL PRIMARY KEY, current_value BIGINT NOT NULL, version BIGINT NOT NULL, updated_at BIGINT NOT NULL)"

// EnsureBusinessIDSchema is an explicit schema-boundary operation. Constructing
// an allocator never executes DDL.
func EnsureBusinessIDSchema(context stdcontext.Context, transport SqlTransport) error {
	if transport == nil {
		return fmt.Errorf("business ID SQL transport must not be nil")
	}
	_, err := transport.ExecuteSql(context, &CompiledQuery{Sql: BusinessIDSchemaSQL})
	if err != nil {
		return fmt.Errorf("ensure Business ID space: %w", err)
	}
	return nil
}

// OptimisticBusinessIDAllocator is a portable CAS allocator over one SQL transport.
type OptimisticBusinessIDAllocator struct {
	Transport SqlTransport
	Dialect   SqlDialect
}

func NewOptimisticBusinessIDAllocator(transport SqlTransport, dialect SqlDialect) *OptimisticBusinessIDAllocator {
	if transport == nil || dialect == nil {
		panic("business ID transport and dialect must not be nil")
	}
	return &OptimisticBusinessIDAllocator{Transport: transport, Dialect: dialect}
}

func (a *OptimisticBusinessIDAllocator) AllocateBusinessID(context stdcontext.Context, plan core.BusinessIDPlan) (core.BusinessIDAllocation, error) {
	scopeKey := plan.Scope().CanonicalKey()
	updatedAt := uint64(plan.BusinessDate().UTC().UnixMilli())
	for attempt := 1; attempt <= MaxIdAllocationAttempts; attempt++ {
		rows, err := a.Transport.FetchAllSql(context, &CompiledQuery{
			Sql:    fmt.Sprintf("SELECT current_value, version FROM teaql_business_id_space WHERE scope_key = %s", a.Dialect.Placeholder(1)),
			Params: []core.Value{core.ValText(scopeKey)},
		})
		if err != nil {
			return core.BusinessIDAllocation{}, fmt.Errorf("Business ID allocation requires explicit EnsureBusinessIDSchema: %w", err)
		}
		if len(rows) == 0 {
			changed, insertErr := a.Transport.ExecuteSql(context, &CompiledQuery{
				Sql:    fmt.Sprintf("INSERT INTO teaql_business_id_space(scope_key, current_value, version, updated_at) VALUES (%s, %s, 1, %s)", a.Dialect.Placeholder(1), a.Dialect.Placeholder(2), a.Dialect.Placeholder(3)),
				Params: []core.Value{core.ValText(scopeKey), core.ValU64(plan.InitialSequence()), core.ValU64(updatedAt)},
			})
			if insertErr == nil {
				if changed == 1 {
					return core.BusinessIDAllocation{Scope: plan.Scope(), Sequence: plan.InitialSequence()}, nil
				}
				return core.BusinessIDAllocation{}, fmt.Errorf("Business ID insert for %s changed %d rows", scopeKey, changed)
			}
			winner, winnerErr := a.Transport.FetchAllSql(context, &CompiledQuery{
				Sql:    fmt.Sprintf("SELECT current_value, version FROM teaql_business_id_space WHERE scope_key = %s", a.Dialect.Placeholder(1)),
				Params: []core.Value{core.ValText(scopeKey)},
			})
			if winnerErr != nil || len(winner) == 0 {
				return core.BusinessIDAllocation{}, fmt.Errorf("insert Business ID space for %s: %w", scopeKey, insertErr)
			}
		} else {
			current, currentOK := rows[0]["current_value"].TryU64()
			version, versionOK := rows[0]["version"].TryU64()
			if !currentOK || !versionOK || version == 0 {
				return core.BusinessIDAllocation{}, fmt.Errorf("invalid Business ID sequence row for %s", scopeKey)
			}
			if current >= plan.MaximumSequence() {
				return core.BusinessIDAllocation{}, &core.BusinessIDError{Code: core.BusinessIDRangeExhausted, Message: "Business ID range exhausted for " + scopeKey}
			}
			next := current + 1
			changed, updateErr := a.Transport.ExecuteSql(context, &CompiledQuery{
				Sql:    fmt.Sprintf("UPDATE teaql_business_id_space SET current_value = %s, version = version + 1, updated_at = %s WHERE scope_key = %s AND version = %s AND current_value = %s", a.Dialect.Placeholder(1), a.Dialect.Placeholder(2), a.Dialect.Placeholder(3), a.Dialect.Placeholder(4), a.Dialect.Placeholder(5)),
				Params: []core.Value{core.ValU64(next), core.ValU64(updatedAt), core.ValText(scopeKey), core.ValU64(version), core.ValU64(current)},
			})
			if updateErr != nil {
				return core.BusinessIDAllocation{}, updateErr
			}
			if changed == 1 {
				return core.BusinessIDAllocation{Scope: plan.Scope(), Sequence: next}, nil
			}
			if changed != 0 {
				return core.BusinessIDAllocation{}, fmt.Errorf("Business ID update for %s changed %d rows", scopeKey, changed)
			}
		}
		time.Sleep(time.Millisecond)
	}
	return core.BusinessIDAllocation{}, &core.BusinessIDError{
		Code:    core.BusinessIDAllocationRetryExhausted,
		Message: fmt.Sprintf("Business ID allocation did not converge for %s", scopeKey),
	}
}
