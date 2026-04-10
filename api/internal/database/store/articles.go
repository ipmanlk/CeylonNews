package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"ipmanlk/cnapi/internal/model"
)

type ArticlesStore struct {
	db *sql.DB
}

func NewArticlesStore(db *sql.DB) *ArticlesStore {
	return &ArticlesStore{db: db}
}

func (s *ArticlesStore) Create(scrapedArticle model.ScrapedArticle) (int64, error) {
	query := `
		INSERT INTO articles (source_id, title, url, content_text, content_html, image_url, language, published_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()

	result, err := s.db.Exec(query,
		scrapedArticle.SourceID,
		scrapedArticle.Title,
		scrapedArticle.URL,
		scrapedArticle.ContentText,
		scrapedArticle.ContentHTML,
		scrapedArticle.ImageURL,
		string(scrapedArticle.Language),
		scrapedArticle.PublishedAt,
		now,
		now,
	)

	if err != nil {
		return 0, fmt.Errorf("failed to create article: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return id, nil
}

func (s *ArticlesStore) Upsert(scrapedArticle model.ScrapedArticle) (int64, error) {
	query := `
		INSERT INTO articles (source_id, title, url, content_text, content_html, image_url, language, published_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(url) DO UPDATE SET
			source_id = excluded.source_id,
			title = excluded.title,
			content_text = excluded.content_text,
			content_html = excluded.content_html,
			image_url = excluded.image_url,
			language = excluded.language,
			published_at = excluded.published_at,
			updated_at = excluded.updated_at
	`

	now := time.Now()

	result, err := s.db.Exec(query,
		scrapedArticle.SourceID,
		scrapedArticle.Title,
		scrapedArticle.URL,
		scrapedArticle.ContentText,
		scrapedArticle.ContentHTML,
		scrapedArticle.ImageURL,
		string(scrapedArticle.Language),
		scrapedArticle.PublishedAt,
		now,
		now,
	)

	if err != nil {
		return 0, fmt.Errorf("failed to upsert article: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected > 0 {
		id, err := result.LastInsertId()
		if err == nil && id > 0 {
			return id, nil
		}

		existingArticle, err := s.GetByURL(scrapedArticle.URL)
		if err != nil {
			return 0, fmt.Errorf("failed to get updated article: %w", err)
		}
		if existingArticle != nil {
			return existingArticle.ID, nil
		}
	}

	return 0, fmt.Errorf("no article was affected by upsert operation")
}

func (s *ArticlesStore) BulkCreate(scrapedArticles []model.ScrapedArticle) ([]int64, error) {
	if len(scrapedArticles) == 0 {
		return []int64{}, nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO articles (source_id, title, url, content_text, content_html, image_url, language, published_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	stmt, err := tx.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	now := time.Now()
	ids := make([]int64, 0, len(scrapedArticles))

	for _, sa := range scrapedArticles {
		result, err := stmt.Exec(
			sa.SourceID,
			sa.Title,
			sa.URL,
			sa.ContentText,
			sa.ContentHTML,
			sa.ImageURL,
			string(sa.Language),
			sa.PublishedAt,
			now,
			now,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to insert article: %w", err)
		}

		id, err := result.LastInsertId()
		if err != nil {
			return nil, fmt.Errorf("failed to get last insert id: %w", err)
		}

		ids = append(ids, id)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return ids, nil
}

func (s *ArticlesStore) BulkUpsert(scrapedArticles []model.ScrapedArticle) ([]int64, error) {
	if len(scrapedArticles) == 0 {
		return []int64{}, nil
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO articles (source_id, title, url, content_text, content_html, image_url, language, published_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(url) DO UPDATE SET
			source_id = excluded.source_id,
			title = excluded.title,
			content_text = excluded.content_text,
			content_html = excluded.content_html,
			image_url = excluded.image_url,
			language = excluded.language,
			published_at = excluded.published_at,
			updated_at = excluded.updated_at
	`

	stmt, err := tx.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	now := time.Now()
	ids := make([]int64, 0, len(scrapedArticles))

	for _, sa := range scrapedArticles {
		result, err := stmt.Exec(
			sa.SourceID,
			sa.Title,
			sa.URL,
			sa.ContentText,
			sa.ContentHTML,
			sa.ImageURL,
			string(sa.Language),
			sa.PublishedAt,
			now,
			now,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to upsert article: %w", err)
		}

		id, err := result.LastInsertId()
		if err == nil && id > 0 {
			ids = append(ids, id)
		} else {
			existingArticle, err := s.GetByURL(sa.URL)
			if err != nil {
				return nil, fmt.Errorf("failed to get updated article: %w", err)
			}
			if existingArticle != nil {
				ids = append(ids, existingArticle.ID)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return ids, nil
}

func (s *ArticlesStore) GetByID(id int64) (*model.Article, error) {
	query := `
		SELECT id, source_id, title, url, content_text, content_html, image_url, language, published_at, created_at, updated_at
		FROM articles
		WHERE id = ?
	`

	article := &model.Article{}
	err := s.db.QueryRow(query, id).Scan(
		&article.ID,
		&article.SourceID,
		&article.Title,
		&article.URL,
		&article.ContentText,
		&article.ContentHTML,
		&article.ImageURL,
		&article.Language,
		&article.PublishedAt,
		&article.CreatedAt,
		&article.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get article by id: %w", err)
	}

	return article, nil
}

func (s *ArticlesStore) GetByIDWithFilter(id int64, filter model.ArticleFilter) (*model.Article, error) {
	selectFields := "id, source_id, title, url"
	if filter.IncludeText {
		selectFields += ", content_text"
	} else {
		selectFields += ", '' as content_text"
	}
	selectFields += ", content_html, image_url, language, published_at, created_at, updated_at"

	query := `
		SELECT ` + selectFields + `
		FROM articles
		WHERE id = ?
	`

	article := &model.Article{}
	err := s.db.QueryRow(query, id).Scan(
		&article.ID,
		&article.SourceID,
		&article.Title,
		&article.URL,
		&article.ContentText,
		&article.ContentHTML,
		&article.ImageURL,
		&article.Language,
		&article.PublishedAt,
		&article.CreatedAt,
		&article.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get article by id: %w", err)
	}

	return article, nil
}

func (s *ArticlesStore) GetByURL(url string) (*model.Article, error) {
	query := `
		SELECT id, source_id, title, url, content_text, content_html, image_url, language, published_at, created_at, updated_at
		FROM articles
		WHERE url = ?
	`

	article := &model.Article{}
	err := s.db.QueryRow(query, url).Scan(
		&article.ID,
		&article.SourceID,
		&article.Title,
		&article.URL,
		&article.ContentText,
		&article.ContentHTML,
		&article.ImageURL,
		&article.Language,
		&article.PublishedAt,
		&article.CreatedAt,
		&article.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get article by url: %w", err)
	}

	return article, nil
}

func (s *ArticlesStore) List(filter model.ArticleFilter) ([]*model.Article, error) {
	query, args := s.buildListQuery(filter)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query articles: %w", err)
	}
	defer rows.Close()

	var articles []*model.Article
	for rows.Next() {
		article := &model.Article{}
		err := rows.Scan(
			&article.ID,
			&article.SourceID,
			&article.Title,
			&article.URL,
			&article.ImageURL,
			&article.Language,
			&article.PublishedAt,
			&article.CreatedAt,
			&article.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan article: %w", err)
		}
		articles = append(articles, article)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return articles, nil
}

func (s *ArticlesStore) Count(filter model.ArticleFilter) (int64, error) {
	query, args := s.buildCountQuery(filter)

	var count int64
	err := s.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count articles: %w", err)
	}

	return count, nil
}

func (s *ArticlesStore) ListPaginated(filter model.ArticleFilter) (*model.Paginated[*model.Article], error) {
	total, err := s.Count(filter)
	if err != nil {
		return nil, err
	}

	articles, err := s.List(filter)
	if err != nil {
		return nil, err
	}

	page := (filter.Offset / filter.Limit) + 1
	if filter.Limit == 0 {
		page = 1
	}

	return model.NewPaginated(articles, total, page, filter.Limit), nil
}

func (s *ArticlesStore) Update(article *model.Article) error {
	query := `
		UPDATE articles 
		SET source_id = ?, title = ?, url = ?, content_text = ?, content_html = ?, image_url = ?, language = ?, 
		    published_at = ?, updated_at = ?
		WHERE id = ?
	`

	article.UpdatedAt = time.Now()

	result, err := s.db.Exec(query,
		article.SourceID,
		article.Title,
		article.URL,
		article.ContentText,
		article.ContentHTML,
		article.ImageURL,
		article.Language,
		article.PublishedAt,
		article.UpdatedAt,
		article.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update article: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("article with id %d not found", article.ID)
	}

	return nil
}

func (s *ArticlesStore) Delete(id int64) error {
	query := `DELETE FROM articles WHERE id = ?`

	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete article: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("article with id %d not found", id)
	}

	return nil
}

func (s *ArticlesStore) ExistsByURL(url string) (bool, error) {
	query := `SELECT 1 FROM articles WHERE url = ? LIMIT 1`

	var exists int
	err := s.db.QueryRow(query, url).Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check article existence: %w", err)
	}

	return true, nil
}

func (s *ArticlesStore) buildListQuery(filter model.ArticleFilter) (string, []interface{}) {
	var conditions []string
	var args []interface{}

	if len(filter.Languages) > 0 {
		placeholders := make([]string, len(filter.Languages))
		for i, lang := range filter.Languages {
			placeholders[i] = "?"
			args = append(args, lang)
		}
		conditions = append(conditions, fmt.Sprintf("language IN (%s)", strings.Join(placeholders, ", ")))
	}

	if len(filter.SourceIDs) > 0 {
		placeholders := make([]string, len(filter.SourceIDs))
		for i, source := range filter.SourceIDs {
			placeholders[i] = "?"
			args = append(args, source)
		}
		conditions = append(conditions, fmt.Sprintf("source_id IN (%s)", strings.Join(placeholders, ", ")))
	}

	if filter.StartDate != nil {
		conditions = append(conditions, "published_at >= ?")
		args = append(args, *filter.StartDate)
	}

	if filter.EndDate != nil {
		conditions = append(conditions, "published_at <= ?")
		args = append(args, *filter.EndDate)
	}

	selectFields := "id, source_id, title, url, image_url, language, published_at, created_at, updated_at"

	query := `
		SELECT ` + selectFields + `
		FROM articles
	`

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY id DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filter.Limit)
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", filter.Offset)
	}

	return query, args
}

func (s *ArticlesStore) buildCountQuery(filter model.ArticleFilter) (string, []interface{}) {
	var conditions []string
	var args []interface{}

	if len(filter.Languages) > 0 {
		placeholders := make([]string, len(filter.Languages))
		for i, lang := range filter.Languages {
			placeholders[i] = "?"
			args = append(args, lang)
		}
		conditions = append(conditions, fmt.Sprintf("language IN (%s)", strings.Join(placeholders, ", ")))
	}

	if len(filter.SourceIDs) > 0 {
		placeholders := make([]string, len(filter.SourceIDs))
		for i, source := range filter.SourceIDs {
			placeholders[i] = "?"
			args = append(args, source)
		}
		conditions = append(conditions, fmt.Sprintf("source_id IN (%s)", strings.Join(placeholders, ", ")))
	}

	if filter.StartDate != nil {
		conditions = append(conditions, "published_at >= ?")
		args = append(args, *filter.StartDate)
	}

	if filter.EndDate != nil {
		conditions = append(conditions, "published_at <= ?")
		args = append(args, *filter.EndDate)
	}

	query := `SELECT COUNT(*) FROM articles`

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	return query, args
}
