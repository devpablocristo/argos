package catalog

import (
	"context"
	"errors"
	"time"

	catalogdomain "github.com/devpablocristo/argos/core/internal/catalog/usecases/domain"
	"github.com/devpablocristo/argos/core/internal/processor"
	"github.com/google/uuid"
)

type Usecases struct {
	repo      Repository
	worker    processor.Processor
	outputDir string
}

func NewUsecases(repo Repository, worker processor.Processor, outputDir string) *Usecases {
	return &Usecases{repo: repo, worker: worker, outputDir: outputDir}
}

func (u *Usecases) CreateDataset(ctx context.Context, name, sourceURI string) (catalogdomain.Dataset, error) {
	return u.repo.CreateDataset(ctx, name, sourceURI)
}

func (u *Usecases) ListDatasets(ctx context.Context) ([]catalogdomain.Dataset, error) {
	return u.repo.ListDatasets(ctx)
}

func (u *Usecases) GetDataset(ctx context.Context, id string) (catalogdomain.Dataset, error) {
	return u.repo.GetDataset(ctx, id)
}

func (u *Usecases) ScanDataset(ctx context.Context, datasetID string) ([]catalogdomain.Capture, []string, string, error) {
	ds, err := u.repo.GetDataset(ctx, datasetID)
	if err != nil {
		return nil, nil, "", err
	}
	if err := u.repo.UpdateDatasetStatus(ctx, datasetID, "processing"); err != nil {
		return nil, nil, "", err
	}
	response, err := u.worker.ProcessCaptureGroups(ctx, ds.SourceURI, datasetOutputPath(u.outputDir, datasetID))
	if err != nil {
		_ = u.repo.UpdateDatasetStatus(ctx, datasetID, "failed")
		return nil, nil, "", err
	}
	captures := convertCaptures(datasetID, response.Captures)
	if err := u.repo.ReplaceCaptures(ctx, datasetID, captures); err != nil {
		_ = u.repo.UpdateDatasetStatus(ctx, datasetID, "failed")
		return nil, nil, "", err
	}
	status := "failed"
	if response.Status == "completed" {
		status = "processed"
	}
	if err := u.repo.UpdateDatasetStatus(ctx, datasetID, status); err != nil && !errors.Is(err, ErrNotFound) {
		return nil, nil, "", err
	}
	return captures, response.Warnings, response.Status, nil
}

func (u *Usecases) ListCaptures(ctx context.Context, datasetID string) ([]catalogdomain.Capture, error) {
	if _, err := u.repo.GetDataset(ctx, datasetID); err != nil {
		return nil, err
	}
	return u.repo.ListCaptures(ctx, datasetID)
}

func (u *Usecases) GetCapture(ctx context.Context, id string) (catalogdomain.Capture, error) {
	return u.repo.GetCapture(ctx, id)
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
