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
