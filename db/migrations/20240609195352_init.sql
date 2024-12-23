-- migrate:up
CREATE TABLE nouns (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    value TEXT NOT NULL,
    from_seed INT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT NULL,
    archived_at DATETIME DEFAULT NULL
);

CREATE TABLE adjectives (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    value TEXT NOT NULL,
    from_seed INT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT NULL,
    archived_at DATETIME DEFAULT NULL
);

-- migrate:down
DROP TABLE nouns;
DROP TABLE adjectives;
