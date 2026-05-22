package domain

import "time"

type Analysis struct {
	ID          string         `json:"id"`
	DatasetID   string         `json:"dataset_id,omitempty"`
	CaptureID   string         `json:"capture_id,omitempty"`
	Kind        string         `json:"kind"`
	Status      string         `json:"status"`
	Metrics     map[string]any `json:"metrics"`
	Warnings    []string       `json:"warnings,omitempty"`
	Outputs     []OutputAsset  `json:"outputs"`
	CreatedAt   time.Time      `json:"created_at"`
	CompletedAt *time.Time     `json:"completed_at,omitempty"`
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
