package core

import (
	"fmt"
	"strings"

	"github.com/shopspring/decimal"
)

type EntityError struct {
	Entity  string
	Message string
}

func NewEntityError(entity, message string) *EntityError {
	return &EntityError{Entity: entity, Message: message}
}

func (e *EntityError) Error() string {
	return fmt.Sprintf("%s: %s", e.Entity, e.Message)
}

type TeaqlEntity interface {
	EntityName() string
	EntityDescriptor() *EntityDescriptor
}

type Entity interface {
	TeaqlEntity
	FromRecord(record Record) error
	IntoRecord() Record
	DirtyFields() []string
	IsMarkedAsDelete() bool
	IsNew() bool
	MarkAsNew()
	GetComment() *string
	SetComment(comment string)
	OriginalValues() Record
	OnLoaded(context any)
	IntoJson() any
}

type Audited[T Entity] struct {
	Inner   T
	Comment string
}

func NewAudited[T Entity](entity T, comment string) *Audited[T] {
	if strings.TrimSpace(comment) == "" {
		panic("audit comment must not be empty")
	}
	return &Audited[T]{Inner: entity, Comment: comment}
}

func (a *Audited[T]) Entity() T {
	return a.Inner
}

func (a *Audited[T]) IntoEntity() T {
	a.Inner.SetComment(a.Comment)
	return a.Inner
}

func (a *Audited[T]) GetComment() string {
	return a.Comment
}

type BaseEntityData struct {
	Id      uint64
	Version int64
	Dynamic Record
}

func NewBaseEntityData() *BaseEntityData {
	return &BaseEntityData{
		Dynamic: make(Record),
	}
}

func (b *BaseEntityData) WithId(id uint64) *BaseEntityData {
	b.Id = id
	return b
}

func (b *BaseEntityData) WithVersion(version int64) *BaseEntityData {
	b.Version = version
	return b
}

func (b *BaseEntityData) WithDynamic(key string, value Value) *BaseEntityData {
	b.Dynamic[key] = value
	return b
}

func (b *BaseEntityData) GetDynamic(key string) (Value, bool) {
	val, ok := b.Dynamic[key]
	return val, ok
}

func (b *BaseEntityData) DynamicI64(key string) (int64, bool) {
	if val, ok := b.GetDynamic(key); ok {
		return val.TryI64()
	}
	return 0, false
}

func (b *BaseEntityData) DynamicU64(key string) (uint64, bool) {
	if val, ok := b.GetDynamic(key); ok {
		return val.TryU64()
	}
	return 0, false
}

func (b *BaseEntityData) DynamicDecimal(key string) (decimal.Decimal, bool) {
	if val, ok := b.GetDynamic(key); ok {
		return val.TryDecimal()
	}
	return decimal.Zero, false
}

func (b *BaseEntityData) DynamicF64(key string) (float64, bool) {
	if val, ok := b.GetDynamic(key); ok {
		return val.TryF64()
	}
	return 0, false
}

func (b *BaseEntityData) DynamicText(key string) (string, bool) {
	if val, ok := b.GetDynamic(key); ok {
		return val.TryText()
	}
	return "", false
}

func (b *BaseEntityData) DynamicBool(key string) (bool, bool) {
	if val, ok := b.GetDynamic(key); ok {
		return val.TryBool()
	}
	return false, false
}

func (b *BaseEntityData) PutDynamic(key string, value Value) {
	b.Dynamic[key] = value
}

func (b *BaseEntityData) RemoveDynamic(key string) {
	delete(b.Dynamic, key)
}

func (b *BaseEntityData) ToRecord() Record {
	record := make(Record)
	record["id"] = ValU64(b.Id)
	record["version"] = ValI64(b.Version)
	for k, v := range b.Dynamic {
		record[k] = v
	}
	return record
}

func BaseEntityDataFromRecord(record Record) (*BaseEntityData, error) {
	b := NewBaseEntityData()
	
	if idVal, ok := record["id"]; ok {
		if id, ok := idVal.TryU64(); ok {
			b.Id = id
		} else if idI, ok := idVal.TryI64(); ok && idI >= 0 {
			b.Id = uint64(idI)
		} else {
			return nil, NewEntityError("BaseEntity", fmt.Sprintf("invalid id field: %v", idVal))
		}
	}
	
	if versionVal, ok := record["version"]; ok {
		if version, ok := versionVal.TryI64(); ok {
			b.Version = version
		} else {
			return nil, NewEntityError("BaseEntity", fmt.Sprintf("invalid version field: %v", versionVal))
		}
	}
	
	for k, v := range record {
		if k != "id" && k != "version" {
			b.Dynamic[k] = v
		}
	}
	
	return b, nil
}

type BaseEntity interface {
	Entity
	Base() *BaseEntityData
}

type IdentifiableEntity interface {
	Entity
	IdValue() Value
}

type VersionedEntity interface {
	Entity
	Version() int64
}

type EntityDescriptorStore interface {
	RegisterDescriptor(descriptor *EntityDescriptor)
}
