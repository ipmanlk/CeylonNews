package store

import (
	"strings"
	"testing"
	"time"

	"ipmanlk/cnapi/internal/model"
)

func TestArticlesStore_Create(t *testing.T) {
	tests := []struct {
		name        string
		article     model.ScrapedArticle
		wantErr     bool
		errContains string
	}{
		{
			name:    "valid article with all fields",
			article: newTestArticle(),
			wantErr: false,
		},
		{
			name: "valid article with nil image URL",
			article: newTestArticle(func(a *model.ScrapedArticle) {
				a.ImageURL = nil
			}),
			wantErr: false,
		},
		{
			name: "valid article in Sinhala",
			article: newTestArticle(func(a *model.ScrapedArticle) {
				a.Language = model.LangSi
				a.SourceID = "Ada Derana"
			}),
			wantErr: false,
		},
		{
			name: "valid article in Tamil",
			article: newTestArticle(func(a *model.ScrapedArticle) {
				a.Language = model.LangTa
				a.SourceID = "BBC Tamil"
			}),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := setupTestDB(t)
			defer db.Close()

			store := NewArticlesStore(db)

			// Create the article
			id, err := store.Create(tt.article)

			// Check error expectations
			if tt.wantErr {
				if err == nil {
					t.Errorf("Create() expected error, got nil")
					return
				}
				if tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("Create() error = %v, want error containing %v", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Fatalf("Create() unexpected error: %v", err)
			}

			// Verify ID is valid
			if id <= 0 {
				t.Errorf("Create() returned invalid id: %d", id)
			}

			// Retrieve the article and verify it matches
			stored, err := store.GetByID(id)
			if err != nil {
				t.Fatalf("GetByID() failed: %v", err)
			}

			if stored == nil {
				t.Fatal("GetByID() returned nil article")
			}

			// Verify stored ID matches returned ID
			if stored.ID != id {
				t.Errorf("Stored ID mismatch: got %d, want %d", stored.ID, id)
			}

			// Verify all fields match
			assertArticleEqual(t, stored, tt.article)
		})
	}
}

func TestArticlesStore_Create_DuplicateURL(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store := NewArticlesStore(db)

	// Create first article
	article := newTestArticle()
	id1, err := store.Create(article)
	if err != nil {
		t.Fatalf("First Create() failed: %v", err)
	}

	// Try to create article with same URL
	article2 := newTestArticle()
	article2.URL = article.URL // Same URL
	article2.Title = "Different Title"

	id2, err := store.Create(article2)
	if err == nil {
		t.Errorf("Create() with duplicate URL should fail, but got id: %d", id2)
	}

	if !strings.Contains(err.Error(), "UNIQUE") && !strings.Contains(err.Error(), "constraint") {
		t.Errorf("Create() error should mention UNIQUE constraint, got: %v", err)
	}

	// Verify first article still exists and unchanged
	stored, err := store.GetByID(id1)
	if err != nil {
		t.Fatalf("GetByID() failed: %v", err)
	}

	if stored.Title != article.Title {
		t.Errorf("Original article was modified: got title %s, want %s", stored.Title, article.Title)
	}
}

func TestArticlesStore_Create_Timestamps(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store := NewArticlesStore(db)

	before := time.Now()
	article := newTestArticle()

	id, err := store.Create(article)
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	after := time.Now()

	stored, err := store.GetByID(id)
	if err != nil {
		t.Fatalf("GetByID() failed: %v", err)
	}

	// Verify CreatedAt and UpdatedAt are set and within reasonable time range
	if stored.CreatedAt.Before(before) || stored.CreatedAt.After(after) {
		t.Errorf("CreatedAt %v is not within expected range [%v, %v]", stored.CreatedAt, before, after)
	}

	if stored.UpdatedAt.Before(before) || stored.UpdatedAt.After(after) {
		t.Errorf("UpdatedAt %v is not within expected range [%v, %v]", stored.UpdatedAt, before, after)
	}

	// For new articles, CreatedAt and UpdatedAt should be very close (within 1 second)
	timeDiff := stored.UpdatedAt.Sub(stored.CreatedAt)
	if timeDiff > time.Second {
		t.Errorf("CreatedAt and UpdatedAt differ by %v, expected them to be nearly equal for new article", timeDiff)
	}
}

