module school-management-service-core-workspace/lib

go 1.18

require (
	github.com/mattn/go-sqlite3 v1.14.16
	github.com/shopspring/decimal v1.4.0
	github.com/teaql/teaql-golang v0.1.11
)

// Local runtime verification comes before publication. Remove this replace only
// for the clean published-artifact verification pass.
replace github.com/teaql/teaql-golang => ../..
