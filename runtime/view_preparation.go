package runtime

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/teaql/teaql-golang/core"
)

// ViewOptions are selected by the trusted host after business work finishes.
type ViewOptions struct {
	Mode            string // ordinary or governed
	BusinessOutcome string // read_only or committed
	ExpiresAt       time.Time
}

// AuthorizedBlobSlot is produced by current authorization, not client JSON.
// Path contains only canonical string field names and nonnegative int indexes.
type AuthorizedBlobSlot struct {
	Path           []any
	Entity         string
	EntityID       string
	Field          string
	Scope          string
	BindingVersion uint64
	Reference      core.BlobRef
}

type AuthorizedResponseGraph struct {
	Presentation any
	BlobSlots    []AuthorizedBlobSlot
}

type ResponseGraphAuthorizer interface {
	Authorize(*UserContext, any, ViewOptions) (AuthorizedResponseGraph, error)
}

type PreparedSecureDocument struct {
	Graph     AuthorizedResponseGraph
	Envelope  any
	ExpiresAt time.Time
}

type SecureDocumentContributor interface {
	Prepare(*UserContext, AuthorizedResponseGraph, ViewOptions) (PreparedSecureDocument, error)
}

type ProjectedBlobURL struct {
	URL       string
	ExpiresAt time.Time
}

type BlobResponseContributor interface {
	ProjectBatch(*UserContext, AuthorizedResponseGraph, time.Time) ([]ProjectedBlobURL, error)
}

type PreparedView struct {
	Presentation any
	Document     any
	ExpiresAt    time.Time
	CacheControl string
}

type ViewPreparationError struct {
	Code          string
	BusinessState string
}

func (e *ViewPreparationError) Error() string { return e.Code }
func (e *ViewPreparationError) Retry() string {
	if e.BusinessState == "committed" {
		return "read"
	}
	return "prepare"
}

type ViewPreparationService struct{}

func (ViewPreparationService) Prepare(context *UserContext, result any, options ViewOptions) (view PreparedView, err error) {
	defer func() {
		if err != nil {
			code := "VIEW_PREPARATION_FAILED"
			var failure *ViewPreparationError
			if errors.As(err, &failure) && safeViewCode(failure.Code) {
				code = failure.Code
			}
			state := "read_only"
			if options.BusinessOutcome == "committed" {
				state = "committed"
			}
			err = &ViewPreparationError{Code: code, BusinessState: state}
		}
	}()
	if context == nil || (options.Mode != "ordinary" && options.Mode != "governed") ||
		(options.BusinessOutcome != "read_only" && options.BusinessOutcome != "committed") {
		return view, viewError("VIEW_OPTIONS_INVALID")
	}
	authorizer, ok := context.GetResource("responseGraphAuthorizer").(ResponseGraphAuthorizer)
	if !ok || missingSPI(authorizer) {
		return view, viewError("VIEW_AUTHORIZER_MISSING")
	}
	graph, err := authorizer.Authorize(context, result, options)
	if err != nil {
		return view, err
	}
	expiry := options.ExpiresAt
	if !expiry.IsZero() && !expiry.After(time.Now()) {
		return view, viewError("BLOB_REFERENCE_EXPIRED")
	}
	budget := 100000
	graph.Presentation, err = copyViewValue(graph.Presentation, 0, &budget)
	if err != nil {
		return view, err
	}
	slots := make([]AuthorizedBlobSlot, len(graph.BlobSlots))
	copy(slots, graph.BlobSlots)
	for index := range slots {
		slots[index].Path = append([]any(nil), slots[index].Path...)
	}
	graph.BlobSlots = slots
	var document any
	if options.Mode == "governed" {
		secure, ok := context.GetResource("secureDocumentContributor").(SecureDocumentContributor)
		if !ok || missingSPI(secure) {
			return view, viewError("DOCUMENT_CONFIGURATION")
		}
		prepared, prepareErr := secure.Prepare(context, graph, options)
		if prepareErr != nil {
			return view, prepareErr
		}
		if prepared.ExpiresAt.IsZero() {
			return view, viewError("DOCUMENT_CONFIGURATION")
		}
		graph, document = prepared.Graph, prepared.Envelope
		expiry = earlier(expiry, prepared.ExpiresAt)
		if !expiry.After(time.Now()) {
			return view, viewError("BLOB_REFERENCE_EXPIRED")
		}
	}
	budget = 100000
	presentation, err := copyViewValue(graph.Presentation, 0, &budget)
	if err != nil {
		return view, err
	}
	paths := make(map[string]struct{}, len(graph.BlobSlots))
	for _, slot := range graph.BlobSlots {
		path, pathErr := viewPathKey(slot.Path)
		if pathErr != nil {
			return view, pathErr
		}
		if _, duplicate := paths[path]; duplicate {
			return view, viewError("VIEW_GRAPH_INVALID")
		}
		paths[path] = struct{}{}
		value, pathErr := viewAtPath(presentation, slot.Path)
		if pathErr != nil {
			return view, pathErr
		}
		reference, ok := value.(core.BlobRef)
		if !ok || reference != slot.Reference {
			return view, viewError("VIEW_GRAPH_INVALID")
		}
	}
	if len(graph.BlobSlots) != 0 {
		blob, ok := context.GetResource("blobResponseContributor").(BlobResponseContributor)
		if !ok || missingSPI(blob) {
			return view, viewError("BLOB_CAPABILITY_MISSING")
		}
		urls, projectErr := blob.ProjectBatch(context, graph, expiry)
		if projectErr != nil {
			return view, projectErr
		}
		if len(urls) != len(graph.BlobSlots) {
			return view, viewError("BLOB_SIGN_FAILED")
		}
		for index, signed := range urls {
			parsed, parseErr := url.Parse(signed.URL)
			if parseErr != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") ||
				!signed.ExpiresAt.After(time.Now()) || (!expiry.IsZero() && signed.ExpiresAt.After(expiry)) {
				return view, viewError("BLOB_SIGN_FAILED")
			}
			presentation, err = replaceViewPath(presentation, graph.BlobSlots[index].Path, signed.URL)
			if err != nil {
				return view, err
			}
			expiry = earlier(expiry, signed.ExpiresAt)
		}
	}
	budget = 100000
	if containsViewBlob(presentation, 0, &budget) {
		return view, viewError("BLOB_CAPABILITY_MISSING")
	}
	return PreparedView{Presentation: presentation, Document: document,
		ExpiresAt: expiry, CacheControl: "private, no-store"}, nil
}

