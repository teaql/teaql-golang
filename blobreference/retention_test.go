package blobreference

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"testing"
)

type policyVector struct {
	ID       string                     `json:"id"`
	Input    map[string]json.RawMessage `json:"input"`
	Expected map[string]json.RawMessage `json:"expected"`
	Error    string                     `json:"error"`
}

func sharedPolicyVectors(t *testing.T) map[string]policyVector {
	t.Helper()
	data, err := os.ReadFile("testdata/blob-policy-cases.json")
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(data)
	if hex.EncodeToString(digest[:]) != "b93cff3f8dabd5690aac5f8f2096633dcd4e158d3aa0424b9a902a16f6d3c2b6" {
		t.Fatal("shared policy fixture drift")
	}
	var document struct {
		Schema string         `json:"schema"`
		Cases  []policyVector `json:"cases"`
	}
	if err := json.Unmarshal(data, &document); err != nil || document.Schema != "teaql.blob-policy-cases/v1" {
		t.Fatalf("invalid shared policy fixture: %v", err)
	}
	vectors := make(map[string]policyVector, len(document.Cases))
	for _, vector := range document.Cases {
		vectors[vector.ID] = vector
	}
	return vectors
}

func vectorInt(t *testing.T, values map[string]json.RawMessage, field string) int64 {
	t.Helper()
	var result int64
	if err := json.Unmarshal(values[field], &result); err != nil {
		t.Fatalf("missing/invalid %s: %v", field, err)
	}
	return result
}

func expectRetentionError(t *testing.T, err error, code string) {
	t.Helper()
	if err == nil || err.Error() != code {
		t.Fatalf("expected %s, got %v", code, err)
	}
}

func TestSharedRetentionDeadlinesAndProtection(t *testing.T) {
	vectors := sharedPolicyVectors(t)
	upload := vectors["upload-clock-independent"]
	completed := vectorInt(t, upload.Input, "completedAtSeconds") * 1000
	eligible, err := UnreferencedEligibleAt(completed, 86400, nil)
	if err != nil || eligible != vectorInt(t, upload.Expected, "unboundEligibleAtSeconds")*1000 {
		t.Fatalf("completion deadline: %d, %v", eligible, err)
	}
	unbound := vectors["unbound-deadline-rejects-bind-before-gc"]
	completed = vectorInt(t, unbound.Input, "completedAtSeconds") * 1000
	now := vectorInt(t, unbound.Input, "nowSeconds") * 1000
	seconds := vectorInt(t, unbound.Input, "unboundRetentionSeconds")
	expectRetentionError(t, RequireBindable("uploaded_unbound", now, &completed, seconds, nil), unbound.Error)
	if err := RequireBindable("uploaded_unbound", now-1, &completed, seconds, nil); err != nil {
		t.Fatal(err)
	}
	active := vectors["active-reference-never-collected"]
	collectable, err := RetentionCollectable("available", vectorInt(t, active.Input, "bindingCount"),
		vectorInt(t, active.Input, "nowSeconds")*1000, nil, nil, 0, &now)
	if err != nil || collectable {
		t.Fatalf("active binding collected: %v, %v", collectable, err)
	}
	deleting := vectors["deleting-rejects-binding"]
	expectRetentionError(t, RequireBindable("deleting", 0, nil, 0, nil), deleting.Error)
	if collectable, err = RetentionCollectable("uploaded_unbound", 0, now, nil, nil, 0, nil); err == nil || collectable {
		t.Fatalf("missing completion time accepted: %v, %v", collectable, err)
	}
}

func TestSharedMaximumRetentionAndFloor(t *testing.T) {
	vectors := sharedPolicyVectors(t)
	maximum := vectors["maximum-retention-across-fields"]
	var accepted []int64
	if err := json.Unmarshal(maximum.Input["acceptedRetentionSeconds"], &accepted); err != nil {
		t.Fatal(err)
	}
	effective, err := EffectiveRetention(accepted, 3600)
	if err != nil || effective != vectorInt(t, maximum.Expected, "effectiveUnreferencedRetentionSeconds") {
		t.Fatalf("maximum retention: %d, %v", effective, err)
	}
	floor := vectorInt(t, maximum.Input, "retentionFloorUntilSeconds") * 1000
	deadline, err := UnreferencedEligibleAt(vectorInt(t, maximum.Input, "lastUnboundAtSeconds")*1000,
		effective, &floor)
	if err != nil || deadline != vectorInt(t, maximum.Expected, "unreferencedEligibleAtSeconds")*1000 {
		t.Fatalf("retention floor: %d, %v", deadline, err)
	}
	profile := vectors["profile-update-does-not-shorten"]
	effective, err = EffectiveRetention([]int64{
		vectorInt(t, profile.Input, "acceptedUnreferencedRetentionSeconds")},
		vectorInt(t, profile.Input, "currentProfileRetentionSeconds"))
	if err != nil || effective != vectorInt(t, profile.Expected, "effectiveUnreferencedRetentionSeconds") {
		t.Fatalf("shorter profile reduced retention: %d, %v", effective, err)
	}
	if _, err := UnreferencedEligibleAt(maxRetentionEpochMillis, 1, nil); err == nil || err.Error() != "BLOB_POLICY_INVALID" {
		t.Fatalf("overflow accepted: %v", err)
	}
	if _, err := NewRetentionPolicy(" default", "server_stream", 1, 1, 0); err == nil || err.Error() != "BLOB_POLICY_INVALID" {
		t.Fatalf("invalid profile accepted: %v", err)
	}
}
