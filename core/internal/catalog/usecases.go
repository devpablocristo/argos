package catalog

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	catalogdomain "github.com/devpablocristo/argos/core/internal/catalog/usecases/domain"
	"github.com/devpablocristo/argos/core/internal/processor"
	"github.com/google/uuid"
)

type Usecases struct {
	repo           Repository
	worker         processor.Processor
	outputDir      string
	fieldValidator FieldValidator
}

type UploadedFile struct {
	Name string
	Open func() (io.ReadCloser, error)
}

type FieldValidator interface {
	EnsureField(ctx context.Context, id string) error
}

func NewUsecases(repo Repository, worker processor.Processor, outputDir string, fieldValidator ...FieldValidator) *Usecases {
	var validator FieldValidator
	if len(fieldValidator) > 0 {
		validator = fieldValidator[0]
	}
	return &Usecases{repo: repo, worker: worker, outputDir: outputDir, fieldValidator: validator}
}

func (u *Usecases) CreateDataset(ctx context.Context, name, sourceURI string) (catalogdomain.Dataset, error) {
	return u.createDataset(ctx, name, sourceURI, nil)
}

func (u *Usecases) CreateDatasetForField(ctx context.Context, fieldID, name, sourceURI string) (catalogdomain.Dataset, error) {
	if err := u.ensureField(ctx, fieldID); err != nil {
		return catalogdomain.Dataset{}, err
	}
	return u.createDataset(ctx, name, sourceURI, &fieldID)
}

func (u *Usecases) createDataset(ctx context.Context, name, sourceURI string, fieldID *string) (catalogdomain.Dataset, error) {
	ds, err := u.repo.CreateDataset(ctx, name, sourceURI, fieldID)
	if err != nil {
		return catalogdomain.Dataset{}, err
	}
	if err := u.RecordDatasetEvent(ctx, ds.ID, "DATASET_CREATED", "completed", "Dataset registered in Argos.", map[string]any{
		"name":       ds.Name,
		"source_uri": ds.SourceURI,
		"field_id":   ds.FieldID,
	}); err != nil {
		return catalogdomain.Dataset{}, err
	}
	classification, err := u.classifyAndStore(ctx, ds, nil)
	if err != nil {
		return catalogdomain.Dataset{}, err
	}
	ds.Classification = &classification
	return ds, nil
}

func (u *Usecases) UploadAndScanDataset(ctx context.Context, name string, files []UploadedFile) (catalogdomain.Dataset, []catalogdomain.Capture, []string, string, error) {
	return u.uploadAndScanDataset(ctx, nil, name, files)
}

func (u *Usecases) UploadAndScanDatasetForField(ctx context.Context, fieldID, name string, files []UploadedFile) (catalogdomain.Dataset, []catalogdomain.Capture, []string, string, error) {
	if err := u.ensureField(ctx, fieldID); err != nil {
		return catalogdomain.Dataset{}, nil, nil, "", err
	}
	return u.uploadAndScanDataset(ctx, &fieldID, name, files)
}

func (u *Usecases) uploadAndScanDataset(ctx context.Context, fieldID *string, name string, files []UploadedFile) (catalogdomain.Dataset, []catalogdomain.Capture, []string, string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		name = "Uploaded dataset"
	}
	if len(files) == 0 {
		return catalogdomain.Dataset{}, nil, nil, "", ErrValidation
	}
	root := filepath.Join(u.outputDir, "uploads", uuid.NewString())
	if err := os.MkdirAll(root, 0o755); err != nil {
		return catalogdomain.Dataset{}, nil, nil, "", err
	}
	if err := writeUploadedFiles(root, files); err != nil {
		_ = os.RemoveAll(root)
		return catalogdomain.Dataset{}, nil, nil, "", err
	}
	ds, err := u.repo.CreateDataset(ctx, name, root, fieldID)
	if err != nil {
		_ = os.RemoveAll(root)
		return catalogdomain.Dataset{}, nil, nil, "", err
	}
	if err := u.RecordDatasetEvent(ctx, ds.ID, "DATASET_CREATED", "completed", "Dataset registered in Argos.", map[string]any{
		"name":       ds.Name,
		"source_uri": ds.SourceURI,
		"field_id":   ds.FieldID,
	}); err != nil {
		return catalogdomain.Dataset{}, nil, nil, "", err
	}
	if err := u.RecordDatasetEvent(ctx, ds.ID, "DATASET_UPLOADED", "completed", "Image files uploaded into Argos storage.", map[string]any{
		"file_count": len(files),
		"source_uri": ds.SourceURI,
	}); err != nil {
		return catalogdomain.Dataset{}, nil, nil, "", err
	}
	if _, err := u.classifyAndStore(ctx, ds, nil); err != nil {
		return catalogdomain.Dataset{}, nil, nil, "", err
	}
	captures, warnings, status, err := u.ScanDataset(ctx, ds.ID)
	if updated, getErr := u.repo.GetDataset(ctx, ds.ID); getErr == nil {
		ds = updated
	}
	return ds, captures, warnings, status, err
}

