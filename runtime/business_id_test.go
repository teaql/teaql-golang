package runtime

import (
	"encoding/csv"
	"encoding/hex"
	"errors"
	"os"
	"regexp"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teaql/teaql-golang/core"
)

func TestBusinessIDPermutationV1GoldenVectors(t *testing.T) {
	input, err := os.Open("testdata/business-id-permutation-v1.csv")
	require.NoError(t, err)
	defer input.Close()
	rows, err := csv.NewReader(input).ReadAll()
	require.NoError(t, err)
	require.Len(t, rows, 11)
	require.Equal(t, []string{"case_id", "key_hex", "key_version", "domain_root_key", "aggregate_type", "namespace", "period_key", "sequence", "expected_code", "expected_business_id"}, rows[0])
	for _, row := range rows[1:] {
		t.Run(row[0], func(t *testing.T) {
			keyBytes, decodeErr := hex.DecodeString(row[1])
			require.NoError(t, decodeErr)
			version, parseErr := strconv.ParseUint(row[2], 10, 32)
			require.NoError(t, parseErr)
			sequence, parseErr := strconv.ParseUint(row[7], 10, 64)
			require.NoError(t, parseErr)
			scope, scopeErr := core.NewBusinessIDScope(row[3], row[4], row[5], row[6])
			require.NoError(t, scopeErr)
			key, keyErr := core.NewBusinessIDEncodingKey(uint32(version), keyBytes)
			require.NoError(t, keyErr)
			actual, encodeErr := EncodeBusinessIDPermutationV1(sequence, scope, key)
			require.NoError(t, encodeErr)
			require.Equal(t, row[8], actual)
			require.Equal(t, row[9], "ORD-"+row[6]+"-"+actual)
		})
	}
}

func TestBusinessIDPermutationV1IsDeterministicUniqueAndCanonical(t *testing.T) {
	scope, err := core.NewBusinessIDScope("tenant-a", "commerce_order", "order_number", "20260925")
	require.NoError(t, err)
	bytes, err := hex.DecodeString("000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f")
	require.NoError(t, err)
	key, err := core.NewBusinessIDEncodingKey(1, bytes)
	require.NoError(t, err)
	values := make(map[string]struct{}, 20_000)
	canonical := regexp.MustCompile(`^[0-9A-Z]{6}$`)
	for sequence := uint64(0); sequence < 20_000; sequence++ {
		first, encodeErr := EncodeBusinessIDPermutationV1(sequence, scope, key)
		require.NoError(t, encodeErr)
		afterRestart, encodeErr := EncodeBusinessIDPermutationV1(sequence, scope, key)
		require.NoError(t, encodeErr)
		require.Equal(t, first, afterRestart)
		require.True(t, canonical.MatchString(first))
		_, exists := values[first]
		require.Falsef(t, exists, "duplicate at sequence %d", sequence)
		values[first] = struct{}{}
	}
}

func TestBusinessIDPermutationV1RejectsInvalidInputs(t *testing.T) {
	scope, err := core.NewBusinessIDScope("tenant-a", "commerce_order", "order_number", "20260925")
	require.NoError(t, err)
	key, err := core.NewBusinessIDEncodingKey(1, make([]byte, 32))
	require.NoError(t, err)
	_, err = EncodeBusinessIDPermutationV1(BusinessIDPermutationV1DomainSize, scope, key)
	var businessErr *core.BusinessIDError
	require.True(t, errors.As(err, &businessErr))
	require.Equal(t, core.BusinessIDRangeExhausted, businessErr.Code)
	_, err = core.NewBusinessIDEncodingKey(0, make([]byte, 32))
	require.Error(t, err)
	_, err = core.NewBusinessIDEncodingKey(1, make([]byte, 31))
	require.Error(t, err)
	_, err = core.NewBusinessIDScope(" ", "commerce_order", "order_number", "20260925")
	require.Error(t, err)
	_, err = EncodeBusinessIDPermutationV1(0, core.BusinessIDScope{}, core.BusinessIDEncodingKey{})
	require.Error(t, err)
}
