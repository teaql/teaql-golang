package securedocumentreference

import (
	"bytes"
	stdcontext "context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"reflect"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/runtime"
	teaqlsql "github.com/teaql/teaql-golang/sql"
)

const maxCommandBytes = 256000

type SubmittedDocumentCommand struct {
	BusinessID, DocumentToken, CommandID string
	Body                                 map[string]any
	// Non-nil selections are part of the durable replay digest.
	BlobSelections []DocumentBlobSelectionSubmission
}

type DocumentBlobSelectionSubmission struct {
	Token, TargetEntity, CanonicalField, Kind, TargetKey string
}

// The command digest uses the same lower-camel selection keys as the shared
// reference wire. The public submission shape remains independent of JSON tags.
type canonicalDocumentSelection struct {
	CanonicalField string `json:"canonicalField"`
	Kind           string `json:"kind"`
	TargetEntity   string `json:"targetEntity"`
	TargetKey      string `json:"targetKey"`
	Token          string `json:"token"`
}

type DocumentCommandReceipt struct {
	PreviousRevision, NewRevision int64
	Response                      string
}

func (DocumentCommandReceipt) String() string { return "DocumentCommandReceipt[redacted]" }

// DocumentCommandAuthority resolves the current trusted actor and current SQL permission.
type DocumentCommandAuthority func(stdcontext.Context, *runtime.UserContext,
	ReferenceDocumentScope, teaqlsql.SqlTransport) (TrustedReferencePrincipal, error)
type DocumentCommandMutation func(teaqlsql.SqlTransport,
	SubmittedDocumentCommand) (DocumentCommandReceipt, error)

type SQLiteDocumentCommandJournal struct{ transport teaqlsql.SqlTransport }

func NewSQLiteDocumentCommandJournal(transport teaqlsql.SqlTransport) (*SQLiteDocumentCommandJournal, error) {
	if transport == nil {
		return nil, errors.New("DOCUMENT_CAPABILITY_MISSING")
	}
	return &SQLiteDocumentCommandJournal{transport: transport}, nil
}

func (j *SQLiteDocumentCommandJournal) EnsureSchema(ctx stdcontext.Context) error {
	if j == nil || j.transport == nil {
		return errors.New("DOCUMENT_CAPABILITY_MISSING")
	}
	_, err := j.transport.ExecuteSql(ctx, &teaqlsql.CompiledQuery{Sql: `CREATE TABLE IF NOT EXISTS teaql_document_command_receipt (
      scope_hash TEXT NOT NULL, actor_hash TEXT NOT NULL, command_id TEXT NOT NULL,
      request_hash TEXT NOT NULL, previous_revision BIGINT NOT NULL,
      new_revision BIGINT NOT NULL, response TEXT NOT NULL,
      PRIMARY KEY (scope_hash, actor_hash, command_id)
    )`})
	if err != nil {
		return errors.New("DOCUMENT_CATALOG_FAILED")
	}
	return nil
}

func (j *SQLiteDocumentCommandJournal) ValidateSchema(ctx stdcontext.Context) error {
	if j == nil || j.transport == nil {
		return errors.New("DOCUMENT_SCHEMA_MISSING")
	}
	rows, err := j.transport.FetchAllSql(ctx, &teaqlsql.CompiledQuery{
		Sql: "PRAGMA table_info(teaql_document_command_receipt)"})
	if err != nil || len(rows) != 7 {
		return errors.New("DOCUMENT_SCHEMA_MISSING")
	}
	expected := []string{"scope_hash", "actor_hash", "command_id", "request_hash",
		"previous_revision", "new_revision", "response"}
	for i, name := range expected {
		found := false
		for _, row := range rows {
			actual, _ := documentSQLText(row["name"])
			if actual != name {
				continue
			}
			notNull, ok1 := row["notnull"].TryI64()
			primary, ok2 := row["pk"].TryI64()
			wantPrimary := int64(0)
			if i < 3 {
				wantPrimary = int64(i + 1)
			}
			if !ok1 || !ok2 || notNull != 1 || primary != wantPrimary {
				return errors.New("DOCUMENT_SCHEMA_MISSING")
			}
			found = true
			break
		}
		if !found {
			return errors.New("DOCUMENT_SCHEMA_MISSING")
		}
	}
	return nil
}

