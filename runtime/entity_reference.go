package runtime

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

const (
	entityReferencePrefix                    = "tqr1."
	unsafeRawReferencePrefix                 = "tqr0."
	EntityReferenceAAD                       = "teaql.entity-reference.v1"
	UnsafeRawEntityReferencesEnvironment     = "TEAQL_UNSAFE_RAW_ENTITY_REFERENCES"
	UnsafeRawEntityReferencesAcknowledgement = "I_UNDERSTAND_RAW_ENTITY_IDS_ARE_VISIBLE_FOR_LOCAL_DEVELOPMENT_ONLY"
)

type EntityReferenceClaims struct {
	EntityType string
	ID         uint64
	Version    int64
	IssuedAt   time.Time
	ExpiresAt  time.Time
	Purpose    string
	KeyVersion uint32
}

type EntityReferenceCodec interface {
	EncodeEntityReference(entityType string, id uint64, version int64, purpose string, ttl time.Duration) (string, error)
	DecodeEntityReference(token, expectedEntityType, purpose string) (EntityReferenceClaims, error)
}

type EntityReferenceTokenError struct{ Code string }

func (e *EntityReferenceTokenError) Error() string { return e.Code }

type AEADReferenceCodec struct {
	activeKeyVersion uint32
	keys             map[uint32][]byte
	now              func() time.Time
	nonce            io.Reader
}

func NewAEADEntityReferenceCodec(activeKeyVersion uint32, keys map[uint32][]byte) (*AEADReferenceCodec, error) {
	if len(keys[activeKeyVersion]) != 32 {
		return nil, errors.New("active entity reference key must contain 32 bytes")
	}
	copyKeys := make(map[uint32][]byte, len(keys))
	for version, key := range keys {
		if len(key) != 32 {
			return nil, fmt.Errorf("entity reference key %d must contain 32 bytes", version)
		}
		copyKeys[version] = append([]byte(nil), key...)
	}
	return &AEADReferenceCodec{activeKeyVersion: activeKeyVersion, keys: copyKeys, now: time.Now, nonce: rand.Reader}, nil
}

func (c *AEADReferenceCodec) WithClock(now func() time.Time) *AEADReferenceCodec {
	c.now = now
	return c
}
func (c *AEADReferenceCodec) WithNonceSource(source io.Reader) *AEADReferenceCodec {
	c.nonce = source
	return c
}

func writeReferenceText(buffer *[]byte, value string) error {
	data := []byte(value)
	if len(data) > 65535 {
		return errors.New("entity reference text exceeds 65535 bytes")
	}
	var length [2]byte
	binary.BigEndian.PutUint16(length[:], uint16(len(data)))
	*buffer = append(*buffer, length[:]...)
	*buffer = append(*buffer, data...)
	return nil
}

func encodeReferenceClaims(claims EntityReferenceClaims) ([]byte, error) {
	data := make([]byte, 0, len(claims.EntityType)+len(claims.Purpose)+36)
	if err := writeReferenceText(&data, claims.EntityType); err != nil {
		return nil, err
	}
	var number [8]byte
	binary.BigEndian.PutUint64(number[:], claims.ID)
	data = append(data, number[:]...)
	binary.BigEndian.PutUint64(number[:], uint64(claims.Version))
	data = append(data, number[:]...)
	binary.BigEndian.PutUint64(number[:], uint64(claims.IssuedAt.Unix()))
	data = append(data, number[:]...)
	binary.BigEndian.PutUint64(number[:], uint64(claims.ExpiresAt.Unix()))
	data = append(data, number[:]...)
	if err := writeReferenceText(&data, claims.Purpose); err != nil {
		return nil, err
	}
	return data, nil
}

func readReferenceText(data []byte, offset *int) (string, error) {
	if *offset+2 > len(data) {
		return "", errors.New("truncated reference")
	}
	length := int(binary.BigEndian.Uint16(data[*offset : *offset+2]))
	*offset += 2
	if *offset+length > len(data) {
		return "", errors.New("truncated reference")
	}
	value := string(data[*offset : *offset+length])
	*offset += length
	return value, nil
}

func decodeReferenceClaims(data []byte) (EntityReferenceClaims, error) {
	offset := 0
	entity, err := readReferenceText(data, &offset)
	if err != nil {
		return EntityReferenceClaims{}, err
	}
	if offset+32 > len(data) {
		return EntityReferenceClaims{}, errors.New("truncated reference")
	}
	id := binary.BigEndian.Uint64(data[offset:])
	offset += 8
	version := int64(binary.BigEndian.Uint64(data[offset:]))
	offset += 8
	issued := int64(binary.BigEndian.Uint64(data[offset:]))
	offset += 8
	expires := int64(binary.BigEndian.Uint64(data[offset:]))
	offset += 8
	purpose, err := readReferenceText(data, &offset)
	if err != nil || offset != len(data) {
		return EntityReferenceClaims{}, errors.New("invalid reference payload")
	}
	if entity == "" || id == 0 {
		return EntityReferenceClaims{}, errors.New("invalid reference payload")
	}
	return EntityReferenceClaims{EntityType: entity, ID: id, Version: version, IssuedAt: time.Unix(issued, 0).UTC(), ExpiresAt: time.Unix(expires, 0).UTC(), Purpose: purpose}, nil
}

