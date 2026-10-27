package securedocumentreference

import (
	"bytes"
	stdcontext "context"
	"encoding/json"
	"errors"
	"io"
	"math"
	"strconv"
	"time"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
	teaqlsql "github.com/teaql/teaql-golang/sql"
)

type DocumentOpenRequest struct {
	AggregateType, BusinessID, Purpose string
	Lifetime                           time.Duration
}

type ContextualDocument struct {
	BusinessID    string         `json:"businessId"`
	DocumentToken string         `json:"documentToken"`
	Body          map[string]any `json:"body"`
}

func (ContextualDocument) String() string { return "ContextualDocument[redacted]" }

type DocumentAcceptRequest struct {
	AggregateType, Purpose string
	Submitted              SubmittedDocumentCommand
}

type AcceptedDocument struct {
	Document                      ContextualDocument
	PreviousRevision, NewRevision int64
	Replayed                      bool
}

func (AcceptedDocument) String() string { return "AcceptedDocument[redacted]" }

type CurrentBoundDocument struct {
	Document BoundDocumentSnapshot
	Body     map[string]any
	Lifetime time.Duration
}

type VerifiedDocumentBlobSelection struct {
	Target          BlobSelectionTarget
	Reference       core.BlobRef
	RowKey          string
	EntityID        string
	ExpectedVersion int64
	NextVersion     int64
}

type DocumentBlobSelectionClaim struct {
	Target    BlobSelectionTarget
	Reference core.BlobRef
}

// Resolve a sealed client key after the application has persisted its child on
// the same SQL lease. The service checks the returned row against both snapshots.
type SQLiteDocumentNewBlobSelectionHost interface {
	ResolveNewBlobSelection(stdcontext.Context, *runtime.UserContext, BoundDocumentSnapshot,
		SubmittedDocumentCommand, DocumentBlobSelectionClaim, teaqlsql.SqlTransport) (string, string, error)
}

// Optional host extension: persist the model field and invoke the actual Blob
// catalog on graph. Its result is checked against the post-mutation projection.
type SQLiteDocumentBlobSelectionHost interface {
	BindBlobSelection(stdcontext.Context, *runtime.UserContext, *runtime.GraphSQLHandle,
		VerifiedDocumentBlobSelection, teaqlsql.SqlTransport) error
}

// SQLiteContextBoundDocumentHost owns the model-specific projection and allowed mutation plan.
// All SQL callbacks receive the same current physical transaction as their read or write.
type SQLiteContextBoundDocumentHost interface {
	Principal(*runtime.UserContext) (TrustedReferencePrincipal, error)
	Authorize(stdcontext.Context, *runtime.UserContext, TrustedReferencePrincipal,
		ReferenceDocumentScope, teaqlsql.SqlTransport) error
	Load(stdcontext.Context, *runtime.UserContext, string, teaqlsql.SqlTransport) (CurrentBoundDocument, error)
	Apply(stdcontext.Context, *runtime.UserContext, BoundDocumentSnapshot,
		SubmittedDocumentCommand, teaqlsql.SqlTransport) error
	Present(stdcontext.Context, *runtime.UserContext, ContextualDocument) error
}

type SQLiteContextBoundDocumentService struct {
	actor     *runtime.UserContext
	transport teaqlsql.SqlTransactionTransport
	module    *SecureDocumentReferenceModule
	journal   *SQLiteDocumentCommandJournal
	host      SQLiteContextBoundDocumentHost
}

func NewSQLiteContextBoundDocumentService(actor *runtime.UserContext,
	transport teaqlsql.SqlTransactionTransport, module *SecureDocumentReferenceModule,
	journal *SQLiteDocumentCommandJournal, host SQLiteContextBoundDocumentHost) (*SQLiteContextBoundDocumentService, error) {
	if actor == nil || transport == nil || module == nil || module.Protocol == nil ||
		journal == nil || host == nil {
		return nil, errors.New("DOCUMENT_CONFIGURATION")
	}
	return &SQLiteContextBoundDocumentService{
		actor: actor, transport: transport, module: module, journal: journal, host: host}, nil
}