func documentSQLText(value core.Value) (string, bool) {
	switch item := value.V.(type) {
	case string:
		return item, true
	case []byte:
		return string(item), true
	default:
		return "", false
	}
}

func documentText(value string, maximum int) bool {
	return value != "" && len(value) <= maximum && utf8.ValidString(value) &&
		value == strings.TrimSpace(value) && strings.IndexFunc(value, unicode.IsControl) < 0
}

func canonicalBody(value any, depth int, remaining *int) error {
	*remaining--
	if *remaining < 0 || depth > 64 {
		return errors.New("DOCUMENT_COMMAND_INVALID")
	}
	switch item := value.(type) {
	case nil, bool, string:
		if text, ok := item.(string); ok && (!utf8.ValidString(text) || len(text) > maxCommandBytes) {
			return errors.New("DOCUMENT_COMMAND_INVALID")
		}
		return nil
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		n := reflect.ValueOf(item)
		if n.Kind() >= reflect.Uint && n.Kind() <= reflect.Uint64 {
			if n.Uint() > (1<<53)-1 {
				return errors.New("DOCUMENT_COMMAND_INVALID")
			}
		} else if n.Int() < -((1<<53)-1) || n.Int() > (1<<53)-1 {
			return errors.New("DOCUMENT_COMMAND_INVALID")
		}
		return nil
	case json.Number:
		n, err := item.Int64()
		if err != nil || n < -((1<<53)-1) || n > (1<<53)-1 {
			return errors.New("DOCUMENT_COMMAND_INVALID")
		}
		return nil
	case []any:
		for _, child := range item {
			if err := canonicalBody(child, depth+1, remaining); err != nil {
				return err
			}
		}
		return nil
	case map[string]any:
		for key, child := range item {
			if !documentText(key, 16384) || key == "__proto__" || key == "constructor" || key == "prototype" {
				return errors.New("DOCUMENT_COMMAND_INVALID")
			}
			if err := canonicalBody(child, depth+1, remaining); err != nil {
				return err
			}
		}
		return nil
	default:
		return errors.New("DOCUMENT_COMMAND_INVALID")
	}
}

func prepareDocumentCommand(command SubmittedDocumentCommand) (string, SubmittedDocumentCommand, error) {
	var empty SubmittedDocumentCommand
	if !documentText(command.BusinessID, 16384) || !documentText(command.DocumentToken, maxCommandBytes) ||
		!documentText(command.CommandID, 256) || command.Body == nil {
		return "", empty, errors.New("DOCUMENT_COMMAND_INVALID")
	}
	remaining := 100000
	if err := canonicalBody(command.Body, 0, &remaining); err != nil {
		return "", empty, err
	}
	if len(command.BlobSelections) > 100 {
		return "", empty, errors.New("DOCUMENT_COMMAND_INVALID")
	}
	for _, selection := range command.BlobSelections {
		if !documentText(selection.Token, maxCommandBytes) ||
			!documentText(selection.TargetEntity, 16384) ||
			!documentText(selection.CanonicalField, 16384) ||
			!documentText(selection.TargetKey, 16384) ||
			(selection.Kind != "EXISTING_ROW" && selection.Kind != "NEW_CLIENT_KEY") {
			return "", empty, errors.New("DOCUMENT_COMMAND_INVALID")
		}
	}
	// Preserve four-element hashes for receipts written before selection support.
	parts := []any{command.BusinessID, command.DocumentToken, command.CommandID, command.Body}
	if command.BlobSelections != nil {
		canonical := make([]canonicalDocumentSelection, len(command.BlobSelections))
		for index, selection := range command.BlobSelections {
			canonical[index] = canonicalDocumentSelection{
				CanonicalField: selection.CanonicalField, Kind: selection.Kind,
				TargetEntity: selection.TargetEntity, TargetKey: selection.TargetKey,
				Token: selection.Token,
			}
		}
		parts = append(parts, canonical)
	}
	encoded, err := json.Marshal(parts)
	if err != nil || len(encoded) > maxCommandBytes {
		return "", empty, errors.New("DOCUMENT_COMMAND_INVALID")
	}
	digest := sha256.Sum256(encoded)
	var values []json.RawMessage
	if err := json.Unmarshal(encoded, &values); err != nil || len(values) != len(parts) {
		return "", empty, errors.New("DOCUMENT_COMMAND_INVALID")
	}
	var snapshot map[string]any
	decoder := json.NewDecoder(bytes.NewReader(values[3]))
	decoder.UseNumber()
	if err := decoder.Decode(&snapshot); err != nil {
		return "", empty, errors.New("DOCUMENT_COMMAND_INVALID")
	}
	var selections []DocumentBlobSelectionSubmission
	if len(values) == 5 {
		if err := json.Unmarshal(values[4], &selections); err != nil || selections == nil {
			return "", empty, errors.New("DOCUMENT_COMMAND_INVALID")
		}
	}
	return hex.EncodeToString(digest[:]), SubmittedDocumentCommand{
		BusinessID: command.BusinessID, DocumentToken: command.DocumentToken,
		CommandID: command.CommandID, Body: snapshot, BlobSelections: selections}, nil
}

