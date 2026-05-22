package catalog

import (
	"context"
	"errors"
	"sync"
	"time"

	catalogdomain "github.com/devpablocristo/argos/core/internal/catalog/usecases/domain"
	"github.com/google/uuid"
)

var ErrNotFound = errors.New("not found")

type Repository interface {
	CreateDataset(ctx context.Context, name, sourceURI string) (catalogdomain.Dataset, error)
	ListDatasets(ctx context.Context) ([]catalogdomain.Dataset, error)
	GetDataset(ctx context.Context, id string) (catalogdomain.Dataset, error)
	UpdateDatasetStatus(ctx context.Context, id, status string) error
	ReplaceCaptures(ctx context.Context, datasetID string, captures []catalogdomain.Capture) error
	ListCaptures(ctx context.Context, datasetID string) ([]catalogdomain.Capture, error)
	GetCapture(ctx context.Context, id string) (catalogdomain.Capture, error)
}

type MemoryRepository struct {
	mu            sync.RWMutex
	datasets      map[string]catalogdomain.Dataset
	captures      map[string]catalogdomain.Capture
	capturesBySet map[string][]string
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		datasets:      map[string]catalogdomain.Dataset{},
		captures:      map[string]catalogdomain.Capture{},
		capturesBySet: map[string][]string{},
	}
}

func (r *MemoryRepository) CreateDataset(_ context.Context, name, sourceURI string) (catalogdomain.Dataset, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now().UTC()
	ds := catalogdomain.Dataset{
		ID:        uuid.NewString(),
		Name:      name,
		SourceURI: sourceURI,
		Status:    "registered",
		CreatedAt: now,
		UpdatedAt: now,
	}
	r.datasets[ds.ID] = ds
	return ds, nil
}

func (r *MemoryRepository) ListDatasets(_ context.Context) ([]catalogdomain.Dataset, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]catalogdomain.Dataset, 0, len(r.datasets))
	for _, ds := range r.datasets {
		out = append(out, ds)
	}
	return out, nil
}

func (r *MemoryRepository) GetDataset(_ context.Context, id string) (catalogdomain.Dataset, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ds, ok := r.datasets[id]
	if !ok {
		return catalogdomain.Dataset{}, ErrNotFound
	}
	return ds, nil
}

func (r *MemoryRepository) UpdateDatasetStatus(_ context.Context, id, status string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	ds, ok := r.datasets[id]
	if !ok {
		return ErrNotFound
	}
	ds.Status = status
	ds.UpdatedAt = time.Now().UTC()
	r.datasets[id] = ds
	return nil
}

func (r *MemoryRepository) ReplaceCaptures(_ context.Context, datasetID string, captures []catalogdomain.Capture) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, id := range r.capturesBySet[datasetID] {
		delete(r.captures, id)
	}
	ids := make([]string, 0, len(captures))
	for _, capture := range captures {
		r.captures[capture.ID] = capture
		ids = append(ids, capture.ID)
	}
	r.capturesBySet[datasetID] = ids
	return nil
}

func (r *MemoryRepository) ListCaptures(_ context.Context, datasetID string) ([]catalogdomain.Capture, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := r.capturesBySet[datasetID]
	out := make([]catalogdomain.Capture, 0, len(ids))
	for _, id := range ids {
		out = append(out, r.captures[id])
	}
	return out, nil
}

func (r *MemoryRepository) GetCapture(_ context.Context, id string) (catalogdomain.Capture, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	capture, ok := r.captures[id]
	if !ok {
		return catalogdomain.Capture{}, ErrNotFound
	}
	return capture, nil
}
