package analyses

import (
	"context"
	"errors"
	"sync"

	analysisdomain "github.com/devpablocristo/argos/core/internal/analyses/usecases/domain"
)

var ErrNotFound = errors.New("not found")

type Repository interface {
	CreateAnalysis(ctx context.Context, analysis analysisdomain.Analysis) (analysisdomain.Analysis, error)
	GetAnalysis(ctx context.Context, id string) (analysisdomain.Analysis, error)
	GetLatestByCaptureKind(ctx context.Context, captureID, kind string) (analysisdomain.Analysis, error)
	UpdateNexusSnapshot(ctx context.Context, analysisID, status, correlationID, errorMessage string, findings []analysisdomain.FindingSnapshot) error
	UpdateCompanionSnapshot(ctx context.Context, analysisID, status, correlationID, errorMessage string, output map[string]any) error
	ListAnalysisOutputs(ctx context.Context, analysisID string) ([]analysisdomain.OutputAsset, error)
	GetOutput(ctx context.Context, id string) (analysisdomain.OutputAsset, error)
}

type MemoryRepository struct {
	mu              sync.RWMutex
	analyses        map[string]analysisdomain.Analysis
	outputs         map[string]analysisdomain.OutputAsset
	outputsByAnalys map[string][]string
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		analyses:        map[string]analysisdomain.Analysis{},
		outputs:         map[string]analysisdomain.OutputAsset{},
		outputsByAnalys: map[string][]string{},
	}
}

func (r *MemoryRepository) CreateAnalysis(_ context.Context, analysis analysisdomain.Analysis) (analysisdomain.Analysis, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, output := range analysis.Outputs {
		r.outputs[output.ID] = output
		r.outputsByAnalys[analysis.ID] = append(r.outputsByAnalys[analysis.ID], output.ID)
	}
	if analysis.NexusSyncStatus == "" {
		analysis.NexusSyncStatus = "pending"
	}
	if analysis.CompanionSyncStatus == "" {
		analysis.CompanionSyncStatus = "pending"
	}
	r.analyses[analysis.ID] = analysis
	return analysis, nil
}

func (r *MemoryRepository) GetAnalysis(_ context.Context, id string) (analysisdomain.Analysis, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	analysis, ok := r.analyses[id]
	if !ok {
		return analysisdomain.Analysis{}, ErrNotFound
	}
	return analysis, nil
}

func (r *MemoryRepository) GetLatestByCaptureKind(_ context.Context, captureID, kind string) (analysisdomain.Analysis, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var latest analysisdomain.Analysis
	for _, analysis := range r.analyses {
		if analysis.CaptureID != captureID || analysis.Kind != kind {
			continue
		}
		if latest.ID == "" || analysis.CreatedAt.After(latest.CreatedAt) {
			latest = analysis
		}
	}
	if latest.ID == "" {
		return analysisdomain.Analysis{}, ErrNotFound
	}
	return latest, nil
}

func (r *MemoryRepository) UpdateNexusSnapshot(_ context.Context, analysisID, status, correlationID, errorMessage string, findings []analysisdomain.FindingSnapshot) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	analysis, ok := r.analyses[analysisID]
	if !ok {
		return ErrNotFound
	}
	analysis.NexusSyncStatus = status
	analysis.NexusCorrelationID = correlationID
	analysis.NexusSyncError = errorMessage
	analysis.NexusFindings = append([]analysisdomain.FindingSnapshot(nil), findings...)
	r.analyses[analysisID] = analysis
	return nil
}

func (r *MemoryRepository) UpdateCompanionSnapshot(_ context.Context, analysisID, status, correlationID, errorMessage string, output map[string]any) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	analysis, ok := r.analyses[analysisID]
	if !ok {
		return ErrNotFound
	}
	analysis.CompanionSyncStatus = status
	analysis.CompanionCorrelationID = correlationID
	analysis.CompanionSyncError = errorMessage
	analysis.CompanionOutput = cloneMap(output)
	r.analyses[analysisID] = analysis
	return nil
}

func (r *MemoryRepository) ListAnalysisOutputs(_ context.Context, analysisID string) ([]analysisdomain.OutputAsset, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := r.outputsByAnalys[analysisID]
	out := make([]analysisdomain.OutputAsset, 0, len(ids))
	for _, id := range ids {
		out = append(out, r.outputs[id])
	}
	return out, nil
}

func (r *MemoryRepository) GetOutput(_ context.Context, id string) (analysisdomain.OutputAsset, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	output, ok := r.outputs[id]
	if !ok {
		return analysisdomain.OutputAsset{}, ErrNotFound
	}
	return output, nil
}

func cloneMap(in map[string]any) map[string]any {
	if in == nil {
		return nil
	}
	out := make(map[string]any, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}
