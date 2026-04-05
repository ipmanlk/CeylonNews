package store

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"ipmanlk/cnapi/internal/model"

	_ "github.com/mattn/go-sqlite3"
)

// setupTestDB creates an in-memory SQLite database with migrations applied
func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	// Create in-memory database with FTS5 support
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Enable foreign keys
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		db.Close()
		t.Fatalf("failed to enable foreign keys: %v", err)
	}

	// Run migrations
	if err := runTestMigrations(db); err != nil {
		db.Close()
		t.Fatalf("failed to run migrations: %v", err)
	}

	return db
}

// TestSQLiteDriverIsAvailable is a dummy test to ensure the sqlite driver is linked
func TestSQLiteDriverIsAvailable(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open sqlite connection: %v", err)
	}
	defer db.Close()

	// Simple test query to verify driver works
	_, err = db.Exec("SELECT 1")
	if err != nil {
		t.Fatalf("failed to execute test query: %v", err)
	}
}

// runTestMigrations applies the database schema for testing
func runTestMigrations(db *sql.DB) error {
	// Execute the migration statements directly
	statements := []string{
		`CREATE TABLE articles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			source_name TEXT NOT NULL,
			title TEXT NOT NULL,
			url TEXT NOT NULL UNIQUE,
			content_text TEXT NOT NULL,
			content_html TEXT,
			image_url TEXT,
			language TEXT NOT NULL CHECK (language IN ('en', 'si', 'ta')),
			published_at DATETIME NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX idx_articles_source_name ON articles(source_name)`,
		`CREATE INDEX idx_articles_language ON articles(language)`,
		`CREATE INDEX idx_articles_published_at ON articles(published_at)`,
		`CREATE INDEX idx_articles_url ON articles(url)`,
		`CREATE VIRTUAL TABLE articles_fts USING fts5(
			title,
			content_text,
			content='articles',
			content_rowid='id'
		)`,
		`CREATE TRIGGER articles_fts_insert AFTER INSERT ON articles BEGIN
			INSERT INTO articles_fts(rowid, title, content_text) VALUES (new.id, new.title, new.content_text);
		END`,
		`CREATE TRIGGER articles_fts_delete AFTER DELETE ON articles BEGIN
			INSERT INTO articles_fts(articles_fts, rowid, title, content_text) VALUES('delete', old.id, old.title, old.content_text);
		END`,
		`CREATE TRIGGER articles_fts_update AFTER UPDATE ON articles BEGIN
			INSERT INTO articles_fts(articles_fts, rowid, title, content_text) VALUES('delete', old.id, old.title, old.content_text);
			INSERT INTO articles_fts(rowid, title, content_text) VALUES (new.id, new.title, new.content_text);
		END`,
	}

	for _, stmt := range statements {
		if _, err := db.Exec(stmt); err != nil {
			return fmt.Errorf("failed to execute migration statement: %w", err)
		}
	}

	return nil
}

// newTestArticle creates a test ScrapedArticle with sensible defaults
func newTestArticle(overrides ...func(*model.ScrapedArticle)) model.ScrapedArticle {
	now := time.Now().Truncate(time.Second)
	imageURL := "https://example.com/image.jpg"

	article := model.ScrapedArticle{
		SourceName:  "TestSource",
		Title:       "Test Article Title",
		URL:         fmt.Sprintf("https://example.com/article-%d", now.UnixNano()),
		ContentText: "This is test article content with some text to make it realistic.",
		ContentHTML: "<p>This is test article content with some text to make it realistic.</p>",
		ImageURL:    &imageURL,
		Language:    model.LangEn,
		PublishedAt: now,
	}

	// Apply overrides
	for _, override := range overrides {
		override(&article)
	}

	return article
}

// newTestArticles creates multiple test articles with unique URLs
func newTestArticles(count int, overrides ...func(int, *model.ScrapedArticle)) []model.ScrapedArticle {
	articles := make([]model.ScrapedArticle, count)
	for i := 0; i < count; i++ {
		articles[i] = newTestArticle()
		articles[i].URL = fmt.Sprintf("https://example.com/article-%d", time.Now().UnixNano()+int64(i))
		articles[i].Title = fmt.Sprintf("Test Article %d", i+1)

		// Apply index-specific overrides
		for _, override := range overrides {
			override(i, &articles[i])
		}
	}
	return articles
}

// assertArticleEqual compares a stored article with the original scraped article
func assertArticleEqual(t *testing.T, stored *model.Article, scraped model.ScrapedArticle) {
	t.Helper()

	if stored.SourceName != scraped.SourceName {
		t.Errorf("SourceName mismatch: got %s, want %s", stored.SourceName, scraped.SourceName)
	}
	if stored.Title != scraped.Title {
		t.Errorf("Title mismatch: got %s, want %s", stored.Title, scraped.Title)
	}
	if stored.URL != scraped.URL {
		t.Errorf("URL mismatch: got %s, want %s", stored.URL, scraped.URL)
	}
	if stored.ContentText != scraped.ContentText {
		t.Errorf("ContentText mismatch: got %s, want %s", stored.ContentText, scraped.ContentText)
	}
	if stored.ContentHTML != scraped.ContentHTML {
		t.Errorf("ContentHTML mismatch: got %s, want %s", stored.ContentHTML, scraped.ContentHTML)
	}
	if stored.Language != string(scraped.Language) {
		t.Errorf("Language mismatch: got %s, want %s", stored.Language, scraped.Language)
	}

	// Check image URL (handle nil cases)
	if (stored.ImageURL == nil) != (scraped.ImageURL == nil) {
		t.Errorf("ImageURL nil mismatch: stored=%v, scraped=%v", stored.ImageURL, scraped.ImageURL)
	} else if stored.ImageURL != nil && scraped.ImageURL != nil && *stored.ImageURL != *scraped.ImageURL {
		t.Errorf("ImageURL mismatch: got %s, want %s", *stored.ImageURL, *scraped.ImageURL)
	}

	// Check published date (truncate to second for comparison since DB might not store nanoseconds)
	if !stored.PublishedAt.Truncate(time.Second).Equal(scraped.PublishedAt.Truncate(time.Second)) {
		t.Errorf("PublishedAt mismatch: got %v, want %v", stored.PublishedAt, scraped.PublishedAt)
	}

	// Verify timestamps are set
	if stored.CreatedAt.IsZero() {
		t.Error("CreatedAt is zero")
	}
	if stored.UpdatedAt.IsZero() {
		t.Error("UpdatedAt is zero")
	}
}
