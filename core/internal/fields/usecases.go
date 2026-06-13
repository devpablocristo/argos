package fields

import (
	"context"
	"errors"
	"strings"

	fielddomain "github.com/devpablocristo/argos/core/internal/fields/usecases/domain"
)

type Usecases struct {
	repo Repository
}

func NewUsecases(repo Repository) *Usecases {
	return &Usecases{repo: repo}
}

func (u *Usecases) CreateField(ctx context.Context, name, notes string) (fielddomain.Field, error) {
	name = strings.TrimSpace(name)
	notes = strings.TrimSpace(notes)
	if name == "" {
		return fielddomain.Field{}, ErrValidation
	}
	return u.repo.CreateField(ctx, name, notes)
}

func (u *Usecases) ListFields(ctx context.Context, includeArchived bool) ([]fielddomain.Field, error) {
	return u.repo.ListFields(ctx, includeArchived)
}

func (u *Usecases) GetField(ctx context.Context, id string) (fielddomain.Field, error) {
	return u.repo.GetField(ctx, id)
}

func (u *Usecases) UpdateField(ctx context.Context, id string, name, notes *string) (fielddomain.Field, error) {
	field, err := u.repo.GetField(ctx, id)
	if err != nil {
		return fielddomain.Field{}, err
	}
	nextName := field.Name
	if name != nil {
		nextName = strings.TrimSpace(*name)
	}
	nextNotes := field.Notes
	if notes != nil {
		nextNotes = strings.TrimSpace(*notes)
	}
	if nextName == "" {
		return fielddomain.Field{}, ErrValidation
	}
	return u.repo.UpdateField(ctx, id, nextName, nextNotes)
}

func (u *Usecases) ArchiveField(ctx context.Context, id string) (fielddomain.Field, error) {
	return u.repo.ArchiveField(ctx, id)
}

func (u *Usecases) RestoreField(ctx context.Context, id string) (fielddomain.Field, error) {
	return u.repo.RestoreField(ctx, id)
}

func (u *Usecases) DeleteField(ctx context.Context, id string) (fielddomain.Field, error) {
	field, err := u.repo.GetField(ctx, id)
	if err != nil {
		return fielddomain.Field{}, err
	}
	if field.ArchivedAt == nil {
		return fielddomain.Field{}, ErrFieldNotArchived
	}
	count, err := u.repo.CountDatasetsByField(ctx, id)
	if err != nil {
		return fielddomain.Field{}, err
	}
	if count > 0 {
		return fielddomain.Field{}, ErrFieldHasDatasets
	}
	return u.repo.DeleteField(ctx, id)
}

func (u *Usecases) EnsureField(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return ErrValidation
	}
	_, err := u.repo.GetField(ctx, id)
	return err
}

var (
	ErrValidation       = errors.New("validation error")
	ErrFieldNotArchived = errors.New("field must be archived before delete")
	ErrFieldHasDatasets = errors.New("field has datasets")
)
