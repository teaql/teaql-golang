module runtime-example-conformance-service-core-workspace

go 1.18

require (
	github.com/teaql/teaql-golang v0.2.1
	runtime-example-conformance-service-core-workspace/lib v0.0.0
)

require (
	github.com/mattn/go-sqlite3 v1.14.16 // indirect
	github.com/shopspring/decimal v1.4.0 // indirect
)

replace runtime-example-conformance-service-core-workspace/lib => ./lib

replace github.com/teaql/teaql-golang => ../..