func (s *SQLiteContextBoundDocumentService) currentPrincipal(actor *runtime.UserContext) (TrustedReferencePrincipal, error) {
	if s == nil || actor == nil || actor != s.actor {
		return TrustedReferencePrincipal{}, errors.New("DOCUMENT_CONFIGURATION")
	}
	principal, err := s.host.Principal(actor)
	if err != nil || validatePrincipal(principal) != nil {
		return TrustedReferencePrincipal{}, errors.New("DOCUMENT_AUTHORIZATION_REQUIRED")
	}
	return principal, nil
}

func verifyCurrentDocument(current CurrentBoundDocument, businessID, aggregateType, purpose string) error {
	if current.Body == nil || current.Lifetime <= 0 ||
		current.Document.Snapshot.BusinessID != businessID ||
		current.Document.Snapshot.Scope.AggregateType != aggregateType ||
		current.Document.Snapshot.Scope.Purpose != purpose {
		return errors.New("DOCUMENT_PROJECTION_VIOLATION")
	}
	if _, err := encodeBound(current.Document, make([]byte, 32), 0, 1); err != nil {
		return errors.New("DOCUMENT_PROJECTION_VIOLATION")
	}
	return nil
}

func copyDocumentBody(body map[string]any) (map[string]any, error) {
	if body == nil {
		return nil, errors.New("DOCUMENT_PROJECTION_VIOLATION")
	}
	remaining := 100000
	if err := canonicalBody(body, 0, &remaining); err != nil {
		return nil, errors.New("DOCUMENT_PROJECTION_VIOLATION")
	}
	encoded, err := json.Marshal(body)
	if err != nil || len(encoded) > maxCommandBytes {
		return nil, errors.New("DOCUMENT_PROJECTION_VIOLATION")
	}
	var copy map[string]any
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.UseNumber()
	if err := decoder.Decode(&copy); err != nil || copy == nil {
		return nil, errors.New("DOCUMENT_PROJECTION_VIOLATION")
	}
	return copy, nil
}

func verifiedDocumentSelections(protocol *ContextBoundReferenceRuntime,
	principal TrustedReferencePrincipal, issued BoundDocumentSnapshot,
	submitted SubmittedDocumentCommand) ([]DocumentBlobSelectionClaim, error) {
	result := make([]DocumentBlobSelectionClaim, 0, len(submitted.BlobSelections))
	seen := make(map[string]bool)
	for _, selection := range submitted.BlobSelections {
		target := BlobSelectionTarget{Scope: issued.Snapshot.Scope,
			TargetEntity: selection.TargetEntity, CanonicalField: selection.CanonicalField,
			Kind: selection.Kind, TargetKey: selection.TargetKey, CommandID: submitted.CommandID}
		pair := target.TargetKey + "\x00" + target.CanonicalField
		if seen[pair] {
			return nil, errors.New("DOCUMENT_COMMAND_INVALID")
		}
		seen[pair] = true
		claim, err := protocol.ConsumeBlobSelection(principal, submitted.DocumentToken,
			selection.Token, target)
		if err != nil {
			return nil, err
		}
		result = append(result, DocumentBlobSelectionClaim{Target: target,
			Reference: claim.Document.Reference})
	}
	return result, nil
}

