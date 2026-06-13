package fields

import (
	"context"
	"errors"
	"testing"
)

func TestFieldLifecycle(t *testing.T) {
	ctx := context.Background()
	uc := NewUsecases(NewMemoryRepository())

	field, err := uc.CreateField(ctx, " Campo Norte ", " Maiz temprano ")
	if err != nil {
		t.Fatal(err)
	}
	if field.Name != "Campo Norte" || field.Notes != "Maiz temprano" {
		t.Fatalf("field = %+v", field)
	}

	if _, err := uc.DeleteField(ctx, field.ID); !errors.Is(err, ErrFieldNotArchived) {
		t.Fatalf("DeleteField before archive error = %v, want %v", err, ErrFieldNotArchived)
	}

	archived, err := uc.ArchiveField(ctx, field.ID)
	if err != nil {
		t.Fatal(err)
	}
	if archived.ArchivedAt == nil {
		t.Fatal("ArchivedAt = nil")
	}

	restored, err := uc.RestoreField(ctx, field.ID)
	if err != nil {
		t.Fatal(err)
	}
	if restored.ArchivedAt != nil {
		t.Fatalf("ArchivedAt = %v, want nil", restored.ArchivedAt)
	}

	archived, err = uc.ArchiveField(ctx, field.ID)
	if err != nil {
		t.Fatal(err)
	}
	deleted, err := uc.DeleteField(ctx, archived.ID)
	if err != nil {
		t.Fatal(err)
	}
	if deleted.ID != field.ID {
		t.Fatalf("deleted ID = %q, want %q", deleted.ID, field.ID)
	}
}

func TestFieldValidation(t *testing.T) {
	ctx := context.Background()
	uc := NewUsecases(NewMemoryRepository())

	if _, err := uc.CreateField(ctx, " ", ""); !errors.Is(err, ErrValidation) {
		t.Fatalf("CreateField error = %v, want %v", err, ErrValidation)
	}

	field, err := uc.CreateField(ctx, "Campo Norte", "")
	if err != nil {
		t.Fatal(err)
	}
	empty := " "
	if _, err := uc.UpdateField(ctx, field.ID, &empty, nil); !errors.Is(err, ErrValidation) {
		t.Fatalf("UpdateField error = %v, want %v", err, ErrValidation)
	}
}
