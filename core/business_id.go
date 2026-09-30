package core

import (
	"fmt"
	"strings"
)

// Stable Business ID failure codes shared by runtime implementations.
const (
	BusinessIDDefinitionInvalid = "BUSINESS_ID_DEFINITION_INVALID"
	BusinessIDRangeExhausted    = "BUSINESS_ID_RANGE_EXHAUSTED"
	BusinessIDEncodingFailed    = "BUSINESS_ID_ENCODING_FAILED"
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
