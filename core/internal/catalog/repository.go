package catalog

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	catalogdomain "github.com/devpablocristo/argos/core/internal/catalog/usecases/domain"
	"github.com/google/uuid"
)

var ErrNotFound = errors.New("not found")

type Repository interface {
	CreateDataset(ctx context.Context, name, sourceURI string, fieldID *string) (catalogdomain.Dataset, error)
	ListDatasets(ctx context.Context, includeArchived bool) ([]catalogdomain.Dataset, error)
	ListDatasetsByField(ctx context.Context, fieldID string, includeArchived bool) ([]catalogdomain.Dataset, error)
	GetDataset(ctx context.Context, id string) (catalogdomain.Dataset, error)
	UpdateDataset(ctx context.Context, id, name, sourceURI string) (catalogdomain.Dataset, error)
	UpdateDatasetField(ctx context.Context, id string, fieldID *string) (catalogdomain.Dataset, error)
	UpdateDatasetStatus(ctx context.Context, id, status string) error
	ArchiveDataset(ctx context.Context, id string) (catalogdomain.Dataset, error)
	RestoreDataset(ctx context.Context, id string) (catalogdomain.Dataset, error)
	DeleteDataset(ctx context.Context, id string) (catalogdomain.Dataset, error)
	UpsertDatasetClassification(ctx context.Context, classification catalogdomain.DatasetClassification) error
	GetDatasetClassification(ctx context.Context, datasetID string) (catalogdomain.DatasetClassification, error)
	AppendDatasetEvent(ctx context.Context, event catalogdomain.DatasetEvent) error
	ListDatasetEvents(ctx context.Context, datasetID string) ([]catalogdomain.DatasetEvent, error)
	ReplaceCaptures(ctx context.Context, datasetID string, captures []catalogdomain.Capture) error
	ListCaptures(ctx context.Context, datasetID string) ([]catalogdomain.Capture, error)
	GetCapture(ctx context.Context, id string) (catalogdomain.Capture, error)
}

type MemoryRepository struct {
	mu              sync.RWMutex
	datasets        map[string]catalogdomain.Dataset
	classifications map[string]catalogdomain.DatasetClassification
	events          map[string][]catalogdomain.DatasetEvent
	captures        map[string]catalogdomain.Capture
	capturesBySet   map[string][]string
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{
		datasets:        map[string]catalogdomain.Dataset{},
		classifications: map[string]catalogdomain.DatasetClassification{},
		events:          map[string][]catalogdomain.DatasetEvent{},
		captures:        map[string]catalogdomain.Capture{},
		capturesBySet:   map[string][]string{},
	}
}

func (r *MemoryRepository) CreateDataset(_ context.Context, name, sourceURI string, fieldID *string) (catalogdomain.Dataset, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now().UTC()
	ds := catalogdomain.Dataset{
		ID:        uuid.NewString(),
		Name:      name,
		SourceURI: sourceURI,
		Status:    "registered",
		FieldID:   cloneStringPtr(fieldID),
		CreatedAt: now,
		UpdatedAt: now,
	}
	r.datasets[ds.ID] = ds
	return ds, nil
}

func (r *MemoryRepository) ListDatasetsByField(_ context.Context, fieldID string, includeArchived bool) ([]catalogdomain.Dataset, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]catalogdomain.Dataset, 0, len(r.datasets))
	for _, ds := range r.datasets {
		if !includeArchived && ds.ArchivedAt != nil {
			continue
		}
		if ds.FieldID == nil || *ds.FieldID != fieldID {
			continue
		}
		ds = r.withClassificationLocked(ds)
		out = append(out, ds)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out, nil
}

func (r *MemoryRepository) ListDatasets(_ context.Context, includeArchived bool) ([]catalogdomain.Dataset, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]catalogdomain.Dataset, 0, len(r.datasets))
	for _, ds := range r.datasets {
		if !includeArchived && ds.ArchivedAt != nil {
			continue
		}
		ds = r.withClassificationLocked(ds)
		out = append(out, ds)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out, nil
}

func (r *MemoryRepository) GetDataset(_ context.Context, id string) (catalogdomain.Dataset, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ds, ok := r.datasets[id]
	if !ok {
		return catalogdomain.Dataset{}, ErrNotFound
	}
	return r.withClassificationLocked(ds), nil
}

func (r *MemoryRepository) UpdateDataset(_ context.Context, id, name, sourceURI string) (catalogdomain.Dataset, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	ds, ok := r.datasets[id]
	if !ok {
		return catalogdomain.Dataset{}, ErrNotFound
	}
	ds.Name = name
	ds.SourceURI = sourceURI
	ds.UpdatedAt = time.Now().UTC()
	r.datasets[id] = ds
	return r.withClassificationLocked(ds), nil
}

