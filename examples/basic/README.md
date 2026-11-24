# Place a first order for tea delivery

The delivery desk first ensures the generated schema and bootstrap data.
It creates one purchase order using the generated Q and audited save APIs,
then reads its actual ID and description. A description longer than 100
characters is rejected at the description field and must leave no order.

- `model.xml` owns the domain model.
- `golang-lib-core/` holds the generated domain library.
- `golang-app-console/src/main.go` runs the business scenario.
- `golang-app-console/calibration/` retains the original runtime probe.

The SDK inventory runs both entry points. The focused Conformance gate runs
the console twice against one fresh SQLite database and independently checks
the persisted records.
