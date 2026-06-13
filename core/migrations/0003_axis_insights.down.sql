ALTER TABLE analyses
    DROP COLUMN IF EXISTS companion_output_json,
    DROP COLUMN IF EXISTS companion_correlation_id,
    DROP COLUMN IF EXISTS companion_sync_error,
    DROP COLUMN IF EXISTS companion_sync_status,
    DROP COLUMN IF EXISTS nexus_findings_json,
    DROP COLUMN IF EXISTS nexus_correlation_id,
    DROP COLUMN IF EXISTS nexus_sync_error,
    DROP COLUMN IF EXISTS nexus_sync_status;
