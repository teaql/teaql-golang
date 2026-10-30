//go:build !linux

package blobreference

import "io"

// LegacyOfficeContentVerifier requires the Linux bubblewrap sandbox.
type LegacyOfficeContentVerifier struct{}

func NewLegacyOfficeContentVerifier(_, _, _, _ string, _, _ int64,
	_ int) (*LegacyOfficeContentVerifier, error) {
	return nil, uploadError("BLOB_CAPABILITY_MISSING")
}

func (*LegacyOfficeContentVerifier) ValidateReady() error {
	return uploadError("BLOB_CAPABILITY_MISSING")
}

func (*LegacyOfficeContentVerifier) Verify(io.Reader, string, int64) (VerifiedContent, error) {
	return VerifiedContent{}, uploadError("BLOB_CAPABILITY_MISSING")
}
