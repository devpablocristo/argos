package fields

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	fielddomain "github.com/devpablocristo/argos/core/internal/fields/usecases/domain"
	"github.com/google/uuid"
)

var ErrNotFound = errors.New("not found")

type Repository interface {
	CreateField(ctx context.Context, name, notes string) (fielddomain.Field, error)
	ListFields(ctx context.Context, includeArchived bool) ([]fielddomain.Field, error)
	GetField(ctx context.Context, id string) (fielddomain.Field, error)
	UpdateField(ctx context.Context, id, name, notes string) (fielddomain.Field, error)
	ArchiveField(ctx context.Context, id string) (fielddomain.Field, error)
	RestoreField(ctx context.Context, id string) (fielddomain.Field, error)
	DeleteField(ctx context.Context, id string) (fielddomain.Field, error)
	CountDatasetsByField(ctx context.Context, id string) (int, error)
}

type MemoryRepository struct {
	mu     sync.RWMutex
	fields map[string]fielddomain.Field
}

func NewMemoryRepository() *MemoryRepository {
	return &MemoryRepository{fields: map[string]fielddomain.Field{}}
}

func (r *MemoryRepository) CreateField(_ context.Context, name, notes string) (fielddomain.Field, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now().UTC()
	field := fielddomain.Field{
		ID:        uuid.NewString(),
		Name:      name,
		Notes:     notes,
		CreatedAt: now,
		UpdatedAt: now,
	}
	r.fields[field.ID] = field
	return field, nil
}

func (r *MemoryRepository) ListFields(_ context.Context, includeArchived bool) ([]fielddomain.Field, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]fielddomain.Field, 0, len(r.fields))
	for _, field := range r.fields {
		if !includeArchived && field.ArchivedAt != nil {
			continue
		}
		out = append(out, field)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].CreatedAt.After(out[j].CreatedAt)
	})
	return out, nil
}

func (r *MemoryRepository) GetField(_ context.Context, id string) (fielddomain.Field, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	field, ok := r.fields[id]
	if !ok {
		return fielddomain.Field{}, ErrNotFound
	}
	return field, nil
}

func (r *MemoryRepository) UpdateField(_ context.Context, id, name, notes string) (fielddomain.Field, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	field, ok := r.fields[id]
	if !ok {
		return fielddomain.Field{}, ErrNotFound
	}
	field.Name = name
	field.Notes = notes
	field.UpdatedAt = time.Now().UTC()
	r.fields[id] = field
	return field, nil
}

func (r *MemoryRepository) ArchiveField(_ context.Context, id string) (fielddomain.Field, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	field, ok := r.fields[id]
	if !ok {
		return fielddomain.Field{}, ErrNotFound
	}
	now := time.Now().UTC()
	if field.ArchivedAt == nil {
		field.ArchivedAt = &now
	}
	field.UpdatedAt = now
	r.fields[id] = field
	return field, nil
}

func (r *MemoryRepository) RestoreField(_ context.Context, id string) (fielddomain.Field, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	field, ok := r.fields[id]
	if !ok {
		return fielddomain.Field{}, ErrNotFound
	}
	field.ArchivedAt = nil
	field.UpdatedAt = time.Now().UTC()
	r.fields[id] = field
	return field, nil
}

func (r *MemoryRepository) DeleteField(_ context.Context, id string) (fielddomain.Field, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	field, ok := r.fields[id]
	if !ok {
		return fielddomain.Field{}, ErrNotFound
	}
	delete(r.fields, id)
	return field, nil
}

func (r *MemoryRepository) CountDatasetsByField(context.Context, string) (int, error) {
	return 0, nil
}
