package runtime

import (
	"fmt"
	"strings"
)

type CheckResult struct {
	RuleID string
	// Location is a legacy presentation retained for source compatibility.
	Location           string
	CanonicalLocation  ObjectLocation
	InputValue         any
	SystemValue        any
	Message            string
	EntityType         string
	SourceInstancePath string
}

func (c *CheckResult) String() string { return c.Message }
func (c *CheckResult) ObjectLocation() ObjectLocation {
	if len(c.CanonicalLocation.Segments) != 0 || c.Location == "" {
		return c.CanonicalLocation
	}
	return LocationFromModelPath(c.Location)
}
func (c CheckResult) PrefixedBy(prefix ObjectLocation) CheckResult {
	c.CanonicalLocation = c.ObjectLocation().PrefixedBy(prefix)
	c.Location = c.CanonicalLocation.ModelPath()
	return c
}
func (c *CheckResult) ModelPath() string    { return c.ObjectLocation().ModelPath() }
func (c *CheckResult) NativePath() string   { return c.ObjectLocation().NativePath() }
func (c *CheckResult) InstancePath() string { return c.ObjectLocation().InstancePath() }

type WireCheckResult struct {
	RuleID             string                `json:"ruleId"`
	EntityType         string                `json:"entityType,omitempty"`
	Location           []WireLocationSegment `json:"location"`
	InstancePath       string                `json:"instancePath"`
	SourceInstancePath string                `json:"sourceInstancePath,omitempty"`
	InputValue         any                   `json:"inputValue,omitempty"`
	SystemValue        any                   `json:"systemValue,omitempty"`
	Message            string                `json:"message,omitempty"`
}

type WireLocationSegment struct {
	Kind  string `json:"kind"`
	Name  string `json:"name,omitempty"`
	Index *int   `json:"index,omitempty"`
}

func (c CheckResult) ToWire(profile JsonFieldNamingProfile) WireCheckResult {
	location := c.ObjectLocation()
	segments := make([]WireLocationSegment, 0, len(location.Segments))
	for _, segment := range location.Segments {
		if segment.Property != nil {
			segments = append(segments, WireLocationSegment{Kind: "property", Name: *segment.Property})
		} else {
			segments = append(segments, WireLocationSegment{Kind: "index", Index: segment.Index})
		}
	}
	return WireCheckResult{c.RuleID, c.EntityType, segments, location.InstancePathWith(profile), c.SourceInstancePath, c.InputValue, c.SystemValue, c.Message}
}

type RuntimeError struct {
	Type                  string
	Message               string
	MissingEntityName     string
	CheckResults          []CheckResult
	MissingRelationEntity string
	MissingRelationName   string
	OptimisticLockEntity  string
	OptimisticLockId      string
	InnerError            error
}

func (e *RuntimeError) Error() string {
	switch e.Type {
	case "MissingEntity":
		return fmt.Sprintf("missing entity descriptor: %s", e.MissingEntityName)
	case "SqlCompile":
		if e.InnerError != nil {
			return e.InnerError.Error()
		}
		return "SQL compile error"
	case "Behavior":
		return fmt.Sprintf("entity data service behavior error: %s", e.Message)
	case "Event":
		return fmt.Sprintf("entity event error: %s", e.Message)
	case "Policy":
		return fmt.Sprintf("request policy error: %s", e.Message)
	case "Check":
		messages := make([]string, len(e.CheckResults))
		for i, res := range e.CheckResults {
			messages[i] = res.String()
		}
		return fmt.Sprintf("check failed: %s", strings.Join(messages, "; "))
	case "Graph":
		return fmt.Sprintf("graph write error: %s", e.Message)
	case "IdGeneration":
		return fmt.Sprintf("id generation error: %s", e.Message)
	case "Language":
		return fmt.Sprintf("language error: %s", e.Message)
	case "Schema":
		return fmt.Sprintf("schema provider error: %s", e.Message)
	case "MissingRelation":
		return fmt.Sprintf("missing relation %s on entity %s", e.MissingRelationName, e.MissingRelationEntity)
	case "OptimisticLockConflict":
		return fmt.Sprintf("optimistic lock conflict on %s(%s)", e.OptimisticLockEntity, e.OptimisticLockId)
	default:
		return e.Message
	}
}

type ContextError struct {
	Type                     string
	MissingResourceName      string
	MissingTypedResourceName string
	MissingEntityDataService string
}

func (e *ContextError) Error() string {
	switch e.Type {
	case "MissingResource":
		return fmt.Sprintf("missing named resource: %s", e.MissingResourceName)
	case "MissingTypedResource":
		return fmt.Sprintf("missing typed resource: %s", e.MissingTypedResourceName)
	case "MissingEntityDataService":
		return fmt.Sprintf("missing entity data service for entity: %s", e.MissingEntityDataService)
	default:
		return "context error"
	}
}

type DataServiceError struct {
	Type          string
	RuntimeError  *RuntimeError
	EntityError   error
	ExecutorError error
}

func (e *DataServiceError) Error() string {
	switch e.Type {
	case "Runtime":
		return e.RuntimeError.Error()
	case "Entity":
		return e.EntityError.Error()
	case "Executor":
		return e.ExecutorError.Error()
	default:
		return "data service error"
	}
}
