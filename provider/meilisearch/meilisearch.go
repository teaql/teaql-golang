package meilisearch

import (
	"bytes"
	stdcontext "context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
)

type MeilisearchProvider struct {
	client *http.Client
	host   string
	apiKey *string
}

func NewMeilisearchProvider(host string, apiKey *string) *MeilisearchProvider {
	return &MeilisearchProvider{
		client: &http.Client{Timeout: 10 * time.Second},
		host:   host,
		apiKey: apiKey,
	}
}

func (p *MeilisearchProvider) Capabilities() data_service.DataServiceCapabilities {
	return data_service.DataServiceCapabilities{
		Query:        true,
		Mutation:     true,
		Transaction:  false,
		Schema:       false,
		IdGeneration: false,
	}
}

func (p *MeilisearchProvider) doRequest(req *http.Request) ([]byte, error) {
	if p.apiKey != nil {
		req.Header.Set("Authorization", "Bearer "+*p.apiKey)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("meilisearch error %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

func (p *MeilisearchProvider) Query(context stdcontext.Context, request *data_service.QueryRequest) (*data_service.QueryResult, error) {
	startedAt := time.Now()
	search := ""
	if request.Query.SearchWithText != nil {
		search = *request.Query.SearchWithText
	}

	url := fmt.Sprintf("%s/indexes/%s/search", p.host, request.Query.Entity)

	payload := map[string]interface{}{
		"q":     search,
		"limit": 100,
	}
	payloadBytes, _ := json.Marshal(payload)

	req, err := http.NewRequestWithContext(context, "POST", url, bytes.NewReader(payloadBytes))
	if err != nil {
		return nil, err
	}

	body, err := p.doRequest(req)
	if err != nil {
		return nil, err
	}

	var response struct {
		Hits []map[string]interface{} `json:"hits"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}

	var rows []core.Record
	for _, hit := range response.Hits {
		record := make(core.Record)
		for k, v := range hit {
			record[k] = core.ValText(fmt.Sprintf("%v", v)) // simplified json mapping
		}
		rows = append(rows, record)
	}

	hitsCount := len(response.Hits)
	debugQuery := fmt.Sprintf("POST %s with %s", url, string(payloadBytes))
	return &data_service.QueryResult{
		Rows: rows,
		Metadata: data_service.ExecutionMetadata{
			Backend:     "meilisearch",
			Operation:   data_service.OpQuery,
			StartedAt:   startedAt,
			EndedAt:     time.Now(),
			ResultCount: &hitsCount,
			DebugQuery:  &debugQuery,
			TraceChain:  request.TraceChain,
			Comment:     request.Comment,
		},
	}, nil
}

func (p *MeilisearchProvider) Mutate(context stdcontext.Context, request data_service.MutationRequest) (*data_service.MutationResult, error) {
	startedAt := time.Now()

	switch req := request.(type) {
	case *data_service.InsertMutation:
		url := fmt.Sprintf("%s/indexes/%s/documents?primaryKey=id", p.host, req.Cmd.Entity)

		doc := make(map[string]interface{})
		for k, v := range req.Cmd.Values {
			doc[k] = v.V
		}

		payloadBytes, _ := json.Marshal([]map[string]interface{}{doc})

		httpReq, err := http.NewRequestWithContext(context, "POST", url, bytes.NewReader(payloadBytes))
		if err != nil {
			return nil, err
		}

		if _, err := p.doRequest(httpReq); err != nil {
			return nil, err
		}

		affected := uint64(1)
		debugQuery := fmt.Sprintf("POST %s to Meilisearch", req.Cmd.Entity)
		return &data_service.MutationResult{
			AffectedRows: 1,
			Metadata: data_service.ExecutionMetadata{
				Backend:      "meilisearch",
				Operation:    data_service.OpUpdate,
				StartedAt:    startedAt,
				EndedAt:      time.Now(),
				AffectedRows: &affected,
				DebugQuery:   &debugQuery,
			},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported mutation type for meilisearch")
	}
}
