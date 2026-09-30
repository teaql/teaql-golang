package main

import (
	stdcontext "context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/provider/sqlite"
	"github.com/teaql/teaql-golang/runtime"
	teaqlsql "github.com/teaql/teaql-golang/sql"
)

type orderNumberSlot struct{ value string }

func (s *orderNumberSlot) CurrentBusinessID() string     { return s.value }
func (*orderNumberSlot) IsNewAggregate() bool            { return true }
func (s *orderNumberSlot) AssignBusinessID(value string) { s.value = value }

func main() {
	directory, err := os.MkdirTemp("", "teaql-business-id-")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(directory)
	db, err := sql.Open("sqlite3", filepath.Join(directory, "business-id.db"))
	if err != nil {
		panic(err)
	}
	defer db.Close()
	transport := sqlite.NewSqliteMutationExecutor(db)
	if err := teaqlsql.EnsureBusinessIDSchema(stdcontext.Background(), transport); err != nil {
		panic(err)
	}
	allocator := teaqlsql.NewOptimisticBusinessIDAllocator(transport, &sqlite.SqliteDialect{})
	keyBytes := make([]byte, 32)
	for index := range keyBytes {
		keyBytes[index] = byte(index)
	}
	key, err := core.NewBusinessIDEncodingKey(1, keyBytes)
	if err != nil {
		panic(err)
	}
	definition, err := core.NewDailyPermutedBusinessIDDefinition("order_number", "ORD", "order_number")
	if err != nil {
		panic(err)
	}
	context := runtime.NewUserContext().
		WithBusinessClock(runtime.NewFixedBusinessClock(time.Date(2026, 10, 1, 8, 30, 0, 0, time.UTC))).
		WithBusinessIDKeyProvider(runtime.NewStaticBusinessIDKeyProvider(key)).
		WithBusinessIDProfileFactory(runtime.DefaultBusinessIDProfileFactory{}).
		WithBusinessIDService(runtime.NewDefaultBusinessIDService(allocator))
	slot := &orderNumberSlot{}
	first, err := context.EnsureBusinessID(definition, "commerce", "commerce_order", slot)
	if err != nil {
		panic(err)
	}
	retry, err := context.EnsureBusinessID(definition, "commerce", "commerce_order", slot)
	if err != nil || first != retry {
		panic("Business ID retry was not idempotent")
	}
	fmt.Println("PASS Go governed Business ID lifecycle example", first.Value)
}
