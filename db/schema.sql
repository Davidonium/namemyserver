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
-- Dbmate schema migrations
INSERT INTO "schema_migrations" (version) VALUES
  ('20240609195352');
