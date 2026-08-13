module order-management-example/app

go 1.18

require (
	github.com/shopspring/decimal v1.4.0
	github.com/teaql/teaql-golang v0.1.1
	order-management-service-core-workspace/lib v0.0.0
)

require github.com/mattn/go-sqlite3 v1.14.16 // indirect

replace order-management-service-core-workspace/lib => ../golang-lib-core
