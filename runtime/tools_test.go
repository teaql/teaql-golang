package runtime

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type stubHTTPTransport struct{ status int }

func (s stubHTTPTransport) Do(_ context.Context, method, url string, _ []byte) (int, []byte, error) {
	return s.status, []byte(fmt.Sprintf("%s:%s", method, url)), nil
}

func TestContextToolsPolicyAndNativeHTTPResponse(t *testing.T) {
	tools := NewContextTools(NewUserContext()).Policy(AllowTools(HTTPToolToken)).
		Provider(HTTPToolProvider{Transport: stubHTTPTransport{status: 200}}).Build()
	httpTool, err := GetTool(tools, HTTPToolToken)
	require.NoError(t, err)
	value, err := httpTool.Get("https://example.com").Purpose("status").Execute(context.Background())
	require.NoError(t, err)
	require.Equal(t, "GET:https://example.com", value)
}

func TestContextToolsNegativesAreExplicit(t *testing.T) {
	denied := NewContextTools(NewUserContext()).Provider(HTTPToolProvider{Transport: stubHTTPTransport{200}}).Build()
	_, err := GetTool(denied, HTTPToolToken)
	require.ErrorContains(t, err, "denied by policy")
	_, err = GetTool(denied, ToolToken[any]{ID: "unknown", Risk: ToolRiskMemoryOnly})
	require.ErrorContains(t, err, "not available")
	allowed := NewContextTools(NewUserContext()).Policy(AllowTools(HTTPToolToken)).
		Provider(HTTPToolProvider{Transport: stubHTTPTransport{200}}).Build()
	httpTool, err := GetTool(allowed, HTTPToolToken)
	require.NoError(t, err)
	_, err = httpTool.Get("https://example.com").Purpose(" ").Execute(context.Background())
	require.ErrorContains(t, err, "non-empty intent")
}
