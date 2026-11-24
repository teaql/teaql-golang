module customer-onboarding-service-core-workspace

go 1.18

require (
	github.com/teaql/teaql-golang v0.6.0
	customer-onboarding-service-core-workspace/lib v0.0.0
)

replace customer-onboarding-service-core-workspace/lib => ../golang-lib-core

replace github.com/teaql/teaql-golang => ../../..
