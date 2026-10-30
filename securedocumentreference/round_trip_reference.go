// Package securedocumentreference implements the optional governed reference wire protocol.
package securedocumentreference

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"io"
	"math"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/crypto/hkdf"
)

const salt = "teaql.round-trip-reference.v1"

type ReferenceError struct{ Code string }

func (e *ReferenceError) Error() string { return e.Code }
func invalid() error                    { return &ReferenceError{Code: "ROUND_TRIP_REFERENCE_INVALID"} }
func configuration() error {
	return &ReferenceError{Code: "ROUND_TRIP_REFERENCE_CONFIGURATION_INVALID"}
}
func expired() error  { return &ReferenceError{Code: "ROUND_TRIP_REFERENCE_EXPIRED"} }
func mismatch() error { return &ReferenceError{Code: "ROUND_TRIP_REFERENCE_SCOPE_MISMATCH"} }
func denied() error   { return &ReferenceError{Code: "ROUND_TRIP_REFERENCE_AUTHORIZATION_REQUIRED"} }

type TrustedReferencePrincipal struct {
	AuthenticationRealm, Subject, DomainRootType string
	DomainRootID                                 uint64
}
type ReferenceDocumentScope struct {
	DocumentID, Purpose, AggregateType string
	AggregateID                        uint64
	AggregateRevision                  int64
}
type ReferenceIdentity struct {
	EntityType string
	ID         uint64
	Version    int64
}
type VerifiedEntityReference struct {
	Identity  ReferenceIdentity
	KeyID     string
	IssuedAt  time.Time
	ExpiresAt time.Time
}

type ReferenceKey struct {
	KeyID     string
	MasterKey []byte
}
type ReferenceKeyProvider interface {
	CurrentKey() (ReferenceKey, error)
	KeyByID(string) (ReferenceKey, bool, error)
}
type StaticReferenceKeyProvider struct {
	current string
	keys    map[string]ReferenceKey
}

func validKeyID(value string) bool {
	if len(value) < 1 || len(value) > 64 {
		return false
	}
	for _, ch := range []byte(value) {
		if !(ch >= 'a' && ch <= 'z' || ch >= 'A' && ch <= 'Z' ||
			ch >= '0' && ch <= '9' || ch == '-' || ch == '_') {
			return false
		}
	}
	return true
}
func validKey(key ReferenceKey) bool { return validKeyID(key.KeyID) && len(key.MasterKey) == 32 }

func NewStaticReferenceKeyProvider(current string, keys ...ReferenceKey) (*StaticReferenceKeyProvider, error) {
	if !validKeyID(current) {
		return nil, configuration()
	}
	retained := make(map[string]ReferenceKey, len(keys))
	for _, key := range keys {
		if !validKey(key) {
			return nil, configuration()
		}
		if _, duplicate := retained[key.KeyID]; duplicate {
			return nil, configuration()
		}
		retained[key.KeyID] = ReferenceKey{KeyID: key.KeyID, MasterKey: append([]byte(nil), key.MasterKey...)}
	}
	if _, ok := retained[current]; !ok {
		return nil, configuration()
	}
	return &StaticReferenceKeyProvider{current: current, keys: retained}, nil
}
func (p *StaticReferenceKeyProvider) CurrentKey() (ReferenceKey, error) {
	key, ok := p.keys[p.current]
	if !ok {
		return ReferenceKey{}, configuration()
	}
	return ReferenceKey{KeyID: key.KeyID, MasterKey: append([]byte(nil), key.MasterKey...)}, nil
}
func (p *StaticReferenceKeyProvider) KeyByID(id string) (ReferenceKey, bool, error) {
	key, ok := p.keys[id]
	if !ok {
		return ReferenceKey{}, false, nil
	}
	return ReferenceKey{KeyID: key.KeyID, MasterKey: append([]byte(nil), key.MasterKey...)}, true, nil
}

type ReferenceAuthorizationPolicy func(TrustedReferencePrincipal, ReferenceDocumentScope, ReferenceIdentity) error

type ContextBoundReferenceRuntime struct {
	service, environment string
	keys                 ReferenceKeyProvider
	authorize            ReferenceAuthorizationPolicy
	now                  func() time.Time
	nonce                io.Reader
}

func NewContextBoundReferenceRuntime(service, environment string, keys ReferenceKeyProvider,
	authorize ReferenceAuthorizationPolicy) (*ContextBoundReferenceRuntime, error) {
	if err := validText(service); err != nil {
		return nil, configuration()
	}
	if err := validText(environment); err != nil || keys == nil || authorize == nil {
		return nil, configuration()
	}
	key, err := keys.CurrentKey()
	if err != nil || !validKey(key) {
		return nil, configuration()
	}
	return &ContextBoundReferenceRuntime{service: service, environment: environment,
		keys: keys, authorize: authorize, now: time.Now, nonce: rand.Reader}, nil
}
func (r *ContextBoundReferenceRuntime) WithClock(now func() time.Time) *ContextBoundReferenceRuntime {
	if now != nil {
		r.now = now
	}
	return r
}
func (r *ContextBoundReferenceRuntime) WithNonceSource(source io.Reader) *ContextBoundReferenceRuntime {
	if source != nil {
		r.nonce = source
	}
	return r
}

