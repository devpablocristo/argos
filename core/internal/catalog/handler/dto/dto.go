package dto

type CreateDatasetRequest struct {
	Name      string `json:"name"`
	SourceURI string `json:"source_uri"`
}

type DatasetListResponse struct {
	Datasets any `json:"datasets"`
}

type CaptureListResponse struct {
	Captures any `json:"captures"`
}

type ScanDatasetResponse struct {
	Status   string   `json:"status"`
	Captures any      `json:"captures"`
	Warnings []string `json:"warnings,omitempty"`
}
