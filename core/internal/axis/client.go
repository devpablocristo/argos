package axis

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	analysisdomain "github.com/devpablocristo/argos/core/internal/analyses/usecases/domain"
)

type Client struct {
	httpClient       *http.Client
	orgID            string
	nexusBaseURL     string
	nexusAPIKey      string
	companionBaseURL string
	companionAPIKey  string
	publicBaseURL    string
}

type Config struct {
	OrgID            string
	NexusBaseURL     string
	NexusAPIKey      string
	CompanionBaseURL string
	CompanionAPIKey  string
	PublicBaseURL    string
}

type FactEvaluationRequest struct {
	OwnerSystem   string         `json:"owner_system"`
	SourceSystem  string         `json:"source_system"`
	FactType      string         `json:"fact_type"`
	SourceEventID string         `json:"source_event_id"`
	SubjectType   string         `json:"subject_type"`
	SubjectID     string         `json:"subject_id"`
	Facts         map[string]any `json:"facts"`
}

type FactEvaluationResponse struct {
	Evaluation FactEvaluation `json:"evaluation"`
	Findings   []Finding      `json:"findings"`
}

type FactEvaluation struct {
	ID string `json:"id"`
}

type Finding struct {
	ID             string         `json:"id"`
	Code           string         `json:"code"`
	Severity       string         `json:"severity"`
	Title          string         `json:"title"`
	Message        string         `json:"message"`
	Recommendation string         `json:"recommendation"`
	Evidence       map[string]any `json:"evidence"`
	Status         string         `json:"status"`
}

type AssistRunRequest struct {
	OwnerSystem    string         `json:"owner_system"`
	ProductSurface string         `json:"product_surface"`
	AssistType     string         `json:"assist_type"`
	SubjectType    string         `json:"subject_type"`
	SubjectID      string         `json:"subject_id"`
	Input          map[string]any `json:"input"`
}

type AssistRunResponse struct {
	ID      string         `json:"id"`
	Status  string         `json:"status"`
	Output  map[string]any `json:"output"`
	Error   string         `json:"error_message"`
	PackID  string         `json:"pack_id"`
	OrgID   string         `json:"org_id"`
	Created string         `json:"created_at"`
}

type ChatRequest struct {
	Message          string         `json:"message"`
	ChatID           string         `json:"chat_id,omitempty"`
	TaskID           string         `json:"task_id,omitempty"`
	AgentID          string         `json:"agent_id,omitempty"`
	ProductSurface   string         `json:"product_surface,omitempty"`
	RouteHint        string         `json:"route_hint,omitempty"`
	ConfirmedActions []string       `json:"confirmed_actions,omitempty"`
	Workspace        map[string]any `json:"workspace,omitempty"`
}

type ChatResponse struct {
	ChatID               string        `json:"chat_id,omitempty"`
	TaskID               string        `json:"task_id,omitempty"`
	RunID                string        `json:"run_id,omitempty"`
	AgentID              string        `json:"agent_id,omitempty"`
	Reply                string        `json:"reply"`
	Blocks               []any         `json:"blocks,omitempty"`
	ToolCalls            []any         `json:"tool_calls,omitempty"`
	PendingConfirmations []any         `json:"pending_confirmations,omitempty"`
	Messages             []ChatMessage `json:"messages,omitempty"`
}

type ChatConversationListResult struct {
	Items []ChatConversationSummary `json:"items"`
}