func (u *Usecases) ListDatasets(ctx context.Context, includeArchived bool) ([]catalogdomain.Dataset, error) {
	datasets, err := u.repo.ListDatasets(ctx, includeArchived)
	return u.ensureDatasetClassifications(ctx, datasets, err)
}

func (u *Usecases) ListDatasetsByField(ctx context.Context, fieldID string, includeArchived bool) ([]catalogdomain.Dataset, error) {
	if err := u.ensureField(ctx, fieldID); err != nil {
		return nil, err
	}
	datasets, err := u.repo.ListDatasetsByField(ctx, fieldID, includeArchived)
	return u.ensureDatasetClassifications(ctx, datasets, err)
}

func (u *Usecases) ensureDatasetClassifications(ctx context.Context, datasets []catalogdomain.Dataset, err error) ([]catalogdomain.Dataset, error) {
	if err != nil {
		return nil, err
	}
	for i := range datasets {
		if datasets[i].Classification != nil {
			continue
		}
		classification, err := u.classifyAndStore(ctx, datasets[i], nil)
		if err != nil {
			return nil, err
		}
		datasets[i].Classification = &classification
	}
	return datasets, nil
}

func (u *Usecases) GetDataset(ctx context.Context, id string) (catalogdomain.Dataset, error) {
	return u.repo.GetDataset(ctx, id)
}

func (u *Usecases) UpdateDataset(ctx context.Context, id string, name, sourceURI *string) (catalogdomain.Dataset, error) {
	ds, err := u.repo.GetDataset(ctx, id)
	if err != nil {
		return catalogdomain.Dataset{}, err
	}
	nextName := ds.Name
	if name != nil {
		nextName = strings.TrimSpace(*name)
	}
	nextSourceURI := ds.SourceURI
	if sourceURI != nil {
		nextSourceURI = strings.TrimSpace(*sourceURI)
	}
	if nextName == "" || nextSourceURI == "" {
		return catalogdomain.Dataset{}, ErrValidation
	}
	updated, err := u.repo.UpdateDataset(ctx, id, nextName, nextSourceURI)
	if err != nil {
		return catalogdomain.Dataset{}, err
	}
	captures, _ := u.repo.ListCaptures(ctx, updated.ID)
	classification, err := u.classifyAndStore(ctx, updated, captures)
	if err != nil {
		return catalogdomain.Dataset{}, err
	}
	updated.Classification = &classification
	return updated, nil
}

func (u *Usecases) UpdateDatasetField(ctx context.Context, id string, fieldID *string) (catalogdomain.Dataset, error) {
	if fieldID != nil {
		trimmed := strings.TrimSpace(*fieldID)
		if trimmed == "" {
			fieldID = nil
		} else {
			if err := u.ensureField(ctx, trimmed); err != nil {
				return catalogdomain.Dataset{}, err
			}
			fieldID = &trimmed
		}
	}
	updated, err := u.repo.UpdateDatasetField(ctx, id, fieldID)
	if err != nil {
		return catalogdomain.Dataset{}, err
	}
	captures, _ := u.repo.ListCaptures(ctx, updated.ID)
	classification, err := u.classifyAndStore(ctx, updated, captures)
	if err != nil {
		return catalogdomain.Dataset{}, err
	}
	updated.Classification = &classification
	return updated, nil
}

func (u *Usecases) ArchiveDataset(ctx context.Context, id string) (catalogdomain.Dataset, error) {
	return u.repo.ArchiveDataset(ctx, id)
}

func (u *Usecases) RestoreDataset(ctx context.Context, id string) (catalogdomain.Dataset, error) {
	return u.repo.RestoreDataset(ctx, id)
}

