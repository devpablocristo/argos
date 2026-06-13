package domain

import "time"

type Analysis struct {
	ID                     string            `json:"id"`
	DatasetID              string            `json:"dataset_id,omitempty"`
	CaptureID              string            `json:"capture_id,omitempty"`
	Kind                   string            `json:"kind"`
	Status                 string            `json:"status"`
	Metrics                map[string]any    `json:"metrics"`
	Warnings               []string          `json:"warnings,omitempty"`
	Outputs                []OutputAsset     `json:"outputs"`
	NexusSyncStatus        string            `json:"nexus_sync_status,omitempty"`
	NexusSyncError         string            `json:"nexus_sync_error,omitempty"`
	NexusCorrelationID     string            `json:"nexus_correlation_id,omitempty"`
	NexusFindings          []FindingSnapshot `json:"nexus_findings,omitempty"`
	CompanionSyncStatus    string            `json:"companion_sync_status,omitempty"`
	CompanionSyncError     string            `json:"companion_sync_error,omitempty"`
	CompanionCorrelationID string            `json:"companion_correlation_id,omitempty"`
	CompanionOutput        map[string]any    `json:"companion_output,omitempty"`
	CreatedAt              time.Time         `json:"created_at"`
	CompletedAt            *time.Time        `json:"completed_at,omitempty"`
}

type OutputAsset struct {
	ID          string         `json:"id"`
	AnalysisID  string         `json:"analysis_id,omitempty"`
	Kind        string         `json:"kind"`
	Path        string         `json:"path"`
	ContentType string         `json:"content_type"`
	ByteSize    int64          `json:"byte_size"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type FindingSnapshot struct {
	ID             string         `json:"id,omitempty"`
	Code           string         `json:"code"`
	Severity       string         `json:"severity"`
	Title          string         `json:"title"`
	Message        string         `json:"message"`
	Recommendation string         `json:"recommendation,omitempty"`
	Status         string         `json:"status"`
	Evidence       map[string]any `json:"evidence,omitempty"`
}