func (r *MemoryRepository) UpdateDatasetField(_ context.Context, id string, fieldID *string) (catalogdomain.Dataset, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	ds, ok := r.datasets[id]
	if !ok {
		return catalogdomain.Dataset{}, ErrNotFound
	}
	ds.FieldID = cloneStringPtr(fieldID)
	ds.UpdatedAt = time.Now().UTC()
	r.datasets[id] = ds
	return r.withClassificationLocked(ds), nil
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

func (r *MemoryRepository) ArchiveDataset(_ context.Context, id string) (catalogdomain.Dataset, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	ds, ok := r.datasets[id]
	if !ok {
		return catalogdomain.Dataset{}, ErrNotFound
	}
	now := time.Now().UTC()
	if ds.ArchivedAt == nil {
		ds.ArchivedAt = &now
	}
	ds.UpdatedAt = now
	r.datasets[id] = ds
	return r.withClassificationLocked(ds), nil
}

func (r *MemoryRepository) RestoreDataset(_ context.Context, id string) (catalogdomain.Dataset, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	ds, ok := r.datasets[id]
	if !ok {
		return catalogdomain.Dataset{}, ErrNotFound
	}
	ds.ArchivedAt = nil
	ds.UpdatedAt = time.Now().UTC()
	r.datasets[id] = ds
	return r.withClassificationLocked(ds), nil
}

func (r *MemoryRepository) DeleteDataset(_ context.Context, id string) (catalogdomain.Dataset, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	ds, ok := r.datasets[id]
	if !ok {
		return catalogdomain.Dataset{}, ErrNotFound
	}
	for _, captureID := range r.capturesBySet[id] {
		delete(r.captures, captureID)
	}
	delete(r.classifications, id)
	delete(r.events, id)
	delete(r.capturesBySet, id)
	delete(r.datasets, id)
	return ds, nil
}

func (r *MemoryRepository) UpsertDatasetClassification(_ context.Context, classification catalogdomain.DatasetClassification) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.datasets[classification.DatasetID]; !ok {
		return ErrNotFound
	}
	classification.CaptureIDs = r.captureIDsLocked(classification.DatasetID)
	classification.AnalysisIDs = append([]string(nil), classification.AnalysisIDs...)
	classification.MissingMetadata = append([]string(nil), classification.MissingMetadata...)
	r.classifications[classification.DatasetID] = classification
	return nil
}

func (r *MemoryRepository) GetDatasetClassification(_ context.Context, datasetID string) (catalogdomain.DatasetClassification, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	classification, ok := r.classifications[datasetID]
	if !ok {
		return catalogdomain.DatasetClassification{}, ErrNotFound
	}
	classification.CaptureIDs = r.captureIDsLocked(datasetID)
	classification.AnalysisIDs = append([]string(nil), classification.AnalysisIDs...)
	classification.MissingMetadata = append([]string(nil), classification.MissingMetadata...)
	return classification, nil
}

func (r *MemoryRepository) AppendDatasetEvent(_ context.Context, event catalogdomain.DatasetEvent) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.datasets[event.DatasetID]; !ok {
		return ErrNotFound
	}
	if event.Details != nil {
		event.Details = cloneAnyMap(event.Details)
	}
	r.events[event.DatasetID] = append(r.events[event.DatasetID], event)
	return nil
}

func (r *MemoryRepository) ListDatasetEvents(_ context.Context, datasetID string) ([]catalogdomain.DatasetEvent, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if _, ok := r.datasets[datasetID]; !ok {
		return nil, ErrNotFound
	}
	events := append([]catalogdomain.DatasetEvent(nil), r.events[datasetID]...)
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.Before(events[j].Timestamp)
	})
	for i := range events {
		events[i].Details = cloneAnyMap(events[i].Details)
	}
	return events, nil
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

func (r *MemoryRepository) withClassificationLocked(ds catalogdomain.Dataset) catalogdomain.Dataset {
	if classification, ok := r.classifications[ds.ID]; ok {
		classification.CaptureIDs = r.captureIDsLocked(ds.ID)
		classification.MissingMetadata = append([]string(nil), classification.MissingMetadata...)
		classification.AnalysisIDs = append([]string(nil), classification.AnalysisIDs...)
		ds.Classification = &classification
	}
	return ds
}

func (r *MemoryRepository) captureIDsLocked(datasetID string) []string {
	ids := append([]string(nil), r.capturesBySet[datasetID]...)
	sort.Strings(ids)
	return ids
}

func cloneAnyMap(in map[string]any) map[string]any {
	if in == nil {
		return nil
	}
	out := make(map[string]any, len(in))
	for key, value := range in {
		out[key] = value
	}
	return out
}

func cloneStringPtr(value *string) *string {
	if value == nil {
		return nil
	}
	clone := *value
	return &clone
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
