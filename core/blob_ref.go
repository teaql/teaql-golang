package core

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"regexp"
	"sort"
	"strings"
)

var blobKeyPattern = regexp.MustCompile(`^blob_[A-Za-z0-9_-]{1,123}$`)
var errBlobReferenceInvalid = errors.New("BLOB_REFERENCE_INVALID")

// BlobRef stores a stable catalog Key. Possession does not authorize reads or binding.
type BlobRef struct{ key string }

func ParseBlobRef(key string) (BlobRef, error) {
	if !blobKeyPattern.MatchString(key) {
		return BlobRef{}, errBlobReferenceInvalid
	}
	return BlobRef{key: key}, nil
}

func (r BlobRef) Key() string { return r.key }

func (r BlobRef) String() string { return "BlobRef[masked]" }

// Generic JSON output must pass through an authorized prepared view.
func (r BlobRef) MarshalJSON() ([]byte, error) {
	return nil, errors.New("BLOB_PRESENTATION_REQUIRED")
}

func (r *BlobRef) UnmarshalJSON(data []byte) error {
	var key string
	if err := json.Unmarshal(data, &key); err != nil {
		return errBlobReferenceInvalid
	}
	parsed, err := ParseBlobRef(key)
	if err != nil {
		return err
	}
	*r = parsed
	return nil
}

func (r BlobRef) Value() (driver.Value, error) {
	if !blobKeyPattern.MatchString(r.key) {
		return nil, errBlobReferenceInvalid
	}
	return r.key, nil
}

func (r *BlobRef) Scan(value any) error {
	var key string
	switch v := value.(type) {
	case string:
		key = v
	case []byte:
		key = string(v)
	default:
		return errBlobReferenceInvalid
	}
	parsed, err := ParseBlobRef(key)
	if err != nil {
		return err
	}
	*r = parsed
	return nil
}

var blobMIME = map[string]string{
	"jpg": "image/jpeg", "jpeg": "image/jpeg", "png": "image/png",
	"mp4": "video/mp4", "mp3": "audio/mpeg",
	"doc":  "application/msword",
	"docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	"xls":  "application/vnd.ms-excel",
	"xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	"pdf":  "application/pdf",
}

// BlobFieldSpec records trusted model format restrictions, not upload authority.
type BlobFieldSpec struct {
	Entity, Field, PolicyProfile string
	Nullable                     bool
	extensions                   []string
	contentTypes                 []string
}

func NewBlobFieldSpec(entity, field string, extensions []string, nullable bool, profile string) (BlobFieldSpec, error) {
	if entity == "" || field == "" || profile == "" ||
		entity != strings.TrimSpace(entity) || field != strings.TrimSpace(field) ||
		profile != strings.TrimSpace(profile) {
		return BlobFieldSpec{}, errors.New("BLOB_POLICY_INVALID")
	}
	unique := make(map[string]struct{}, len(extensions))
	mimes := make(map[string]struct{}, len(extensions))
	for _, extension := range extensions {
		mime, ok := blobMIME[extension]
		if !ok {
			return BlobFieldSpec{}, errors.New("BLOB_FORMAT_NOT_ALLOWED")
		}
		unique[extension] = struct{}{}
		mimes[mime] = struct{}{}
	}
	if len(unique) == 0 {
		return BlobFieldSpec{}, errors.New("BLOB_FORMAT_NOT_ALLOWED")
	}
	spec := BlobFieldSpec{Entity: entity, Field: field, Nullable: nullable, PolicyProfile: profile}
	for extension := range unique {
		spec.extensions = append(spec.extensions, extension)
	}
	for mime := range mimes {
		spec.contentTypes = append(spec.contentTypes, mime)
	}
	sort.Strings(spec.extensions)
	sort.Strings(spec.contentTypes)
	return spec, nil
}

func (s BlobFieldSpec) ProtocolVersion() string     { return "blob.v1" }
func (s BlobFieldSpec) MaxFiles() int               { return 1 }
func (s BlobFieldSpec) AllowedExtensions() []string { return append([]string(nil), s.extensions...) }
func (s BlobFieldSpec) AllowedContentTypes() []string {
	return append([]string(nil), s.contentTypes...)
}