func (u *Usecases) DeleteDataset(ctx context.Context, id string) (catalogdomain.Dataset, error) {
	current, err := u.repo.GetDataset(ctx, id)
	if err != nil {
		return catalogdomain.Dataset{}, err
	}
	if current.ArchivedAt == nil {
		return catalogdomain.Dataset{}, ErrDatasetNotArchived
	}
	ds, err := u.repo.DeleteDataset(ctx, id)
	if err != nil {
		return catalogdomain.Dataset{}, err
	}
	if u.outputDir != "" {
		if err := os.RemoveAll(datasetOutputPath(u.outputDir, id)); err != nil {
			return catalogdomain.Dataset{}, err
		}
	}
	return ds, nil
}

func (u *Usecases) ScanDataset(ctx context.Context, datasetID string) ([]catalogdomain.Capture, []string, string, error) {
	ds, err := u.repo.GetDataset(ctx, datasetID)
	if err != nil {
		return nil, nil, "", err
	}
	if ds.ArchivedAt != nil {
		return nil, nil, "", ErrDatasetArchived
	}
	if err := u.repo.UpdateDatasetStatus(ctx, datasetID, "processing"); err != nil {
		return nil, nil, "", err
	}
	if err := u.RecordDatasetEvent(ctx, datasetID, "ANALYSIS_STARTED", "running", "Processing worker started dataset analysis.", map[string]any{
		"source_uri": ds.SourceURI,
	}); err != nil {
		return nil, nil, "", err
	}
	response, err := u.worker.ProcessCaptureGroups(ctx, ds.SourceURI, datasetOutputPath(u.outputDir, datasetID))
	if err != nil {
		_ = u.repo.UpdateDatasetStatus(ctx, datasetID, "failed")
		_ = u.RecordDatasetEvent(ctx, datasetID, "ANALYSIS_COMPLETED", "failed", "Processing worker failed dataset analysis.", map[string]any{
			"error": err.Error(),
		})
		_ = u.RecordDatasetEvent(ctx, datasetID, "ERROR", "failed", "Dataset processing failed.", map[string]any{
			"error": err.Error(),
		})
		return nil, nil, "", err
	}
	captures := convertCaptures(datasetID, response.Captures)
	if err := u.repo.ReplaceCaptures(ctx, datasetID, captures); err != nil {
		_ = u.repo.UpdateDatasetStatus(ctx, datasetID, "failed")
		_ = u.RecordDatasetEvent(ctx, datasetID, "ERROR", "failed", "Persisting detected captures failed.", map[string]any{
			"error": err.Error(),
		})
		return nil, nil, "", err
	}
	if err := u.RecordDatasetEvent(ctx, datasetID, "METADATA_EXTRACTED", "completed", "Capture metadata extracted from source files.", map[string]any{
		"input_path":  response.InputPath,
		"output_path": response.OutputPath,
		"warnings":    response.Warnings,
	}); err != nil {
		return nil, nil, "", err
	}
	if err := u.RecordDatasetEvent(ctx, datasetID, "CAPTURES_DETECTED", "completed", "Capture groups detected for dataset.", map[string]any{
		"capture_count": len(captures),
		"capture_ids":   captureIDs(captures),
	}); err != nil {
		return nil, nil, "", err
	}
	if _, err := u.classifyAndStore(ctx, ds, captures); err != nil {
		return nil, nil, "", err
	}
	status := "failed"
	if response.Status == "completed" {
		status = "processed"
	}
	if err := u.repo.UpdateDatasetStatus(ctx, datasetID, status); err != nil && !errors.Is(err, ErrNotFound) {
		return nil, nil, "", err
	}
	eventStatus := "failed"
	if status == "processed" {
		eventStatus = "completed"
	}
	if err := u.RecordDatasetEvent(ctx, datasetID, "ANALYSIS_COMPLETED", eventStatus, "Processing worker completed dataset analysis.", map[string]any{
		"capture_count": len(captures),
		"worker_status": response.Status,
	}); err != nil {
		return nil, nil, "", err
	}
	for _, event := range indexGeneratedEvents(datasetID, captures) {
		if err := u.repo.AppendDatasetEvent(ctx, event); err != nil {
			return nil, nil, "", err
		}
	}
	return captures, response.Warnings, response.Status, nil
}

var (
	ErrDatasetArchived    = errors.New("dataset archived")
	ErrDatasetNotArchived = errors.New("dataset must be archived before delete")
	ErrValidation         = errors.New("validation error")
)

