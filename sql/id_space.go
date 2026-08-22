package sql

import (
	stdcontext "context"
	"fmt"
	"math"

	"github.com/teaql/teaql-golang/core"
)

const MaxIdAllocationAttempts = 100

// NextOptimisticId implements the portable TeaQL ID-space contract. Providers
// may optimize internally only when they preserve these observable semantics.
func NextOptimisticId(context stdcontext.Context, transport SqlTransport, dialect SqlDialect, entity string) (uint64, error) {
	if entity == "" {
		return 0, fmt.Errorf("ID space type name must not be empty")
	}
	_, err := transport.ExecuteSql(context, &CompiledQuery{Sql: "CREATE TABLE IF NOT EXISTS teaql_id_space (type_name VARCHAR(255) NOT NULL PRIMARY KEY, current_level BIGINT NOT NULL)"})
	if err != nil {
		return 0, fmt.Errorf("ensure ID space for %s: %w", entity, err)
	}
	for attempt := 1; attempt <= MaxIdAllocationAttempts; attempt++ {
		rows, readErr := transport.FetchAllSql(context, &CompiledQuery{
			Sql:    fmt.Sprintf("SELECT current_level FROM teaql_id_space WHERE type_name = %s", dialect.Placeholder(1)),
			Params: []core.Value{core.ValText(entity)},
		})
		if readErr != nil {
			return 0, fmt.Errorf("read ID space for %s on attempt %d: %w", entity, attempt, readErr)
		}
		if len(rows) == 0 {
			changed, insertErr := transport.ExecuteSql(context, &CompiledQuery{
				Sql:    fmt.Sprintf("INSERT INTO teaql_id_space(type_name, current_level) VALUES (%s, 1)", dialect.Placeholder(1)),
				Params: []core.Value{core.ValText(entity)},
			})
			if insertErr == nil {
				if changed == 1 {
					return 1, nil
				}
				return 0, fmt.Errorf("ID space insert for %s changed %d rows", entity, changed)
			}
			// A competing instance may have inserted the primary-key row.
			winner, winnerErr := transport.FetchAllSql(context, &CompiledQuery{
				Sql:    fmt.Sprintf("SELECT current_level FROM teaql_id_space WHERE type_name = %s", dialect.Placeholder(1)),
				Params: []core.Value{core.ValText(entity)},
			})
			if winnerErr != nil || len(winner) == 0 {
				return 0, fmt.Errorf("insert ID space for %s: %w", entity, insertErr)
			}
			continue
		}
		current, ok := rows[0]["current_level"].TryU64()
		if !ok {
			return 0, fmt.Errorf("ID space current_level for %s is not an unsigned integer", entity)
		}
		if current >= math.MaxInt64 {
			return 0, fmt.Errorf("ID space overflow for %s", entity)
		}
		next := current + 1
		changed, updateErr := transport.ExecuteSql(context, &CompiledQuery{
			Sql:    fmt.Sprintf("UPDATE teaql_id_space SET current_level = %s WHERE type_name = %s AND current_level = %s", dialect.Placeholder(1), dialect.Placeholder(2), dialect.Placeholder(3)),
			Params: []core.Value{core.ValU64(next), core.ValText(entity), core.ValU64(current)},
		})
		if updateErr != nil {
			return 0, fmt.Errorf("update ID space for %s on attempt %d: %w", entity, attempt, updateErr)
		}
		if changed == 1 {
			return next, nil
		}
		if changed != 0 {
			return 0, fmt.Errorf("ID space update for %s changed %d rows on attempt %d", entity, changed, attempt)
		}
	}
	return 0, fmt.Errorf("unable to allocate ID for %s after %d optimistic-lock attempts", entity, MaxIdAllocationAttempts)
}

// EnsureOptimisticIdFloor advances an ID space after model-defined root or
// constant records are bootstrapped. It never moves an existing space back.
func EnsureOptimisticIdFloor(context stdcontext.Context, transport SqlTransport, dialect SqlDialect, entity string, floor uint64) error {
	if floor > math.MaxInt64 {
		return fmt.Errorf("ID space floor %d for %s exceeds BIGINT", floor, entity)
	}
	_, err := transport.ExecuteSql(context, &CompiledQuery{Sql: "CREATE TABLE IF NOT EXISTS teaql_id_space (type_name VARCHAR(255) NOT NULL PRIMARY KEY, current_level BIGINT NOT NULL)"})
	if err != nil {
		return fmt.Errorf("ensure ID space for %s: %w", entity, err)
	}
	for attempt := 1; attempt <= MaxIdAllocationAttempts; attempt++ {
		rows, readErr := transport.FetchAllSql(context, &CompiledQuery{
			Sql:    fmt.Sprintf("SELECT current_level FROM teaql_id_space WHERE type_name = %s", dialect.Placeholder(1)),
			Params: []core.Value{core.ValText(entity)},
		})
		if readErr != nil {
			return readErr
		}
		if len(rows) == 0 {
			changed, insertErr := transport.ExecuteSql(context, &CompiledQuery{
				Sql:    fmt.Sprintf("INSERT INTO teaql_id_space(type_name, current_level) VALUES (%s, %s)", dialect.Placeholder(1), dialect.Placeholder(2)),
				Params: []core.Value{core.ValText(entity), core.ValU64(floor)},
			})
			if insertErr == nil && changed == 1 {
				return nil
			}
			if insertErr != nil {
				winner, _ := transport.FetchAllSql(context, &CompiledQuery{
					Sql:    fmt.Sprintf("SELECT current_level FROM teaql_id_space WHERE type_name = %s", dialect.Placeholder(1)),
					Params: []core.Value{core.ValText(entity)},
				})
				if len(winner) == 0 {
					return insertErr
				}
			}
			continue
		}
		current, ok := rows[0]["current_level"].TryU64()
		if !ok {
			return fmt.Errorf("ID space current_level for %s is invalid", entity)
		}
		if current >= floor {
			return nil
		}
		changed, updateErr := transport.ExecuteSql(context, &CompiledQuery{
			Sql:    fmt.Sprintf("UPDATE teaql_id_space SET current_level = %s WHERE type_name = %s AND current_level = %s", dialect.Placeholder(1), dialect.Placeholder(2), dialect.Placeholder(3)),
			Params: []core.Value{core.ValU64(floor), core.ValText(entity), core.ValU64(current)},
		})
		if updateErr != nil {
			return updateErr
		}
		if changed == 1 {
			return nil
		}
		if changed != 0 {
			return fmt.Errorf("ID space floor update for %s changed %d rows", entity, changed)
		}
	}
	return fmt.Errorf("unable to synchronize ID space floor for %s after %d optimistic-lock attempts", entity, MaxIdAllocationAttempts)
}
