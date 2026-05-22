package catalog

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	catalogdomain "github.com/devpablocristo/argos/core/internal/catalog/usecases/domain"
	"github.com/google/uuid"
)

type SQLRepository struct {
	db *sql.DB
}

type captureMetadata struct {
	CRS      string                       `json:"crs,omitempty"`
	Warnings []string                     `json:"warnings,omitempty"`
	Errors   []string                     `json:"errors,omitempty"`
	Analysis *catalogdomain.AnalysisDraft `json:"analysis,omitempty"`
}

func NewSQLRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) CreateDataset(ctx context.Context, name, sourceURI string) (catalogdomain.Dataset, error) {
	ds := catalogdomain.Dataset{
		ID:        uuid.NewString(),
		Name:      name,
		SourceURI: sourceURI,
		Status:    "registered",
	}
	err := r.db.QueryRowContext(
		ctx,
		`INSERT INTO datasets (id, name, source_uri, status)
		 VALUES ($1, $2, $3, $4)
		 RETURNING created_at, updated_at`,
		ds.ID,
		ds.Name,
		ds.SourceURI,
		ds.Status,
	).Scan(&ds.CreatedAt, &ds.UpdatedAt)
	return ds, err
}

func (r *SQLRepository) ListDatasets(ctx context.Context) ([]catalogdomain.Dataset, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, org_id, name, source_uri, status, created_at, updated_at
		 FROM datasets
		 ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []catalogdomain.Dataset
	for rows.Next() {
		ds, err := scanDataset(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, ds)
	}
	return out, rows.Err()
}

