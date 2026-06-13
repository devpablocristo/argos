package catalog

import (
	"path/filepath"
	"sort"
	"strings"
	"time"

	catalogdomain "github.com/devpablocristo/argos/core/internal/catalog/usecases/domain"
)

const (
	DatasetTypeSample              = "sample"
	DatasetTypeUploadedFolder      = "uploaded_folder"
	DatasetTypeFlightDataset       = "flight_dataset"
	DatasetTypeSingleCapture       = "single_capture"
	DatasetTypeMultiCaptureDataset = "multi_capture_dataset"
	DatasetTypeSectorCapture       = "sector_capture"
	DatasetTypeUnknown             = "unknown"

	ScopeGlobal  = "global"
	ScopeField   = "field"
	ScopeFlight  = "flight"
	ScopeDataset = "dataset"
)

func classifyDataset(ds catalogdomain.Dataset, captures []catalogdomain.Capture, outputDir string, now time.Time) catalogdomain.DatasetClassification {
	classification := catalogdomain.DatasetClassification{
		DatasetID:    ds.ID,
		DatasetName:  ds.Name,
		DatasetType:  DatasetTypeUnknown,
		Scope:        ScopeGlobal,
		Confidence:   0.25,
		Reason:       "No hay suficiente informacion para clasificar semanticamente el dataset.",
		ClassifiedAt: now.UTC(),
	}
	classification.CaptureIDs = captureIDs(captures)

	switch {
	case isSampleSource(ds.SourceURI):
		classification.DatasetType = DatasetTypeSample
		classification.Scope = ScopeGlobal
		classification.Confidence = 1
		classification.Reason = "El origen del dataset corresponde a /data/sample."
	case len(captures) == 0 && isUploadedSource(ds.SourceURI, outputDir):
		classification.DatasetType = DatasetTypeUploadedFolder
		classification.Scope = ScopeGlobal
		classification.Confidence = 0.60
		classification.Reason = "Es una carpeta subida manualmente y todavia no hay capturas detectadas."
	case len(captures) == 0:
		classification.DatasetType = DatasetTypeUnknown
		classification.Scope = ScopeGlobal
		classification.Confidence = 0.25
		classification.Reason = "No hay capturas ni metadata suficiente para clasificar el dataset."
	case hasSectorMetadata(captures):
		classification.DatasetType = DatasetTypeSectorCapture
		classification.Scope = ScopeDataset
		classification.Confidence = 0.78
		classification.Reason = "La metadata indica que el dataset representa un sector parcial."
	case hasFlightEvidence(captures):
		classification.DatasetType = DatasetTypeFlightDataset
		classification.Scope = ScopeFlight
		classification.Confidence = 0.78
		classification.Reason = "Hay multiples capturas con evidencia de fecha, GPS o estructura de mision de dron."
	case len(captures) == 1:
		classification.DatasetType = DatasetTypeSingleCapture
		classification.Scope = ScopeDataset
		classification.Confidence = 0.85
		classification.Reason = "El scan detecto un unico grupo coherente de imagenes/bandas."
	default:
		classification.DatasetType = DatasetTypeMultiCaptureDataset
		classification.Scope = ScopeDataset
		classification.Confidence = 0.82
		classification.Reason = "El scan detecto varios grupos coherentes de imagenes/bandas."
	}

	if ds.FieldID != nil && *ds.FieldID != "" {
		classification.FieldID = ds.FieldID
		classification.Scope = ScopeField
	}
	classification.MissingMetadata = missingMetadata(classification, captures)
	return classification
}

func captureIDs(captures []catalogdomain.Capture) []string {
	ids := make([]string, 0, len(captures))
	for _, capture := range captures {
		if capture.ID != "" {
			ids = append(ids, capture.ID)
		}
	}
	sort.Strings(ids)
	return ids
}

func isSampleSource(sourceURI string) bool {
	normalized := normalizeSourceURI(sourceURI)
	return normalized == "/data/sample" ||
		normalized == "../sample" ||
		normalized == "sample" ||
		strings.HasSuffix(normalized, "/data/sample")
}