func validText(value string) error {
	if value == "" || strings.TrimSpace(value) != value || !utf8.ValidString(value) || len(value) > 4096 {
		return invalid()
	}
	return nil
}
func writeText(buffer *bytes.Buffer, value string) error {
	if err := validText(value); err != nil {
		return err
	}
	var length [2]byte
	binary.BigEndian.PutUint16(length[:], uint16(len(value)))
	buffer.Write(length[:])
	buffer.WriteString(value)
	return nil
}
func writeU64(buffer *bytes.Buffer, value uint64) {
	var bytes [8]byte
	binary.BigEndian.PutUint64(bytes[:], value)
	buffer.Write(bytes[:])
}
func writeI64(buffer *bytes.Buffer, value int64) {
	writeU64(buffer, uint64(value))
}
func validatePrincipal(value TrustedReferencePrincipal) error {
	for _, text := range []string{value.AuthenticationRealm, value.Subject, value.DomainRootType} {
		if err := validText(text); err != nil {
			return err
		}
	}
	if value.DomainRootID == 0 {
		return invalid()
	}
	return nil
}
func validateScope(value ReferenceDocumentScope) error {
	for _, text := range []string{value.DocumentID, value.Purpose, value.AggregateType} {
		if err := validText(text); err != nil {
			return err
		}
	}
	if value.AggregateID == 0 || value.AggregateRevision < 0 {
		return invalid()
	}
	return nil
}
func validateIdentity(value ReferenceIdentity) error {
	if err := validText(value.EntityType); err != nil || value.ID == 0 || value.Version < 0 {
		return invalid()
	}
	return nil
}

func (r *ContextBoundReferenceRuntime) derive(key ReferenceKey, label string) ([]byte, error) {
	var info bytes.Buffer
	for _, text := range []string{label, r.service, r.environment, key.KeyID} {
		if err := writeText(&info, text); err != nil {
			return nil, configuration()
		}
	}
	output := make([]byte, 32)
	if _, err := io.ReadFull(hkdf.New(sha256.New, key.MasterKey, []byte(salt), info.Bytes()), output); err != nil {
		return nil, configuration()
	}
	return output, nil
}
func (r *ContextBoundReferenceRuntime) actor(key ReferenceKey, principal TrustedReferencePrincipal) ([]byte, error) {
	secret, err := r.derive(key, "actor-fingerprint")
	if err != nil {
		return nil, err
	}
	var canonical bytes.Buffer
	for _, text := range []string{principal.AuthenticationRealm, principal.Subject, principal.DomainRootType} {
		if err := writeText(&canonical, text); err != nil {
			return nil, err
		}
	}
	writeU64(&canonical, principal.DomainRootID)
	mac := hmac.New(sha256.New, secret)
	mac.Write(canonical.Bytes())
	return mac.Sum(nil), nil
}
func (r *ContextBoundReferenceRuntime) aad(key ReferenceKey, purpose string) ([]byte, error) {
	var result bytes.Buffer
	for _, text := range []string{"tqr1", key.KeyID, r.service, r.environment, purpose} {
		if err := writeText(&result, text); err != nil {
			return nil, err
		}
	}
	return result.Bytes(), nil
}
func (r *ContextBoundReferenceRuntime) checkAuthorization(principal TrustedReferencePrincipal,
	scope ReferenceDocumentScope, identity ReferenceIdentity) error {
	if err := r.authorize(principal, scope, identity); err != nil {
		return denied()
	}
	return nil
}

func (r *ContextBoundReferenceRuntime) Issue(principal TrustedReferencePrincipal,
	scope ReferenceDocumentScope, identity ReferenceIdentity, lifetime time.Duration) (string, error) {
	if err := validatePrincipal(principal); err != nil {
		return "", err
	}
	if err := validateScope(scope); err != nil {
		return "", err
	}
	if err := validateIdentity(identity); err != nil {
		return "", err
	}
	if lifetime < time.Second || lifetime%time.Second != 0 {
		return "", invalid()
	}
	if err := r.checkAuthorization(principal, scope, identity); err != nil {
		return "", err
	}
	key, err := r.keys.CurrentKey()
	if err != nil || !validKey(key) {
		return "", configuration()
	}
	issued := r.now().Unix()
	seconds := int64(lifetime / time.Second)
	if seconds <= 0 || issued > math.MaxInt64-seconds {
		return "", invalid()
	}
	expires := issued + seconds
	actor, err := r.actor(key, principal)
	if err != nil {
		return "", err
	}
	var plaintext bytes.Buffer
	plaintext.Write(actor)
	if err := writeText(&plaintext, identity.EntityType); err != nil {
		return "", err
	}
	writeU64(&plaintext, identity.ID)
	writeI64(&plaintext, identity.Version)
	if err := writeText(&plaintext, scope.DocumentID); err != nil {
		return "", err
	}
	if err := writeText(&plaintext, scope.AggregateType); err != nil {
		return "", err
	}
	writeU64(&plaintext, scope.AggregateID)
	writeI64(&plaintext, scope.AggregateRevision)
	writeI64(&plaintext, issued)
	writeI64(&plaintext, expires)
	secret, err := r.derive(key, "entity-reference-encryption")
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(secret)
	if err != nil {
		return "", configuration()
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return "", configuration()
	}
	nonce := make([]byte, 12)
	if _, err := io.ReadFull(r.nonce, nonce); err != nil {
		return "", configuration()
	}
	aad, err := r.aad(key, scope.Purpose)
	if err != nil {
		return "", err
	}
	envelope := aead.Seal(append([]byte(nil), nonce...), nonce, plaintext.Bytes(), aad)
	token := "tqr1." + key.KeyID + "." + base64.RawURLEncoding.EncodeToString(envelope)
	if len(token) > 24000 {
		return "", invalid()
	}
	return token, nil
}

