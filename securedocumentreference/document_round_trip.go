package securedocumentreference

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"encoding/base64"
	"encoding/binary"
	"io"
	"math"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

const maxDocumentToken = 256000

type DocumentEntitySnapshot struct {
	RowKey                                         string
	Identity                                       ReferenceIdentity
	LoadedFields, WritableFields, AllowedMutations []string
}
type DocumentRelationSnapshot struct {
	Name             string
	RowKeys          []string
	Completeness     string
	Writable         bool
	AllowedMutations []string
}
type DocumentSnapshot struct {
	ModelFingerprint, BusinessID string
	Scope                        ReferenceDocumentScope
	Aggregate                    ReferenceIdentity
	Entities                     []DocumentEntitySnapshot
	Relations                    []DocumentRelationSnapshot
}
type DocumentBlobBindingSnapshot struct {
	RowKey, CanonicalField, EntityType string
	StableKey                          *string
	BindingVersion                     uint64
	Nullable, Writable                 bool
}
type BoundDocumentSnapshot struct {
	Snapshot DocumentSnapshot
	Bindings []DocumentBlobBindingSnapshot
}
type VerifiedDocument[T any] struct {
	Document            T
	KeyID               string
	IssuedAt, ExpiresAt time.Time
}

func documentViolation() error { return &ReferenceError{Code: "DOCUMENT_PROJECTION_VIOLATION"} }
func docText(s string) error {
	if len(s) == 0 || len(s) > 16384 || strings.TrimSpace(s) == "" || !utf8.ValidString(s) {
		return invalid()
	}
	return nil
}
func putDocText(b *bytes.Buffer, s string) error {
	if err := docText(s); err != nil {
		return err
	}
	putCount(b, len(s))
	b.WriteString(s)
	return nil
}
func putCount(b *bytes.Buffer, n int) {
	var data [4]byte
	binary.BigEndian.PutUint32(data[:], uint32(n))
	b.Write(data[:])
}
func putBool(b *bytes.Buffer, v bool) {
	if v {
		b.WriteByte(1)
	} else {
		b.WriteByte(0)
	}
}
func putScope(b *bytes.Buffer, s ReferenceDocumentScope) error {
	if err := validateScope(s); err != nil {
		return err
	}
	for _, v := range []string{s.DocumentID, s.Purpose, s.AggregateType} {
		if err := putDocText(b, v); err != nil {
			return err
		}
	}
	writeU64(b, s.AggregateID)
	writeI64(b, s.AggregateRevision)
	return nil
}
func putIdentity(b *bytes.Buffer, v ReferenceIdentity) error {
	if err := validateIdentity(v); err != nil {
		return err
	}
	if err := putDocText(b, v.EntityType); err != nil {
		return err
	}
	writeU64(b, v.ID)
	writeI64(b, v.Version)
	return nil
}
func normalizedSet(values []string, maximum int) ([]string, error) {
	if len(values) > maximum {
		return nil, documentViolation()
	}
	copyValues := append([]string(nil), values...)
	sort.Strings(copyValues)
	for i, value := range copyValues {
		if err := docText(value); err != nil {
			return nil, documentViolation()
		}
		if i > 0 && value == copyValues[i-1] {
			return nil, documentViolation()
		}
	}
	return copyValues, nil
}
func putSet(b *bytes.Buffer, values []string, maximum int) error {
	ordered, err := normalizedSet(values, maximum)
	if err != nil {
		return err
	}
	putCount(b, len(ordered))
	for _, v := range ordered {
		if err := putDocText(b, v); err != nil {
			return err
		}
	}
	return nil
}

var mutationCode = map[string]byte{"UPDATE": 1, "CREATE_CHILD": 2, "REMOVE_CHILD": 3}
var mutationName = map[byte]string{1: "UPDATE", 2: "CREATE_CHILD", 3: "REMOVE_CHILD"}

