package runtime

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"

	"github.com/teaql/teaql-golang/core"
)

const (
	BusinessIDPermutationV1Width       = 6
	BusinessIDPermutationV1DomainSize  = uint64(2_176_782_336)
	BusinessIDPermutationV1MaxSequence = BusinessIDPermutationV1DomainSize - 1
	BusinessIDPermutationV1Alphabet    = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
)

var businessIDPermutationV1Magic = []byte("teaql-business-id-fp-v1\x00")

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