func mapExistingBlobSelection(issued BoundDocumentSnapshot,
	selection DocumentBlobSelectionClaim) (VerifiedDocumentBlobSelection, error) {
	target := selection.Target
	var row *DocumentEntitySnapshot
	var binding *DocumentBlobBindingSnapshot
	for i := range issued.Snapshot.Entities {
		if issued.Snapshot.Entities[i].RowKey == target.TargetKey {
			row = &issued.Snapshot.Entities[i]
			break
		}
	}
	for i := range issued.Bindings {
		if issued.Bindings[i].RowKey == target.TargetKey &&
			issued.Bindings[i].CanonicalField == target.CanonicalField {
			binding = &issued.Bindings[i]
			break
		}
	}
	if row == nil || binding == nil || row.Identity.EntityType != target.TargetEntity ||
		!binding.Writable || binding.BindingVersion >= math.MaxInt64 {
		return VerifiedDocumentBlobSelection{}, errors.New("DOCUMENT_PROJECTION_VIOLATION")
	}
	return VerifiedDocumentBlobSelection{Target: target, Reference: selection.Reference,
		RowKey: target.TargetKey, EntityID: strconv.FormatUint(row.Identity.ID, 10),
		ExpectedVersion: int64(binding.BindingVersion),
		NextVersion:     int64(binding.BindingVersion) + 1}, nil
}

func mapNewBlobSelection(issued BoundDocumentSnapshot, selection DocumentBlobSelectionClaim,
	rowKey, entityID string) (VerifiedDocumentBlobSelection, error) {
	id, err := strconv.ParseUint(entityID, 10, 64)
	if !documentText(rowKey, 16384) || err != nil || id == 0 ||
		strconv.FormatUint(id, 10) != entityID {
		return VerifiedDocumentBlobSelection{}, errors.New("DOCUMENT_PROJECTION_VIOLATION")
	}
	for _, row := range issued.Snapshot.Entities {
		if row.RowKey == rowKey {
			return VerifiedDocumentBlobSelection{}, errors.New("DOCUMENT_PROJECTION_VIOLATION")
		}
	}
	return VerifiedDocumentBlobSelection{Target: selection.Target, Reference: selection.Reference,
		RowKey: rowKey, EntityID: entityID, ExpectedVersion: 0, NextVersion: 1}, nil
}

func validNewBlobSelectionChild(issued, after BoundDocumentSnapshot,
	selection VerifiedDocumentBlobSelection) bool {
	for _, row := range after.Snapshot.Entities {
		if row.RowKey != selection.RowKey || row.Identity.EntityType != selection.Target.TargetEntity ||
			strconv.FormatUint(row.Identity.ID, 10) != selection.EntityID ||
			!includes(row.WritableFields, selection.Target.CanonicalField) {
			continue
		}
		for _, relation := range after.Snapshot.Relations {
			if relation.Completeness != "COMPLETE" || !relation.Writable ||
				!includes(relation.AllowedMutations, "CREATE_CHILD") ||
				!includes(relation.RowKeys, selection.RowKey) {
				continue
			}
			for _, previous := range issued.Snapshot.Relations {
				if previous.Name == relation.Name && previous.Completeness == "COMPLETE" &&
					previous.Writable && includes(previous.AllowedMutations, "CREATE_CHILD") &&
					!includes(previous.RowKeys, selection.RowKey) {
					return true
				}
			}
		}
	}
	return false
}

