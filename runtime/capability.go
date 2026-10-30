package runtime

import (
	"errors"
	"sort"
	"strings"
)

// RuntimeCapabilityResource is installed by a concrete optional module after
// it has validated its SQL owner, schema, policy, storage, and service graph.
type RuntimeCapabilityResource interface {
	Capability() string
	ValidateReady(*UserContext) error
}

func canonicalCapability(name string) bool {
	return name != "" && name == strings.TrimSpace(name) && !strings.ContainsAny(name, "\x00\r\n\t ")
}

func capabilitySlot(name string) string { return "teaql.runtime.capability:" + name }

func (c *UserContext) InstallRuntimeCapability(resource RuntimeCapabilityResource) error {
	if resource == nil || missingSPI(resource) || !canonicalCapability(resource.Capability()) {
		return errors.New("RUNTIME_CAPABILITY_INVALID")
	}
	name := resource.Capability()
	if c.GetResource(capabilitySlot(name)) != nil {
		return errors.New("RUNTIME_CAPABILITY_DUPLICATE")
	}
	c.InsertResource(capabilitySlot(name), resource)
	return nil
}

// ValidateCapabilities is explicit startup readiness and a save-time fail-closed gate.
func (c *UserContext) ValidateCapabilities() error {
	names := make([]string, 0, len(c.requiredCapabilities))
	for name := range c.requiredCapabilities {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		resource, ok := c.GetResource(capabilitySlot(name)).(RuntimeCapabilityResource)
		if !ok || resource.Capability() != name {
			return missingCapability(name)
		}
		if err := resource.ValidateReady(c); err != nil {
			return missingCapability(name)
		}
	}
	return nil
}

func missingCapability(name string) error {
	switch name {
	case "blob.v1":
		return errors.New("BLOB_CAPABILITY_MISSING")
	case "secure_document.v1":
		return errors.New("DOCUMENT_CONFIGURATION")
	default:
		return errors.New("RUNTIME_CAPABILITY_MISSING")
	}
}
