package runtime

import (
	stdcontext "context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/teaql/teaql-golang/core"
)

const (
	BusinessIDPermutationV1Width       = 6
	BusinessIDPermutationV1DomainSize  = uint64(2_176_782_336)
	BusinessIDPermutationV1MaxSequence = BusinessIDPermutationV1DomainSize - 1
	BusinessIDPermutationV1Alphabet    = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

var businessIDPermutationV1Magic = []byte("teaql-business-id-fp-v1\x00")

type BusinessIDKeyProvider interface {
	CurrentBusinessIDKey(*UserContext, core.BusinessIDDefinition, core.BusinessIDScope) (core.BusinessIDEncodingKey, error)
}

type BusinessIDProfile interface {
	PlanBusinessID(core.BusinessIDGenerationRequest) (core.BusinessIDPlan, error)
	FormatBusinessID(core.BusinessIDPlan, core.BusinessIDAllocation) (core.BusinessIDValue, error)
	ValidateBusinessID(core.BusinessIDDefinition, string) (core.BusinessIDValue, error)
}

type BusinessIDProfileFactory interface {
	CreateBusinessIDProfile(*UserContext, core.BusinessIDDefinition) (BusinessIDProfile, error)
}

type BusinessIDService interface {
	EnsureBusinessID(*UserContext, core.BusinessIDDefinition, string, string, core.BusinessIDSlot) (core.BusinessIDValue, error)
}

type StaticBusinessIDKeyProvider struct{ key core.BusinessIDEncodingKey }

func NewStaticBusinessIDKeyProvider(key core.BusinessIDEncodingKey) *StaticBusinessIDKeyProvider {
	return &StaticBusinessIDKeyProvider{key: key}
}

func (p *StaticBusinessIDKeyProvider) CurrentBusinessIDKey(_ *UserContext, _ core.BusinessIDDefinition, _ core.BusinessIDScope) (core.BusinessIDEncodingKey, error) {
	return p.key, nil
}

type InMemoryBusinessIDAllocator struct {
	mu       sync.Mutex
	counters map[core.BusinessIDScope]uint64
}

func NewInMemoryBusinessIDAllocator() *InMemoryBusinessIDAllocator {
	return &InMemoryBusinessIDAllocator{counters: make(map[core.BusinessIDScope]uint64)}
}

func (a *InMemoryBusinessIDAllocator) AllocateBusinessID(_ stdcontext.Context, plan core.BusinessIDPlan) (core.BusinessIDAllocation, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	current, found := a.counters[plan.Scope()]
	if !found {
		current = plan.InitialSequence()
	}
	if current > plan.MaximumSequence() {
		return core.BusinessIDAllocation{}, &core.BusinessIDError{
			Code:    core.BusinessIDRangeExhausted,
			Message: "Business ID range exhausted for " + plan.Scope().CanonicalKey(),
		}
	}
	a.counters[plan.Scope()] = current + 1
	return core.BusinessIDAllocation{Scope: plan.Scope(), Sequence: current}, nil
}

type PermutedDailyBusinessIDProfile struct {
	context     *UserContext
	keyProvider BusinessIDKeyProvider
}

func NewPermutedDailyBusinessIDProfile(context *UserContext, keyProvider BusinessIDKeyProvider) *PermutedDailyBusinessIDProfile {
	return &PermutedDailyBusinessIDProfile{context: context, keyProvider: keyProvider}
}

func (p *PermutedDailyBusinessIDProfile) PlanBusinessID(request core.BusinessIDGenerationRequest) (core.BusinessIDPlan, error) {
	definition := request.Definition
	if definition.Profile() != core.DefaultBusinessIDProfile || definition.Reset() != "daily" ||
		definition.DateFormat() != core.DefaultBusinessIDDateFormat || definition.Digits() != BusinessIDPermutationV1Width {
		return core.BusinessIDPlan{}, &core.BusinessIDError{
			Code:    core.BusinessIDDefinitionInvalid,
			Message: "daily-permuted-v1 requires reset=daily, dateFormat=yyyyMMdd and digits=6",
		}
	}
	dateText := request.BusinessDate.Format("20060102")
	scope, err := core.NewBusinessIDScope(request.DomainRootKey, request.AggregateType, definition.Namespace(), dateText)
	if err != nil {
		return core.BusinessIDPlan{}, err
	}
	return core.NewBusinessIDPlan(definition, scope, request.BusinessDate, dateText, 0, BusinessIDPermutationV1MaxSequence)
}

func (p *PermutedDailyBusinessIDProfile) FormatBusinessID(plan core.BusinessIDPlan, allocation core.BusinessIDAllocation) (core.BusinessIDValue, error) {
	if allocation.Scope != plan.Scope() {
		return core.BusinessIDValue{}, &core.BusinessIDError{Code: core.BusinessIDDefinitionInvalid, Message: "Allocation scope does not match Business ID plan"}
	}
	key, err := p.keyProvider.CurrentBusinessIDKey(p.context, plan.Definition(), plan.Scope())
	if err != nil {
		return core.BusinessIDValue{}, err
	}
	code, err := EncodeBusinessIDPermutationV1(allocation.Sequence, plan.Scope(), key)
	if err != nil {
		return core.BusinessIDValue{}, err
	}
	definition := plan.Definition()
	return core.BusinessIDValue{
		Value:   strings.Join([]string{definition.Prefix(), plan.DateText(), code}, definition.Separator()),
		Profile: definition.Profile(), PolicyVersion: definition.PolicyVersion(),
	}, nil
}

func (p *PermutedDailyBusinessIDProfile) ValidateBusinessID(definition core.BusinessIDDefinition, value string) (core.BusinessIDValue, error) {
	pattern := "^" + regexp.QuoteMeta(definition.Prefix()) + regexp.QuoteMeta(definition.Separator()) + `(\d{8})` + regexp.QuoteMeta(definition.Separator()) + `([0-9A-Z]{6})$`
	match := regexp.MustCompile(pattern).FindStringSubmatch(value)
	if len(match) != 3 {
		return core.BusinessIDValue{}, invalidBusinessIDFormat(value)
	}
	if _, err := time.Parse("20060102", match[1]); err != nil {
		return core.BusinessIDValue{}, invalidBusinessIDFormat(value)
	}
	return core.BusinessIDValue{Value: value, Profile: definition.Profile(), PolicyVersion: definition.PolicyVersion()}, nil
}

func invalidBusinessIDFormat(value string) error {
	return &core.BusinessIDError{Code: core.BusinessIDFormatInvalid, Message: "Invalid daily-permuted-v1 Business ID: " + value}
}

type DefaultBusinessIDProfileFactory struct{}

func (DefaultBusinessIDProfileFactory) CreateBusinessIDProfile(context *UserContext, definition core.BusinessIDDefinition) (BusinessIDProfile, error) {
	if definition.Profile() != core.DefaultBusinessIDProfile {
		return nil, &core.BusinessIDError{Code: core.BusinessIDProfileNotFound, Message: "Business ID profile is not registered: " + definition.Profile()}
	}
	provider := context.BusinessIDKeyProvider()
	if provider == nil {
		return nil, &core.BusinessIDError{Code: core.BusinessIDKeyNotFound, Message: "Business ID key provider is not registered"}
	}
	return NewPermutedDailyBusinessIDProfile(context, provider), nil
}

type DefaultBusinessIDService struct{ allocator core.BusinessIDAllocator }

func NewDefaultBusinessIDService(allocator core.BusinessIDAllocator) *DefaultBusinessIDService {
	if allocator == nil {
		panic("business ID allocator must not be nil")
	}
	return &DefaultBusinessIDService{allocator: allocator}
}

func (s *DefaultBusinessIDService) EnsureBusinessID(context *UserContext, definition core.BusinessIDDefinition, domainRootKey, aggregateType string, slot core.BusinessIDSlot) (core.BusinessIDValue, error) {
	if slot == nil {
		return core.BusinessIDValue{}, fmt.Errorf("business ID slot must not be nil")
	}
	factory := context.BusinessIDProfileFactory()
	if factory == nil {
		return core.BusinessIDValue{}, &core.BusinessIDError{Code: core.BusinessIDProfileNotFound, Message: "Business ID profile factory is not registered"}
	}
	profile, err := factory.CreateBusinessIDProfile(context, definition)
	if err != nil {
		return core.BusinessIDValue{}, err
	}
	if current := slot.CurrentBusinessID(); strings.TrimSpace(current) != "" {
		return profile.ValidateBusinessID(definition, current)
	}
	if !slot.IsNewAggregate() {
		return core.BusinessIDValue{}, &core.BusinessIDError{Code: core.BusinessIDImmutable, Message: "An established Aggregate cannot be assigned a new Business ID"}
	}
	plan, err := profile.PlanBusinessID(core.BusinessIDGenerationRequest{
		Definition: definition, DomainRootKey: domainRootKey,
		AggregateType: aggregateType, BusinessDate: context.BusinessDate(),
	})
	if err != nil {
		return core.BusinessIDValue{}, err
	}
	allocation, err := s.allocator.AllocateBusinessID(context, plan)
	if err != nil {
		return core.BusinessIDValue{}, err
	}
	value, err := profile.FormatBusinessID(plan, allocation)
	if err != nil {
		return core.BusinessIDValue{}, err
	}
	slot.AssignBusinessID(value.Value)
	return value, nil
}

// EncodeBusinessIDPermutationV1 maps one sequence to the canonical six-character code.
func EncodeBusinessIDPermutationV1(sequence uint64, scope core.BusinessIDScope, key core.BusinessIDEncodingKey) (string, error) {
	if sequence >= BusinessIDPermutationV1DomainSize {
		return "", &core.BusinessIDError{
			Code: core.BusinessIDRangeExhausted, Message: "Business ID V1 sequence must be in 0..2176782335",
		}
	}
	validatedScope, err := core.NewBusinessIDScope(scope.DomainRootKey(), scope.AggregateType(), scope.Namespace(), scope.PeriodKey())
	if err != nil {
		return "", err
	}
	validatedKey, err := core.NewBusinessIDEncodingKey(key.Version(), key.Bytes())
	if err != nil {
		return "", err
	}
	tweak := businessIDPermutationV1Tweak(validatedScope, validatedKey.Version())
	keyBytes := validatedKey.Bytes()
	candidate := sequence
	for {
		candidate = uint64(businessIDPermutationV1Permute32(uint32(candidate), tweak, keyBytes))
		if candidate < BusinessIDPermutationV1DomainSize {
			break
		}
	}
	encoded := make([]byte, BusinessIDPermutationV1Width)
	for index := len(encoded) - 1; index >= 0; index-- {
		encoded[index] = BusinessIDPermutationV1Alphabet[candidate%36]
		candidate /= 36
	}
	return string(encoded), nil
}

func businessIDPermutationV1Permute32(value uint32, tweak, key []byte) uint32 {
	left := uint16(value >> 16)
	right := uint16(value)
	for round := byte(0); round < 8; round++ {
		mac := hmac.New(sha256.New, key)
		_, _ = mac.Write(tweak)
		_, _ = mac.Write([]byte{round, byte(right >> 8), byte(right)})
		digest := mac.Sum(nil)
		output := binary.BigEndian.Uint16(digest[:2])
		left, right = right, left^output
	}
	return uint32(left)<<16 | uint32(right)
}

func businessIDPermutationV1Tweak(scope core.BusinessIDScope, keyVersion uint32) []byte {
	capacity := len(businessIDPermutationV1Magic) + 5 + len(scope.DomainRootKey()) + len(scope.AggregateType()) + len(scope.Namespace()) + len(scope.PeriodKey()) + 16
	result := make([]byte, 0, capacity)
	result = append(result, businessIDPermutationV1Magic...)
	result = append(result, 1)
	var number [4]byte
	binary.BigEndian.PutUint32(number[:], keyVersion)
	result = append(result, number[:]...)
	for _, value := range []string{scope.DomainRootKey(), scope.AggregateType(), scope.Namespace(), scope.PeriodKey()} {
		binary.BigEndian.PutUint32(number[:], uint32(len([]byte(value))))
		result = append(result, number[:]...)
		result = append(result, []byte(value)...)
	}
	return result
}