func viewError(code string) error { return &ViewPreparationError{Code: code} }

func safeViewCode(code string) bool {
	switch code {
	case "VIEW_OPTIONS_INVALID", "VIEW_AUTHORIZER_MISSING", "VIEW_GRAPH_INVALID",
		"DOCUMENT_CONFIGURATION", "BLOB_CAPABILITY_MISSING", "BLOB_SIGN_FAILED",
		"BLOB_READ_DENIED", "BLOB_REFERENCE_EXPIRED", "BLOB_SCHEMA_MISSING", "BLOB_CATALOG_FAILED":
		return true
	}
	return false
}

func earlier(left, right time.Time) time.Time {
	if left.IsZero() || right.Before(left) {
		return right
	}
	return left
}

func copyViewValue(value any, depth int, budget *int) (any, error) {
	*budget--
	if *budget < 0 || depth > 64 {
		return nil, viewError("VIEW_GRAPH_INVALID")
	}
	switch original := value.(type) {
	case nil, bool, string, int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64, float32, float64, core.BlobRef, time.Time:
		return original, nil
	case map[string]any:
		if len(original) > *budget {
			return nil, viewError("VIEW_GRAPH_INVALID")
		}
		copy := make(map[string]any, len(original))
		for key, child := range original {
			var err error
			copy[key], err = copyViewValue(child, depth+1, budget)
			if err != nil {
				return nil, err
			}
		}
		return copy, nil
	case []any:
		if len(original) > *budget {
			return nil, viewError("VIEW_GRAPH_INVALID")
		}
		copy := make([]any, len(original))
		for index, child := range original {
			var err error
			copy[index], err = copyViewValue(child, depth+1, budget)
			if err != nil {
				return nil, err
			}
		}
		return copy, nil
	default:
		return nil, viewError("VIEW_GRAPH_INVALID")
	}
}

func viewPathKey(path []any) (string, error) {
	if len(path) == 0 || len(path) > 64 {
		return "", viewError("VIEW_GRAPH_INVALID")
	}
	var key strings.Builder
	for _, part := range path {
		switch typed := part.(type) {
		case string:
			if typed == "" {
				return "", viewError("VIEW_GRAPH_INVALID")
			}
			key.WriteString("s")
			key.WriteString(strconv.Itoa(len(typed)))
			key.WriteString(":")
			key.WriteString(typed)
		case int:
			if typed < 0 {
				return "", viewError("VIEW_GRAPH_INVALID")
			}
			key.WriteString("i")
			key.WriteString(strconv.Itoa(typed))
			key.WriteString(":")
		default:
			return "", viewError("VIEW_GRAPH_INVALID")
		}
	}
	return key.String(), nil
}

func viewAtPath(root any, path []any) (any, error) {
	var current = root
	for _, part := range path {
		switch typed := part.(type) {
		case string:
			object, ok := current.(map[string]any)
			if !ok {
				return nil, viewError("VIEW_GRAPH_INVALID")
			}
			value, exists := object[typed]
			if !exists {
				return nil, viewError("VIEW_GRAPH_INVALID")
			}
			current = value
		case int:
			array, ok := current.([]any)
			if !ok || typed < 0 || typed >= len(array) {
				return nil, viewError("VIEW_GRAPH_INVALID")
			}
			current = array[typed]
		default:
			return nil, viewError("VIEW_GRAPH_INVALID")
		}
	}
	return current, nil
}

func replaceViewPath(root any, path []any, value string) (any, error) {
	if _, err := viewAtPath(root, path); err != nil {
		return nil, err
	}
	current := root
	for _, part := range path[:len(path)-1] {
		switch typed := part.(type) {
		case string:
			current = current.(map[string]any)[typed]
		case int:
			current = current.([]any)[typed]
		}
	}
	switch typed := path[len(path)-1].(type) {
	case string:
		current.(map[string]any)[typed] = value
	case int:
		current.([]any)[typed] = value
	}
	return root, nil
}

func containsViewBlob(value any, depth int, budget *int) bool {
	*budget--
	if *budget < 0 || depth > 64 {
		return true
	}
	switch typed := value.(type) {
	case core.BlobRef:
		return true
	case map[string]any:
		for _, child := range typed {
			if containsViewBlob(child, depth+1, budget) {
				return true
			}
		}
	case []any:
		for _, child := range typed {
			if containsViewBlob(child, depth+1, budget) {
				return true
			}
		}
	}
	return false
}
