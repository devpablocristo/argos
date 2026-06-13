CREATE TABLE IF NOT EXISTS fields (
    id UUID PRIMARY KEY,
    org_id TEXT NOT NULL DEFAULT '',
    name TEXT NOT NULL,
    notes TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    archived_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_fields_org_created
    ON fields(org_id, created_at DESC);

ALTER TABLE datasets
    ADD COLUMN IF NOT EXISTS field_id UUID REFERENCES fields(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_datasets_field_created
    ON datasets(field_id, created_at DESC);