func legacyDocumentSelectionHash(command SubmittedDocumentCommand) (string, error) {
	if command.BlobSelections == nil {
		return "", nil
	}
	parts := []any{command.BusinessID, command.DocumentToken, command.CommandID,
		command.Body, command.BlobSelections}
	encoded, err := json.Marshal(parts)
	if err != nil || len(encoded) > maxCommandBytes {
		return "", errors.New("DOCUMENT_COMMAND_INVALID")
	}
	digest := sha256.Sum256(encoded)
	return hex.EncodeToString(digest[:]), nil
}

func documentScopeHash(scope ReferenceDocumentScope) (string, error) {
	if !documentText(scope.DocumentID, 16384) || !documentText(scope.Purpose, 16384) ||
		!documentText(scope.AggregateType, 16384) || scope.AggregateID == 0 || scope.AggregateRevision < 0 {
		return "", errors.New("DOCUMENT_COMMAND_INVALID")
	}
	encoded, _ := json.Marshal([]string{scope.DocumentID, scope.Purpose,
		scope.AggregateType, strconv.FormatUint(scope.AggregateID, 10)})
	digest := sha256.Sum256(encoded)
	return hex.EncodeToString(digest[:]), nil
}

func documentActorHash(principal TrustedReferencePrincipal) (string, error) {
	if !documentText(principal.AuthenticationRealm, 16384) || !documentText(principal.Subject, 16384) ||
		!documentText(principal.DomainRootType, 16384) || principal.DomainRootID == 0 {
		return "", errors.New("DOCUMENT_COMMAND_INVALID")
	}
	encoded, _ := json.Marshal([]string{principal.AuthenticationRealm, principal.Subject,
		principal.DomainRootType, strconv.FormatUint(principal.DomainRootID, 10)})
	digest := sha256.Sum256(encoded)
	return hex.EncodeToString(digest[:]), nil
}

func checkedDocumentReceipt(receipt DocumentCommandReceipt) (DocumentCommandReceipt, error) {
	if receipt.PreviousRevision < 0 || receipt.NewRevision < receipt.PreviousRevision ||
		!documentText(receipt.Response, maxCommandBytes) {
		return DocumentCommandReceipt{}, errors.New("DOCUMENT_COMMAND_INVALID")
	}
	return receipt, nil
}