func putMutations(b *bytes.Buffer, values []string) error {
	if len(values) > 3 {
		return documentViolation()
	}
	seen := map[byte]bool{}
	codes := make([]int, 0, len(values))
	for _, v := range values {
		code := mutationCode[v]
		if code == 0 || seen[code] {
			return documentViolation()
		}
		seen[code] = true
		codes = append(codes, int(code))
	}
	sort.Ints(codes)
	putCount(b, len(codes))
	for _, v := range codes {
		b.WriteByte(byte(v))
	}
	return nil
}
func includes(values []string, value string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}

func encodeSnapshot(snapshot DocumentSnapshot, version byte, actor []byte, issued, expires int64) ([]byte, error) {
	if err := validateScope(snapshot.Scope); err != nil {
		return nil, err
	}
	if err := validateIdentity(snapshot.Aggregate); err != nil {
		return nil, err
	}
	if snapshot.Aggregate.EntityType != snapshot.Scope.AggregateType || snapshot.Aggregate.ID != snapshot.Scope.AggregateID || snapshot.Aggregate.Version != snapshot.Scope.AggregateRevision || len(snapshot.Entities) > 10000 || len(snapshot.Relations) > 1024 {
		return nil, documentViolation()
	}
	var b bytes.Buffer
	b.WriteByte(version)
	b.Write(actor)
	writeI64(&b, issued)
	writeI64(&b, expires)
	for _, v := range []string{snapshot.ModelFingerprint, snapshot.BusinessID} {
		if err := putDocText(&b, v); err != nil {
			return nil, err
		}
	}
	if err := putScope(&b, snapshot.Scope); err != nil {
		return nil, err
	}
	if err := putIdentity(&b, snapshot.Aggregate); err != nil {
		return nil, err
	}
	rows := map[string]DocumentEntitySnapshot{}
	putCount(&b, len(snapshot.Entities))
	for _, entity := range snapshot.Entities {
		if _, exists := rows[entity.RowKey]; exists || len(entity.AllowedMutations) == 0 {
			return nil, documentViolation()
		}
		rows[entity.RowKey] = entity
		for _, field := range entity.WritableFields {
			if !includes(entity.LoadedFields, field) {
				return nil, documentViolation()
			}
		}
		if err := putDocText(&b, entity.RowKey); err != nil {
			return nil, err
		}
		if err := putIdentity(&b, entity.Identity); err != nil {
			return nil, err
		}
		if err := putSet(&b, entity.LoadedFields, 1024); err != nil {
			return nil, err
		}
		if err := putSet(&b, entity.WritableFields, 1024); err != nil {
			return nil, err
		}
		if err := putMutations(&b, entity.AllowedMutations); err != nil {
			return nil, err
		}
	}
	putCount(&b, len(snapshot.Relations))
	for _, relation := range snapshot.Relations {
		if len(relation.RowKeys) > 10000 || (!relation.Writable && len(relation.AllowedMutations) != 0) {
			return nil, documentViolation()
		}
		if err := putDocText(&b, relation.Name); err != nil {
			return nil, err
		}
		switch relation.Completeness {
		case "COMPLETE":
			b.WriteByte(1)
		case "PARTIAL":
			b.WriteByte(2)
		default:
			return nil, documentViolation()
		}
		putBool(&b, relation.Writable)
		if err := putMutations(&b, relation.AllowedMutations); err != nil {
			return nil, err
		}
		putCount(&b, len(relation.RowKeys))
		seen := map[string]bool{}
		for _, key := range relation.RowKeys {
			if _, ok := rows[key]; !ok || seen[key] {
				return nil, documentViolation()
			}
			seen[key] = true
			if err := putDocText(&b, key); err != nil {
				return nil, err
			}
		}
	}
	return b.Bytes(), nil
}

type docReader struct {
	data   []byte
	offset int
}

