package runtime

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/teaql/teaql-golang/core"
)

// ContextFixResolver is a trusted application SPI. Arguments are model-derived,
// never executable request code. The default supports now and uniStr.
type ContextFixResolver interface {
	ResolveContextFix(*UserContext, *CheckAndFixInput, string, string, []string) (core.Value, error)
}

func (c *UserContext) captureInputContextFix(input *CheckAndFixInput) func() {
	fields, before := []string{}, make(core.Record)
	if descriptor := c.Entity(input.Entity); descriptor != nil {
		for _, property := range descriptor.Properties {
			if property.CreateFunction == "" && property.UpdateFunction == "" {
				continue
			}
			field, err := core.CanonicalLedgerField(descriptor, property.Name)
			if err != nil {
				continue
			}
			fields = append(fields, field)
			if value, present := input.Values[field]; present {
				before[field] = core.CloneValue(value)
			}
		}
	}
	return func() {
		for _, field := range fields {
			if value, present := before[field]; present {
				if input.Values == nil {
					input.Values = make(core.Record)
				}
				input.Values[field] = core.CloneValue(value)
			} else {
				delete(input.Values, field)
			}
		}
	}
}

func (c *UserContext) contextFixEligible(input *CheckAndFixInput) bool {
	if input.Operation == core.MutationInsert {
		return true
	}
	if input.Operation != core.MutationUpdate {
		return false
	}
	managed := make(map[string]bool)
	if descriptor := c.Entity(input.Entity); descriptor != nil {
		for _, property := range descriptor.Properties {
			if property.IsId || property.IsVersion || property.CreateFunction != "" || property.UpdateFunction != "" {
				field, err := core.CanonicalLedgerField(descriptor, property.Name)
				if err == nil {
					managed[field] = true
				}
			}
		}
	}
	for field, value := range input.Values {
		if !input.FieldChanged(field) || field == "id" || field == "version" || managed[field] || strings.HasPrefix(field, "_") {
			continue
		}
		if original, known := input.OldValues[field]; known && core.ExactValueEqual(original, value) {
			continue
		}
		return true
	}
	return false
}

// ApplyContextFix stages one model-declared field. Provider exceptions and
// panics become safe, structured field errors; their raw text is not retained.
func (c *UserContext) ApplyContextFix(input *CheckAndFixInput, field, function string, arguments ...string) (failure *CheckResult) {
	if !input.ContextFixAllowed() {
		return nil
	}
	descriptor := c.Entity(input.Entity)
	property := (*core.PropertyDescriptor)(nil)
	if descriptor != nil {
		storage, err := core.StorageLedgerField(descriptor, field)
		if err == nil {
			property = descriptor.PropertyByName(storage)
		}
	}
	invalid := func() *CheckResult {
		return &CheckResult{RuleID: "invalid_type", EntityType: input.Entity,
			CanonicalLocation: Location().Property(field), Message: "Context Fix failed at " + input.Entity + "." + field}
	}
	if property == nil {
		return invalid()
	}
	declared := property.CreateFunction
	if input.Operation == core.MutationUpdate {
		declared = property.UpdateFunction
	}
	if declared == "" || declared != function {
		return invalid()
	}
	defer func() {
		if recover() != nil {
			failure = invalid()
		}
	}()
	var value core.Value
	var err error
	if resolver, ok := c.GetResource("contextFixResolver").(ContextFixResolver); ok && resolver != nil {
		value, err = resolver.ResolveContextFix(c, input, field, function, append([]string(nil), arguments...))
	} else {
		switch function {
		case "now":
			value = core.ValTimestamp(input.Now.UnixMilli())
		case "uniStr":
			var bytes [16]byte
			if _, err = rand.Read(bytes[:]); err == nil {
				prefix := ""
				if len(arguments) > 0 {
					prefix = arguments[0]
				}
				value = core.ValText(prefix + hex.EncodeToString(bytes[:]))
			}
		default:
			err = fmt.Errorf("Context function requires a configured resolver")
		}
	}
	if err != nil {
		return invalid()
	}
	if !contextFixValueValid(property.DataType, value) {
		return invalid()
	}
	source, label := FixEvidenceContext, "contextFunction:"+function
	if function == "now" {
		source, label = FixEvidenceClock, "graphClock"
	}
	if err = c.RecordFixEvidence(FixEvidence{EntityType: input.Entity, ModelPath: field, Source: source, SourceLabel: label}); err != nil {
		return invalid()
	}
	if input.Values == nil {
		input.Values = make(core.Record)
	}
	input.Values[field] = value
	return nil
}

func contextFixValueValid(kind core.DataType, value core.Value) bool {
	if value.IsNull() {
		return true
	} // Required remains a model Check result.
	switch kind {
	case core.TypeText, core.TypeLargeText:
		_, ok := value.V.(string)
		return ok
	case core.TypeTimestamp:
		_, ok := value.TryTime()
		return ok
	case core.TypeDate:
		_, ok := value.TryTime()
		return ok
	case core.TypeBool:
		_, ok := value.V.(bool)
		return ok
	case core.TypeI64:
		_, ok := value.TryI64()
		return ok
	case core.TypeU64:
		_, ok := value.TryU64()
		return ok
	case core.TypeF64:
		_, ok := value.TryF64()
		return ok
	case core.TypeDecimal:
		_, ok := value.TryDecimal()
		return ok
	case core.TypeTime:
		_, ok := value.TryTimeOfDay()
		return ok
	case core.TypeJson:
		return true
	default:
		return false
	}
}

// CaptureContextFixUndo is generated save infrastructure. It captures only
// model-declared automatic fields; ordinary caller edits are never rewound.
// Its lifetime is a single prepared save attempt, not a shared Context ledger.
func (c *UserContext) CaptureContextFixUndo(entity string, currentBase func() *core.BaseEntityData, ledger *core.EntityRoot, owner core.EntityKey) {
	base := currentBase()
	c.graphSaveMu.Lock()
	defer c.graphSaveMu.Unlock()
	state := c.preparedGraphSnapshot
	if state == nil || state.sealed || base == nil || ledger == nil {
		return
	}
	descriptor := c.Entity(entity)
	if descriptor == nil {
		return
	}
	fields := make([]string, 0)
	storageNames := make(map[string]string)
	values := make(core.Record)
	for _, property := range descriptor.Properties {
		if property.CreateFunction == "" && property.UpdateFunction == "" {
			continue
		}
		field, err := core.CanonicalLedgerField(descriptor, property.Name)
		if err != nil {
			continue
		}
		fields = append(fields, field)
		storageNames[field] = property.Name
		if value, present := base.GetDynamic(property.Name); present {
			values[property.Name] = core.CloneValue(value)
		}
	}
	if len(fields) == 0 {
		return
	}
	undoLedger := ledger.CaptureFields(owner, fields)
	state.contextFixUndo = append(state.contextFixUndo, func() {
		base := currentBase()
		for _, field := range fields {
			storage := storageNames[field]
			if value, present := values[storage]; present {
				base.PutDynamic(storage, core.CloneValue(value))
			} else {
				base.RemoveDynamic(storage)
			}
		}
		actual := owner
		if assigned, ok := state.assignments[preparedEntityKey(owner)]; ok {
			actual = assigned
		}
		undoLedger(actual)
	})
}