func (s *SQLiteContextBoundDocumentService) Open(ctx stdcontext.Context,
	actor *runtime.UserContext, request DocumentOpenRequest) (ContextualDocument, error) {
	var empty ContextualDocument
	if ctx == nil || !documentText(request.AggregateType, 16384) ||
		!documentText(request.BusinessID, 16384) || !documentText(request.Purpose, 16384) ||
		request.Lifetime <= 0 {
		return empty, errors.New("DOCUMENT_COMMAND_INVALID")
	}
	principal, err := s.currentPrincipal(actor)
	if err != nil {
		return empty, err
	}
	tx, err := s.transport.BeginSql(ctx)
	if err != nil {
		return empty, errors.New("DOCUMENT_CATALOG_FAILED")
	}
	committed := false
	defer func() {
		if !committed {
			_ = tx.RollbackSql(ctx)
		}
	}()
	current, err := s.host.Load(ctx, actor, request.BusinessID, tx)
	if err != nil {
		return empty, err
	}
	if err = verifyCurrentDocument(current, request.BusinessID, request.AggregateType, request.Purpose); err != nil {
		return empty, err
	}
	if latest, e := s.currentPrincipal(actor); e != nil || latest != principal {
		return empty, errors.New("DOCUMENT_AUTHORIZATION_REQUIRED")
	}
	if err = s.host.Authorize(ctx, actor, principal, current.Document.Snapshot.Scope, tx); err != nil {
		return empty, errors.New("DOCUMENT_AUTHORIZATION_REQUIRED")
	}
	if latest, e := s.currentPrincipal(actor); e != nil || latest != principal {
		return empty, errors.New("DOCUMENT_AUTHORIZATION_REQUIRED")
	}
	lifetime := request.Lifetime
	if current.Lifetime < lifetime {
		lifetime = current.Lifetime
	}
	token, err := s.module.Protocol.IssueBoundDocumentSnapshot(principal, current.Document, lifetime)
	if err != nil {
		return empty, err
	}
	body, err := copyDocumentBody(current.Body)
	if err != nil {
		return empty, err
	}
	result := ContextualDocument{BusinessID: request.BusinessID, DocumentToken: token, Body: body}
	if err = tx.CommitSql(ctx); err != nil {
		return empty, errors.New("DOCUMENT_COMMIT_UNKNOWN")
	}
	committed = true
	if err = s.host.Present(ctx, actor, result); err != nil {
		return empty, err
	}
	return result, nil
}

