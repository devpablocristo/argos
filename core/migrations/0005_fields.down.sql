DROP INDEX IF EXISTS idx_datasets_field_created;

ALTER TABLE datasets
    DROP COLUMN IF EXISTS field_id;

DROP INDEX IF EXISTS idx_fields_org_created;
DROP TABLE IF EXISTS fields;