// ApplyInGraph keeps an authorized command receipt on the caller's physical SQL owner.
// The embedding service still authenticates tqd2 and validates the model's mutation plan.
func (j *SQLiteDocumentCommandJournal) ApplyInGraph(ctx stdcontext.Context,
	graph *runtime.GraphSQLHandle, actor *runtime.UserContext, scope ReferenceDocumentScope,
	command SubmittedDocumentCommand, authority DocumentCommandAuthority,
	mutate DocumentCommandMutation) (DocumentCommandReceipt, bool, error) {
	var receipt DocumentCommandReceipt
	var replayed bool
	if j == nil || graph == nil || ctx == nil || actor == nil {
		return receipt, false, errors.New("DOCUMENT_TRANSACTION_REQUIRED")
	}
	err := graph.WithSQL(func(sql teaqlsql.SqlTransport) error {
		if !graph.BelongsTo(actor) || authority == nil || mutate == nil {
			return errors.New("DOCUMENT_AUTHORIZATION_REQUIRED")
		}
		scopeKey, err := documentScopeHash(scope)
		if err != nil {
			return err
		}
		requestHash, frozen, err := prepareDocumentCommand(command)
		if err != nil {
			return err
		}
		current := func() (string, error) {
			principal, err := authority(ctx, actor, scope, sql)
			if err != nil {
				return "", errors.New("DOCUMENT_AUTHORIZATION_REQUIRED")
			}
			return documentActorHash(principal)
		}
		actorKey, err := current()
		if err != nil {
			return err
		}
		values := []core.Value{core.ValText(scopeKey), core.ValText(actorKey), core.ValText(command.CommandID)}
		rows, err := sql.FetchAllSql(ctx, &teaqlsql.CompiledQuery{
			Sql: `SELECT request_hash,previous_revision,new_revision,response
              FROM teaql_document_command_receipt
              WHERE scope_hash = ? AND actor_hash = ? AND command_id = ?`, Params: values})
		if err != nil || len(rows) > 1 {
			return errors.New("DOCUMENT_CATALOG_FAILED")
		}
		if len(rows) == 1 {
			hash, ok1 := documentSQLText(rows[0]["request_hash"])
			previous, ok2 := rows[0]["previous_revision"].TryI64()
			next, ok3 := rows[0]["new_revision"].TryI64()
			response, ok4 := documentSQLText(rows[0]["response"])
			if !ok1 || !ok2 || !ok3 || !ok4 {
				return errors.New("DOCUMENT_CATALOG_FAILED")
			}
			legacyHash, legacyErr := legacyDocumentSelectionHash(frozen)
			if legacyErr != nil || (hash != requestHash && hash != legacyHash) {
				return errors.New("DOCUMENT_REPLAY_CONFLICT")
			}
			receipt, err = checkedDocumentReceipt(DocumentCommandReceipt{previous, next, response})
			if err != nil {
				return errors.New("DOCUMENT_CATALOG_FAILED")
			}
			second, err := current()
			if err != nil || second != actorKey {
				return errors.New("DOCUMENT_AUTHORIZATION_REQUIRED")
			}
			replayed = true
			return nil
		}
		receipt, err = mutate(sql, frozen)
		if err != nil {
			return err
		}
		receipt, err = checkedDocumentReceipt(receipt)
		if err != nil {
			return err
		}
		frozenHash, _, err := prepareDocumentCommand(frozen)
		if err != nil || frozenHash != requestHash {
			return errors.New("DOCUMENT_COMMAND_INVALID")
		}
		checkHash, _, err := prepareDocumentCommand(command)
		if err != nil || checkHash != requestHash {
			return errors.New("DOCUMENT_COMMAND_INVALID")
		}
		second, err := current()
		if err != nil || second != actorKey {
			return errors.New("DOCUMENT_AUTHORIZATION_REQUIRED")
		}
		params := append(values, core.ValText(requestHash), core.ValI64(receipt.PreviousRevision),
			core.ValI64(receipt.NewRevision), core.ValText(receipt.Response))
		affected, err := sql.ExecuteSql(ctx, &teaqlsql.CompiledQuery{Sql: `INSERT INTO teaql_document_command_receipt
          (scope_hash,actor_hash,command_id,request_hash,previous_revision,new_revision,response)
          VALUES (?,?,?,?,?,?,?)`, Params: params})
		if err != nil || affected != 1 {
			return errors.New("DOCUMENT_CATALOG_FAILED")
		}
		return nil
	})
	if err != nil {
		return DocumentCommandReceipt{}, false, err
	}
	return receipt, replayed, nil
}
