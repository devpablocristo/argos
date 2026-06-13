package fields

import (
	"context"
	"database/sql"
	"errors"

	fielddomain "github.com/devpablocristo/argos/core/internal/fields/usecases/domain"
	"github.com/google/uuid"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) CreateField(ctx context.Context, name, notes string) (fielddomain.Field, error) {
	field, err := scanField(r.db.QueryRowContext(
		ctx,
		`INSERT INTO fields (id, name, notes)
		 VALUES ($1, $2, $3)
		 RETURNING id, org_id, name, notes, created_at, updated_at, archived_at`,
		uuid.NewString(),
		name,
		notes,
	))
	if err != nil {
		return fielddomain.Field{}, err
	}
	return field, nil
}

func (r *SQLRepository) ListFields(ctx context.Context, includeArchived bool) ([]fielddomain.Field, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, org_id, name, notes, created_at, updated_at, archived_at
		 FROM fields
		 WHERE ($1::boolean OR archived_at IS NULL)
		 ORDER BY created_at DESC`,
		includeArchived,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []fielddomain.Field{}
	for rows.Next() {
		field, err := scanField(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, field)
	}
	return out, rows.Err()
}

func (r *SQLRepository) GetField(ctx context.Context, id string) (fielddomain.Field, error) {
	field, err := scanField(r.db.QueryRowContext(
		ctx,
		`SELECT id, org_id, name, notes, created_at, updated_at, archived_at
		 FROM fields
		 WHERE id = $1`,
		id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return fielddomain.Field{}, ErrNotFound
	}
	return field, err
}

func (r *SQLRepository) UpdateField(ctx context.Context, id, name, notes string) (fielddomain.Field, error) {
	field, err := scanField(r.db.QueryRowContext(
		ctx,
		`UPDATE fields
		 SET name = $2, notes = $3, updated_at = now()
		 WHERE id = $1
		 RETURNING id, org_id, name, notes, created_at, updated_at, archived_at`,
		id,
		name,
		notes,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return fielddomain.Field{}, ErrNotFound
	}
	return field, err
}

func (r *SQLRepository) ArchiveField(ctx context.Context, id string) (fielddomain.Field, error) {
	field, err := scanField(r.db.QueryRowContext(
		ctx,
		`UPDATE fields
		 SET archived_at = COALESCE(archived_at, now()), updated_at = now()
		 WHERE id = $1
		 RETURNING id, org_id, name, notes, created_at, updated_at, archived_at`,
		id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return fielddomain.Field{}, ErrNotFound
	}
	return field, err
}

func (r *SQLRepository) RestoreField(ctx context.Context, id string) (fielddomain.Field, error) {
	field, err := scanField(r.db.QueryRowContext(
		ctx,
		`UPDATE fields
		 SET archived_at = NULL, updated_at = now()
		 WHERE id = $1
		 RETURNING id, org_id, name, notes, created_at, updated_at, archived_at`,
		id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return fielddomain.Field{}, ErrNotFound
	}
	return field, err
}

func (r *SQLRepository) DeleteField(ctx context.Context, id string) (fielddomain.Field, error) {
	field, err := scanField(r.db.QueryRowContext(
		ctx,
		`DELETE FROM fields
		 WHERE id = $1
		 RETURNING id, org_id, name, notes, created_at, updated_at, archived_at`,
		id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return fielddomain.Field{}, ErrNotFound
	}
	return field, err
}

func (r *SQLRepository) CountDatasetsByField(ctx context.Context, id string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT count(*) FROM datasets WHERE field_id = $1`, id).Scan(&count)
	return count, err
}

func scanField(row interface{ Scan(dest ...any) error }) (fielddomain.Field, error) {
	var field fielddomain.Field
	var archivedAt sql.NullTime
	err := row.Scan(
		&field.ID,
		&field.OrgID,
		&field.Name,
		&field.Notes,
		&field.CreatedAt,
		&field.UpdatedAt,
		&archivedAt,
	)
	if archivedAt.Valid {
		field.ArchivedAt = &archivedAt.Time
	}
	return field, err
}