func (s *SQLiteContextBoundDocumentService) Accept(ctx stdcontext.Context,
	actor *runtime.UserContext, request DocumentAcceptRequest) (AcceptedDocument, error) {
	var empty AcceptedDocument
	if ctx == nil || !documentText(request.AggregateType, 16384) ||
		!documentText(request.Purpose, 16384) || !documentText(request.Submitted.BusinessID, 16384) ||
		!documentText(request.Submitted.CommandID, 256) {
		return empty, errors.New("DOCUMENT_COMMAND_INVALID")
	}
	principal, err := s.currentPrincipal(actor)
	if err != nil {
		return empty, err
	}
	verified, err := s.module.Protocol.ConsumeBoundDocumentSnapshot(principal,
		request.Submitted.DocumentToken, request.AggregateType, request.Purpose)
	if err != nil {
		return empty, err
	}
	issued := verified.Document
	if issued.Snapshot.BusinessID != request.Submitted.BusinessID {
		return empty, errors.New("DOCUMENT_PROJECTION_VIOLATION")
	}
	scope := issued.Snapshot.Scope
	comment := "accept protected document command"
	intent, err := core.NewMutationIntent(&comment)
	if err != nil {
		return empty, errors.New("DOCUMENT_COMMAND_INVALID")
	}
	var receipt DocumentCommandReceipt
	var replayed bool
	err = actor.ExecuteGraphSaveWithSQL(intent, func(graph *runtime.GraphSQLHandle) error {
		var inner error
		receipt, replayed, inner = s.journal.ApplyInGraph(ctx, graph, actor, scope,
			request.Submitted, func(currentCtx stdcontext.Context, currentActor *runtime.UserContext,
				expected ReferenceDocumentScope, sql teaqlsql.SqlTransport) (TrustedReferencePrincipal, error) {
				current, e := s.currentPrincipal(currentActor)
				if e != nil || current != principal {
					return TrustedReferencePrincipal{}, errors.New("DOCUMENT_AUTHORIZATION_REQUIRED")
				}
				if e = s.host.Authorize(currentCtx, currentActor, current, expected, sql); e != nil {
					return TrustedReferencePrincipal{}, errors.New("DOCUMENT_AUTHORIZATION_REQUIRED")
				}
				return current, nil
			}, func(sql teaqlsql.SqlTransport, submitted SubmittedDocumentCommand) (DocumentCommandReceipt, error) {
				before, e := s.host.Load(ctx, actor, submitted.BusinessID, sql)
				if e != nil {
					return DocumentCommandReceipt{}, e
				}
				if e = verifyCurrentDocument(before, submitted.BusinessID,
					request.AggregateType, request.Purpose); e != nil {
					return DocumentCommandReceipt{}, e
				}
				currentBytes, e := encodeBound(before.Document, make([]byte, 32), 0, 1)
				if e != nil {
					return DocumentCommandReceipt{}, e
				}
				issuedBytes, e := encodeBound(issued, make([]byte, 32), 0, 1)
				if e != nil || !bytes.Equal(currentBytes, issuedBytes) {
					return DocumentCommandReceipt{}, errors.New("DOCUMENT_REVISION_CONFLICT")
				}
				selections, e := verifiedDocumentSelections(s.module.Protocol, principal, issued, submitted)
				if e != nil {
					return DocumentCommandReceipt{}, e
				}
				binder, supported := s.host.(SQLiteDocumentBlobSelectionHost)
				if len(selections) > 0 && !supported {
					return DocumentCommandReceipt{}, errors.New("DOCUMENT_CAPABILITY_MISSING")
				}
				resolver, hasResolver := s.host.(SQLiteDocumentNewBlobSelectionHost)
				for _, selection := range selections {
					if selection.Target.Kind == "NEW_CLIENT_KEY" && !hasResolver {
						return DocumentCommandReceipt{}, errors.New("DOCUMENT_CAPABILITY_MISSING")
					}
				}
				if e = s.host.Apply(ctx, actor, issued, submitted, sql); e != nil {
					return DocumentCommandReceipt{}, e
				}
				mapped := make([]VerifiedDocumentBlobSelection, 0, len(selections))
				mappedTargets := make(map[string]bool)
				for _, selection := range selections {
					var bound VerifiedDocumentBlobSelection
					if selection.Target.Kind == "EXISTING_ROW" {
						bound, e = mapExistingBlobSelection(issued, selection)
					} else {
						var rowKey, entityID string
						rowKey, entityID, e = resolver.ResolveNewBlobSelection(ctx, actor,
							issued, submitted, selection, sql)
						if e == nil {
							bound, e = mapNewBlobSelection(issued, selection, rowKey, entityID)
						}
					}
					if e != nil {
						return DocumentCommandReceipt{}, e
					}
					pair := bound.RowKey + "\x00" + bound.Target.CanonicalField
					if mappedTargets[pair] {
						return DocumentCommandReceipt{}, errors.New("DOCUMENT_COMMAND_INVALID")
					}
					mappedTargets[pair] = true
					mapped = append(mapped, bound)
				}
				for _, selection := range mapped {
					if e = binder.BindBlobSelection(ctx, actor, graph, selection, sql); e != nil {
						return DocumentCommandReceipt{}, e
					}
				}
				after, e := s.host.Load(ctx, actor, submitted.BusinessID, sql)
				if e != nil {
					return DocumentCommandReceipt{}, e
				}
				if e = verifyCurrentDocument(after, submitted.BusinessID,
					request.AggregateType, request.Purpose); e != nil {
					return DocumentCommandReceipt{}, e
				}
				nextScope := after.Document.Snapshot.Scope
				if nextScope.DocumentID != scope.DocumentID || nextScope.AggregateID != scope.AggregateID ||
					nextScope.AggregateType != scope.AggregateType || nextScope.Purpose != scope.Purpose ||
					nextScope.AggregateRevision < scope.AggregateRevision {
					return DocumentCommandReceipt{}, errors.New("DOCUMENT_PROJECTION_VIOLATION")
				}
				for _, selection := range mapped {
					matched := false
					for _, binding := range after.Document.Bindings {
						if binding.RowKey == selection.RowKey &&
							binding.CanonicalField == selection.Target.CanonicalField &&
							binding.EntityType == selection.Target.TargetEntity &&
							binding.StableKey != nil && *binding.StableKey == selection.Reference.Key() &&
							binding.BindingVersion == uint64(selection.NextVersion) {
							matched = true
							break
						}
					}
					if !matched {
						return DocumentCommandReceipt{}, errors.New("DOCUMENT_PROJECTION_VIOLATION")
					}
					if selection.Target.Kind == "NEW_CLIENT_KEY" &&
						!validNewBlobSelectionChild(issued, after.Document, selection) {
						return DocumentCommandReceipt{}, errors.New("DOCUMENT_PROJECTION_VIOLATION")
					}
				}
				if e = s.host.Authorize(ctx, actor, principal, nextScope, sql); e != nil {
					return DocumentCommandReceipt{}, errors.New("DOCUMENT_AUTHORIZATION_REQUIRED")
				}
				token, e := s.module.Protocol.IssueBoundDocumentSnapshot(principal,
					after.Document, after.Lifetime)
				if e != nil {
					return DocumentCommandReceipt{}, e
				}
				body, e := copyDocumentBody(after.Body)
				if e != nil {
					return DocumentCommandReceipt{}, e
				}
				response, e := json.Marshal(ContextualDocument{
					BusinessID: submitted.BusinessID, DocumentToken: token, Body: body})
				if e != nil {
					return DocumentCommandReceipt{}, errors.New("DOCUMENT_PROJECTION_VIOLATION")
				}
				return DocumentCommandReceipt{PreviousRevision: scope.AggregateRevision,
					NewRevision: nextScope.AggregateRevision, Response: string(response)}, nil
			})
		return inner
	})
	if err != nil {
		return empty, err
	}
	var document ContextualDocument
	decoder := json.NewDecoder(bytes.NewBufferString(receipt.Response))
	decoder.UseNumber()
	if err = decoder.Decode(&document); err != nil || document.BusinessID != request.Submitted.BusinessID ||
		!documentText(document.DocumentToken, maxDocumentToken) || document.Body == nil {
		return empty, &runtime.GraphCommittedError{Cause: errors.New("DOCUMENT_CATALOG_FAILED")}
	}
	var trailing any
	if decoder.Decode(&trailing) != io.EOF {
		return empty, &runtime.GraphCommittedError{Cause: errors.New("DOCUMENT_CATALOG_FAILED")}
	}
	if latest, e := s.currentPrincipal(actor); e != nil || latest != principal {
		return empty, &runtime.GraphCommittedError{Cause: errors.New("DOCUMENT_AUTHORIZATION_REQUIRED")}
	}
	verifiedResult, err := s.module.Protocol.ConsumeBoundDocumentSnapshot(principal,
		document.DocumentToken, request.AggregateType, request.Purpose)
	if err != nil {
		return empty, &runtime.GraphCommittedError{Cause: err}
	}
	resultScope := verifiedResult.Document.Snapshot.Scope
	if resultScope.DocumentID != scope.DocumentID || resultScope.AggregateID != scope.AggregateID ||
		resultScope.AggregateType != scope.AggregateType || resultScope.Purpose != scope.Purpose ||
		resultScope.AggregateRevision != receipt.NewRevision ||
		verifiedResult.Document.Snapshot.BusinessID != document.BusinessID ||
		receipt.PreviousRevision != scope.AggregateRevision {
		return empty, &runtime.GraphCommittedError{Cause: errors.New("DOCUMENT_CATALOG_FAILED")}
	}
	if document.Body, err = copyDocumentBody(document.Body); err != nil {
		return empty, &runtime.GraphCommittedError{Cause: err}
	}
	if err = s.host.Present(ctx, actor, document); err != nil {
		return empty, &runtime.GraphCommittedError{Cause: err}
	}
	return AcceptedDocument{Document: document,
		PreviousRevision: receipt.PreviousRevision, NewRevision: receipt.NewRevision,
		Replayed: replayed}, nil
}