func TestArticlesStore_GetByID(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store := NewArticlesStore(db)

	t.Run("existing article", func(t *testing.T) {
		article := newTestArticle()
		id, err := store.Create(article)
		if err != nil {
			t.Fatalf("Create() failed: %v", err)
		}

		stored, err := store.GetByID(id)
		if err != nil {
			t.Fatalf("GetByID() failed: %v", err)
		}

		if stored == nil {
			t.Fatal("GetByID() returned nil")
		}

		assertArticleEqual(t, stored, article)
	})

	t.Run("non-existing article", func(t *testing.T) {
		stored, err := store.GetByID(99999)
		if err != nil {
			t.Fatalf("GetByID() should not error for non-existing ID: %v", err)
		}

		if stored != nil {
			t.Errorf("GetByID() should return nil for non-existing ID, got: %+v", stored)
		}
	})
}

func TestArticlesStore_GetByURL(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store := NewArticlesStore(db)

	t.Run("existing article", func(t *testing.T) {
		article := newTestArticle()
		_, err := store.Create(article)
		if err != nil {
			t.Fatalf("Create() failed: %v", err)
		}

		stored, err := store.GetByURL(article.URL)
		if err != nil {
			t.Fatalf("GetByURL() failed: %v", err)
		}

		if stored == nil {
			t.Fatal("GetByURL() returned nil")
		}

		assertArticleEqual(t, stored, article)
	})

	t.Run("non-existing article", func(t *testing.T) {
		stored, err := store.GetByURL("https://nonexistent.com/article")
		if err != nil {
			t.Fatalf("GetByURL() should not error for non-existing URL: %v", err)
		}

		if stored != nil {
			t.Errorf("GetByURL() should return nil for non-existing URL, got: %+v", stored)
		}
	})
}

func TestArticlesStore_ExistsByURL(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	store := NewArticlesStore(db)

	article := newTestArticle()
	_, err := store.Create(article)
	if err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	t.Run("existing article", func(t *testing.T) {
		exists, err := store.ExistsByURL(article.URL)
		if err != nil {
			t.Fatalf("ExistsByURL() failed: %v", err)
		}

		if !exists {
			t.Error("ExistsByURL() should return true for existing URL")
		}
	})

	t.Run("non-existing article", func(t *testing.T) {
		exists, err := store.ExistsByURL("https://nonexistent.com/article")
		if err != nil {
			t.Fatalf("ExistsByURL() failed: %v", err)
		}

		if exists {
			t.Error("ExistsByURL() should return false for non-existing URL")
		}
	})
}