func (r *docReader) take(n int) ([]byte, error) {
	if n < 0 || n > len(r.data)-r.offset {
		return nil, invalid()
	}
	v := r.data[r.offset : r.offset+n]
	r.offset += n
	return v, nil
}
func (r *docReader) byte() (byte, error) {
	v, e := r.take(1)
	if e != nil {
		return 0, e
	}
	return v[0], nil
}
func (r *docReader) count(max int) (int, error) {
	v, e := r.take(4)
	if e != nil {
		return 0, e
	}
	n := binary.BigEndian.Uint32(v)
	if n > uint32(max) {
		return 0, invalid()
	}
	return int(n), nil
}
func (r *docReader) text() (string, error) {
	n, e := r.count(16384)
	if e != nil {
		return "", e
	}
	v, e := r.take(n)
	if e != nil {
		return "", e
	}
	s := string(v)
	return s, docText(s)
}
func (r *docReader) u64() (uint64, error) {
	v, e := r.take(8)
	if e != nil {
		return 0, e
	}
	return binary.BigEndian.Uint64(v), nil
}
func (r *docReader) i64() (int64, error) { v, e := r.u64(); return int64(v), e }
func (r *docReader) flag() (bool, error) {
	v, e := r.byte()
	if e != nil || v > 1 {
		return false, invalid()
	}
	return v == 1, nil
}
func (r *docReader) scope() (ReferenceDocumentScope, error) {
	var s ReferenceDocumentScope
	var e error
	if s.DocumentID, e = r.text(); e != nil {
		return s, e
	}
	if s.Purpose, e = r.text(); e != nil {
		return s, e
	}
	if s.AggregateType, e = r.text(); e != nil {
		return s, e
	}
	if s.AggregateID, e = r.u64(); e != nil {
		return s, e
	}
	s.AggregateRevision, e = r.i64()
	if e != nil {
		return s, e
	}
	return s, validateScope(s)
}
func (r *docReader) identity() (ReferenceIdentity, error) {
	var v ReferenceIdentity
	var e error
	if v.EntityType, e = r.text(); e != nil {
		return v, e
	}
	if v.ID, e = r.u64(); e != nil {
		return v, e
	}
	v.Version, e = r.i64()
	if e != nil {
		return v, e
	}
	return v, validateIdentity(v)
}
func (r *docReader) set(max int) ([]string, error) {
	n, e := r.count(max)
	if e != nil {
		return nil, e
	}
	values := make([]string, n)
	seen := map[string]bool{}
	for i := range values {
		values[i], e = r.text()
		if e != nil || seen[values[i]] {
			return nil, invalid()
		}
		seen[values[i]] = true
	}
	return values, nil
}
func (r *docReader) mutations() ([]string, error) {
	n, e := r.count(3)
	if e != nil {
		return nil, e
	}
	values := make([]string, n)
	seen := map[byte]bool{}
	for i := range values {
		v, e := r.byte()
		if e != nil || seen[v] || mutationName[v] == "" {
			return nil, invalid()
		}
		seen[v] = true
		values[i] = mutationName[v]
	}
	return values, nil
}

