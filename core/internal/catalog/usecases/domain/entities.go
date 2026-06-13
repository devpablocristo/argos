package domain

import "time"

type Dataset struct {
	ID             string                 `json:"id"`
	OrgID          string                 `json:"org_id,omitempty"`
	Name           string                 `json:"name"`
	SourceURI      string                 `json:"source_uri"`
	Status         string                 `json:"status"`
	FieldID        *string                `json:"field_id,omitempty"`
	Classification *DatasetClassification `json:"classification,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	ArchivedAt     *time.Time             `json:"archived_at,omitempty"`
}

type DatasetClassification struct {
	DatasetID       string    `json:"dataset_id"`
	DatasetName     string    `json:"dataset_name"`
	DatasetType     string    `json:"dataset_type"`
	Scope           string    `json:"scope"`
	FieldID         *string   `json:"field_id"`
	LotID           *string   `json:"lot_id"`
	CampaignID      *string   `json:"campaign_id"`
	FlightID        *string   `json:"flight_id"`
	CaptureIDs      []string  `json:"capture_ids"`
	AnalysisIDs     []string  `json:"analysis_ids"`
	Confidence      float64   `json:"confidence"`
	MissingMetadata []string  `json:"missing_metadata"`
	Reason          string    `json:"reason"`
	ClassifiedAt    time.Time `json:"classified_at"`
}

type DatasetEvent struct {
	EventID   string         `json:"event_id"`
	DatasetID string         `json:"dataset_id"`
	EventType string         `json:"event_type"`
	Timestamp time.Time      `json:"timestamp"`
	Status    string         `json:"status"`
	Message   string         `json:"message"`
	Details   map[string]any `json:"details"`
}

type Capture struct {
	ID               string         `json:"id"`
	DatasetID        string         `json:"dataset_id"`
	CaptureKey       string         `json:"capture_key"`
	CapturedAt       string         `json:"captured_at,omitempty"`
	Location         Location       `json:"location"`
	ValidationStatus string         `json:"validation_status"`
	Warnings         []string       `json:"warnings,omitempty"`
	Errors           []string       `json:"errors,omitempty"`
	Assets           []BandAsset    `json:"assets"`
	Analysis         *AnalysisDraft `json:"analysis,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
}

type Location struct {
	Lat  *float64 `json:"lat,omitempty"`
	Lon  *float64 `json:"lon,omitempty"`
	AltM *float64 `json:"alt_m,omitempty"`
	CRS  string   `json:"crs"`
}

type BandAsset struct {
	ID             string         `json:"id"`
	CaptureID      string         `json:"capture_id"`
	Band           string         `json:"band"`
	Role           string         `json:"role"`
	Path           string         `json:"path"`
	ChecksumSHA256 string         `json:"checksum_sha256"`
	MimeType       string         `json:"mime_type"`
	Width          int            `json:"width"`
	Height         int            `json:"height"`
	BitDepth       int            `json:"bit_depth"`
	WavelengthNM   *int           `json:"wavelength_nm,omitempty"`
	FWHMNM         *int           `json:"fwhm_nm,omitempty"`
	Metadata       map[string]any `json:"metadata,omitempty"`
}

// AnalysisDraft is a processing result attached to a scanned capture. The
// analyses module persists it as a first-class Analysis when requested.
type AnalysisDraft struct {
	Kind     string         `json:"kind"`
	Status   string         `json:"status"`
	Metrics  map[string]any `json:"metrics"`
	Warnings []string       `json:"warnings,omitempty"`
	Outputs  []OutputAsset  `json:"outputs"`
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
