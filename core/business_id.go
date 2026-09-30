package core

import (
	stdcontext "context"
	"fmt"
	"strings"
	"time"
)

// Stable Business ID failure codes shared by runtime implementations.
const (
	BusinessIDProfileNotFound          = "BUSINESS_ID_PROFILE_NOT_FOUND"
	BusinessIDDefinitionInvalid        = "BUSINESS_ID_DEFINITION_INVALID"
	BusinessIDRangeExhausted           = "BUSINESS_ID_RANGE_EXHAUSTED"
	BusinessIDAllocationRetryExhausted = "BUSINESS_ID_ALLOCATION_RETRY_EXHAUSTED"
	BusinessIDFormatInvalid            = "BUSINESS_ID_FORMAT_INVALID"
	BusinessIDDuplicate                = "BUSINESS_ID_DUPLICATE"
	BusinessIDImmutable                = "BUSINESS_ID_IMMUTABLE"
	BusinessIDKeyNotFound              = "BUSINESS_ID_KEY_NOT_FOUND"
	BusinessIDEncodingFailed           = "BUSINESS_ID_ENCODING_FAILED"
)

// BusinessIDError classifies a portable Business ID failure.
type BusinessIDError struct {
	Code    string
	Message string
}

func (e *BusinessIDError) Error() string { return e.Message }

// BusinessIDScope is the complete logical allocation and permutation scope.
type BusinessIDScope struct {
	domainRootKey string
	aggregateType string
	namespace     string
	periodKey     string
}

func NewBusinessIDScope(domainRootKey, aggregateType, namespace, periodKey string) (BusinessIDScope, error) {
	values := []struct {
		name  string
		value string
	}{
		{"domainRootKey", domainRootKey},
		{"aggregateType", aggregateType},
		{"namespace", namespace},
		{"periodKey", periodKey},
	}
	for _, field := range values {
		if strings.TrimSpace(field.value) == "" {
			return BusinessIDScope{}, &BusinessIDError{
				Code:    BusinessIDDefinitionInvalid,
				Message: fmt.Sprintf("%s must not be blank", field.name),
			}
		}
	}
	return BusinessIDScope{domainRootKey, aggregateType, namespace, periodKey}, nil
}

func (s BusinessIDScope) DomainRootKey() string { return s.domainRootKey }
func (s BusinessIDScope) AggregateType() string { return s.aggregateType }
func (s BusinessIDScope) Namespace() string     { return s.namespace }
func (s BusinessIDScope) PeriodKey() string     { return s.periodKey }

func (s BusinessIDScope) CanonicalKey() string {
	escape := func(value string) string {
		return strings.ReplaceAll(strings.ReplaceAll(value, "%", "%25"), "|", "%7C")
	}
	return strings.Join([]string{
		escape(s.domainRootKey), escape(s.aggregateType),
		escape(s.namespace), escape(s.periodKey),
	}, "|")
}

// BusinessIDEncodingKey contains one versioned 256-bit application-owned key.
type BusinessIDEncodingKey struct {
	version uint32
	bytes   [32]byte
}

func NewBusinessIDEncodingKey(version uint32, key []byte) (BusinessIDEncodingKey, error) {
	if version == 0 {
		return BusinessIDEncodingKey{}, &BusinessIDError{
			Code: BusinessIDDefinitionInvalid, Message: "Business ID key version must be positive",
		}
	}
	if len(key) != 32 {
		return BusinessIDEncodingKey{}, &BusinessIDError{
			Code: BusinessIDDefinitionInvalid, Message: "Business ID V1 key must contain exactly 32 bytes",
		}
	}
	result := BusinessIDEncodingKey{version: version}
	copy(result.bytes[:], key)
	return result, nil
}

func (k BusinessIDEncodingKey) Version() uint32 { return k.version }

func (k BusinessIDEncodingKey) Bytes() []byte {
	result := make([]byte, len(k.bytes))
	copy(result, k.bytes[:])
	return result
}

const (
	DefaultBusinessIDProfile    = "daily-permuted-v1"
	DefaultBusinessIDDateFormat = "yyyyMMdd"
	DefaultBusinessIDDigits     = 6
)

// BusinessIDDefinition is the immutable model-derived numbering policy.
type BusinessIDDefinition struct {
	fieldName     string
	profile       string
	prefix        string
	dateFormat    string
	reset         string
	digits        int
	separator     string
	namespace     string
	policyVersion uint32
}

