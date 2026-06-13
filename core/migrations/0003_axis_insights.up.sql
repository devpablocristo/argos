ALTER TABLE analyses
    ADD COLUMN IF NOT EXISTS nexus_sync_status TEXT NOT NULL DEFAULT 'pending',
    ADD COLUMN IF NOT EXISTS nexus_sync_error TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS nexus_correlation_id TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS nexus_findings_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS companion_sync_status TEXT NOT NULL DEFAULT 'pending',
    ADD COLUMN IF NOT EXISTS companion_sync_error TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS companion_correlation_id TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS companion_output_json JSONB NOT NULL DEFAULT '{}'::jsonb;
