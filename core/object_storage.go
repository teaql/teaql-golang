package core

import (
	"fmt"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const MaxObjectPutSize int64 = 5 * 1024 * 1024 * 1024

type StorageErrorCode string

const (
	StorageInvalidInput  StorageErrorCode = "InvalidInput"
	StorageNotConfigured StorageErrorCode = "NotConfigured"
	StorageNotFound      StorageErrorCode = "NotFound"
	StorageAccessDenied  StorageErrorCode = "AccessDenied"
	StorageConflict      StorageErrorCode = "Conflict"
	StorageTimeout       StorageErrorCode = "Timeout"
	StorageUnavailable   StorageErrorCode = "Unavailable"
	StorageUnsupported   StorageErrorCode = "Unsupported"
	StorageTooLarge      StorageErrorCode = "TooLarge"
	StorageIO            StorageErrorCode = "Io"
)

// No wrapped provider error: Error and diagnostic formatting contain only a code.
type StorageError struct{ Code StorageErrorCode }

func (e *StorageError) Error() string { return "Object storage: " + string(e.Code) }

type ObjectKey struct{ value string }

func NewObjectKey(value string) (ObjectKey, error) {
	if value == "" || len(value) > 1024 || !utf8.ValidString(value) || strings.IndexFunc(value, unicode.IsControl) >= 0 {
		return ObjectKey{}, &StorageError{StorageInvalidInput}
	}
	return ObjectKey{value}, nil
}
func (k ObjectKey) Value() string    { return k.value }
func (k ObjectKey) Valid() bool      { return k.value != "" }
func (k ObjectKey) String() string   { return "ObjectKey[masked]" }
func (k ObjectKey) GoString() string { return k.String() }

type ObjectWriteMode string

const (
	ObjectCreateOnly ObjectWriteMode = "create_only"
	ObjectReplace    ObjectWriteMode = "replace"
)

type ObjectPutOptions struct {
	size        int64
	mode        ObjectWriteMode
	contentType string
	metadata    map[string]string
}

func NewObjectPutOptions(size int64, mode ObjectWriteMode) (ObjectPutOptions, error) {
	if size < 0 || (mode != ObjectCreateOnly && mode != ObjectReplace) {
		return ObjectPutOptions{}, &StorageError{StorageInvalidInput}
	}
	if size > MaxObjectPutSize {
		return ObjectPutOptions{}, &StorageError{StorageTooLarge}
	}
	return ObjectPutOptions{size, mode, "application/octet-stream", nil}, nil
}
func objectHeader(value string) bool {
	if value == "" {
		return false
	}
	for _, c := range value {
		if c < 32 || c > 126 {
			return false
		}
	}
	return true
}
func cloneObjectMetadata(values map[string]string) map[string]string {
	result := make(map[string]string, len(values))
	for k, v := range values {
		result[k] = v
	}
	return result
}
func (o ObjectPutOptions) WithContentType(value string) (ObjectPutOptions, error) {
	if !objectHeader(value) {
		return o, &StorageError{StorageInvalidInput}
	}
	o.contentType = value
	return o, nil
}
func (o ObjectPutOptions) WithMetadata(values map[string]string) (ObjectPutOptions, error) {
	for k, v := range values {
		if k == "" || k[0] == '-' || !objectHeader(v) {
			return o, &StorageError{StorageInvalidInput}
		}
		for _, c := range k {
			if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-') {
				return o, &StorageError{StorageInvalidInput}
			}
		}
	}
	o.metadata = cloneObjectMetadata(values)
	return o, nil
}
func (o ObjectPutOptions) Size() int64                 { return o.size }
func (o ObjectPutOptions) Mode() ObjectWriteMode       { return o.mode }
func (o ObjectPutOptions) ContentType() string         { return o.contentType }
func (o ObjectPutOptions) Metadata() map[string]string { return cloneObjectMetadata(o.metadata) }
func (o ObjectPutOptions) Valid() bool                 { return o.mode == ObjectCreateOnly || o.mode == ObjectReplace }
func (o ObjectPutOptions) String() string {
	return fmt.Sprintf("ObjectPutOptions[size=%d, mode=%s, metadata=masked]", o.size, o.mode)
}
func (o ObjectPutOptions) GoString() string { return o.String() }