func TestArticlesStore_Upsert(t *testing.T) {
	t.Run("insert new article", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()

		store := NewArticlesStore(db)

		article := newTestArticle()
		id, err := store.Upsert(article)
		if err != nil {
			t.Fatalf("Upsert() failed: %v", err)
		}

		if id <= 0 {
			t.Errorf("Upsert() returned invalid id: %d", id)
		}

		// Verify article was inserted
		stored, err := store.GetByID(id)
		if err != nil {
			t.Fatalf("GetByID() failed: %v", err)
		}

		if stored == nil {
			t.Fatal("GetByID() returned nil")
		}

		assertArticleEqual(t, stored, article)
	})

	t.Run("update existing article", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()

		store := NewArticlesStore(db)

		// Create initial article
		original := newTestArticle()
		id1, err := store.Create(original)
		if err != nil {
			t.Fatalf("Create() failed: %v", err)
		}

		// Small delay to ensure different timestamps
		time.Sleep(10 * time.Millisecond)

		// Upsert with same URL but different content
		updated := newTestArticle()
		updated.URL = original.URL // Same URL
		updated.Title = "Updated Title"
		updated.ContentText = "Updated content with new information"
		updated.ContentHTML = "<p>Updated content with new information</p>"
		newImageURL := "https://example.com/new-image.jpg"
		updated.ImageURL = &newImageURL

		id2, err := store.Upsert(updated)
		if err != nil {
			t.Fatalf("Upsert() failed: %v", err)
		}

		// ID should be the same (it's an update, not a new insert)
		if id2 != id1 {
			t.Errorf("Upsert() returned different ID: got %d, want %d", id2, id1)
		}

		// Verify article was updated
		stored, err := store.GetByID(id1)
		if err != nil {
			t.Fatalf("GetByID() failed: %v", err)
		}

		if stored.Title != updated.Title {
			t.Errorf("Title not updated: got %s, want %s", stored.Title, updated.Title)
		}
		if stored.ContentText != updated.ContentText {
			t.Errorf("ContentText not updated: got %s, want %s", stored.ContentText, updated.ContentText)
		}
		if stored.ContentHTML != updated.ContentHTML {
			t.Errorf("ContentHTML not updated: got %s, want %s", stored.ContentHTML, updated.ContentHTML)
		}
		if stored.ImageURL == nil || *stored.ImageURL != *updated.ImageURL {
			t.Errorf("ImageURL not updated")
		}

		// Verify only one article exists (no duplicate was created)
		count, err := store.Count(model.ArticleFilter{})
		if err != nil {
			t.Fatalf("Count() failed: %v", err)
		}
		if count != 1 {
			t.Errorf("Expected 1 article, found %d", count)
		}
	})

	t.Run("update preserves CreatedAt but updates UpdatedAt", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()

		store := NewArticlesStore(db)

		// Create initial article
		original := newTestArticle()
		id, err := store.Create(original)
		if err != nil {
			t.Fatalf("Create() failed: %v", err)
		}

		firstStored, err := store.GetByID(id)
		if err != nil {
			t.Fatalf("GetByID() failed: %v", err)
		}

		// Wait a bit to ensure timestamp difference
		time.Sleep(10 * time.Millisecond)

		// Upsert the same URL with updated content
		updated := newTestArticle()
		updated.URL = original.URL
		updated.Title = "Updated Title"

		_, err = store.Upsert(updated)
		if err != nil {
			t.Fatalf("Upsert() failed: %v", err)
		}

		secondStored, err := store.GetByID(id)
		if err != nil {
			t.Fatalf("GetByID() failed: %v", err)
		}

		// CreatedAt should remain the same
		if !secondStored.CreatedAt.Equal(firstStored.CreatedAt) {
			t.Errorf("CreatedAt changed: before=%v, after=%v", firstStored.CreatedAt, secondStored.CreatedAt)
		}

		// UpdatedAt should be later
		if !secondStored.UpdatedAt.After(firstStored.UpdatedAt) {
			t.Errorf("UpdatedAt not updated: before=%v, after=%v", firstStored.UpdatedAt, secondStored.UpdatedAt)
		}
	})

	t.Run("upsert multiple times with same URL", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()

		store := NewArticlesStore(db)

		article := newTestArticle()

		// First upsert (insert)
		id1, err := store.Upsert(article)
		if err != nil {
			t.Fatalf("First Upsert() failed: %v", err)
		}

		// Second upsert (update)
		article.Title = "Updated Once"
		id2, err := store.Upsert(article)
		if err != nil {
			t.Fatalf("Second Upsert() failed: %v", err)
		}

		// Third upsert (update again)
		article.Title = "Updated Twice"
		id3, err := store.Upsert(article)
		if err != nil {
			t.Fatalf("Third Upsert() failed: %v", err)
		}

		// All IDs should be the same
		if id1 != id2 || id2 != id3 {
			t.Errorf("IDs should be same: id1=%d, id2=%d, id3=%d", id1, id2, id3)
		}

		// Verify only one article exists
		count, err := store.Count(model.ArticleFilter{})
		if err != nil {
			t.Fatalf("Count() failed: %v", err)
		}
		if count != 1 {
			t.Errorf("Expected 1 article after multiple upserts, found %d", count)
		}

		// Verify final title
		stored, err := store.GetByID(id1)
		if err != nil {
			t.Fatalf("GetByID() failed: %v", err)
		}
		if stored.Title != "Updated Twice" {
			t.Errorf("Title not updated to final value: got %s", stored.Title)
		}
	})
}

