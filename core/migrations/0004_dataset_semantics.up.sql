CREATE TABLE IF NOT EXISTS dataset_classifications (
    dataset_id UUID PRIMARY KEY REFERENCES datasets(id) ON DELETE CASCADE,
    dataset_name TEXT NOT NULL,
    dataset_type TEXT NOT NULL CHECK (dataset_type IN (
        'sample',
        'uploaded_folder',
        'flight_dataset',
        'single_capture',
        'multi_capture_dataset',
        'sector_capture',
        'unknown'
    )),
    scope TEXT NOT NULL CHECK (scope IN (
        'global',
        'field',
        'lot',
        'campaign',
        'flight',
        'dataset'
    )),
    field_id UUID,
    lot_id UUID,
    campaign_id UUID,
    flight_id UUID REFERENCES flights(id) ON DELETE SET NULL,
    confidence DOUBLE PRECISION NOT NULL DEFAULT 0 CHECK (confidence >= 0 AND confidence <= 1),
    missing_metadata_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    reason TEXT NOT NULL DEFAULT '',
    classified_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_dataset_classifications_type
    ON dataset_classifications(dataset_type);

CREATE INDEX IF NOT EXISTS idx_dataset_classifications_scope
    ON dataset_classifications(scope);

CREATE TABLE IF NOT EXISTS dataset_events (
    event_id UUID PRIMARY KEY,
    dataset_id UUID NOT NULL REFERENCES datasets(id) ON DELETE CASCADE,
    event_type TEXT NOT NULL CHECK (event_type IN (
        'DATASET_CREATED',
        'DATASET_UPLOADED',
        'DATASET_CLASSIFIED',
        'METADATA_EXTRACTED',
        'CAPTURES_DETECTED',
        'ANALYSIS_STARTED',
        'ANALYSIS_COMPLETED',
        'INDEX_GENERATED',
        'REPORT_GENERATED',
        'ERROR'
    )),
    status TEXT NOT NULL CHECK (status IN ('pending', 'running', 'completed', 'failed')),
    message TEXT NOT NULL DEFAULT '',
    details_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_dataset_events_dataset_created
    ON dataset_events(dataset_id, created_at);

CREATE INDEX IF NOT EXISTS idx_dataset_events_type_created
    ON dataset_events(event_type, created_at);
