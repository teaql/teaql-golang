package runtime

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

type ToolRisk string

const (
	ToolRiskMemoryOnly       ToolRisk = "MEMORY_ONLY"
	ToolRiskExternalResource ToolRisk = "EXTERNAL_RESOURCE"
	ToolRiskPrivileged       ToolRisk = "PRIVILEGED"
)

type ToolToken[T any] struct {
	ID   string
	Risk ToolRisk
}

type ToolProvider interface {
	ToolID() string
	Create(*UserContext) any
}

type ToolPolicy struct {
	allowed         map[string]struct{}
	allowMemoryOnly bool
}

func StandardToolPolicy() ToolPolicy {
	return ToolPolicy{allowed: map[string]struct{}{}, allowMemoryOnly: true}
}
func DenyAllTools() ToolPolicy { return ToolPolicy{allowed: map[string]struct{}{}} }
func AllowTools(tokens ...ToolDescriptor) ToolPolicy {
	allowed := make(map[string]struct{}, len(tokens))
	for _, token := range tokens {
		allowed[token.ToolID()] = struct{}{}
	}
	return ToolPolicy{allowed: allowed, allowMemoryOnly: true}
}
func (p ToolPolicy) allows(id string, risk ToolRisk) bool {
	_, explicit := p.allowed[id]
	return explicit || (risk == ToolRiskMemoryOnly && p.allowMemoryOnly)
}

type ToolDescriptor interface{ ToolID() string }

func (t ToolToken[T]) ToolID() string { return t.ID }

type Tools struct {
	ctx       *UserContext
	policy    ToolPolicy
	providers map[string]ToolProvider
}

func (t *Tools) Has(descriptor ToolDescriptor) bool {
	_, ok := t.providers[descriptor.ToolID()]
	return ok
}
func (t *Tools) Descriptors() []string {
	result := make([]string, 0, len(t.providers))
	for id := range t.providers {
		result = append(result, id)
	}
	return result
}
func GetTool[T any](tools *Tools, token ToolToken[T]) (T, error) {
	var zero T
	provider, ok := tools.providers[token.ID]
	if !ok {
		return zero, fmt.Errorf("tool not available: %s", token.ID)
	}
	if !tools.policy.allows(token.ID, token.Risk) {
		return zero, fmt.Errorf("tool denied by policy: %s", token.ID)
	}
	value, ok := provider.Create(tools.ctx).(T)
	if !ok {
		return zero, fmt.Errorf("tool provider type mismatch: %s", token.ID)
	}
	return value, nil
}

type ContextToolsBuilder struct {
	ctx       *UserContext
	policy    ToolPolicy
	providers []ToolProvider
}

func NewContextTools(ctx *UserContext) *ContextToolsBuilder {
	return &ContextToolsBuilder{ctx: ctx, policy: StandardToolPolicy()}
}
func (b *ContextToolsBuilder) Policy(policy ToolPolicy) *ContextToolsBuilder {
	b.policy = policy
	return b
}
func (b *ContextToolsBuilder) Provider(provider ToolProvider) *ContextToolsBuilder {
	b.providers = append(b.providers, provider)
	return b
}
func (b *ContextToolsBuilder) Build() *Tools {
	providers := make(map[string]ToolProvider, len(b.providers))
	for _, provider := range b.providers {
		providers[provider.ToolID()] = provider
	}
	return &Tools{ctx: b.ctx, policy: b.policy, providers: providers}
}

type HTTPTransport interface {
	Do(context.Context, string, string, []byte) (int, []byte, error)
}
type HTTPTool interface {
	Get(string) *HTTPIntentPhase
	Post(string, any) (*HTTPIntentPhase, error)
}

var HTTPToolToken = ToolToken[HTTPTool]{ID: "http", Risk: ToolRiskExternalResource}

type HTTPToolProvider struct{ Transport HTTPTransport }

func (p HTTPToolProvider) ToolID() string            { return HTTPToolToken.ID }
func (p HTTPToolProvider) Create(_ *UserContext) any { return &httpTool{transport: p.Transport} }

type httpTool struct{ transport HTTPTransport }

func (h *httpTool) Get(url string) *HTTPIntentPhase {
	return &HTTPIntentPhase{transport: h.transport, method: "GET", url: url}
}
func (h *httpTool) Post(url string, body any) (*HTTPIntentPhase, error) {
	encoded, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	return &HTTPIntentPhase{transport: h.transport, method: "POST", url: url, body: encoded}, nil
}

type HTTPIntentPhase struct {
	transport   HTTPTransport
	method, url string
	body        []byte
}

func (p *HTTPIntentPhase) Purpose(intent string) *ExecutableHTTPTool { return p.executable(intent) }
func (p *HTTPIntentPhase) AuditAs(intent string) *ExecutableHTTPTool { return p.executable(intent) }
func (p *HTTPIntentPhase) executable(intent string) *ExecutableHTTPTool {
	return &ExecutableHTTPTool{transport: p.transport, method: p.method, url: p.url, body: p.body, intent: intent}
}

type ExecutableHTTPTool struct {
	transport   HTTPTransport
	method, url string
	body        []byte
	intent      string
}

func (e *ExecutableHTTPTool) Execute(ctx context.Context) (string, error) {
	if strings.TrimSpace(e.intent) == "" {
		return "", errors.New("HTTP tool execution requires non-empty intent")
	}
	status, body, err := e.transport.Do(ctx, e.method, e.url, e.body)
	if err != nil {
		return "", err
	}
	if status < 200 || status >= 300 {
		return "", fmt.Errorf("HTTP tool failed: %d", status)
	}
	return string(body), nil
}