func (u *Usecases) ListCaptures(ctx context.Context, datasetID string) ([]catalogdomain.Capture, error) {
	if _, err := u.repo.GetDataset(ctx, datasetID); err != nil {
		return nil, err
	}
	return u.repo.ListCaptures(ctx, datasetID)
}

func (u *Usecases) GetCapture(ctx context.Context, id string) (catalogdomain.Capture, error) {
	return u.repo.GetCapture(ctx, id)
}

func (u *Usecases) ClassifyDataset(ctx context.Context, datasetID string) (catalogdomain.DatasetClassification, error) {
	ds, err := u.repo.GetDataset(ctx, datasetID)
	if err != nil {
		return catalogdomain.DatasetClassification{}, err
	}
	captures, err := u.repo.ListCaptures(ctx, datasetID)
	if err != nil {
		return catalogdomain.DatasetClassification{}, err
	}
	return u.classifyAndStore(ctx, ds, captures)
}

func (u *Usecases) GetDatasetClassification(ctx context.Context, datasetID string) (catalogdomain.DatasetClassification, error) {
	if _, err := u.repo.GetDataset(ctx, datasetID); err != nil {
		return catalogdomain.DatasetClassification{}, err
	}
	classification, err := u.repo.GetDatasetClassification(ctx, datasetID)
	if errors.Is(err, ErrNotFound) {
		return u.ClassifyDataset(ctx, datasetID)
	}
	return classification, err
}

func (u *Usecases) RecordDatasetEvent(ctx context.Context, datasetID, eventType, status, message string, details map[string]any) error {
	if details == nil {
		details = map[string]any{}
	}
	return u.repo.AppendDatasetEvent(ctx, catalogdomain.DatasetEvent{
		EventID:   uuid.NewString(),
		DatasetID: datasetID,
		EventType: eventType,
		Timestamp: time.Now().UTC(),
		Status:    status,
		Message:   message,
		Details:   details,
	})
}

func (u *Usecases) ListDatasetEvents(ctx context.Context, datasetID string) ([]catalogdomain.DatasetEvent, error) {
	return u.repo.ListDatasetEvents(ctx, datasetID)
}

func (u *Usecases) GetCaptureAsset(ctx context.Context, captureID, band string) (catalogdomain.BandAsset, error) {
	capture, err := u.repo.GetCapture(ctx, captureID)
	if err != nil {
		return catalogdomain.BandAsset{}, err
	}
	for _, asset := range capture.Assets {
		if asset.Band == band {
			return asset, nil
		}
	}
	return catalogdomain.BandAsset{}, ErrNotFound
}

func (u *Usecases) classifyAndStore(ctx context.Context, ds catalogdomain.Dataset, captures []catalogdomain.Capture) (catalogdomain.DatasetClassification, error) {
	if captures == nil {
		currentCaptures, err := u.repo.ListCaptures(ctx, ds.ID)
		if err == nil {
			captures = currentCaptures
		}
	}
	classification := classifyDataset(ds, captures, u.outputDir, time.Now().UTC())
	if err := u.repo.UpsertDatasetClassification(ctx, classification); err != nil {
		return catalogdomain.DatasetClassification{}, err
	}
	if err := u.RecordDatasetEvent(ctx, ds.ID, "DATASET_CLASSIFIED", "completed", "Dataset semantic classification updated.", map[string]any{
		"dataset_type":     classification.DatasetType,
		"scope":            classification.Scope,
		"confidence":       classification.Confidence,
		"missing_metadata": classification.MissingMetadata,
	}); err != nil {
		return catalogdomain.DatasetClassification{}, err
	}
	stored, err := u.repo.GetDatasetClassification(ctx, ds.ID)
	if err != nil {
		return classification, nil
	}
	return stored, nil
}

func (u *Usecases) ensureField(ctx context.Context, fieldID string) error {
	fieldID = strings.TrimSpace(fieldID)
	if fieldID == "" {
		return ErrValidation
	}
	if u.fieldValidator == nil {
		return nil
	}
	if err := u.fieldValidator.EnsureField(ctx, fieldID); err != nil {
		return ErrNotFound
	}
	return nil
}

