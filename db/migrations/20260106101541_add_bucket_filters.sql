-- migrate:up
ALTER TABLE buckets ADD COLUMN filter_length_enabled INTEGER NOT NULL DEFAULT 0;
ALTER TABLE buckets ADD COLUMN filter_length_mode TEXT DEFAULT 'upto';
ALTER TABLE buckets ADD COLUMN filter_length_value INTEGER DEFAULT NULL;

-- migrate:down
ALTER TABLE buckets DROP COLUMN filter_length_enabled;
ALTER TABLE buckets DROP COLUMN filter_length_mode;
ALTER TABLE buckets DROP COLUMN filter_length_value;