type ObjectMetadata struct {
	key               ObjectKey
	size              int64
	contentType, etag string
	lastModified      *time.Time
	metadata          map[string]string
}

func NewObjectMetadata(key ObjectKey, size int64, contentType, etag string, lastModified *time.Time, metadata map[string]string) (*ObjectMetadata, error) {
	if size < 0 {
		return nil, &StorageError{StorageUnavailable}
	}
	var modified *time.Time
	if lastModified != nil {
		copy := *lastModified
		modified = &copy
	}
	return &ObjectMetadata{key, size, contentType, etag, modified, cloneObjectMetadata(metadata)}, nil
}
func (m ObjectMetadata) Key() ObjectKey      { return m.key }
func (m ObjectMetadata) Size() int64         { return m.size }
func (m ObjectMetadata) ContentType() string { return m.contentType }
func (m ObjectMetadata) ETag() string        { return m.etag }
func (m ObjectMetadata) LastModified() *time.Time {
	if m.lastModified == nil {
		return nil
	}
	copy := *m.lastModified
	return &copy
}
func (m ObjectMetadata) CustomMetadata() map[string]string { return cloneObjectMetadata(m.metadata) }
func (m ObjectMetadata) String() string {
	return fmt.Sprintf("ObjectMetadata[size=%d, details=masked]", m.size)
}
func (m ObjectMetadata) GoString() string { return m.String() }

type ObjectReadURLOptions struct {
	lifetime    time.Duration
	disposition string
}

func DefaultObjectReadURLOptions() ObjectReadURLOptions {
	return ObjectReadURLOptions{lifetime: 5 * time.Minute}
}
func NewObjectReadURLOptions(lifetime time.Duration, disposition, filename string) (ObjectReadURLOptions, error) {
	if lifetime <= 0 || lifetime > 7*24*time.Hour || lifetime%time.Second != 0 || (disposition != "" && disposition != "inline" && disposition != "attachment") {
		return ObjectReadURLOptions{}, &StorageError{StorageInvalidInput}
	}
	if filename != "" {
		if disposition == "" || len(filename) > 1024 || !utf8.ValidString(filename) ||
			strings.IndexFunc(filename, unicode.IsControl) >= 0 || strings.ContainsAny(filename, "/\\\"") {
			return ObjectReadURLOptions{}, &StorageError{StorageInvalidInput}
		}
		var fallback strings.Builder
		ascii := true
		for _, character := range filename {
			if character > 126 {
				ascii = false
				fallback.WriteByte('_')
			} else {
				fallback.WriteRune(character)
			}
		}
		if ascii {
			disposition += "; filename=\"" + filename + "\""
		} else {
			const digits = "0123456789ABCDEF"
			var encoded strings.Builder
			for i := 0; i < len(filename); i++ {
				b := filename[i]
				if b >= 'A' && b <= 'Z' || b >= 'a' && b <= 'z' ||
					b >= '0' && b <= '9' || b == '-' || b == '.' || b == '_' || b == '~' {
					encoded.WriteByte(b)
				} else {
					encoded.WriteByte('%')
					encoded.WriteByte(digits[b>>4])
					encoded.WriteByte(digits[b&15])
				}
			}
			disposition += "; filename=\"" + fallback.String() + "\"; filename*=UTF-8''" + encoded.String()
		}
	}
	return ObjectReadURLOptions{lifetime, disposition}, nil
}
func (o ObjectReadURLOptions) Lifetime() time.Duration    { return o.lifetime }
func (o ObjectReadURLOptions) ContentDisposition() string { return o.disposition }
func (o ObjectReadURLOptions) String() string             { return "ObjectReadURLOptions[details=masked]" }
func (o ObjectReadURLOptions) GoString() string           { return o.String() }

type ObjectReadURL struct {
	url       string
	expiresAt time.Time
}

func NewObjectReadURL(url string, expiresAt time.Time) ObjectReadURL {
	return ObjectReadURL{url, expiresAt}
}
func (u ObjectReadURL) URL() string          { return u.url }
func (u ObjectReadURL) ExpiresAt() time.Time { return u.expiresAt }
func (u ObjectReadURL) String() string       { return "ObjectReadURL[url=masked]" }
func (u ObjectReadURL) GoString() string     { return u.String() }
