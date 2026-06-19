package axis

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClientSubmitFactsSendsServicePrincipalContext(t *testing.T) {
	var gotOrg string
	var gotKey string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotOrg = r.Header.Get("X-Org-ID")
		gotKey = r.Header.Get("X-API-Key")
		if r.URL.Path != "/v1/fact-evaluations" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(FactEvaluationResponse{
			Evaluation: FactEvaluation{ID: "eval-1"},
			Findings:   []Finding{{ID: "finding-1", Code: "argos.test", Severity: "info", Title: "Test", Message: "ok", Status: "open"}},
		})
	}))
	defer server.Close()

	client := NewClient(Config{OrgID: "argos-local-org", NexusBaseURL: server.URL, NexusAPIKey: "secret"})
	out, err := client.SubmitFacts(context.Background(), FactEvaluationRequest{Facts: map[string]any{"mean": 0.4}})
	if err != nil {
		t.Fatalf("SubmitFacts: %v", err)
	}
	if gotOrg != "argos-local-org" {
		t.Fatalf("org header mismatch: %q", gotOrg)
	}
	if gotKey != "secret" {
		t.Fatalf("api key header mismatch: %q", gotKey)
	}
	if out.Evaluation.ID != "eval-1" || len(out.Findings) != 1 {
		t.Fatalf("unexpected response: %+v", out)
	}
}

func TestClientRunAssistRequiresConfiguration(t *testing.T) {
	client := NewClient(Config{OrgID: "argos-local-org"})
	if _, err := client.RunAssist(context.Background(), AssistRunRequest{}); !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("expected ErrNotConfigured, got %v", err)
	}
}

func TestClientChatSendsArgosProductContext(t *testing.T) {
	var gotOrg string
	var gotProduct string
	var gotBody ChatRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotOrg = r.Header.Get("X-Org-ID")
		gotProduct = r.Header.Get("X-Product-Surface")
		if r.URL.Path != "/v1/chat" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		_ = json.NewEncoder(w).Encode(ChatResponse{ChatID: "chat-1", Reply: "ok"})
	}))
	defer server.Close()

	client := NewClient(Config{OrgID: "argos-local-org", CompanionBaseURL: server.URL, CompanionAPIKey: "secret"})
	out, err := client.Chat(context.Background(), ChatRequest{Message: "explain"})
	if err != nil {
		t.Fatalf("Chat: %v", err)
	}
	if out.ChatID != "chat-1" || out.Reply != "ok" {
		t.Fatalf("unexpected response: %+v", out)
	}
	if gotOrg != "argos-local-org" {
		t.Fatalf("org header mismatch: %q", gotOrg)
	}
	if gotProduct != "argos" {
		t.Fatalf("product header mismatch: %q", gotProduct)
	}
	if gotBody.ProductSurface != "argos" {
		t.Fatalf("product body mismatch: %q", gotBody.ProductSurface)
	}
}