type ChatConversationSummary struct {
	ID        string `json:"id"`
	Title     string `json:"title,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

type ChatConversationDetail struct {
	ID        string        `json:"id"`
	Title     string        `json:"title,omitempty"`
	Messages  []ChatMessage `json:"messages"`
	UpdatedAt string        `json:"updated_at,omitempty"`
	CreatedAt string        `json:"created_at,omitempty"`
}

type ChatMessage struct {
	ID        string         `json:"id,omitempty"`
	Role      string         `json:"role"`
	Content   string         `json:"content"`
	CreatedAt string         `json:"created_at,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

func NewClient(cfg Config) *Client {
	return &Client{
		httpClient:       &http.Client{Timeout: 20 * time.Second},
		orgID:            strings.TrimSpace(cfg.OrgID),
		nexusBaseURL:     strings.TrimRight(strings.TrimSpace(cfg.NexusBaseURL), "/"),
		nexusAPIKey:      strings.TrimSpace(cfg.NexusAPIKey),
		companionBaseURL: strings.TrimRight(strings.TrimSpace(cfg.CompanionBaseURL), "/"),
		companionAPIKey:  strings.TrimSpace(cfg.CompanionAPIKey),
		publicBaseURL:    strings.TrimRight(strings.TrimSpace(cfg.PublicBaseURL), "/"),
	}
}

func (c *Client) OrgID() string {
	if c == nil || c.orgID == "" {
		return "argos-local-org"
	}
	return c.orgID
}

func (c *Client) PublicBaseURL() string {
	if c == nil {
		return ""
	}
	return c.publicBaseURL
}

func (c *Client) NexusConfigured() bool {
	return c != nil && c.nexusBaseURL != "" && c.nexusAPIKey != ""
}

func (c *Client) CompanionConfigured() bool {
	return c != nil && c.companionBaseURL != "" && c.companionAPIKey != ""
}

func (c *Client) SubmitFacts(ctx context.Context, req FactEvaluationRequest) (FactEvaluationResponse, error) {
	if !c.NexusConfigured() {
		return FactEvaluationResponse{}, ErrNotConfigured
	}
	var out FactEvaluationResponse
	err := c.postJSON(ctx, c.nexusBaseURL, c.nexusAPIKey, "/v1/fact-evaluations", req, &out)
	return out, err
}

func (c *Client) RunAssist(ctx context.Context, req AssistRunRequest) (AssistRunResponse, error) {
	if !c.CompanionConfigured() {
		return AssistRunResponse{}, ErrNotConfigured
	}
	var out AssistRunResponse
	err := c.postJSON(ctx, c.companionBaseURL, c.companionAPIKey, "/v1/assist-runs", req, &out)
	return out, err
}

func (c *Client) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	if !c.CompanionConfigured() {
		return ChatResponse{}, ErrNotConfigured
	}
	if strings.TrimSpace(req.ProductSurface) == "" {
		req.ProductSurface = "argos"
	}
	var out ChatResponse
	err := c.postJSON(ctx, c.companionBaseURL, c.companionAPIKey, "/v1/chat", req, &out)
	return out, err
}

func (c *Client) ListChatConversations(ctx context.Context, limit int) (ChatConversationListResult, error) {
	if !c.CompanionConfigured() {
		return ChatConversationListResult{}, ErrNotConfigured
	}
	values := url.Values{}
	values.Set("product_surface", "argos")
	if limit > 0 {
		values.Set("limit", fmt.Sprintf("%d", limit))
	}
	var out ChatConversationListResult
	err := c.getJSON(ctx, c.companionBaseURL, c.companionAPIKey, "/v1/chat/conversations", values, &out)
	return out, err
}

func (c *Client) GetChatConversation(ctx context.Context, id string) (ChatConversationDetail, error) {
	if !c.CompanionConfigured() {
		return ChatConversationDetail{}, ErrNotConfigured
	}
	values := url.Values{}
	values.Set("product_surface", "argos")
	var out ChatConversationDetail
	err := c.getJSON(ctx, c.companionBaseURL, c.companionAPIKey, "/v1/chat/conversations/"+strings.TrimSpace(id), values, &out)
	return out, err
}

func (c *Client) postJSON(ctx context.Context, baseURL, apiKey, path string, body any, out any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	endpoint, err := url.JoinPath(baseURL, path)
	if err != nil {
		return err
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-Key", apiKey)
	httpReq.Header.Set("X-Org-ID", c.OrgID())
	httpReq.Header.Set("X-Product-Surface", "argos")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("axis request failed: status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	if out == nil {
		return nil
	}
	if len(respBody) == 0 {
		return nil
	}
	return json.Unmarshal(respBody, out)
}

func (c *Client) getJSON(ctx context.Context, baseURL, apiKey, path string, query url.Values, out any) error {
	endpoint, err := url.JoinPath(baseURL, path)
	if err != nil {
		return err
	}
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	httpReq.Header.Set("X-API-Key", apiKey)
	httpReq.Header.Set("X-Org-ID", c.OrgID())
	httpReq.Header.Set("X-Product-Surface", "argos")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("axis request failed: status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	if out == nil || len(respBody) == 0 {
		return nil
	}
	return json.Unmarshal(respBody, out)
}

func FindingsSnapshot(items []Finding) []analysisdomain.FindingSnapshot {
	out := make([]analysisdomain.FindingSnapshot, 0, len(items))
	for _, item := range items {
		out = append(out, analysisdomain.FindingSnapshot{
			ID:             item.ID,
			Code:           item.Code,
			Severity:       item.Severity,
			Title:          item.Title,
			Message:        item.Message,
			Recommendation: item.Recommendation,
			Status:         item.Status,
			Evidence:       item.Evidence,
		})
	}
	return out
}

var ErrNotConfigured = fmt.Errorf("axis integration is not configured")
