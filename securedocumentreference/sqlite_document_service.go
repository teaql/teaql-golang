package securedocumentreference

import (
	"bytes"
	stdcontext "context"
	"encoding/json"
	"errors"
	"io"
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
				if e = s.host.Apply(ctx, actor, issued, submitted, sql); e != nil {
					return DocumentCommandReceipt{}, e
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
