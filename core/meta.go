package core

type PropertyDescriptor struct {
	Name      string
	DataType  DataType
	Nullable  bool
	ColName   string
	IsId      bool
	IsVersion bool
}

func NewPropertyDescriptor(name string, dataType DataType) *PropertyDescriptor {
	return &PropertyDescriptor{
		Name:      name,
		DataType:  dataType,
		Nullable:  true,
		ColName:   name,
		IsId:      false,
		IsVersion: false,
	}
}

func (p *PropertyDescriptor) ColumnName(colName string) *PropertyDescriptor {
	p.ColName = colName
	return p
}

func (p *PropertyDescriptor) NotNull() *PropertyDescriptor {
	p.Nullable = false
	return p
}

func (p *PropertyDescriptor) Id() *PropertyDescriptor {
	p.IsId = true
	return p
}

func (p *PropertyDescriptor) Version() *PropertyDescriptor {
	p.IsVersion = true
	return p
}


type RelationDescriptor struct {
	Name            string
	TargetEntity    string
	LocKey          string
	ForKey          string
	IsMany          bool
	IsAttach        bool
	IsDeleteMissing bool
}

func NewRelationDescriptor(name, targetEntity string) *RelationDescriptor {
	return &RelationDescriptor{
		Name:            name,
		TargetEntity:    targetEntity,
		LocKey:          "id",
		ForKey:          "id",
		IsMany:          false,
		IsAttach:        true,
		IsDeleteMissing: true,
	}
}

func (r *RelationDescriptor) LocalKey(key string) *RelationDescriptor {
	r.LocKey = key
	return r
}

func (r *RelationDescriptor) ForeignKey(key string) *RelationDescriptor {
	r.ForKey = key
	return r
}

func (r *RelationDescriptor) Many() *RelationDescriptor {
	r.IsMany = true
	return r
}

func (r *RelationDescriptor) Attach() *RelationDescriptor {
	r.IsAttach = true
	return r
}

func (r *RelationDescriptor) Detached() *RelationDescriptor {
	r.IsAttach = false
	return r
}

func (r *RelationDescriptor) DeleteMissing() *RelationDescriptor {
	r.IsDeleteMissing = true
	return r
}

func (r *RelationDescriptor) KeepMissing() *RelationDescriptor {
	r.IsDeleteMissing = false
	return r
}


type EntityDescriptor struct {
	Name             string
	TabName          string
	DataSvc          *string
	Properties       []*PropertyDescriptor
	Relations        []*RelationDescriptor
	AuditMaskFlds    []string
	AuditValueMaxL   *int
}

func NewEntityDescriptor(name string) *EntityDescriptor {
	return &EntityDescriptor{
		Name:          name,
		TabName:       DefaultTableName(name),
		DataSvc:       nil,
		Properties:    make([]*PropertyDescriptor, 0),
		Relations:     make([]*RelationDescriptor, 0),
		AuditMaskFlds: make([]string, 0),
		AuditValueMaxL: nil,
	}
}

func (e *EntityDescriptor) TableName(tableName string) *EntityDescriptor {
	e.TabName = tableName
	return e
}

func (e *EntityDescriptor) DataService(dataService string) *EntityDescriptor {
	e.DataSvc = &dataService
	return e
}

func (e *EntityDescriptor) Property(property *PropertyDescriptor) *EntityDescriptor {
	e.Properties = append(e.Properties, property)
	return e
}

func (e *EntityDescriptor) Relation(relation *RelationDescriptor) *EntityDescriptor {
	e.Relations = append(e.Relations, relation)
	return e
}

func (e *EntityDescriptor) AuditMaskFields(fields []string) *EntityDescriptor {
	e.AuditMaskFlds = fields
	return e
}

func (e *EntityDescriptor) AuditValueMaxLen(maxLen int) *EntityDescriptor {
	e.AuditValueMaxL = &maxLen
	return e
}


// Lookups

func (e *EntityDescriptor) PropertyByName(name string) *PropertyDescriptor {
	for _, p := range e.Properties {
		if p.Name == name {
			return p
		}
	}
	return nil
}

func (e *EntityDescriptor) RelationByName(name string) *RelationDescriptor {
	for _, r := range e.Relations {
		if r.Name == name {
			return r
		}
	}
	return nil
}

func (e *EntityDescriptor) IdProperty() *PropertyDescriptor {
	for _, p := range e.Properties {
		if p.IsId {
			return p
		}
	}
	return nil
}

func (e *EntityDescriptor) VersionProperty() *PropertyDescriptor {
	for _, p := range e.Properties {
		if p.IsVersion {
			return p
		}
	}
	return nil
}

func (e *EntityDescriptor) WritableProperties() []*PropertyDescriptor {
	var writable []*PropertyDescriptor
	for _, p := range e.Properties {
		if !p.IsId {
			writable = append(writable, p)
		}
	}
	return writable
}

// Getters for optional values
func (e *EntityDescriptor) GetDataService() *string {
	return e.DataSvc
}

func (e *EntityDescriptor) GetAuditValueMaxLen() *int {
	return e.AuditValueMaxL
}