func (c *AEADReferenceCodec) EncodeEntityReference(entityType string, id uint64, version int64, purpose string, ttl time.Duration) (string, error) {
	if strings.TrimSpace(entityType) == "" || id == 0 || ttl <= 0 {
		return "", &EntityReferenceTokenError{Code: "ENTITY_REFERENCE_INVALID"}
	}
	now := c.now().UTC()
	plain, err := encodeReferenceClaims(EntityReferenceClaims{EntityType: entityType, ID: id, Version: version, IssuedAt: now, ExpiresAt: now.Add(ttl), Purpose: purpose})
	if err != nil {
		return "", err
	}
	block, _ := aes.NewCipher(c.keys[c.activeKeyVersion])
	aead, _ := cipher.NewGCM(block)
	nonce := make([]byte, aead.NonceSize())
	if _, err = io.ReadFull(c.nonce, nonce); err != nil {
		return "", err
	}
	ciphertext := aead.Seal(nil, nonce, plain, []byte(EntityReferenceAAD))
	envelope := make([]byte, 4, 4+len(nonce)+len(ciphertext))
	binary.BigEndian.PutUint32(envelope, c.activeKeyVersion)
	envelope = append(envelope, nonce...)
	envelope = append(envelope, ciphertext...)
	return entityReferencePrefix + base64.RawURLEncoding.EncodeToString(envelope), nil
}

func (c *AEADReferenceCodec) DecodeEntityReference(token, expectedEntityType, purpose string) (EntityReferenceClaims, error) {
	fail := func() (EntityReferenceClaims, error) {
		return EntityReferenceClaims{}, &EntityReferenceTokenError{Code: "ENTITY_REFERENCE_INVALID"}
	}
	if !strings.HasPrefix(token, entityReferencePrefix) {
		return fail()
	}
	envelope, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(token, entityReferencePrefix))
	if err != nil || len(envelope) < 4+12+16 {
		return fail()
	}
	keyVersion := binary.BigEndian.Uint32(envelope[:4])
	key, ok := c.keys[keyVersion]
	if !ok {
		return fail()
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return fail()
	}
	aead, _ := cipher.NewGCM(block)
	if len(envelope) < 4+aead.NonceSize()+aead.Overhead() {
		return fail()
	}
	plain, err := aead.Open(nil, envelope[4:4+aead.NonceSize()], envelope[4+aead.NonceSize():], []byte(EntityReferenceAAD))
	if err != nil {
		return fail()
	}
	claims, err := decodeReferenceClaims(plain)
	if err != nil {
		return fail()
	}
	claims.KeyVersion = keyVersion
	now := c.now().UTC()
	if !claims.ExpiresAt.After(now) || claims.IssuedAt.After(now.Add(time.Minute)) || claims.EntityType != expectedEntityType || claims.Purpose != purpose {
		return fail()
	}
	return claims, nil
}

func rawEntityReferencesEnabled() bool {
	return os.Getenv(UnsafeRawEntityReferencesEnvironment) == UnsafeRawEntityReferencesAcknowledgement
}

func (c *UserContext) WithEntityReferenceCodec(codec EntityReferenceCodec) *UserContext {
	c.entityReferenceCodec = codec
	return c
}
func (c *UserContext) EncodeEntityReference(entityType string, id uint64, version int64, purpose string, ttl time.Duration) (string, error) {
	if c.entityReferenceCodec != nil {
		return c.entityReferenceCodec.EncodeEntityReference(entityType, id, version, purpose, ttl)
	}
	if !rawEntityReferencesEnabled() {
		return "", &EntityReferenceTokenError{Code: "ENTITY_REFERENCE_CODEC_REQUIRED"}
	}
	now := time.Now().UTC()
	plain, err := encodeReferenceClaims(EntityReferenceClaims{EntityType: entityType, ID: id, Version: version, IssuedAt: now, ExpiresAt: now.Add(ttl), Purpose: purpose})
	if err != nil {
		return "", err
	}
	return unsafeRawReferencePrefix + base64.RawURLEncoding.EncodeToString(plain), nil
}
func (c *UserContext) DecodeEntityReference(token, expectedEntityType, purpose string) (EntityReferenceClaims, error) {
	if c.entityReferenceCodec != nil {
		return c.entityReferenceCodec.DecodeEntityReference(token, expectedEntityType, purpose)
	}
	if !rawEntityReferencesEnabled() || !strings.HasPrefix(token, unsafeRawReferencePrefix) {
		return EntityReferenceClaims{}, &EntityReferenceTokenError{Code: "ENTITY_REFERENCE_CODEC_REQUIRED"}
	}
	plain, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(token, unsafeRawReferencePrefix))
	if err != nil {
		return EntityReferenceClaims{}, &EntityReferenceTokenError{Code: "ENTITY_REFERENCE_INVALID"}
	}
	claims, err := decodeReferenceClaims(plain)
	if err != nil || claims.EntityType != expectedEntityType || claims.Purpose != purpose || !claims.ExpiresAt.After(time.Now().UTC()) {
		return EntityReferenceClaims{}, &EntityReferenceTokenError{Code: "ENTITY_REFERENCE_INVALID"}
	}
	return claims, nil
}