type reader struct {
	data   []byte
	offset int
}

func (r *reader) take(n int) ([]byte, error) {
	if n < 0 || n > len(r.data)-r.offset {
		return nil, invalid()
	}
	value := r.data[r.offset : r.offset+n]
	r.offset += n
	return value, nil
}
func (r *reader) text() (string, error) {
	length, err := r.take(2)
	if err != nil {
		return "", err
	}
	value, err := r.take(int(binary.BigEndian.Uint16(length)))
	if err != nil {
		return "", err
	}
	result := string(value)
	if err := validText(result); err != nil {
		return "", err
	}
	return result, nil
}
func (r *reader) u64() (uint64, error) {
	value, err := r.take(8)
	if err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint64(value), nil
}
func (r *reader) i64() (int64, error) {
	value, err := r.u64()
	return int64(value), err
}

func (r *ContextBoundReferenceRuntime) Consume(principal TrustedReferencePrincipal,
	scope ReferenceDocumentScope, token, expectedEntityType string) (VerifiedEntityReference, error) {
	var empty VerifiedEntityReference
	if err := validatePrincipal(principal); err != nil {
		return empty, err
	}
	if err := validateScope(scope); err != nil {
		return empty, err
	}
	if err := validText(expectedEntityType); err != nil {
		return empty, err
	}
	if len(token) > 24000 {
		return empty, invalid()
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 || parts[0] != "tqr1" || !validKeyID(parts[1]) || parts[2] == "" {
		return empty, invalid()
	}
	key, ok, err := r.keys.KeyByID(parts[1])
	if err != nil || !ok || !validKey(key) || key.KeyID != parts[1] {
		return empty, invalid()
	}
	envelope, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil || len(envelope) < 28 || base64.RawURLEncoding.EncodeToString(envelope) != parts[2] {
		return empty, invalid()
	}
	secret, err := r.derive(key, "entity-reference-encryption")
	if err != nil {
		return empty, err
	}
	block, err := aes.NewCipher(secret)
	if err != nil {
		return empty, configuration()
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return empty, configuration()
	}
	aad, err := r.aad(key, scope.Purpose)
	if err != nil {
		return empty, err
	}
	plaintext, err := aead.Open(nil, envelope[:12], envelope[12:], aad)
	if err != nil {
		return empty, invalid()
	}
	decoded := reader{data: plaintext}
	actor, err := decoded.take(32)
	if err != nil {
		return empty, err
	}
	entityType, err := decoded.text()
	if err != nil {
		return empty, err
	}
	id, err := decoded.u64()
	if err != nil {
		return empty, err
	}
	version, err := decoded.i64()
	if err != nil {
		return empty, err
	}
	documentID, err := decoded.text()
	if err != nil {
		return empty, err
	}
	aggregateType, err := decoded.text()
	if err != nil {
		return empty, err
	}
	aggregateID, err := decoded.u64()
	if err != nil {
		return empty, err
	}
	aggregateRevision, err := decoded.i64()
	if err != nil {
		return empty, err
	}
	issued, err := decoded.i64()
	if err != nil {
		return empty, err
	}
	expires, err := decoded.i64()
	if err != nil || decoded.offset != len(decoded.data) {
		return empty, invalid()
	}
	identity := ReferenceIdentity{EntityType: entityType, ID: id, Version: version}
	if err := validateIdentity(identity); err != nil {
		return empty, err
	}
	now := r.now().Unix()
	if expires <= now {
		return empty, expired()
	}
	if issued > now+60 {
		return empty, invalid()
	}
	expectedActor, err := r.actor(key, principal)
	if err != nil {
		return empty, err
	}
	if !hmac.Equal(actor, expectedActor) || documentID != scope.DocumentID ||
		aggregateType != scope.AggregateType || aggregateID != scope.AggregateID ||
		aggregateRevision != scope.AggregateRevision || entityType != expectedEntityType {
		return empty, mismatch()
	}
	if err := r.checkAuthorization(principal, scope, identity); err != nil {
		return empty, err
	}
	return VerifiedEntityReference{Identity: identity, KeyID: key.KeyID,
		IssuedAt: time.Unix(issued, 0).UTC(), ExpiresAt: time.Unix(expires, 0).UTC()}, nil
}
