CREATE EXTENSION IF NOT EXISTS postgis;

CREATE TABLE IF NOT EXISTS datasets (
    id UUID PRIMARY KEY,
    org_id TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL,
    source_uri TEXT NOT NULL,
    status TEXT NOT NULL,
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS flights (
    id UUID PRIMARY KEY,
    dataset_id UUID NOT NULL REFERENCES datasets(id) ON DELETE CASCADE,
    name TEXT NOT NULL DEFAULT '',
    platform TEXT NOT NULL DEFAULT '',
    sensor TEXT NOT NULL DEFAULT '',
    started_at TIMESTAMPTZ,
    ended_at TIMESTAMPTZ,
    footprint GEOMETRY(POLYGON, 4326),
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS captures (
    id UUID PRIMARY KEY,
    dataset_id UUID NOT NULL REFERENCES datasets(id) ON DELETE CASCADE,
    flight_id UUID REFERENCES flights(id) ON DELETE SET NULL,
    capture_key TEXT NOT NULL,
    captured_at TIMESTAMPTZ,
    location GEOGRAPHY(POINT, 4326),
    altitude_m DOUBLE PRECISION,
    validation_status TEXT NOT NULL,
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(dataset_id, capture_key)
);

CREATE TABLE IF NOT EXISTS band_assets (
    id UUID PRIMARY KEY,
    capture_id UUID NOT NULL REFERENCES captures(id) ON DELETE CASCADE,
    band TEXT NOT NULL,
    role TEXT NOT NULL,
    uri TEXT NOT NULL,
    checksum_sha256 TEXT NOT NULL,
    mime_type TEXT NOT NULL,
    width INTEGER NOT NULL,
    height INTEGER NOT NULL,
    bit_depth INTEGER NOT NULL,
    wavelength_nm INTEGER,
    fwhm_nm INTEGER,
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(capture_id, band)
);

CREATE TABLE IF NOT EXISTS analyses (
    id UUID PRIMARY KEY,
    dataset_id UUID REFERENCES datasets(id) ON DELETE CASCADE,
    capture_id UUID REFERENCES captures(id) ON DELETE CASCADE,
    kind TEXT NOT NULL,
    status TEXT NOT NULL,
    params_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    metrics_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    warnings_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    completed_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS analysis_outputs (
    id UUID PRIMARY KEY,
    analysis_id UUID NOT NULL REFERENCES analyses(id) ON DELETE CASCADE,
    kind TEXT NOT NULL,
    content_type TEXT NOT NULL,
    uri TEXT NOT NULL,
    byte_size BIGINT NOT NULL DEFAULT 0,
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

