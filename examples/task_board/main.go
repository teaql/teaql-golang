package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	gen "robot-kanban-service-core-workspace/lib"
)

// Generated metadata/bootstrap/checkers remain library-owned; application declares intent.
func run() error {
	database := os.Getenv("TEAQL_TASK_BOARD_DB")
	if database == "" {
		file, err := os.CreateTemp("", "teaql-task-board-*.db")
		if err != nil {
			return err
		}
		database = file.Name()
		if err := file.Close(); err != nil {
			return err
		}
	}
	if err := os.Setenv("ROBOT_KANBAN_SERVICE_CORE_DATABASE_URL", database); err != nil {
		return err
	}
	context, err := gen.ServiceRuntimeFromEnv()
	if err != nil {
		return err
	}
	defer context.GetResource("db").(*sql.DB).Close()
	if err := gen.EnsureSchema(context); err != nil {
		return err
	}
	if err := gen.EnsureSchema(context); err != nil {
		return err
	}

	task := gen.Q.Tasks().Comment("prepare task").Purpose("demonstrate governed task authoring").NewEntity(context)
	task.UpdateName("Build Robot Arm").UpdateStatusToPlanned().UpdatePlatformId(1)
	if _, err := task.AuditAs("create demo robot task").Save(context); err != nil {
		return err
	}
	loaded, err := gen.Q.Tasks().WithIdIs(task.Id()).Limit(1).
		SelectId().SelectName().SelectPlatform().SelectVersion().
		SelectStatusWith(gen.Q.TaskStatuses().Limit(1)).
		Comment("load task and planned status").Purpose("review complete task before editing").ExecuteForOne(context)
	if err != nil {
		return err
	}
	if loaded == nil {
		return fmt.Errorf("created task was not loaded")
	}
	if name, present := gen.E.Task(loaded).Name().Eval(); !present || name != "Build Robot Arm" {
		return fmt.Errorf("typed E name mismatch")
	}
	if status, present := gen.E.Task(loaded).Status().Name().Eval(); !present || status != "Planned" {
		return fmt.Errorf("typed E status relation mismatch")
	}
	loaded.UpdateName("Build Robot Arm V2")
	if _, err := loaded.AuditAs("rename demo robot task").Save(context); err != nil {
		return err
	}
	updated, err := gen.Q.Tasks().WithIdIs(task.Id()).Limit(1).
		Comment("read renamed task").Purpose("verify persisted mutation").ExecuteForOne(context)
	if err != nil {
		return err
	}
	if updated == nil || updated.Name() != "Build Robot Arm V2" {
		return fmt.Errorf("rename not persisted")
	}

	entry := gen.Q.TaskExecutionLogs().Comment("prepare task log").Purpose("record the rename").NewEntity(context)
	entry.UpdateTaskId(task.Id()).UpdateAction("RENAME").UpdateDetail("PRIVATE-TASK-DETAIL")
	if _, err := entry.AuditAs("record PRIVATE-TASK-DETAIL rename evidence").Save(context); err != nil {
		return err
	}
	rows, err := gen.Q.TaskExecutionLogs().WithIdIs(entry.Id()).WithDetailIs("PRIVATE-TASK-DETAIL").Limit(1).
		Comment("read PRIVATE-TASK-DETAIL evidence").Purpose("verify PRIVATE-TASK-DETAIL is persisted").ExecuteForList(context)
	if err != nil {
		return err
	}
	if len(rows.Data) != 1 || rows.Data[0].Detail() != "PRIVATE-TASK-DETAIL" {
		return fmt.Errorf("masking changed the stored detail")
	}
	fmt.Println("PASS task board: governed Q/E/mutation and masked detail")
	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
