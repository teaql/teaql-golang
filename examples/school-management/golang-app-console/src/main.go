package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
	src "school-management-service-core-workspace/lib"
)

func main() {
	if err := showcase(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func showcase() error {
	database := os.Getenv("SCHOOL_MANAGEMENT_SERVICE_CORE_DATABASE_URL")
	if database == "" {
		database = ".local/school.sqlite"
		if err := os.Setenv("SCHOOL_MANAGEMENT_SERVICE_CORE_DATABASE_URL", database); err != nil {
			return err
		}
	}
	if err := os.MkdirAll(filepath.Dir(database), 0755); err != nil {
		return err
	}
	context, err := src.ServiceRuntimeFromEnv()
	if err != nil {
		return err
	}
	defer context.GetResource("db").(*sql.DB).Close()
	for round := 0; round < 2; round++ {
		if err := src.EnsureSchema(context); err != nil {
			return err
		}
	}
	roots, err := src.Q.Platforms().Limit(10).Comment("Read the School platform").Purpose("Verify School startup").ExecuteForList(context)
	if err != nil {
		return err
	}
	kinds, err := src.Q.SchoolTypes().OrderByIdAsc().Limit(10).Comment("Read the School types").Purpose("Verify repeatable School startup").ExecuteForList(context)
	if err != nil {
		return err
	}
	if len(roots.Data) != 1 || roots.Data[0].Id() != 1 || len(kinds.Data) != 2 || kinds.Data[0].Id() != 1001 || kinds.Data[1].Id() != 1002 {
		return fmt.Errorf("School startup did not retain the expected platform and two types")
	}
	const name = "Showcase Riverside Primary School"
	before, err := src.Q.Schools().WithNameIs(name).Limit(2).Comment("Find the showcase School").Purpose("Reuse the same School on repeated startup").ExecuteForList(context)
	if err != nil {
		return err
	}
	created := 0
	if len(before.Data) == 0 {
		entity, err := src.Q.Schools().Comment("Initialize Riverside School").Purpose("Show an audited School create after bootstrap").NewEntity(context)
		if err != nil {
			return err
		}
		entity.UpdatePlatformId(1)
		entity.UpdateSchoolTypeToPrimary()
		entity.UpdateName(name)
		entity.UpdateAddress("12 River Road, Springfield")
		entity.UpdateEstablishedDate(time.Date(1995, 9, 1, 0, 0, 0, 0, time.UTC))
		entity.UpdateStudentCapacity(800)
		entity.UpdateActive(true)
		if _, err := entity.AuditAs("Register the School startup showcase").Save(context); err != nil {
			return err
		}
		created = 1
	}
	rows, err := src.Q.Schools().WithNameIs(name).Limit(2).Comment("Read the persisted showcase School").Purpose("Confirm the School business result").ExecuteForList(context)
	if err != nil {
		return err
	}
	if len(rows.Data) != 1 || rows.Data[0].Name() != name || rows.Data[0].Version() != 1 {
		return fmt.Errorf("School create/query did not return exactly one unchanged School")
	}
	fmt.Printf("School startup: one platform, two types; %s is registered (created=%d).\n", rows.Data[0].Name(), created)
	result, err := json.Marshal(map[string]interface{}{
		"roots": len(roots.Data), "types": len(kinds.Data), "schools": len(rows.Data),
		"created": created, "schoolId": rows.Data[0].Id(), "version": rows.Data[0].Version(),
	})
	if err != nil {
		return err
	}
	fmt.Println("SHOWCASE_RESULT " + string(result))
	return nil
}
