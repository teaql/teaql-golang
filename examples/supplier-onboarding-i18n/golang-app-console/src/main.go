package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/teaql/teaql-golang/runtime"
	domain "supplier-onboarding-service-core-workspace/lib"
)

type registration struct {
	Name                 *string
	MinimumOrderQuantity int64 `json:"minimum_order_quantity"`
}

type step struct {
	Operation string
	Context   string
	Locale    string
	Path      string
	Case      string
	Input     registration
}

type scenario struct {
	Database  string
	Warehouse struct{ ID uint64 }
	Steps     []step
}

func emit(step step, context *runtime.UserContext, event string, details map[string]any) error {
	details["type"] = event
	details["context"] = step.Context
	details["effectiveLocale"] = context.Language()
	data, err := json.Marshal(details)
	if err != nil {
		return err
	}
	fmt.Printf("TEAQL_I18N_SHOWCASE %s\n", data)
	return nil
}

func verifyFailure(step step, context *runtime.UserContext, warehouseID uint64) error {
	supplier := domain.Q.Suppliers().Comment("Prepare supplier registration").
		Purpose("Register a supplier for warehouse purchasing").NewEntity(context)
	supplier.UpdateWarehouseId(warehouseID)
	if step.Input.Name != nil {
		supplier.UpdateName(*step.Input.Name)
	}
	supplier.UpdateMinimumOrderQuantity(step.Input.MinimumOrderQuantity)
	_, err := supplier.AuditAs("Register the supplier for purchasing").Save(context)
	if err == nil {
		return fmt.Errorf("invalid supplier registration was stored: %s", step.Case)
	}
	var failure *runtime.RuntimeError
	if !errors.As(err, &failure) || failure.Type != "Check" || len(failure.CheckResults) != 1 {
		return err
	}
	violation := failure.CheckResults[0]
	rule := strings.ToLower(violation.RuleID)
	arguments := map[string]int{}
	switch rule {
	case "required":
	case "min":
		minimum, parseErr := strconv.Atoi(fmt.Sprint(violation.SystemValue))
		if parseErr != nil {
			return fmt.Errorf("read native bound: %w", parseErr)
		}
		arguments["minimum"] = minimum
	default:
		return err
	}
	return emit(step, context, "validation", map[string]any{
		"case": step.Case, "messageKey": "checker." + rule, "nativeRule": violation.RuleID,
		"location": violation.ModelPath(), "message": violation.Message, "arguments": arguments,
		"nativeArguments": map[string]any{"inputValue": violation.InputValue, "systemValue": violation.SystemValue},
	})
}

func run() error {
	var plan scenario
	if err := json.NewDecoder(os.Stdin).Decode(&plan); err != nil {
		return err
	}
	if err := os.Setenv("SUPPLIER_ONBOARDING_SERVICE_CORE_DATABASE_URL", plan.Database); err != nil {
		return err
	}
	contexts := make(map[string]*runtime.UserContext)
	for _, name := range []string{"main", "independent"} {
		context, err := domain.ServiceRuntimeFromEnv()
		if err != nil {
			return err
		}
		contexts[name] = context
	}
	if err := domain.EnsureSchema(contexts["main"]); err != nil {
		return err
	}
	for _, step := range plan.Steps {
		context := contexts[step.Context]
		if context == nil {
			return fmt.Errorf("unknown example context %q", step.Context)
		}
		if err := runStep(step, context, plan.Warehouse.ID); err != nil {
			return err
		}
	}
	return nil
}

func runStep(step step, context *runtime.UserContext, warehouseID uint64) error {
	switch step.Operation {
	case "select":
		result := map[string]any{"requestedLocale": step.Locale, "accepted": true}
		if err := context.SetLocaleCode(step.Locale); err != nil {
			var unsupported *runtime.UnsupportedLocaleError
			if !errors.As(err, &unsupported) {
				return err
			}
			result["accepted"] = false
			result["nativeError"] = "UnsupportedLocaleError"
		}
		return emit(step, context, "locale", result)
	case "validate":
		return verifyFailure(step, context, warehouseID)
	case "install-catalog":
		data, err := os.ReadFile(step.Path)
		if err != nil {
			return err
		}
		catalog, err := runtime.ParseI18nCatalog(data, runtime.BuiltinI18nCatalog())
		if err != nil {
			return err
		}
		if err := context.InstallI18nCatalog(catalog); err != nil {
			return err
		}
		return emit(step, context, "catalog-installed", map[string]any{})
	case "verify-unchanged":
		rows, err := domain.Q.SuppliersMinimal().Limit(1).Comment("Check the rejected registrations").
			Purpose("Verify no invalid supplier was stored").ExecuteForList(context)
		if err != nil {
			return err
		}
		return emit(step, context, "storage", map[string]any{"supplierCount": rows.Len()})
	default:
		return fmt.Errorf("unknown example operation %q", step.Operation)
	}
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