func isUploadedSource(sourceURI, outputDir string) bool {
	normalized := normalizeSourceURI(sourceURI)
	if strings.Contains(normalized, "/uploads/") || strings.HasSuffix(normalized, "/uploads") {
		return true
	}
	if outputDir == "" {
		return false
	}
	uploadRoot := strings.TrimRight(normalizeSourceURI(filepath.Join(outputDir, "uploads")), "/")
	return normalized == uploadRoot || strings.HasPrefix(normalized, uploadRoot+"/")
}

func normalizeSourceURI(value string) string {
	value = filepath.ToSlash(strings.TrimSpace(value))
	if value == "" {
		return ""
	}
	if strings.Contains(value, "://") {
		return strings.TrimRight(value, "/")
	}
	cleaned := filepath.ToSlash(filepath.Clean(value))
	if value == "/" {
		return "/"
	}
	return strings.TrimRight(cleaned, "/")
}

func hasSectorMetadata(captures []catalogdomain.Capture) bool {
	for _, capture := range captures {
		if metadataMarksSector(capture.Warnings) || metadataMarksSector(capture.Errors) {
			return true
		}
		for _, asset := range capture.Assets {
			if metadataMapMarksSector(asset.Metadata) {
				return true
			}
		}
	}
	return false
}

func metadataMarksSector(values []string) bool {
	for _, value := range values {
		value = strings.ToLower(value)
		if strings.Contains(value, "sector") || strings.Contains(value, "partial") {
			return true
		}
	}
	return false
}

func metadataMapMarksSector(metadata map[string]any) bool {
	for _, key := range []string{"scope", "dataset_scope", "capture_scope", "area_type"} {
		value, ok := metadata[key]
		if !ok {
			continue
		}
		text := strings.ToLower(strings.TrimSpace(toString(value)))
		if text == "sector" || text == "partial" || text == "sector_capture" {
			return true
		}
	}
	return false
}

func hasFlightEvidence(captures []catalogdomain.Capture) bool {
	if len(captures) < 2 {
		return false
	}
	score := 0
	if hasAnyCaptureTime(captures) {
		score++
	}
	if hasAnyGPS(captures) {
		score++
	}
	if hasAnyDroneMetadata(captures) {
		score++
	}
	return score >= 2
}

func hasAnyCaptureTime(captures []catalogdomain.Capture) bool {
	for _, capture := range captures {
		if strings.TrimSpace(capture.CapturedAt) != "" {
			return true
		}
	}
	return false
}

func hasAnyGPS(captures []catalogdomain.Capture) bool {
	for _, capture := range captures {
		if capture.Location.Lat != nil && capture.Location.Lon != nil {
			return true
		}
	}
	return false
}

func hasAnyDroneMetadata(captures []catalogdomain.Capture) bool {
	for _, capture := range captures {
		if strings.HasPrefix(strings.ToUpper(capture.CaptureKey), "DJI_") {
			return true
		}
		for _, asset := range capture.Assets {
			if _, ok := asset.Metadata["camera"]; ok {
				return true
			}
			if _, ok := asset.Metadata["exif"]; ok {
				return true
			}
		}
	}
	return false
}

func missingMetadata(classification catalogdomain.DatasetClassification, captures []catalogdomain.Capture) []string {
	var missing []string
	if classification.FieldID == nil {
		missing = append(missing, "field_id")
	}
	missing = append(missing, "lot_id", "campaign_id")
	if classification.FlightID == nil {
		missing = append(missing, "flight_id")
	}
	if len(captures) == 0 || !hasAnyGPS(captures) {
		missing = append(missing, "gps")
	}
	if len(captures) == 0 || !hasAnyCaptureTime(captures) {
		missing = append(missing, "capture_time")
	}
	if len(captures) == 0 || !hasAnyDroneMetadata(captures) {
		missing = append(missing, "drone_metadata")
	}
	return missing
}

func toString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	default:
		return ""
	}
}
