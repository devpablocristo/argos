package analyses

import (
	"context"
	"errors"
	"time"

	analysisdomain "github.com/devpablocristo/argos/core/internal/analyses/usecases/domain"
	"github.com/devpablocristo/argos/core/internal/catalog"
	catalogdomain "github.com/devpablocristo/argos/core/internal/catalog/usecases/domain"
	"github.com/google/uuid"
)

type CaptureGetter interface {
	GetCapture(ctx context.Context, id string) (catalogdomain.Capture, error)
}

type Usecases struct {
	repo    Repository
	catalog CaptureGetter
}

func NewUsecases(repo Repository, catalog CaptureGetter) *Usecases {
	return &Usecases{repo: repo, catalog: catalog}
}

func (u *Usecases) CreateNDVIForCapture(ctx context.Context, captureID string) (analysisdomain.Analysis, error) {
	capture, err := u.catalog.GetCapture(ctx, captureID)
	if err != nil {
		if errors.Is(err, catalog.ErrNotFound) {
			return analysisdomain.Analysis{}, ErrNotFound
		}
		return analysisdomain.Analysis{}, err
	}
	if capture.Analysis == nil {
		return analysisdomain.Analysis{}, ErrAnalysisNotAvailable
	}
	now := time.Now().UTC()
	analysis := analysisdomain.Analysis{
		ID:          uuid.NewString(),
		DatasetID:   capture.DatasetID,
		CaptureID:   capture.ID,
		Kind:        capture.Analysis.Kind,
		Status:      capture.Analysis.Status,
		Metrics:     capture.Analysis.Metrics,
		Warnings:    append([]string(nil), capture.Analysis.Warnings...),
		CreatedAt:   now,
		CompletedAt: &now,
	}
	for _, output := range capture.Analysis.Outputs {
		analysis.Outputs = append(analysis.Outputs, analysisdomain.OutputAsset{
			ID:          uuid.NewString(),
			AnalysisID:  analysis.ID,
			Kind:        output.Kind,
			Path:        output.Path,
			ContentType: output.ContentType,
			ByteSize:    output.ByteSize,
			Metadata:    output.Metadata,
		})
	}
	return u.repo.CreateAnalysis(ctx, analysis)
}

func (u *Usecases) GetAnalysis(ctx context.Context, id string) (analysisdomain.Analysis, error) {
	return u.repo.GetAnalysis(ctx, id)
}

func (u *Usecases) ListAnalysisOutputs(ctx context.Context, analysisID string) ([]analysisdomain.OutputAsset, error) {
	if _, err := u.repo.GetAnalysis(ctx, analysisID); err != nil {
		return nil, err
	}
	return u.repo.ListAnalysisOutputs(ctx, analysisID)
}

func (u *Usecases) GetOutput(ctx context.Context, id string) (analysisdomain.OutputAsset, error) {
	return u.repo.GetOutput(ctx, id)
}

var ErrAnalysisNotAvailable = errors.New("analysis not available")
