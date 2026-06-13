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
	findingsJSON, err := marshalJSON(analysis.NexusFindings)
	if err != nil {
		return analysisdomain.Analysis{}, err
	}
	companionOutputJSON, err := marshalJSON(analysis.CompanionOutput)
	if err != nil {
		return analysisdomain.Analysis{}, err
	}
	if analysis.NexusSyncStatus == "" {
		analysis.NexusSyncStatus = "pending"
	}
	if analysis.CompanionSyncStatus == "" {
		analysis.CompanionSyncStatus = "pending"
	}
	if _, err := tx.ExecContext(
		ctx,
		`INSERT INTO analyses (
			id, dataset_id, capture_id, kind, status, metrics_json,
			warnings_json, nexus_sync_status, nexus_sync_error, nexus_correlation_id,
			nexus_findings_json, companion_sync_status, companion_sync_error,
			companion_correlation_id, companion_output_json, created_at, completed_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)`,
		analysis.ID,
		emptyStringAsNil(analysis.DatasetID),
		emptyStringAsNil(analysis.CaptureID),
		analysis.Kind,
		analysis.Status,
		metricsJSON,
		warningsJSON,
		analysis.NexusSyncStatus,
		analysis.NexusSyncError,
		analysis.NexusCorrelationID,
		findingsJSON,
		analysis.CompanionSyncStatus,
		analysis.CompanionSyncError,
		analysis.CompanionCorrelationID,
		companionOutputJSON,
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
			kind, status, metrics_json, warnings_json,
			nexus_sync_status, nexus_sync_error, nexus_correlation_id, nexus_findings_json,
			companion_sync_status, companion_sync_error, companion_correlation_id, companion_output_json,
			created_at, completed_at
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

func (r *SQLRepository) GetLatestByCaptureKind(ctx context.Context, captureID, kind string) (analysisdomain.Analysis, error) {
	analysis, err := r.scanAnalysis(r.db.QueryRowContext(
		ctx,
		`SELECT
			id, COALESCE(dataset_id::text, ''), COALESCE(capture_id::text, ''),
			kind, status, metrics_json, warnings_json,
			nexus_sync_status, nexus_sync_error, nexus_correlation_id, nexus_findings_json,
			companion_sync_status, companion_sync_error, companion_correlation_id, companion_output_json,
			created_at, completed_at
		 FROM analyses
		 WHERE capture_id = $1 AND kind = $2
		 ORDER BY created_at DESC
		 LIMIT 1`,
		captureID,
		kind,
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

func (r *SQLRepository) UpdateNexusSnapshot(ctx context.Context, analysisID, status, correlationID, errorMessage string, findings []analysisdomain.FindingSnapshot) error {
	findingsJSON, err := marshalJSON(findings)
	if err != nil {
		return err
	}
	result, err := r.db.ExecContext(
		ctx,
		`UPDATE analyses
		 SET nexus_sync_status = $2,
		     nexus_sync_error = $3,
		     nexus_correlation_id = $4,
		     nexus_findings_json = $5
		 WHERE id = $1`,
		analysisID,
		status,
		errorMessage,
		correlationID,
		findingsJSON,
	)
	if err != nil {
		return err
	}
	return requireRowsAffected(result)
}

func (r *SQLRepository) UpdateCompanionSnapshot(ctx context.Context, analysisID, status, correlationID, errorMessage string, output map[string]any) error {
	outputJSON, err := marshalJSON(output)
	if err != nil {
		return err
	}
	result, err := r.db.ExecContext(
		ctx,
		`UPDATE analyses
		 SET companion_sync_status = $2,
		     companion_sync_error = $3,
		     companion_correlation_id = $4,
		     companion_output_json = $5
		 WHERE id = $1`,
		analysisID,
		status,
		errorMessage,
		correlationID,
		outputJSON,
	)
	if err != nil {
		return err
	}
	return requireRowsAffected(result)
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
	var findingsBytes []byte
	var companionOutputBytes []byte
	var completedAt sql.NullTime
	if err := row.Scan(
		&analysis.ID,
		&analysis.DatasetID,
		&analysis.CaptureID,
		&analysis.Kind,
		&analysis.Status,
		&metricsBytes,
		&warningsBytes,
		&analysis.NexusSyncStatus,
		&analysis.NexusSyncError,
		&analysis.NexusCorrelationID,
		&findingsBytes,
		&analysis.CompanionSyncStatus,
		&analysis.CompanionSyncError,
		&analysis.CompanionCorrelationID,
		&companionOutputBytes,
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
	if err := unmarshalJSON(findingsBytes, &analysis.NexusFindings); err != nil {
		return analysisdomain.Analysis{}, err
	}
	if err := unmarshalJSON(companionOutputBytes, &analysis.CompanionOutput); err != nil {
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

func requireRowsAffected(result sql.Result) error {
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrNotFound
	}
	return nil
}
