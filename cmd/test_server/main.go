package main

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
	teaql_sql "github.com/teaql/teaql-golang/sql"
	"github.com/teaql/teaql-golang/provider/sqlite"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/tfp_endpoint"
)

type DummySchemaProvider struct{}

func (p *DummySchemaProvider) GetEntity(name string) *core.EntityDescriptor {
	d := core.NewEntityDescriptor(name)
	d.TabName = name // Use exact name since we created tables as User, Product, "Order"

	if name == "User" {
		d.Properties = []*core.PropertyDescriptor{
			core.NewPropertyDescriptor("id", core.TypeI64).ColumnName("id").Id(),
			core.NewPropertyDescriptor("username", core.TypeText),
			core.NewPropertyDescriptor("email", core.TypeText),
			core.NewPropertyDescriptor("isActive", core.TypeBool),
			core.NewPropertyDescriptor("createdAt", core.TypeText),
			core.NewPropertyDescriptor("version", core.TypeI64).Version(),
		}
	} else if name == "Product" {
		d.Properties = []*core.PropertyDescriptor{
			core.NewPropertyDescriptor("id", core.TypeI64).ColumnName("id").Id(),
			core.NewPropertyDescriptor("name", core.TypeText),
			core.NewPropertyDescriptor("price", core.TypeI64),
			core.NewPropertyDescriptor("stock", core.TypeI64),
			core.NewPropertyDescriptor("version", core.TypeI64).Version(),
		}
	} else if name == "Order" {
		d.TabName = `"Order"`
		d.Properties = []*core.PropertyDescriptor{
			core.NewPropertyDescriptor("id", core.TypeI64).ColumnName("id").Id(),
			core.NewPropertyDescriptor("customerName", core.TypeText),
			core.NewPropertyDescriptor("totalAmount", core.TypeI64),
			core.NewPropertyDescriptor("status", core.TypeText),
			core.NewPropertyDescriptor("version", core.TypeI64).Version(),
		}
	}

	return d
}

func main() {
	// Initialize in-memory SQLite database
	db, err := sql.Open("sqlite3", "file:testdb?mode=memory&cache=shared")
	if err != nil {
		log.Fatalf("failed to open sqlite: %v", err)
	}
	defer db.Close()

	// Create tables
	_, err = db.Exec(`
		CREATE TABLE User (id INTEGER PRIMARY KEY AUTOINCREMENT, username TEXT, email TEXT, isActive BOOLEAN, createdAt TEXT, version INTEGER DEFAULT 1);
		CREATE TABLE Product (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, price INTEGER, stock INTEGER, version INTEGER DEFAULT 1);
		CREATE TABLE "Order" (id INTEGER PRIMARY KEY AUTOINCREMENT, customerName TEXT, totalAmount INTEGER, status TEXT, version INTEGER DEFAULT 1);
	`)
	if err != nil {
		log.Fatalf("failed to create tables: %v", err)
	}

	// Insert mock data
	db.Exec(`INSERT INTO User (username, email, isActive, createdAt) VALUES ('admin', 'admin@teaql.com', 1, '2026-08-11')`)
	db.Exec(`INSERT INTO User (username, email, isActive, createdAt) VALUES ('johndoe', 'john@example.com', 1, '2026-08-10')`)
	db.Exec(`INSERT INTO Product (name, price, stock) VALUES ('TeaQL Enterprise License', 999, 99)`)
	db.Exec(`INSERT INTO Product (name, price, stock) VALUES ('Cloud Starter Pack', 299, 50)`)
	db.Exec(`INSERT INTO "Order" (customerName, totalAmount, status) VALUES ('John Doe', 1298, 'Completed')`)

	// Setup TeaQL Data Service Executor
	transport := sqlite.NewSqliteMutationExecutor(db)
	executor := teaql_sql.NewSqlDataServiceExecutor(&sqlite.SqliteDialect{}, transport, &DummySchemaProvider{})
	endpoint := tfp_endpoint.NewTfpEndpoint(executor, executor)

	// Setup HTTP Handlers
	mux := http.NewServeMux()

	corsMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
			if r.Method == "OPTIONS" {
				return
			}
			next(w, r)
		}
	}

	mux.HandleFunc("/tfp/query", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		res, err := endpoint.HandleQuery(r.Context(), body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
	}))

	mux.HandleFunc("/tfp/mutate", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		res, err := endpoint.HandleMutation(r.Context(), body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(res)
	}))

	log.Println("Starting TeaQL Test Server on :9090...")
	log.Fatal(http.ListenAndServe(":9090", mux))
}
