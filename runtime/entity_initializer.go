package runtime

import "strings"

// EntityInitializer applies trusted context-owned defaults during generated
// entity creation. Business input must not provide tenant, actor, policy,
// provider, or audit infrastructure values directly.
type EntityInitializer interface {
	InitializeEntity(ctx *UserContext, entityName string, entity any) any
}

// InitializeEntity invokes the optional trusted initializer registered in the
// UserContext and otherwise preserves the generated concrete entity.
func (c *UserContext) InitializeEntity(entityName string, entity any) any {
	if c == nil {
		panic("UserContext is required for entity creation")
	}
	if strings.TrimSpace(entityName) == "" {
		panic("entityName must not be empty")
	}
	if entity == nil {
		panic("entity must not be nil")
	}
	if initializer, ok := c.GetResource("entityInitializer").(EntityInitializer); ok {
		initialized := initializer.InitializeEntity(c, entityName, entity)
		if initialized == nil {
			panic("entity initializer returned nil")
		}
		return initialized
	}
	return entity
}
