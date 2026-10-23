package runtime

import (
	"encoding/json"
	"fmt"
	"strings"
)

type expressionStep struct {
	call     string
	location ObjectLocation
}

// ExpressionPath records generated native calls separately from canonical keys.
// Value receivers and copied slices keep sibling traversals independent.
type ExpressionPath struct {
	root       string
	profile    JsonFieldNamingProfile
	location   ObjectLocation
	steps      []expressionStep
	missing    int
	hasMissing bool
}

func NewExpressionPath(root string, profile JsonFieldNamingProfile) ExpressionPath {
	return ExpressionPath{root: root, profile: profile, location: Location()}
}
func (p ExpressionPath) Member(name, method string) ExpressionPath {
	return p.append(method+"()", p.location.Property(name))
}
func (p ExpressionPath) At(index int, first bool) ExpressionPath {
	call := fmt.Sprintf("Get(%d)", index)
	if first {
		call = "First()"
	}
	return p.append(call, p.location.At(index))
}
func (p ExpressionPath) Operation(call string) ExpressionPath { return p.append(call, p.location) }
func (p ExpressionPath) append(call string, location ObjectLocation) ExpressionPath {
	p.steps = append(append([]expressionStep{}, p.steps...), expressionStep{call, location})
	p.location = location
	return p
}
func (p ExpressionPath) MarkMissing() ExpressionPath {
	if !p.hasMissing && len(p.steps) > 0 {
		p.missing = len(p.steps) - 1
		p.hasMissing = true
	}
	return p
}
func (p ExpressionPath) WithFailureFrom(failed ExpressionPath) ExpressionPath {
	p.missing, p.hasMissing = failed.missing, failed.hasMissing
	return p
}
func (p ExpressionPath) Render(broken bool) string {
	var result strings.Builder
	result.WriteString(p.root)
	for index, step := range p.steps {
		result.WriteByte('.')
		result.WriteString(step.call)
		if broken && p.hasMissing && index == p.missing {
			result.WriteString("<broken>")
		}
	}
	return result.String()
}

type ExpressionPreloadGuidance struct {
	Action   string                `json:"action"`
	Location []WireLocationSegment `json:"location"`
}
type NotLoadedDiagnostic struct {
	Code                      string                    `json:"code"`
	Location                  []WireLocationSegment     `json:"location"`
	InstancePath              string                    `json:"instancePath"`
	BreakPoint                string                    `json:"breakPoint"`
	AttemptedLocation         []WireLocationSegment     `json:"attemptedLocation"`
	AttemptedInstancePath     string                    `json:"attemptedInstancePath"`
	BreakPointIndex           int                       `json:"breakPointIndex"`
	OriginalExpression        string                    `json:"originalExpression"`
	ExpressionWithBrokenPoint string                    `json:"expressionWithBrokenPoint"`
	SuggestedFix              ExpressionPreloadGuidance `json:"suggestedFix"`
}

func (p ExpressionPath) Diagnostic(fallback string) NotLoadedDiagnostic {
	failure := p.location
	if p.hasMissing {
		failure = p.steps[p.missing].location
	}
	breakPoint := fallback
	if len(failure.Segments) > 0 && failure.Segments[len(failure.Segments)-1].Property != nil {
		breakPoint = *failure.Segments[len(failure.Segments)-1].Property
	}
	segments := (CheckResult{CanonicalLocation: failure}).ToWire(p.profile).Location
	attempted := (CheckResult{CanonicalLocation: p.location}).ToWire(p.profile).Location
	index := len(segments) - 1
	if index < 0 {
		index = 0
	}
	return NotLoadedDiagnostic{"NOT_LOADED", segments, failure.InstancePathWith(p.profile), breakPoint,
		attempted, p.location.InstancePathWith(p.profile), index, p.Render(false), p.Render(true),
		ExpressionPreloadGuidance{"preload", segments}}
}

type NotLoadedError struct {
	path       ExpressionPath
	breakPoint string
}

func NewNotLoadedError(path ExpressionPath, breakPoint string) *NotLoadedError {
	return &NotLoadedError{path, breakPoint}
}
func (e *NotLoadedError) WithPath(path ExpressionPath) *NotLoadedError {
	return NewNotLoadedError(path.WithFailureFrom(e.path), e.breakPoint)
}
func (e *NotLoadedError) Terminal(call string) *NotLoadedError {
	return NewNotLoadedError(e.path.Operation(call), e.breakPoint)
}
func (e *NotLoadedError) ToWire() NotLoadedDiagnostic { return e.path.Diagnostic(e.breakPoint) }
func (e *NotLoadedError) Error() string {
	var value strings.Builder
	encoder := json.NewEncoder(&value)
	encoder.SetEscapeHTML(false)   // Native source and <broken> remain readable in logs.
	_ = encoder.Encode(e.ToWire()) // all fields are closed scalar/location types
	return strings.TrimSuffix(value.String(), "\n")
}