func decodeSnapshot(r *docReader, version byte) (DocumentSnapshot, []byte, int64, int64, error) {
	var s DocumentSnapshot
	v, e := r.byte()
	if e != nil || v != version {
		return s, nil, 0, 0, invalid()
	}
	actor, e := r.take(32)
	if e != nil {
		return s, nil, 0, 0, e
	}
	actor = append([]byte(nil), actor...)
	issued, e := r.i64()
	if e != nil {
		return s, nil, 0, 0, e
	}
	expires, e := r.i64()
	if e != nil {
		return s, nil, 0, 0, e
	}
	if s.ModelFingerprint, e = r.text(); e != nil {
		return s, nil, 0, 0, e
	}
	if s.BusinessID, e = r.text(); e != nil {
		return s, nil, 0, 0, e
	}
	if s.Scope, e = r.scope(); e != nil {
		return s, nil, 0, 0, e
	}
	if s.Aggregate, e = r.identity(); e != nil {
		return s, nil, 0, 0, e
	}
	n, e := r.count(10000)
	if e != nil {
		return s, nil, 0, 0, e
	}
	s.Entities = make([]DocumentEntitySnapshot, n)
	for i := range s.Entities {
		item := &s.Entities[i]
		if item.RowKey, e = r.text(); e != nil {
			return s, nil, 0, 0, e
		}
		if item.Identity, e = r.identity(); e != nil {
			return s, nil, 0, 0, e
		}
		if item.LoadedFields, e = r.set(1024); e != nil {
			return s, nil, 0, 0, e
		}
		if item.WritableFields, e = r.set(1024); e != nil {
			return s, nil, 0, 0, e
		}
		if item.AllowedMutations, e = r.mutations(); e != nil {
			return s, nil, 0, 0, e
		}
	}
	n, e = r.count(1024)
	if e != nil {
		return s, nil, 0, 0, e
	}
	s.Relations = make([]DocumentRelationSnapshot, n)
	for i := range s.Relations {
		item := &s.Relations[i]
		if item.Name, e = r.text(); e != nil {
			return s, nil, 0, 0, e
		}
		v, e := r.byte()
		if e != nil {
			return s, nil, 0, 0, e
		}
		if v == 1 {
			item.Completeness = "COMPLETE"
		} else if v == 2 {
			item.Completeness = "PARTIAL"
		} else {
			return s, nil, 0, 0, invalid()
		}
		if item.Writable, e = r.flag(); e != nil {
			return s, nil, 0, 0, e
		}
		if item.AllowedMutations, e = r.mutations(); e != nil {
			return s, nil, 0, 0, e
		}
		count, e := r.count(10000)
		if e != nil {
			return s, nil, 0, 0, e
		}
		item.RowKeys = make([]string, count)
		for j := range item.RowKeys {
			item.RowKeys[j], e = r.text()
			if e != nil {
				return s, nil, 0, 0, e
			}
		}
	}
	if _, e = encodeSnapshot(s, version, actor, issued, expires); e != nil {
		return s, nil, 0, 0, invalid()
	}
	return s, actor, issued, expires, nil
}