func (r *SQLRepository) GetDataset(ctx context.Context, id string) (catalogdomain.Dataset, error) {
	ds, err := scanDataset(r.db.QueryRowContext(
		ctx,
		`SELECT id, org_id, name, source_uri, status, created_at, updated_at
		 FROM datasets
		 WHERE id = $1`,
		id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return catalogdomain.Dataset{}, ErrNotFound
	}
	return ds, err
}

func (r *SQLRepository) UpdateDatasetStatus(ctx context.Context, id, status string) error {
	result, err := r.db.ExecContext(
		ctx,
		`UPDATE datasets SET status = $2, updated_at = now() WHERE id = $1`,
		id,
		status,
	)
	if err != nil {
		return err
	}
	count, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if count == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *SQLRepository) ReplaceCaptures(ctx context.Context, datasetID string, captures []catalogdomain.Capture) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `DELETE FROM captures WHERE dataset_id = $1`, datasetID); err != nil {
		return err
	}
	for _, capture := range captures {
		metadata := captureMetadata{
			CRS:      capture.Location.CRS,
			Warnings: capture.Warnings,
			Errors:   capture.Errors,
			Analysis: capture.Analysis,
		}
		metadataJSON, err := marshalJSON(metadata)
		if err != nil {
			return err
		}
		var lat any
		var lon any
		var alt any
		if capture.Location.Lat != nil {
			lat = *capture.Location.Lat
		}
		if capture.Location.Lon != nil {
			lon = *capture.Location.Lon
		}
		if capture.Location.AltM != nil {
			alt = *capture.Location.AltM
		}
		if _, err := tx.ExecContext(
			ctx,
			`INSERT INTO captures (
				id, dataset_id, capture_key, captured_at, location, altitude_m,
				validation_status, metadata_json, created_at
			)
			VALUES (
				$1, $2, $3, $4,
				CASE WHEN $5::double precision IS NULL OR $6::double precision IS NULL
					THEN NULL
					ELSE ST_SetSRID(ST_MakePoint($6, $5), 4326)::geography
				END,
				$7, $8, $9, $10
			)`,
			capture.ID,
			capture.DatasetID,
			capture.CaptureKey,
			parseCapturedAt(capture.CapturedAt),
			lat,
			lon,
			alt,
			capture.ValidationStatus,
			metadataJSON,
			capture.CreatedAt,
		); err != nil {
			return err
		}
		for _, asset := range capture.Assets {
			metadataJSON, err := marshalJSON(asset.Metadata)
			if err != nil {
				return err
			}
			if _, err := tx.ExecContext(
				ctx,
				`INSERT INTO band_assets (
					id, capture_id, band, role, uri, checksum_sha256, mime_type,
					width, height, bit_depth, wavelength_nm, fwhm_nm, metadata_json
				)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
				asset.ID,
				asset.CaptureID,
				asset.Band,
				asset.Role,
				asset.Path,
				asset.ChecksumSHA256,
				asset.MimeType,
				asset.Width,
				asset.Height,
				asset.BitDepth,
				asset.WavelengthNM,
				asset.FWHMNM,
				metadataJSON,
			); err != nil {
				return err
			}
		}
	}
	return tx.Commit()
}

func (r *SQLRepository) ListCaptures(ctx context.Context, datasetID string) ([]catalogdomain.Capture, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT
			id, dataset_id, capture_key, captured_at,
			ST_Y(location::geometry), ST_X(location::geometry), altitude_m,
			validation_status, metadata_json, created_at
		 FROM captures
		 WHERE dataset_id = $1
		 ORDER BY capture_key`,
		datasetID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []catalogdomain.Capture
	for rows.Next() {
		capture, err := r.scanCapture(ctx, rows)
		if err != nil {
			return nil, err
		}
		out = append(out, capture)
	}
	return out, rows.Err()
}

func (r *SQLRepository) GetCapture(ctx context.Context, id string) (catalogdomain.Capture, error) {
	capture, err := r.scanCapture(ctx, r.db.QueryRowContext(
		ctx,
		`SELECT
			id, dataset_id, capture_key, captured_at,
			ST_Y(location::geometry), ST_X(location::geometry), altitude_m,
			validation_status, metadata_json, created_at
		 FROM captures
		 WHERE id = $1`,
		id,
	))
	if errors.Is(err, sql.ErrNoRows) {
		return catalogdomain.Capture{}, ErrNotFound
	}
	return capture, err
}

func (r *SQLRepository) scanCapture(ctx context.Context, row rowScanner) (catalogdomain.Capture, error) {
	var capture catalogdomain.Capture
	var capturedAt sql.NullTime
	var lat sql.NullFloat64
	var lon sql.NullFloat64
	var alt sql.NullFloat64
	var metadataBytes []byte
	if err := row.Scan(
		&capture.ID,
		&capture.DatasetID,
		&capture.CaptureKey,
		&capturedAt,
		&lat,
		&lon,
		&alt,
		&capture.ValidationStatus,
		&metadataBytes,
		&capture.CreatedAt,
	); err != nil {
		return catalogdomain.Capture{}, err
	}
	var metadata captureMetadata
	if err := unmarshalJSON(metadataBytes, &metadata); err != nil {
		return catalogdomain.Capture{}, err
	}
	if capturedAt.Valid {
		capture.CapturedAt = capturedAt.Time.UTC().Format(time.RFC3339)
	}
	capture.Location.CRS = metadata.CRS
	if capture.Location.CRS == "" {
		capture.Location.CRS = "EPSG:4326"
	}
	if lat.Valid {
		capture.Location.Lat = &lat.Float64
	}
	if lon.Valid {
		capture.Location.Lon = &lon.Float64
	}
	if alt.Valid {
		capture.Location.AltM = &alt.Float64
	}
	capture.Warnings = metadata.Warnings
	capture.Errors = metadata.Errors
	capture.Analysis = metadata.Analysis

	assets, err := r.listAssets(ctx, capture.ID)
	if err != nil {
		return catalogdomain.Capture{}, err
	}
	capture.Assets = assets
	return capture, nil
}

func (r *SQLRepository) listAssets(ctx context.Context, captureID string) ([]catalogdomain.BandAsset, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT
			id, capture_id, band, role, uri, checksum_sha256, mime_type,
			width, height, bit_depth, wavelength_nm, fwhm_nm, metadata_json
		 FROM band_assets
		 WHERE capture_id = $1
		 ORDER BY band`,
		captureID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []catalogdomain.BandAsset
	for rows.Next() {
		var asset catalogdomain.BandAsset
		var metadataBytes []byte
		var wavelength sql.NullInt64
		var fwhm sql.NullInt64
		if err := rows.Scan(
			&asset.ID,
			&asset.CaptureID,
			&asset.Band,
			&asset.Role,
			&asset.Path,
			&asset.ChecksumSHA256,
			&asset.MimeType,
			&asset.Width,
			&asset.Height,
			&asset.BitDepth,
			&wavelength,
			&fwhm,
			&metadataBytes,
		); err != nil {
			return nil, err
		}
		if wavelength.Valid {
			value := int(wavelength.Int64)
			asset.WavelengthNM = &value
		}
		if fwhm.Valid {
			value := int(fwhm.Int64)
			asset.FWHMNM = &value
		}
		if err := unmarshalJSON(metadataBytes, &asset.Metadata); err != nil {
			return nil, err
		}
		out = append(out, asset)
	}
	return out, rows.Err()
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanDataset(row rowScanner) (catalogdomain.Dataset, error) {
	var ds catalogdomain.Dataset
	err := row.Scan(&ds.ID, &ds.OrgID, &ds.Name, &ds.SourceURI, &ds.Status, &ds.CreatedAt, &ds.UpdatedAt)
	return ds, err
}

func parseCapturedAt(value string) any {
	if value == "" {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return nil
	}
	return parsed
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
