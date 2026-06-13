package dto

type CreateDatasetRequest struct {
	Name      string `json:"name"`
	SourceURI string `json:"source_uri"`
}

type UpdateDatasetRequest struct {
	Name      *string `json:"name,omitempty"`
	SourceURI *string `json:"source_uri,omitempty"`
}

type UpdateDatasetFieldRequest struct {
	FieldID *string `json:"field_id"`
}

type DatasetListResponse struct {
	Datasets any `json:"datasets"`
}

type CaptureListResponse struct {
	Captures any `json:"captures"`
}

type DatasetEventListResponse struct {
	Events any `json:"events"`
}

type ScanDatasetResponse struct {
	Status   string   `json:"status"`
	Captures any      `json:"captures"`
	Warnings []string `json:"warnings,omitempty"`
}

type UploadScanDatasetResponse struct {
	Dataset  any      `json:"dataset"`
	Status   string   `json:"status"`
	Captures any      `json:"captures"`
	Warnings []string `json:"warnings,omitempty"`
}
