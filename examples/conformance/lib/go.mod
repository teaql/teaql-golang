module runtime-example-conformance-service-core-workspace/lib

go 1.18

require (
	github.com/mattn/go-sqlite3 v1.14.16
	github.com/shopspring/decimal v1.4.0
	github.com/teaql/teaql-golang v0.1.10
)

replace github.com/teaql/teaql-golang => ../../..