func indexGeneratedEvents(datasetID string, captures []catalogdomain.Capture) []catalogdomain.DatasetEvent {
	var events []catalogdomain.DatasetEvent
	for _, capture := range captures {
		if capture.Analysis == nil {
			continue
		}
		for _, output := range capture.Analysis.Outputs {
			events = append(events, catalogdomain.DatasetEvent{
				EventID:   uuid.NewString(),
				DatasetID: datasetID,
				EventType: "INDEX_GENERATED",
				Timestamp: time.Now().UTC(),
				Status:    "completed",
				Message:   "Processing worker generated an analysis output.",
				Details: map[string]any{
					"capture_id":  capture.ID,
					"capture_key": capture.CaptureKey,
					"kind":        output.Kind,
					"path":        output.Path,
				},
			})
		}
	}
	return events
}

func convertCaptures(datasetID string, in []processor.CaptureResult) []catalogdomain.Capture {
	out := make([]catalogdomain.Capture, 0, len(in))
	for _, item := range in {
		captureID := uuid.NewString()
		assets := make([]catalogdomain.BandAsset, 0, len(item.Assets))
		for _, asset := range item.Assets {
			assets = append(assets, catalogdomain.BandAsset{
				ID:             uuid.NewString(),
				CaptureID:      captureID,
				Band:           asset.Band,
				Role:           asset.Role,
				Path:           asset.Path,
				ChecksumSHA256: asset.ChecksumSHA256,
				MimeType:       asset.MimeType,
				Width:          asset.Width,
				Height:         asset.Height,
				BitDepth:       asset.BitDepth,
				WavelengthNM:   asset.WavelengthNM,
				FWHMNM:         asset.FWHMNM,
				Metadata:       asset.SourceMetadata,
			})
		}
		var analysis *catalogdomain.AnalysisDraft
		if item.Analysis != nil {
			outputs := make([]catalogdomain.OutputAsset, 0, len(item.Analysis.Outputs))
			for _, output := range item.Analysis.Outputs {
				outputs = append(outputs, catalogdomain.OutputAsset{
					Kind:        output.Kind,
					Path:        output.Path,
					ContentType: output.ContentType,
					ByteSize:    output.ByteSize,
					Metadata:    output.Metadata,
				})
			}
			analysis = &catalogdomain.AnalysisDraft{
				Kind:     item.Analysis.Kind,
				Status:   item.Analysis.Status,
				Metrics:  item.Analysis.Metrics,
				Warnings: item.Analysis.Warnings,
				Outputs:  outputs,
			}
		}
		out = append(out, catalogdomain.Capture{
			ID:               captureID,
			DatasetID:        datasetID,
			CaptureKey:       item.CaptureKey,
			CapturedAt:       item.CapturedAt,
			Location:         catalogdomain.Location{Lat: item.Location.Lat, Lon: item.Location.Lon, AltM: item.Location.AltM, CRS: item.Location.CRS},
			ValidationStatus: item.Validation.Status,
			Warnings:         item.Validation.Warnings,
			Errors:           item.Validation.Errors,
			Assets:           assets,
			Analysis:         analysis,
			CreatedAt:        time.Now().UTC(),
		})
	}
	return out
}

func writeUploadedFiles(root string, files []UploadedFile) error {
	wrote := 0
	for _, file := range files {
		relative, ok := cleanUploadPath(file.Name)
		if !ok || !isSupportedUploadName(relative) {
			continue
		}
		src, err := file.Open()
		if err != nil {
			return err
		}
		target := filepath.Join(root, relative)
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			_ = src.Close()
			return err
		}
		dst, err := os.OpenFile(target, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
		if err != nil {
			_ = src.Close()
			return err
		}
		_, copyErr := io.Copy(dst, src)
		closeErr := dst.Close()
		srcCloseErr := src.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
		if srcCloseErr != nil {
			return srcCloseErr
		}
		wrote++
	}
	if wrote == 0 {
		return ErrValidation
	}
	return nil
}

func cleanUploadPath(name string) (string, bool) {
	name = filepath.ToSlash(strings.TrimSpace(name))
	name = strings.TrimLeft(name, "/")
	if name == "" {
		return "", false
	}
	cleaned := filepath.Clean(filepath.FromSlash(name))
	if cleaned == "." || cleaned == ".." || filepath.IsAbs(cleaned) {
		return "", false
	}
	if strings.HasPrefix(cleaned, ".."+string(os.PathSeparator)) {
		return "", false
	}
	return cleaned, true
}

func isSupportedUploadName(name string) bool {
	switch strings.ToLower(filepath.Ext(name)) {
	case ".jpg", ".jpeg", ".tif", ".tiff", ".png":
		return true
	default:
		return false
	}
}
