package blobreference

import "errors"

// RetentionPolicy is a validated host-selected input. SQLite upload/binding
// integration must snapshot and recheck it on their physical SQL owners.
type RetentionPolicy struct {
	profile                      string
	uploadMode                   string
	unboundRetentionSeconds      int64
	unreferencedRetentionSeconds int64
	minimumRetentionSeconds      int64
}

func (p RetentionPolicy) Profile() string                     { return p.profile }
func (p RetentionPolicy) UploadMode() string                  { return p.uploadMode }
func (p RetentionPolicy) UnboundRetentionSeconds() int64      { return p.unboundRetentionSeconds }
func (p RetentionPolicy) UnreferencedRetentionSeconds() int64 { return p.unreferencedRetentionSeconds }
func (p RetentionPolicy) MinimumRetentionSeconds() int64      { return p.minimumRetentionSeconds }

func (p RetentionPolicy) Valid() bool {
	return validScope(p.profile) && (p.uploadMode == "server_stream" || p.uploadMode == "direct") &&
		validRetentionSeconds(p.unboundRetentionSeconds) &&
		validRetentionSeconds(p.unreferencedRetentionSeconds) &&
		validRetentionSeconds(p.minimumRetentionSeconds)
}

const maxRetentionEpochMillis int64 = 8_640_000_000_000_000

func validRetentionSeconds(seconds int64) bool {
	return seconds >= 0 && seconds <= maxRetentionEpochMillis/1000
}

func NewRetentionPolicy(profile, uploadMode string, unbound, unreferenced, minimum int64) (RetentionPolicy, error) {
	if !validScope(profile) || (uploadMode != "server_stream" && uploadMode != "direct") ||
		!validRetentionSeconds(unbound) || !validRetentionSeconds(unreferenced) ||
		!validRetentionSeconds(minimum) {
		return RetentionPolicy{}, errors.New("BLOB_POLICY_INVALID")
	}
	return RetentionPolicy{profile, uploadMode, unbound, unreferenced, minimum}, nil
}

func validRetentionTime(millis int64) bool {
	return millis >= 0 && millis <= maxRetentionEpochMillis
}

func retentionDeadline(baseMillis, seconds int64) (int64, error) {
	if !validRetentionTime(baseMillis) {
		return 0, errors.New("BLOB_REFERENCE_INVALID")
	}
	if !validRetentionSeconds(seconds) || seconds > (maxRetentionEpochMillis-baseMillis)/1000 {
		return 0, errors.New("BLOB_POLICY_INVALID")
	}
	return baseMillis + seconds*1000, nil
}

func EffectiveRetention(accepted []int64, proposed int64) (int64, error) {
	if !validRetentionSeconds(proposed) {
		return 0, errors.New("BLOB_POLICY_INVALID")
	}
	for _, duration := range accepted {
		if !validRetentionSeconds(duration) {
			return 0, errors.New("BLOB_POLICY_INVALID")
		}
		if duration > proposed {
			proposed = duration
		}
	}
	return proposed, nil
}

func UnreferencedEligibleAt(lastUnboundMillis, seconds int64, floorUntilMillis *int64) (int64, error) {
	eligible, err := retentionDeadline(lastUnboundMillis, seconds)
	if err != nil {
		return 0, err
	}
	if floorUntilMillis != nil {
		if !validRetentionTime(*floorUntilMillis) {
			return 0, errors.New("BLOB_REFERENCE_INVALID")
		}
		if *floorUntilMillis > eligible {
			eligible = *floorUntilMillis
		}
	}
	return eligible, nil
}

func RequireBindable(status string, nowMillis int64, completedMillis *int64,
	unboundSeconds int64, unreferencedEligibleMillis *int64) error {
	if !validRetentionTime(nowMillis) {
		return errors.New("BLOB_REFERENCE_INVALID")
	}
	switch status {
	case "initiated":
		return errors.New("BLOB_UPLOAD_INCOMPLETE")
	case "deleting", "deleted":
		return errors.New("BLOB_REFERENCE_EXPIRED")
	case "uploaded_unbound":
		if completedMillis == nil {
			return errors.New("BLOB_REFERENCE_INVALID")
		}
		deadline, err := retentionDeadline(*completedMillis, unboundSeconds)
		if err != nil {
			return err
		}
		if nowMillis >= deadline {
			return errors.New("BLOB_REFERENCE_EXPIRED")
		}
	case "available":
		if unreferencedEligibleMillis != nil {
			if !validRetentionTime(*unreferencedEligibleMillis) {
				return errors.New("BLOB_REFERENCE_INVALID")
			}
			if nowMillis >= *unreferencedEligibleMillis {
				return errors.New("BLOB_REFERENCE_EXPIRED")
			}
		}
	default:
		return errors.New("BLOB_REFERENCE_INVALID")
	}
	return nil
}

func RetentionCollectable(status string, bindingCount int64, nowMillis int64,
	uploadExpiresMillis, completedMillis *int64, unboundSeconds int64,
	unreferencedEligibleMillis *int64) (bool, error) {
	if !validRetentionTime(nowMillis) || bindingCount < 0 {
		return false, errors.New("BLOB_REFERENCE_INVALID")
	}
	if bindingCount > 0 || status == "deleted" {
		return false, nil
	}
	var deadline int64
	switch status {
	case "initiated":
		if uploadExpiresMillis == nil || !validRetentionTime(*uploadExpiresMillis) {
			return false, errors.New("BLOB_REFERENCE_INVALID")
		}
		deadline = *uploadExpiresMillis
	case "uploaded_unbound":
		if completedMillis == nil {
			return false, errors.New("BLOB_REFERENCE_INVALID")
		}
		var err error
		deadline, err = retentionDeadline(*completedMillis, unboundSeconds)
		if err != nil {
			return false, err
		}
	case "available":
		if unreferencedEligibleMillis == nil {
			return false, nil
		}
		if !validRetentionTime(*unreferencedEligibleMillis) {
			return false, errors.New("BLOB_REFERENCE_INVALID")
		}
		deadline = *unreferencedEligibleMillis
	case "deleting":
		return true, nil
	default:
		return false, errors.New("BLOB_REFERENCE_INVALID")
	}
	return nowMillis >= deadline, nil
}
