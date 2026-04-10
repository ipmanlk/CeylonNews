-- +goose Up
-- +goose StatementBegin
CREATE TABLE articles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    source_id TEXT NOT NULL,
    title TEXT NOT NULL,
    url TEXT NOT NULL UNIQUE,
    content_text TEXT NOT NULL,
    content_html TEXT,
    image_url TEXT,
    language TEXT NOT NULL CHECK (language IN ('en', 'si', 'ta')),
    published_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose StatementBegin
CREATE INDEX idx_articles_source_id ON articles(source_id);
CREATE INDEX idx_articles_language ON articles(language);
CREATE INDEX idx_articles_published_at ON articles(published_at);
CREATE INDEX idx_articles_url ON articles(url);
-- +goose StatementEnd

-- +goose StatementBegin
-- Create FTS (Full Text Search) virtual table for article search
CREATE VIRTUAL TABLE articles_fts USING fts5(
    title,
    content_text,
    content='articles',
    content_rowid='id'
);
-- +goose StatementEnd

-- +goose StatementBegin
-- Create triggers to maintain FTS index
CREATE TRIGGER articles_fts_insert AFTER INSERT ON articles BEGIN
    INSERT INTO articles_fts(rowid, title, content_text) VALUES (new.id, new.title, new.content_text);
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER articles_fts_delete AFTER DELETE ON articles BEGIN
    INSERT INTO articles_fts(articles_fts, rowid, title, content_text) VALUES('delete', old.id, old.title, old.content_text);
END;
-- +goose StatementEnd

-- +goose StatementBegin
CREATE TRIGGER articles_fts_update AFTER UPDATE ON articles BEGIN
    INSERT INTO articles_fts(articles_fts, rowid, title, content_text) VALUES('delete', old.id, old.title, old.content_text);
    INSERT INTO articles_fts(rowid, title, content_text) VALUES (new.id, new.title, new.content_text);
END;
-- +goose StatementEnd

-- +goose Down

-- +goose StatementBegin
-- Drop triggers first
DROP TRIGGER IF EXISTS articles_fts_update;
DROP TRIGGER IF EXISTS articles_fts_delete;
DROP TRIGGER IF EXISTS articles_fts_insert;
-- +goose StatementEnd

-- +goose StatementBegin
-- Drop FTS virtual table
DROP TABLE IF EXISTS articles_fts;
-- +goose StatementEnd

-- +goose StatementBegin
-- Drop indexes
DROP INDEX IF EXISTS idx_articles_url;
DROP INDEX IF EXISTS idx_articles_published_at;
DROP INDEX IF EXISTS idx_articles_language;
DROP INDEX IF EXISTS idx_articles_source_id;
-- +goose StatementEnd

-- +goose StatementBegin
-- Drop tables
DROP TABLE IF EXISTS articles;
-- +goose StatementEnd