func (r *ContextBoundReferenceRuntime) sealDocument(prefix, label, purpose string, key ReferenceKey, plaintext []byte) (string, error) {
	secret, e := r.derive(key, label)
	if e != nil {
		return "", e
	}
	block, e := aes.NewCipher(secret)
	if e != nil {
		return "", configuration()
	}
	aead, e := cipher.NewGCM(block)
	if e != nil {
		return "", configuration()
	}
	var aad bytes.Buffer
	for _, v := range []string{prefix, key.KeyID, r.service, r.environment, purpose} {
		if e = putDocText(&aad, v); e != nil {
			return "", e
		}
	}
	nonce := make([]byte, 12)
	if _, e = io.ReadFull(r.nonce, nonce); e != nil {
		return "", configuration()
	}
	token := prefix + "." + key.KeyID + "." + base64.RawURLEncoding.EncodeToString(aead.Seal(nonce, nonce, plaintext, aad.Bytes()))
	if len(token) > maxDocumentToken {
		return "", invalid()
	}
	return token, nil
}
func (r *ContextBoundReferenceRuntime) openDocument(prefix, label, purpose, token string) (ReferenceKey, []byte, error) {
	var empty ReferenceKey
	if len(token) > maxDocumentToken {
		return empty, nil, invalid()
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 || parts[0] != prefix || !validKeyID(parts[1]) || parts[2] == "" {
		return empty, nil, invalid()
	}
	key, ok, e := r.keys.KeyByID(parts[1])
	if e != nil || !ok || !validKey(key) || key.KeyID != parts[1] {
		return empty, nil, invalid()
	}
	envelope, e := base64.RawURLEncoding.DecodeString(parts[2])
	if e != nil || len(envelope) < 28 || base64.RawURLEncoding.EncodeToString(envelope) != parts[2] {
		return empty, nil, invalid()
	}
	secret, e := r.derive(key, label)
	if e != nil {
		return empty, nil, e
	}
	block, e := aes.NewCipher(secret)
	if e != nil {
		return empty, nil, configuration()
	}
	aead, e := cipher.NewGCM(block)
	if e != nil {
		return empty, nil, configuration()
	}
	var aad bytes.Buffer
	for _, v := range []string{prefix, key.KeyID, r.service, r.environment, purpose} {
		if e = putDocText(&aad, v); e != nil {
			return empty, nil, e
		}
	}
	plain, e := aead.Open(nil, envelope[:12], envelope[12:], aad.Bytes())
	if e != nil {
		return empty, nil, invalid()
	}
	return key, plain, nil
}
func (r *ContextBoundReferenceRuntime) documentTimes(lifetime time.Duration) (int64, int64, error) {
	if lifetime < time.Second || lifetime%time.Second != 0 {
		return 0, 0, invalid()
	}
	issued := r.now().Unix()
	seconds := int64(lifetime / time.Second)
	if issued > math.MaxInt64-seconds {
		return 0, 0, invalid()
	}
	return issued, issued + seconds, nil
}
func (r *ContextBoundReferenceRuntime) authorizeDocument(principal TrustedReferencePrincipal, s DocumentSnapshot) error {
	if e := r.checkAuthorization(principal, s.Scope, s.Aggregate); e != nil {
		return e
	}
	for _, v := range s.Entities {
		if e := r.checkAuthorization(principal, s.Scope, v.Identity); e != nil {
			return e
		}
	}
	return nil
}
func (r *ContextBoundReferenceRuntime) verifyDocument(principal TrustedReferencePrincipal, key ReferenceKey, actor []byte, issued, expires int64, scope ReferenceDocumentScope, aggregateType, purpose string) error {
	now := r.now().Unix()
	if expires <= now {
		return expired()
	}
	if issued > now+60 {
		return invalid()
	}
	expected, e := r.actor(key, principal)
	if e != nil {
		return e
	}
	if !hmac.Equal(actor, expected) || scope.AggregateType != aggregateType || scope.Purpose != purpose {
		return mismatch()
	}
	return nil
}
func (r *ContextBoundReferenceRuntime) IssueDocumentSnapshot(principal TrustedReferencePrincipal, s DocumentSnapshot, lifetime time.Duration) (string, error) {
	if e := validatePrincipal(principal); e != nil {
		return "", e
	}
	issued, expires, e := r.documentTimes(lifetime)
	if e != nil {
		return "", e
	}
	key, e := r.keys.CurrentKey()
	if e != nil || !validKey(key) {
		return "", configuration()
	}
	actor, e := r.actor(key, principal)
	if e != nil {
		return "", e
	}
	plain, e := encodeSnapshot(s, 1, actor, issued, expires)
	if e != nil {
		return "", e
	}
	if e = r.authorizeDocument(principal, s); e != nil {
		return "", e
	}
	return r.sealDocument("tqd1", "document-snapshot-encryption", s.Scope.Purpose, key, plain)
}
func (r *ContextBoundReferenceRuntime) ConsumeDocumentSnapshot(principal TrustedReferencePrincipal, token, aggregateType, purpose string) (VerifiedDocument[DocumentSnapshot], error) {
	var result VerifiedDocument[DocumentSnapshot]
	if e := validatePrincipal(principal); e != nil {
		return result, e
	}
	if e := docText(aggregateType); e != nil {
		return result, e
	}
	if e := docText(purpose); e != nil {
		return result, e
	}
	key, plain, e := r.openDocument("tqd1", "document-snapshot-encryption", purpose, token)
	if e != nil {
		return result, e
	}
	reader := docReader{data: plain}
	snapshot, actor, issued, expires, e := decodeSnapshot(&reader, 1)
	if e != nil || reader.offset != len(plain) {
		return result, invalid()
	}
	if e = r.verifyDocument(principal, key, actor, issued, expires, snapshot.Scope, aggregateType, purpose); e != nil {
		return result, e
	}
	if e = r.authorizeDocument(principal, snapshot); e != nil {
		return result, e
	}
	return VerifiedDocument[DocumentSnapshot]{snapshot, key.KeyID, time.Unix(issued, 0).UTC(), time.Unix(expires, 0).UTC()}, nil
}
