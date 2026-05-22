package processor

type Response struct {
	SchemaVersion string          `json:"schema_version"`
	Status        string          `json:"status"`
	InputPath     string          `json:"input_path"`
	OutputPath    string          `json:"output_path"`
	Captures      []CaptureResult `json:"captures"`
	Warnings      []string        `json:"warnings"`
}

type CaptureResult struct {
	SchemaVersion string          `json:"schema_version"`
	Vendor        string          `json:"vendor"`
	Platform      string          `json:"platform"`
	CaptureKey    string          `json:"capture_key"`
	CapturedAt    string          `json:"captured_at"`
	Location      LocationResult  `json:"location"`
	Assets        []BandResult    `json:"assets"`
	Validation    Validation      `json:"validation"`
	Analysis      *AnalysisResult `json:"analysis"`
}

type LocationResult struct {
	Lat  *float64 `json:"lat"`
	Lon  *float64 `json:"lon"`
	AltM *float64 `json:"alt_m"`
	CRS  string   `json:"crs"`
}

type BandResult struct {
	Band           string         `json:"band"`
	Role           string         `json:"role"`
	Path           string         `json:"path"`
	ChecksumSHA256 string         `json:"checksum_sha256"`
	MimeType       string         `json:"mime_type"`
	Width          int            `json:"width"`
	Height         int            `json:"height"`
	BitDepth       int            `json:"bit_depth"`
	WavelengthNM   *int           `json:"wavelength_nm"`
	FWHMNM         *int           `json:"fwhm_nm"`
	SourceMetadata map[string]any `json:"source_metadata"`
}

type Validation struct {
	Status   string   `json:"status"`
	Warnings []string `json:"warnings"`
	Errors   []string `json:"errors"`
}

type AnalysisResult struct {
	SchemaVersion string         `json:"schema_version"`
	Kind          string         `json:"kind"`
	Status        string         `json:"status"`
	Outputs       []OutputResult `json:"outputs"`
	Metrics       map[string]any `json:"metrics"`
	Warnings      []string       `json:"warnings"`
}

type OutputResult struct {
	Kind        string         `json:"kind"`
	Path        string         `json:"path"`
	ContentType string         `json:"content_type"`
	ByteSize    int64          `json:"byte_size"`
	Metadata    map[string]any `json:"metadata"`
}
