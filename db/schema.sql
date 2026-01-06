CREATE TABLE IF NOT EXISTS "schema_migrations" (version varchar(128) primary key);
CREATE TABLE nouns (
    id INTEGER PRIMARY KEY,
    value TEXT NOT NULL,
    from_seed INT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT NULL
);
CREATE TABLE adjectives (
    id INTEGER PRIMARY KEY,
    value TEXT NOT NULL,
    from_seed INT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT NULL
);
CREATE UNIQUE INDEX idx_unique_value_nouns ON nouns(value);
CREATE UNIQUE INDEX idx_unique_value_adjectives ON adjectives(value);
CREATE TABLE buckets (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    cursor INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT NULL,
    archived_at DATETIME
, filter_length_enabled INTEGER NOT NULL DEFAULT 0, filter_length_mode TEXT DEFAULT 'upto', filter_length_value INTEGER DEFAULT NULL);
CREATE UNIQUE INDEX idx_unique_name_buckets ON buckets(name);
CREATE TABLE bucket_values (
    id INTEGER PRIMARY KEY,
    bucket_id INTEGER NOT NULL,
    order_id INTEGER NOT NULL,
    value TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT NULL,
    FOREIGN KEY (bucket_id) REFERENCES buckets(id) ON DELETE CASCADE
);
CREATE INDEX idx_bucket_values_bucket_id ON bucket_values(bucket_id, order_id);
-- Dbmate schema migrations
INSERT INTO "schema_migrations" (version) VALUES
  ('20240609195352'),
  ('20260106101541');
