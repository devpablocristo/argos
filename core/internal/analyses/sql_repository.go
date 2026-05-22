package analyses

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	analysisdomain "github.com/devpablocristo/argos/core/internal/analyses/usecases/domain"
)

type SQLRepository struct {
	db *sql.DB
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) CreateAnalysis(ctx context.Context, analysis analysisdomain.Analysis) (analysisdomain.Analysis, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return analysisdomain.Analysis{}, err
	}
	defer func() { _ = tx.Rollback() }()

	metricsJSON, err := marshalJSON(analysis.Metrics)
	if err != nil {
		return analysisdomain.Analysis{}, err
	}
	warningsJSON, err := marshalJSON(analysis.Warnings)
	if err != nil {
		return analysisdomain.Analysis{}, err
	}
	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO analyses (
			id, dataset_id, capture_id, kind, status, metrics_json,
			warnings_json, created_at, completed_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		analysis.ID,
		emptyStringAsNil(analysis.DatasetID),
		emptyStringAsNil(analysis.CaptureID),
		analysis.Kind,
		analysis.Status,
		metricsJSON,
		warningsJSON,
		analysis.CreatedAt,
		analysis.CompletedAt,
	); err != nil {
		return analysisdomain.Analysis{}, err
	}
	for _, output := range analysis.Outputs {
		metadataJSON, err := marshalJSON(output.Metadata)
		if err != nil {
			return analysisdomain.Analysis{}, err
		}
		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO analysis_outputs (
				id, analysis_id, kind, content_type, uri, byte_size, metadata_json
			)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			output.ID,
			analysis.ID,
			output.Kind,
			output.ContentType,
			output.Path,
			output.ByteSize,
			metadataJSON,
		); err != nil {
			return analysisdomain.Analysis{}, err
		}
	}
	if err := tx.Commit(); err != nil {
		return analysisdomain.Analysis{}, err
	}
	return analysis, nil
}

func (r *SQLRepository) GetAnalysis(ctx context.Context, id string) (analysisdomain.Analysis, error) {
	analysis, err := r.scanAnalysis(r.db.QueryRowContext(
		ctx,
		`SELECT
			id, COALESCE(dataset_id::text, ''), COALESCE(capture_id::text, ''),
			kind, status, metrics_json, warnings_json, created_at, completed_at
		 FROM analyses
		 WHERE id = $1`,
		id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return analysisdomain.Analysis{}, ErrNotFound
	}
	if err != nil {
		return analysisdomain.Analysis{}, err
	}
	outputs, err := r.ListAnalysisOutputs(ctx, analysis.ID)
	if err != nil {
		return analysisdomain.Analysis{}, err
	}
	analysis.Outputs = outputs
	return analysis, nil
}

func (r *SQLRepository) ListAnalysisOutputs(ctx context.Context, analysisID string) ([]analysisdomain.OutputAsset, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, analysis_id, kind, uri, content_type, byte_size, metadata_json
		 FROM analysis_outputs
		 WHERE analysis_id = $1
		 ORDER BY created_at, kind`,
		analysisID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []analysisdomain.OutputAsset
	for rows.Next() {
		output, err := scanOutput(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, output)
	}
	return out, rows.Err()
}

func (r *SQLRepository) GetOutput(ctx context.Context, id string) (analysisdomain.OutputAsset, error) {
	output, err := scanOutput(r.db.QueryRowContext(
		ctx,
		`SELECT id, analysis_id, kind, uri, content_type, byte_size, metadata_json
		 FROM analysis_outputs
		 WHERE id = $1`,
		id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return analysisdomain.OutputAsset{}, ErrNotFound
	}
	return output, err
}

func (r *SQLRepository) scanAnalysis(row rowScanner) (analysisdomain.Analysis, error) {
	var analysis analysisdomain.Analysis
	var metricsBytes []byte
	var warningsBytes []byte
	var completedAt sql.NullTime
	if err := row.Scan(
		&analysis.ID,
		&analysis.DatasetID,
		&analysis.CaptureID,
		&analysis.Kind,
		&analysis.Status,
		&metricsBytes,
		&warningsBytes,
		&analysis.CreatedAt,
		&completedAt,
	); err != nil {
		return analysisdomain.Analysis{}, err
	}
	if completedAt.Valid {
		analysis.CompletedAt = &completedAt.Time
	}
	if err := unmarshalJSON(metricsBytes, &analysis.Metrics); err != nil {
		return analysisdomain.Analysis{}, err
	}
	if err := unmarshalJSON(warningsBytes, &analysis.Warnings); err != nil {
		return analysisdomain.Analysis{}, err
	}
	return analysis, nil
}

func scanOutput(row rowScanner) (analysisdomain.OutputAsset, error) {
	var output analysisdomain.OutputAsset
	var metadataBytes []byte
	if err := row.Scan(
		&output.ID,
		&output.AnalysisID,
		&output.Kind,
		&output.Path,
		&output.ContentType,
		&output.ByteSize,
		&metadataBytes,
	); err != nil {
		return analysisdomain.OutputAsset{}, err
	}
	if err := unmarshalJSON(metadataBytes, &output.Metadata); err != nil {
		return analysisdomain.OutputAsset{}, err
	}
	return output, nil
}

type rowScanner interface {
	Scan(dest ...any) error
}

func emptyStringAsNil(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func marshalJSON(value any) ([]byte, error) {
	if value == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(value)
}

func unmarshalJSON(data []byte, dst any) error {
	if len(data) == 0 {
		data = []byte("{}")
	}
	return json.Unmarshal(data, dst)
}