func NewBusinessIDDefinition(fieldName, profile, prefix, dateFormat, reset string, digits int, separator, namespace string, policyVersion uint32) (BusinessIDDefinition, error) {
	for name, value := range map[string]string{
		"fieldName": fieldName, "profile": profile, "prefix": prefix,
		"dateFormat": dateFormat, "reset": reset, "separator": separator,
		"namespace": namespace,
	} {
		if strings.TrimSpace(value) == "" {
			return BusinessIDDefinition{}, &BusinessIDError{Code: BusinessIDDefinitionInvalid, Message: name + " must not be blank"}
		}
	}
	if profile == DefaultBusinessIDProfile && digits != DefaultBusinessIDDigits {
		return BusinessIDDefinition{}, &BusinessIDError{Code: BusinessIDDefinitionInvalid, Message: "daily-permuted-v1 requires exactly 6 digits"}
	}
	if digits < 1 || digits > 18 {
		return BusinessIDDefinition{}, &BusinessIDError{Code: BusinessIDDefinitionInvalid, Message: "digits must be between 1 and 18"}
	}
	if policyVersion == 0 {
		return BusinessIDDefinition{}, &BusinessIDError{Code: BusinessIDDefinitionInvalid, Message: "policyVersion must be positive"}
	}
	return BusinessIDDefinition{fieldName, profile, prefix, dateFormat, reset, digits, separator, namespace, policyVersion}, nil
}

func NewDailyPermutedBusinessIDDefinition(fieldName, prefix, namespace string) (BusinessIDDefinition, error) {
	return NewBusinessIDDefinition(fieldName, DefaultBusinessIDProfile, prefix,
		DefaultBusinessIDDateFormat, "daily", DefaultBusinessIDDigits, "-", namespace, 1)
}

func (d BusinessIDDefinition) FieldName() string     { return d.fieldName }
func (d BusinessIDDefinition) Profile() string       { return d.profile }
func (d BusinessIDDefinition) Prefix() string        { return d.prefix }
func (d BusinessIDDefinition) DateFormat() string    { return d.dateFormat }
func (d BusinessIDDefinition) Reset() string         { return d.reset }
func (d BusinessIDDefinition) Digits() int           { return d.digits }
func (d BusinessIDDefinition) Separator() string     { return d.separator }
func (d BusinessIDDefinition) Namespace() string     { return d.namespace }
func (d BusinessIDDefinition) PolicyVersion() uint32 { return d.policyVersion }
func (d BusinessIDDefinition) MaximumSequence() uint64 {
	if d.profile == DefaultBusinessIDProfile {
		return 2_176_782_335
	}
	value := uint64(1)
	for i := 0; i < d.digits; i++ {
		value *= 10
	}
	return value - 1
}

type BusinessIDGenerationRequest struct {
	Definition    BusinessIDDefinition
	DomainRootKey string
	AggregateType string
	BusinessDate  time.Time
}

type BusinessIDPlan struct {
	definition      BusinessIDDefinition
	scope           BusinessIDScope
	businessDate    time.Time
	dateText        string
	initialSequence uint64
	maximumSequence uint64
}

func NewBusinessIDPlan(definition BusinessIDDefinition, scope BusinessIDScope, businessDate time.Time, dateText string, initialSequence, maximumSequence uint64) (BusinessIDPlan, error) {
	if maximumSequence < initialSequence {
		return BusinessIDPlan{}, &BusinessIDError{Code: BusinessIDDefinitionInvalid, Message: "Business ID allocation range must satisfy initialSequence <= maximumSequence"}
	}
	return BusinessIDPlan{definition, scope, businessDate, dateText, initialSequence, maximumSequence}, nil
}

func (p BusinessIDPlan) Definition() BusinessIDDefinition { return p.definition }
func (p BusinessIDPlan) Scope() BusinessIDScope           { return p.scope }
func (p BusinessIDPlan) BusinessDate() time.Time          { return p.businessDate }
func (p BusinessIDPlan) DateText() string                 { return p.dateText }
func (p BusinessIDPlan) InitialSequence() uint64          { return p.initialSequence }
func (p BusinessIDPlan) MaximumSequence() uint64          { return p.maximumSequence }

type BusinessIDAllocation struct {
	Scope    BusinessIDScope
	Sequence uint64
}

type BusinessIDValue struct {
	Value         string
	Profile       string
	PolicyVersion uint32
}

type BusinessIDSlot interface {
	CurrentBusinessID() string
	IsNewAggregate() bool
	AssignBusinessID(string)
}

type BusinessIDAllocator interface {
	AllocateBusinessID(stdcontext.Context, BusinessIDPlan) (BusinessIDAllocation, error)
}
