package main

import (
	"fmt"
	"os"

	"runtime-example-conformance-service-core-workspace/lib"
)

func main() {
	if os.Getenv("RUNTIME_EXAMPLE_CONFORMANCE_SERVICE_CORE_DATABASE_URL") == "" {
		os.Setenv("RUNTIME_EXAMPLE_CONFORMANCE_SERVICE_CORE_DATABASE_URL", "./runtime-example.db")
	}
	context, err := lib.ServiceRuntimeFromEnv()
	if err != nil {
		panic(err)
	}
	if err := lib.EnsureSchema(context); err != nil {
		panic(err)
	}

	platform := lib.Q.Platforms().Comment("create the example root").Purpose("initialize retained example data").NewEntity(context)
	platform.UpdateName("Runtime Example")
	platform, err = platform.AuditAs("create the example root").Save(context)
	if err != nil {
		panic(err)
	}

	item := lib.Q.WorkItems().Comment("create a conformance work item").Purpose("verify generated mutation API").NewEntity(context)
	item.UpdateTitle("verify Go runtime")
	item.UpdatePlatformId(platform.Id())
	item, err = item.AuditAs("create a conformance work item").Save(context)
	if err != nil {
		panic(err)
	}

	items, err := lib.Q.WorkItems().OrderByIdAsc().Comment("load retained example rows").Purpose("verify typed Q and SmartList").ExecuteForList(context)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Go retained conformance example passed: %d item(s)\n", len(items.Data))
}