func TestArticlesStore_BulkCreate(t *testing.T) {
	t.Run("create multiple valid articles", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()

		store := NewArticlesStore(db)

		articles := newTestArticles(5)
		ids, err := store.BulkCreate(articles)
		if err != nil {
			t.Fatalf("BulkCreate() failed: %v", err)
		}

		if len(ids) != 5 {
			t.Errorf("Expected 5 IDs, got %d", len(ids))
		}

		// Verify all articles were created
		for i, id := range ids {
			if id <= 0 {
				t.Errorf("Invalid ID at index %d: %d", i, id)
			}

			stored, err := store.GetByID(id)
			if err != nil {
				t.Fatalf("GetByID(%d) failed: %v", id, err)
			}

			if stored == nil {
				t.Fatalf("Article %d not found", id)
			}

			assertArticleEqual(t, stored, articles[i])
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()

		store := NewArticlesStore(db)

		ids, err := store.BulkCreate([]model.ScrapedArticle{})
		if err != nil {
			t.Fatalf("BulkCreate() with empty slice failed: %v", err)
		}

		if len(ids) != 0 {
			t.Errorf("Expected 0 IDs for empty slice, got %d", len(ids))
		}
	})

	t.Run("transaction rollback on duplicate URL", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()

		store := NewArticlesStore(db)

		// Create first article
		firstArticle := newTestArticle()
		_, err := store.Create(firstArticle)
		if err != nil {
			t.Fatalf("Create() failed: %v", err)
		}

		// Try to bulk create with one duplicate URL
		articles := newTestArticles(3)
		articles[1].URL = firstArticle.URL // Duplicate URL in the middle

		ids, err := store.BulkCreate(articles)
		if err == nil {
			t.Errorf("BulkCreate() should fail with duplicate URL, got ids: %v", ids)
		}

		// Verify none of the bulk articles were created (transaction rollback)
		count, err := store.Count(model.ArticleFilter{})
		if err != nil {
			t.Fatalf("Count() failed: %v", err)
		}

		// Only the first article should exist
		if count != 1 {
			t.Errorf("Expected 1 article (transaction should rollback), found %d", count)
		}
	})

	t.Run("large batch", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()

		store := NewArticlesStore(db)

		articles := newTestArticles(100)
		ids, err := store.BulkCreate(articles)
		if err != nil {
			t.Fatalf("BulkCreate() with 100 articles failed: %v", err)
		}

		if len(ids) != 100 {
			t.Errorf("Expected 100 IDs, got %d", len(ids))
		}

		// Verify count
		count, err := store.Count(model.ArticleFilter{})
		if err != nil {
			t.Fatalf("Count() failed: %v", err)
		}

		if count != 100 {
			t.Errorf("Expected 100 articles in DB, found %d", count)
		}
	})

	t.Run("mixed languages", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()

		store := NewArticlesStore(db)

		articles := newTestArticles(3, func(i int, a *model.ScrapedArticle) {
			switch i {
			case 0:
				a.Language = model.LangEn
			case 1:
				a.Language = model.LangSi
			case 2:
				a.Language = model.LangTa
			}
		})

		ids, err := store.BulkCreate(articles)
		if err != nil {
			t.Fatalf("BulkCreate() failed: %v", err)
		}

		if len(ids) != 3 {
			t.Errorf("Expected 3 IDs, got %d", len(ids))
		}

		// Verify languages
		for i, id := range ids {
			stored, err := store.GetByID(id)
			if err != nil {
				t.Fatalf("GetByID(%d) failed: %v", id, err)
			}

			expectedLang := string(articles[i].Language)
			if stored.Language != expectedLang {
				t.Errorf("Article %d: expected language %s, got %s", i, expectedLang, stored.Language)
			}
		}
	})
}

