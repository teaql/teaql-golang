package core

type MutationKind int

const (
	MutationInsert MutationKind = iota
	MutationUpdate
	MutationDelete
	MutationRecover
)

type InsertCommand struct {
	Entity     string
	Values     Record
	TraceChain []*TraceNode
}

func NewInsertCommand(entity string) *InsertCommand {
	return &InsertCommand{
		Entity:     entity,
		Values:     make(Record),
		TraceChain: make([]*TraceNode, 0),
	}
}

func (c *InsertCommand) Value(field string, value Value) *InsertCommand {
	c.Values[field] = value
	return c
}

type UpdateCommand struct {
	Entity          string
	Id              Value
	ExpectedVersion *int64
	Values          Record
	TraceChain      []*TraceNode
	OldValues       Record
}

func NewUpdateCommand(entity string, id Value) *UpdateCommand {
	return &UpdateCommand{
		Entity:     entity,
		Id:         id,
		Values:     make(Record),
		TraceChain: make([]*TraceNode, 0),
	}
}

func (c *UpdateCommand) WithExpectedVersion(version int64) *UpdateCommand {
	c.ExpectedVersion = &version
	return c
}

func (c *UpdateCommand) Value(field string, value Value) *UpdateCommand {
	c.Values[field] = value
	return c
}

type BatchInsertCommand struct {
	Entity      string
	BatchValues []Record
	TraceChains [][]*TraceNode
}

func NewBatchInsertCommand(entity string) *BatchInsertCommand {
	return &BatchInsertCommand{
		Entity:      entity,
		BatchValues: make([]Record, 0),
		TraceChains: make([][]*TraceNode, 0),
	}
}

type BatchUpdateCommand struct {
	Entity                string
	BatchIds              []Value
	BatchExpectedVersions []*int64
	BatchValues           []Record
	UpdateFields          []string
	TraceChains           [][]*TraceNode
	BatchOldValues        []Record
}

func NewBatchUpdateCommand(entity string, updateFields []string) *BatchUpdateCommand {
	return &BatchUpdateCommand{
		Entity:                entity,
		BatchIds:              make([]Value, 0),
		BatchExpectedVersions: make([]*int64, 0),
		BatchValues:           make([]Record, 0),
		UpdateFields:          updateFields,
		TraceChains:           make([][]*TraceNode, 0),
		BatchOldValues:        make([]Record, 0),
	}
}

type DeleteCommand struct {
	Entity          string
	Id              Value
	ExpectedVersion *int64
	SoftDelete      bool
	TraceChain      []*TraceNode
}

func NewDeleteCommand(entity string, id Value) *DeleteCommand {
	return &DeleteCommand{
		Entity:     entity,
		Id:         id,
		SoftDelete: true,
		TraceChain: make([]*TraceNode, 0),
	}
}

func (c *DeleteCommand) WithExpectedVersion(version int64) *DeleteCommand {
	c.ExpectedVersion = &version
	return c
}

func (c *DeleteCommand) HardDelete() *DeleteCommand {
	c.SoftDelete = false
	return c
}

type RecoverCommand struct {
	Entity          string
	Id              Value
	ExpectedVersion int64
	TraceChain      []*TraceNode
}

func NewRecoverCommand(entity string, id Value, expectedVersion int64) *RecoverCommand {
	return &RecoverCommand{
		Entity:          entity,
		Id:              id,
		ExpectedVersion: expectedVersion,
		TraceChain:      make([]*TraceNode, 0),
	}
}
