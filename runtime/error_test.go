package runtime

import (
	"errors"
	"testing"
)

func TestRuntimeErrorError(t *testing.T) {
	tests := []struct {
		name     string
		err      RuntimeError
		expected string
	}{
		{
			name: "MissingEntity",
			err: RuntimeError{
				Type:              "MissingEntity",
				MissingEntityName: "User",
			},
			expected: "missing entity descriptor: User",
		},
		{
			name: "Behavior",
			err: RuntimeError{
				Type:    "Behavior",
				Message: "validation failed",
			},
			expected: "entity data service behavior error: validation failed",
		},
		{
			name: "MissingRelation",
			err: RuntimeError{
				Type:                  "MissingRelation",
				MissingRelationEntity: "User",
				MissingRelationName:   "Profile",
			},
			expected: "missing relation Profile on entity User",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.err.Error(); got != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}

func TestContextErrorError(t *testing.T) {
	tests := []struct {
		name     string
		err      ContextError
		expected string
	}{
		{
			name: "MissingResource",
			err: ContextError{
				Type:                "MissingResource",
				MissingResourceName: "config",
			},
			expected: "missing named resource: config",
		},
		{
			name: "MissingEntityDataService",
			err: ContextError{
				Type:                     "MissingEntityDataService",
				MissingEntityDataService: "User",
			},
			expected: "missing entity data service for entity: User",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.err.Error(); got != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}

func TestRuntimeErrorErrorAll(t *testing.T) {
	tests := []struct {
		name     string
		err      RuntimeError
		expected string
	}{
		{
			name: "Event",
			err: RuntimeError{
				Type:    "Event",
				Message: "foo",
			},
			expected: "entity event error: foo",
		},
		{
			name: "Policy",
			err: RuntimeError{
				Type:    "Policy",
				Message: "foo",
			},
			expected: "request policy error: foo",
		},
		{
			name: "SqlCompile",
			err: RuntimeError{
				Type:       "SqlCompile",
				InnerError: errors.New("unknown entity: User"),
			},
			expected: "unknown entity: User",
		},
		{
			name: "SqlCompile (nil InnerError)",
			err: RuntimeError{
				Type: "SqlCompile",
			},
			expected: "SQL compile error",
		},
		{
			name: "Check",
			err: RuntimeError{
				Type: "Check",
				CheckResults: []CheckResult{
					{Message: "Error at $: bar"},
				},
			},
			expected: "check failed: Error at $: bar",
		},
		{
			name: "Graph",
			err: RuntimeError{
				Type:    "Graph",
				Message: "foo",
			},
			expected: "graph write error: foo",
		},
		{
			name: "IdGeneration",
			err: RuntimeError{
				Type:    "IdGeneration",
				Message: "foo",
			},
			expected: "id generation error: foo",
		},
		{
			name: "Language",
			err: RuntimeError{
				Type:    "Language",
				Message: "foo",
			},
			expected: "language error: foo",
		},
		{
			name: "Schema",
			err: RuntimeError{
				Type:    "Schema",
				Message: "foo",
			},
			expected: "schema provider error: foo",
		},
		{
			name: "OptimisticLockConflict",
			err: RuntimeError{
				Type:                 "OptimisticLockConflict",
				OptimisticLockEntity: "User",
				OptimisticLockId:     "1",
			},
			expected: "optimistic lock conflict on User(1)",
		},
		{
			name: "Default",
			err: RuntimeError{
				Type:    "UnknownType",
				Message: "fallback message",
			},
			expected: "fallback message",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.err.Error(); got != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}

func TestContextErrorErrorAll(t *testing.T) {
	tests := []struct {
		name     string
		err      ContextError
		expected string
	}{
		{
			name: "MissingTypedResource",
			err: ContextError{
				Type:                     "MissingTypedResource",
				MissingTypedResourceName: "foo",
			},
			expected: "missing typed resource: foo",
		},
		{
			name: "Default",
			err: ContextError{
				Type: "UnknownType",
			},
			expected: "context error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.err.Error(); got != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}

func TestDataServiceErrorError(t *testing.T) {
	tests := []struct {
		name     string
		err      DataServiceError
		expected string
	}{
		{
			name: "Runtime",
			err: DataServiceError{
				Type: "Runtime",
				RuntimeError: &RuntimeError{
					Type:              "MissingEntity",
					MissingEntityName: "E",
				},
			},
			expected: "missing entity descriptor: E",
		},
		{
			name: "Entity",
			err: DataServiceError{
				Type:        "Entity",
				EntityError: errors.New("entity E error: missing field: f"),
			},
			expected: "entity E error: missing field: f",
		},
		{
			name: "Executor",
			err: DataServiceError{
				Type:          "Executor",
				ExecutorError: errors.New("err"),
			},
			expected: "err",
		},
		{
			name: "Default",
			err: DataServiceError{
				Type: "UnknownType",
			},
			expected: "data service error",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.err.Error(); got != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, got)
			}
		})
	}
}
