package core

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPropertyDescriptorBuilder(t *testing.T) {
	prop := NewPropertyDescriptor("username", TypeText).
		ColumnName("user_name").
		NotNull().
		Id().
		Version()

	assert.Equal(t, "username", prop.Name)
	assert.Equal(t, "user_name", prop.ColName) // Using ColName because of method conflict
	assert.Equal(t, TypeText, prop.DataType)
	assert.False(t, prop.Nullable)
	assert.True(t, prop.IsId)
	assert.True(t, prop.IsVersion)
}

func TestRelationDescriptorBuilder(t *testing.T) {
	rel := NewRelationDescriptor("orders", "Order").
		LocalKey("user_id").
		ForeignKey("customer_id").
		Many().
		Detached().
		KeepMissing()

	assert.Equal(t, "orders", rel.Name)
	assert.Equal(t, "Order", rel.TargetEntity)
	assert.Equal(t, "user_id", rel.LocKey) // Using LocKey because of method conflict
	assert.Equal(t, "customer_id", rel.ForKey) // Using ForKey
	assert.True(t, rel.IsMany)
	assert.False(t, rel.IsAttach)
	assert.False(t, rel.IsDeleteMissing)
}

func TestEntityDescriptorBuilderAndLookups(t *testing.T) {
	entity := NewEntityDescriptor("User").
		TableName("users").
		DataService("auth_db").
		AuditMaskFields([]string{"password"}).
		AuditValueMaxLen(255)

	idProp := NewPropertyDescriptor("id", TypeI64).Id()
	nameProp := NewPropertyDescriptor("name", TypeText)
	versionProp := NewPropertyDescriptor("version", TypeI64).Version()

	ordersRel := NewRelationDescriptor("orders", "Order")

	entity = entity.
		Property(idProp).
		Property(nameProp).
		Property(versionProp).
		Relation(ordersRel)

	assert.Equal(t, "User", entity.Name)
	assert.Equal(t, "users", entity.TabName) // Using TabName because of method conflict
	
	ds := entity.GetDataService()
	assert.NotNil(t, ds)
	assert.Equal(t, "auth_db", *ds)
	
	assert.Equal(t, []string{"password"}, entity.AuditMaskFlds)
	
	maxLen := entity.GetAuditValueMaxLen()
	assert.NotNil(t, maxLen)
	assert.Equal(t, 255, *maxLen)

	// Lookups
	foundNameProp := entity.PropertyByName("name")
	assert.NotNil(t, foundNameProp)
	assert.Equal(t, nameProp.Name, foundNameProp.Name)
	
	assert.Nil(t, entity.PropertyByName("missing"))

	foundOrdersRel := entity.RelationByName("orders")
	assert.NotNil(t, foundOrdersRel)
	assert.Equal(t, ordersRel.Name, foundOrdersRel.Name)
	
	assert.Nil(t, entity.RelationByName("missing"))

	foundIdProp := entity.IdProperty()
	assert.NotNil(t, foundIdProp)
	assert.Equal(t, idProp.Name, foundIdProp.Name)

	foundVersionProp := entity.VersionProperty()
	assert.NotNil(t, foundVersionProp)
	assert.Equal(t, versionProp.Name, foundVersionProp.Name)

	writable := entity.WritableProperties()
	assert.Len(t, writable, 2)
	
	names := []string{writable[0].Name, writable[1].Name}
	assert.Contains(t, names, "name")
	assert.Contains(t, names, "version")
}

func TestRelationDescriptorAttachDeleteMissing(t *testing.T) {
	rel := NewRelationDescriptor("orders", "Order").
		Attach().
		DeleteMissing()
	assert.True(t, rel.IsAttach)
	assert.True(t, rel.IsDeleteMissing)
}

func TestEntityDescriptorMissingIdVersion(t *testing.T) {
	entity := NewEntityDescriptor("User")
	// no id or version property added
	assert.Nil(t, entity.IdProperty())
	assert.Nil(t, entity.VersionProperty())
}
