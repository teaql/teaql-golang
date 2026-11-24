package main

import (
	"fmt"
	"os"
	src "natural-message-service-core-workspace/lib"
)

func main() {
		// SQLite can use a local default; server providers require an explicit URL.
		if os.Getenv("NATURAL_MESSAGE_SERVICE_CORE_DATABASE_URL") == "" {
			os.Setenv("NATURAL_MESSAGE_SERVICE_CORE_DATABASE_URL", "./teaql.db")
		}
	context, err := src.ServiceRuntimeFromEnv()
	if err != nil {
		panic(err)
	}
	if err := src.EnsureSchema(context); err != nil {
		panic(err)
	}
	fmt.Println("natural-message-service TeaQL console ready")
}
