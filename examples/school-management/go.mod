module school-management-service-core-workspace/lib

go 1.18

require (
	github.com/mattn/go-sqlite3 v1.14.16
	github.com/shopspring/decimal v1.4.0
	github.com/teaql/teaql-golang v0.2.1
)

// Runtime-owned examples always verify the current checkout. Published-package
// verification is performed from the external conformance workspace.
replace github.com/teaql/teaql-golang => ../..
