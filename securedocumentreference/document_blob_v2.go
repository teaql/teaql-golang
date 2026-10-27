package securedocumentreference

import (
	"bytes"
	"strings"
	"time"

	"github.com/teaql/teaql-golang/core"
)

type BlobSelectionTarget struct {
	Scope                                                    ReferenceDocumentScope
	TargetEntity, CanonicalField, Kind, TargetKey, CommandID string
}
type BlobSelectionClaim struct {
	Target    BlobSelectionTarget
	Reference core.BlobRef
}

func validateBound(d BoundDocumentSnapshot) error {
	if len(d.Bindings) > 10000 {
		return documentViolation()
	}
	previousRow, previousField := "", ""
	for _, b := range d.Bindings {
		for _, v := range []string{b.RowKey, b.CanonicalField, b.EntityType} {
			if docText(v) != nil || strings.TrimSpace(v) != v {
				return documentViolation()
			}
		}
		if previousRow != "" && (b.RowKey < previousRow ||
			(b.RowKey == previousRow && b.CanonicalField <= previousField)) {
			return documentViolation()
		}
		previousRow, previousField = b.RowKey, b.CanonicalField
		var entity *DocumentEntitySnapshot
		for i := range d.Snapshot.Entities {
			if d.Snapshot.Entities[i].RowKey == b.RowKey {
				entity = &d.Snapshot.Entities[i]
				break
			}
		}
		if entity == nil || entity.Identity.EntityType != b.EntityType || !includes(entity.LoadedFields, b.CanonicalField) || (b.Writable && !includes(entity.WritableFields, b.CanonicalField)) {
			return documentViolation()
		}
		if b.StableKey != nil {
			if _, e := core.ParseBlobRef(*b.StableKey); e != nil {
				return documentViolation()
			}
		}
	}
	return nil
}
func encodeBound(d BoundDocumentSnapshot, actor []byte, issued, expires int64) ([]byte, error) {
	if e := validateBound(d); e != nil {
		return nil, e
	}
	plain, e := encodeSnapshot(d.Snapshot, 2, actor, issued, expires)
	if e != nil {
		return nil, e
	}
	var b bytes.Buffer
	b.Write(plain)
	putCount(&b, len(d.Bindings))
	for _, v := range d.Bindings {
		for _, s := range []string{v.RowKey, v.CanonicalField, v.EntityType} {
			if e = putDocText(&b, s); e != nil {
				return nil, e
			}
		}
		putBool(&b, v.StableKey != nil)
		if v.StableKey != nil {
			if e = putDocText(&b, *v.StableKey); e != nil {
				return nil, e
			}
		}
		writeU64(&b, v.BindingVersion)
		putBool(&b, v.Nullable)
		putBool(&b, v.Writable)
	}
	return b.Bytes(), nil
}
func decodeBound(plain []byte) (BoundDocumentSnapshot, []byte, int64, int64, error) {
	var d BoundDocumentSnapshot
	r := docReader{data: plain}
	s, actor, issued, expires, e := decodeSnapshot(&r, 2)
	if e != nil {
		return d, nil, 0, 0, e
	}
	d.Snapshot = s
	n, e := r.count(10000)
	if e != nil {
		return d, nil, 0, 0, e
	}
	d.Bindings = make([]DocumentBlobBindingSnapshot, n)
	for i := range d.Bindings {
		v := &d.Bindings[i]
		if v.RowKey, e = r.text(); e != nil {
			return d, nil, 0, 0, e
		}
		if v.CanonicalField, e = r.text(); e != nil {
			return d, nil, 0, 0, e
		}
		if v.EntityType, e = r.text(); e != nil {
			return d, nil, 0, 0, e
		}
		has, e := r.flag()
		if e != nil {
			return d, nil, 0, 0, e
		}
		if has {
			key, e := r.text()
			if e != nil {
				return d, nil, 0, 0, e
			}
			v.StableKey = &key
		}
		if v.BindingVersion, e = r.u64(); e != nil {
			return d, nil, 0, 0, e
		}
		if v.Nullable, e = r.flag(); e != nil {
			return d, nil, 0, 0, e
		}
		if v.Writable, e = r.flag(); e != nil {
			return d, nil, 0, 0, e
		}
	}
	if r.offset != len(plain) || validateBound(d) != nil {
		return d, nil, 0, 0, invalid()
	}
	return d, actor, issued, expires, nil
}
func (r *ContextBoundReferenceRuntime) IssueBoundDocumentSnapshot(principal TrustedReferencePrincipal, d BoundDocumentSnapshot, lifetime time.Duration) (string, error) {
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
	plain, e := encodeBound(d, actor, issued, expires)
	if e != nil {
		return "", e
	}
	if e = r.authorizeDocument(principal, d.Snapshot); e != nil {
		return "", e
	}
	return r.sealDocument("tqd2", "document-snapshot-v2-encryption", d.Snapshot.Scope.Purpose, key, plain)
}
func (r *ContextBoundReferenceRuntime) ConsumeBoundDocumentSnapshot(principal TrustedReferencePrincipal, token, aggregateType, purpose string) (VerifiedDocument[BoundDocumentSnapshot], error) {
	var result VerifiedDocument[BoundDocumentSnapshot]
	if e := validatePrincipal(principal); e != nil {
		return result, e
	}
	if e := docText(aggregateType); e != nil {
		return result, e
	}
	if e := docText(purpose); e != nil {
		return result, e
	}
	key, plain, e := r.openDocument("tqd2", "document-snapshot-v2-encryption", purpose, token)
	if e != nil {
		return result, e
	}
	d, actor, issued, expires, e := decodeBound(plain)
	if e != nil {
		return result, e
	}
	if e = r.verifyDocument(principal, key, actor, issued, expires, d.Snapshot.Scope, aggregateType, purpose); e != nil {
		return result, e
	}
	if e = r.authorizeDocument(principal, d.Snapshot); e != nil {
		return result, e
	}
	return VerifiedDocument[BoundDocumentSnapshot]{d, key.KeyID, time.Unix(issued, 0).UTC(), time.Unix(expires, 0).UTC()}, nil
}
func validTarget(target BlobSelectionTarget) error {
	if e := validateScope(target.Scope); e != nil {
		return e
	}
	for _, v := range []string{target.TargetEntity, target.CanonicalField, target.TargetKey, target.CommandID} {
		if docText(v) != nil || strings.TrimSpace(v) != v {
			return invalid()
		}
	}
	if target.Kind != "EXISTING_ROW" && target.Kind != "NEW_CLIENT_KEY" {
		return invalid()
	}
	return nil
}
func requireTarget(d BoundDocumentSnapshot, t BlobSelectionTarget) error {
	if e := validTarget(t); e != nil {
		return e
	}
	if d.Snapshot.Scope != t.Scope {
		return mismatch()
	}
	if t.Kind == "EXISTING_ROW" {
		rowAllowed, bindingAllowed := false, false
		for _, row := range d.Snapshot.Entities {
			if row.RowKey == t.TargetKey && row.Identity.EntityType == t.TargetEntity && includes(row.WritableFields, t.CanonicalField) {
				rowAllowed = true
				break
			}
		}
		for _, binding := range d.Bindings {
			if binding.RowKey == t.TargetKey && binding.CanonicalField == t.CanonicalField && binding.Writable {
				bindingAllowed = true
				break
			}
		}
		if !rowAllowed || !bindingAllowed {
			return documentViolation()
		}
		return nil
	}
	for _, relation := range d.Snapshot.Relations {
		if relation.Completeness == "COMPLETE" && relation.Writable && includes(relation.AllowedMutations, "CREATE_CHILD") {
			return nil
		}
	}
	return documentViolation()
}
func encodeSelection(c BlobSelectionClaim, actor []byte, issued, expires int64) ([]byte, error) {
	if e := validTarget(c.Target); e != nil {
		return nil, e
	}
	var b bytes.Buffer
	b.WriteByte(1)
	b.Write(actor)
	writeI64(&b, issued)
	writeI64(&b, expires)
	if e := putScope(&b, c.Target.Scope); e != nil {
		return nil, e
	}
	for _, v := range []string{c.Target.TargetEntity, c.Target.CanonicalField} {
		if e := putDocText(&b, v); e != nil {
			return nil, e
		}
	}
	if c.Target.Kind == "EXISTING_ROW" {
		b.WriteByte(1)
	} else {
		b.WriteByte(2)
	}
	if e := putDocText(&b, c.Target.TargetKey); e != nil {
		return nil, e
	}
	if e := putDocText(&b, c.Reference.Key()); e != nil {
		return nil, e
	}
	if e := putDocText(&b, c.Target.CommandID); e != nil {
		return nil, e
	}
	return b.Bytes(), nil
}
func decodeSelection(plain []byte) (BlobSelectionClaim, []byte, int64, int64, error) {
	var claim BlobSelectionClaim
	r := docReader{data: plain}
	version, e := r.byte()
	if e != nil || version != 1 {
		return claim, nil, 0, 0, invalid()
	}
	actor, e := r.take(32)
	if e != nil {
		return claim, nil, 0, 0, e
	}
	actor = append([]byte(nil), actor...)
	issued, e := r.i64()
	if e != nil {
		return claim, nil, 0, 0, e
	}
	expires, e := r.i64()
	if e != nil {
		return claim, nil, 0, 0, e
	}
	t := &claim.Target
	if t.Scope, e = r.scope(); e != nil {
		return claim, nil, 0, 0, e
	}
	if t.TargetEntity, e = r.text(); e != nil {
		return claim, nil, 0, 0, e
	}
	if t.CanonicalField, e = r.text(); e != nil {
		return claim, nil, 0, 0, e
	}
	kind, e := r.byte()
	if e != nil {
		return claim, nil, 0, 0, e
	}
	if kind == 1 {
		t.Kind = "EXISTING_ROW"
	} else if kind == 2 {
		t.Kind = "NEW_CLIENT_KEY"
	} else {
		return claim, nil, 0, 0, invalid()
	}
	if t.TargetKey, e = r.text(); e != nil {
		return claim, nil, 0, 0, e
	}
	key, e := r.text()
	if e != nil {
		return claim, nil, 0, 0, e
	}
	claim.Reference, e = core.ParseBlobRef(key)
	if e != nil {
		return claim, nil, 0, 0, invalid()
	}
	if t.CommandID, e = r.text(); e != nil {
		return claim, nil, 0, 0, e
	}
	if r.offset != len(plain) || validTarget(*t) != nil {
		return claim, nil, 0, 0, invalid()
	}
	return claim, actor, issued, expires, nil
}
func (r *ContextBoundReferenceRuntime) IssueBlobSelection(principal TrustedReferencePrincipal, documentToken string, claim BlobSelectionClaim, lifetime time.Duration) (string, error) {
	if e := validTarget(claim.Target); e != nil {
		return "", e
	}
	document, e := r.ConsumeBoundDocumentSnapshot(principal, documentToken, claim.Target.Scope.AggregateType, claim.Target.Scope.Purpose)
	if e != nil {
		return "", e
	}
	if e = requireTarget(document.Document, claim.Target); e != nil {
		return "", e
	}
	issued, expires, e := r.documentTimes(lifetime)
	if e != nil {
		return "", e
	}
	if expires > document.ExpiresAt.Unix() {
		return "", invalid()
	}
	key, e := r.keys.CurrentKey()
	if e != nil || !validKey(key) {
		return "", configuration()
	}
	actor, e := r.actor(key, principal)
	if e != nil {
		return "", e
	}
	plain, e := encodeSelection(claim, actor, issued, expires)
	if e != nil {
		return "", e
	}
	return r.sealDocument("tqs1", "blob-selection-v1-encryption", claim.Target.Scope.Purpose, key, plain)
}
func (r *ContextBoundReferenceRuntime) ConsumeBlobSelection(principal TrustedReferencePrincipal, documentToken, token string, target BlobSelectionTarget) (VerifiedDocument[BlobSelectionClaim], error) {
	var result VerifiedDocument[BlobSelectionClaim]
	if e := validTarget(target); e != nil {
		return result, e
	}
	document, e := r.ConsumeBoundDocumentSnapshot(principal, documentToken, target.Scope.AggregateType, target.Scope.Purpose)
	if e != nil {
		return result, e
	}
	if e = requireTarget(document.Document, target); e != nil {
		return result, e
	}
	key, plain, e := r.openDocument("tqs1", "blob-selection-v1-encryption", target.Scope.Purpose, token)
	if e != nil {
		return result, e
	}
	claim, actor, issued, expires, e := decodeSelection(plain)
	if e != nil {
		return result, e
	}
	if expires > document.ExpiresAt.Unix() {
		return result, invalid()
	}
	if claim.Target != target {
		return result, mismatch()
	}
	if e = r.verifyDocument(principal, key, actor, issued, expires, claim.Target.Scope, target.Scope.AggregateType, target.Scope.Purpose); e != nil {
		return result, e
	}
	return VerifiedDocument[BlobSelectionClaim]{claim, key.KeyID, time.Unix(issued, 0).UTC(), time.Unix(expires, 0).UTC()}, nil
}