func TestArticlesStore_BulkUpsert(t *testing.T) {
	t.Run("insert all new articles", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()

		store := NewArticlesStore(db)

		articles := newTestArticles(5)
		ids, err := store.BulkUpsert(articles)
		if err != nil {
			t.Fatalf("BulkUpsert() failed: %v", err)
		}

		if len(ids) != 5 {
			t.Errorf("Expected 5 IDs, got %d", len(ids))
		}

		// Verify all articles were created
		count, err := store.Count(model.ArticleFilter{})
		if err != nil {
			t.Fatalf("Count() failed: %v", err)
		}

		if count != 5 {
			t.Errorf("Expected 5 articles in DB, found %d", count)
		}
	})

	t.Run("update all existing articles", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()

		store := NewArticlesStore(db)

		// Create initial articles
		original := newTestArticles(3)
		_, err := store.BulkCreate(original)
		if err != nil {
			t.Fatalf("BulkCreate() failed: %v", err)
		}

		time.Sleep(10 * time.Millisecond)

		// Update all with same URLs
		updated := make([]model.ScrapedArticle, 3)
		for i := range updated {
			updated[i] = newTestArticle()
			updated[i].URL = original[i].URL // Keep same URL
			updated[i].Title = "Updated Title " + string(rune('A'+i))
			updated[i].ContentText = "Updated content " + string(rune('A'+i))
			updated[i].ContentHTML = "<p>Updated content " + string(rune('A'+i)) + "</p>"
		}

		updatedIDs, err := store.BulkUpsert(updated)
		if err != nil {
			t.Fatalf("BulkUpsert() failed: %v", err)
		}

		if len(updatedIDs) != 3 {
			t.Errorf("Expected 3 IDs, got %d", len(updatedIDs))
		}

		// Verify only 3 articles exist (no duplicates created)
		count, err := store.Count(model.ArticleFilter{})
		if err != nil {
			t.Fatalf("Count() failed: %v", err)
		}

		if count != 3 {
			t.Errorf("Expected 3 articles (no duplicates), found %d", count)
		}

		// Verify updates were applied by checking URLs
		for i, updatedArticle := range updated {
			stored, err := store.GetByURL(updatedArticle.URL)
			if err != nil {
				t.Fatalf("GetByURL() failed: %v", err)
			}

			if stored == nil {
				t.Fatalf("Article %d not found by URL", i)
			}

			if stored.Title != updatedArticle.Title {
				t.Errorf("Article %d title not updated: got %s, want %s", i, stored.Title, updatedArticle.Title)
			}
			if stored.ContentText != updatedArticle.ContentText {
				t.Errorf("Article %d content text not updated: got %s, want %s", i, stored.ContentText, updatedArticle.ContentText)
			}
			if stored.ContentHTML != updatedArticle.ContentHTML {
				t.Errorf("Article %d content html not updated: got %s, want %s", i, stored.ContentHTML, updatedArticle.ContentHTML)
			}
		}
	})

	t.Run("mix of new and existing articles", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()

		store := NewArticlesStore(db)

		// Create 2 existing articles
		existing := newTestArticles(2)
		_, err := store.BulkCreate(existing)
		if err != nil {
			t.Fatalf("BulkCreate() failed: %v", err)
		}

		time.Sleep(10 * time.Millisecond)

		// Prepare upsert batch: 2 updates + 3 new inserts
		articles := newTestArticles(5)
		articles[0].URL = existing[0].URL // Update first existing
		articles[0].Title = "Updated First"
		articles[2].URL = existing[1].URL // Update second existing
		articles[2].Title = "Updated Second"
		// articles[1], [3], [4] are new

		ids, err := store.BulkUpsert(articles)
		if err != nil {
			t.Fatalf("BulkUpsert() failed: %v", err)
		}

		if len(ids) != 5 {
			t.Errorf("Expected 5 IDs, got %d", len(ids))
		}

		// Verify total count (2 existing + 3 new = 5)
		count, err := store.Count(model.ArticleFilter{})
		if err != nil {
			t.Fatalf("Count() failed: %v", err)
		}

		if count != 5 {
			t.Errorf("Expected 5 articles total, found %d", count)
		}

		// Verify updates were applied by checking URLs
		stored, err := store.GetByURL(articles[0].URL)
		if err != nil {
			t.Fatalf("GetByURL() failed: %v", err)
		}
		if stored == nil {
			t.Fatal("First updated article not found")
		}
		if stored.Title != "Updated First" {
			t.Errorf("First article not updated: got %s", stored.Title)
		}

		stored, err = store.GetByURL(articles[2].URL)
		if err != nil {
			t.Fatalf("GetByURL() failed: %v", err)
		}
		if stored == nil {
			t.Fatal("Second updated article not found")
		}
		if stored.Title != "Updated Second" {
			t.Errorf("Second article not updated: got %s", stored.Title)
		}

		// Verify new articles were created
		for _, idx := range []int{1, 3, 4} {
			stored, err := store.GetByURL(articles[idx].URL)
			if err != nil {
				t.Fatalf("GetByURL() failed for new article %d: %v", idx, err)
			}
			if stored == nil {
				t.Errorf("New article %d not found", idx)
			}
		}
	})

	t.Run("empty slice", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()

		store := NewArticlesStore(db)

		ids, err := store.BulkUpsert([]model.ScrapedArticle{})
		if err != nil {
			t.Fatalf("BulkUpsert() with empty slice failed: %v", err)
		}

		if len(ids) != 0 {
			t.Errorf("Expected 0 IDs for empty slice, got %d", len(ids))
		}
	})

	t.Run("large batch with mixed operations", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()

		store := NewArticlesStore(db)

		// Create 50 existing articles
		existing := newTestArticles(50)
		_, err := store.BulkCreate(existing)
		if err != nil {
			t.Fatalf("BulkCreate() failed: %v", err)
		}

		time.Sleep(10 * time.Millisecond)

		// Prepare batch: 50 updates + 50 new inserts
		articles := newTestArticles(100)
		for i := 0; i < 50; i++ {
			articles[i].URL = existing[i].URL // Reuse URLs for update
			articles[i].Title = "Updated " + articles[i].Title
		}
		// articles[50:100] are new

		ids, err := store.BulkUpsert(articles)
		if err != nil {
			t.Fatalf("BulkUpsert() with 100 articles failed: %v", err)
		}

		if len(ids) != 100 {
			t.Errorf("Expected 100 IDs, got %d", len(ids))
		}

		// Verify total count (50 updated + 50 new = 100 total)
		count, err := store.Count(model.ArticleFilter{})
		if err != nil {
			t.Fatalf("Count() failed: %v", err)
		}

		if count != 100 {
			t.Errorf("Expected 100 articles total, found %d", count)
		}
	})

	t.Run("verify CreatedAt preserved on update", func(t *testing.T) {
		db := setupTestDB(t)
		defer db.Close()

		store := NewArticlesStore(db)

		// Create initial articles
		original := newTestArticles(2)
		originalIDs, err := store.BulkCreate(original)
		if err != nil {
			t.Fatalf("BulkCreate() failed: %v", err)
		}

		// Get original timestamps
		firstStored, err := store.GetByID(originalIDs[0])
		if err != nil {
			t.Fatalf("GetByID() failed: %v", err)
		}

		time.Sleep(10 * time.Millisecond)

		// Upsert with same URLs
		updated := newTestArticles(2)
		for i := range updated {
			updated[i].URL = original[i].URL
			updated[i].Title = "Updated Title"
		}

		_, err = store.BulkUpsert(updated)
		if err != nil {
			t.Fatalf("BulkUpsert() failed: %v", err)
		}

		// Verify CreatedAt unchanged but UpdatedAt changed
		secondStored, err := store.GetByID(originalIDs[0])
		if err != nil {
			t.Fatalf("GetByID() failed: %v", err)
		}

		if !secondStored.CreatedAt.Equal(firstStored.CreatedAt) {
			t.Errorf("CreatedAt changed: before=%v, after=%v", firstStored.CreatedAt, secondStored.CreatedAt)
		}

		if !secondStored.UpdatedAt.After(firstStored.UpdatedAt) {
			t.Errorf("UpdatedAt not updated: before=%v, after=%v", firstStored.UpdatedAt, secondStored.UpdatedAt)
		}
	})
}
