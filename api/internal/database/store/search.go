package store

import (
	"database/sql"
	"fmt"
	"strings"

	"ipmanlk/cnapi/internal/model"
)

type SourceInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type SearchStore struct {
	db *sql.DB
}

func NewSearchStore(db *sql.DB) *SearchStore {
	return &SearchStore{db: db}
}

func (s *SearchStore) Search(filter model.SearchFilter) (*model.Paginated[*model.SearchResult], error) {
	total, err := s.CountSearchResults(filter)
	if err != nil {
		return nil, err
	}

	query, args := s.buildSearchQuery(filter)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search articles: %w", err)
	}
	defer rows.Close()

	var results []*model.SearchResult
	for rows.Next() {
		result := &model.SearchResult{}
		err := rows.Scan(
			&result.ID,
			&result.SourceID,
			&result.Title,
			&result.URL,
			&result.ImageURL,
			&result.Language,
			&result.PublishedAt,
			&result.CreatedAt,
			&result.UpdatedAt,
			&result.RelevanceScore,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan search result: %w", err)
		}
		results = append(results, result)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating search rows: %w", err)
	}

	page := (filter.Offset / filter.Limit) + 1
	if filter.Limit == 0 {
		page = 1
	}

	return model.NewPaginated(results, total, page, filter.Limit), nil
}

func (s *SearchStore) CountSearchResults(filter model.SearchFilter) (int64, error) {
	query, args := s.buildSearchCountQuery(filter)

	var count int64
	err := s.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count search results: %w", err)
	}

	return count, nil
}

func (s *SearchStore) GetAvailableSources() ([]string, error) {
	query := `SELECT DISTINCT source_id FROM articles ORDER BY source_id`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get available sources: %w", err)
	}
	defer rows.Close()

	var sources []string
	for rows.Next() {
		var source string
		err := rows.Scan(&source)
		if err != nil {
			return nil, fmt.Errorf("failed to scan source: %w", err)
		}
		sources = append(sources, source)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating source rows: %w", err)
	}

	return sources, nil
}

func (s *SearchStore) GetAvailableLanguages() ([]string, error) {
	query := `SELECT DISTINCT language FROM articles ORDER BY language`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get available languages: %w", err)
	}
	defer rows.Close()

	var languages []string
	for rows.Next() {
		var language string
		err := rows.Scan(&language)
		if err != nil {
			return nil, fmt.Errorf("failed to scan language: %w", err)
		}
		languages = append(languages, language)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating language rows: %w", err)
	}

	return languages, nil
}

func (s *SearchStore) GetSourcesByLanguage(language string) ([]SourceInfo, error) {
	query := `SELECT DISTINCT source_id FROM articles WHERE language = ? ORDER BY source_id`

	rows, err := s.db.Query(query, language)
	if err != nil {
		return nil, fmt.Errorf("failed to get sources by language: %w", err)
	}
	defer rows.Close()

	var sources []SourceInfo
	for rows.Next() {
		var sourceID string
		err := rows.Scan(&sourceID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan source: %w", err)
		}
		sources = append(sources, SourceInfo{ID: sourceID})
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating source rows: %w", err)
	}

	return sources, nil
}

func (s *SearchStore) buildSearchQuery(filter model.SearchFilter) (string, []interface{}) {
	var conditions []string
	var args []interface{}

	if filter.Query != "" {
		escapedQuery := s.escapeFTSQuery(filter.Query)
		conditions = append(conditions, "articles_fts MATCH ?")
		args = append(args, escapedQuery)
	}

	if len(filter.Languages) > 0 {
		placeholders := make([]string, len(filter.Languages))
		for i, lang := range filter.Languages {
			placeholders[i] = "?"
			args = append(args, lang)
		}
		conditions = append(conditions, fmt.Sprintf("a.language IN (%s)", strings.Join(placeholders, ",")))
	}

	if len(filter.SourceIDs) > 0 {
		placeholders := make([]string, len(filter.SourceIDs))
		for i, source := range filter.SourceIDs {
			placeholders[i] = "?"
			args = append(args, source)
		}
		conditions = append(conditions, fmt.Sprintf("a.source_id IN (%s)", strings.Join(placeholders, ",")))
	}

	if filter.StartDate != nil {
		conditions = append(conditions, "a.published_at >= ?")
		args = append(args, *filter.StartDate)
	}

	if filter.EndDate != nil {
		conditions = append(conditions, "a.published_at <= ?")
		args = append(args, *filter.EndDate)
	}

	selectFields := "a.id, a.source_id, a.title, a.url, a.image_url, a.language, a.published_at, a.created_at, a.updated_at, articles_fts.rank"

	query := `
		SELECT ` + selectFields + `
		FROM articles a
		JOIN articles_fts ON a.id = articles_fts.rowid
	`

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	if filter.Query != "" {
		query += " ORDER BY articles_fts.rank, a.id DESC"
	} else {
		query += " ORDER BY a.id DESC"
	}

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", filter.Limit)
	}

	if filter.Offset > 0 {
		query += fmt.Sprintf(" OFFSET %d", filter.Offset)
	}

	return query, args
}

func (s *SearchStore) buildSearchCountQuery(filter model.SearchFilter) (string, []interface{}) {
	var conditions []string
	var args []interface{}

	if filter.Query != "" {
		escapedQuery := s.escapeFTSQuery(filter.Query)
		conditions = append(conditions, "articles_fts MATCH ?")
		args = append(args, escapedQuery)
	}

	if len(filter.Languages) > 0 {
		placeholders := make([]string, len(filter.Languages))
		for i, lang := range filter.Languages {
			placeholders[i] = "?"
			args = append(args, lang)
		}
		conditions = append(conditions, fmt.Sprintf("a.language IN (%s)", strings.Join(placeholders, ",")))
	}

	if len(filter.SourceIDs) > 0 {
		placeholders := make([]string, len(filter.SourceIDs))
		for i, source := range filter.SourceIDs {
			placeholders[i] = "?"
			args = append(args, source)
		}
		conditions = append(conditions, fmt.Sprintf("a.source_id IN (%s)", strings.Join(placeholders, ",")))
	}

	if filter.StartDate != nil {
		conditions = append(conditions, "a.published_at >= ?")
		args = append(args, *filter.StartDate)
	}

	if filter.EndDate != nil {
		conditions = append(conditions, "a.published_at <= ?")
		args = append(args, *filter.EndDate)
	}

	query := `
		SELECT COUNT(*)
		FROM articles a
		JOIN articles_fts ON a.id = articles_fts.rowid
	`

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	return query, args
}

func (s *SearchStore) escapeFTSQuery(query string) string {
	// Escape double quotes for FTS5 phrase search syntax
	escaped := strings.ReplaceAll(query, `"`, `""`)
	return `"` + escaped + `"`
}

func (s *SearchStore) GetRecentArticles(languages []string, sourceIDs []string, limit int) ([]*model.Article, error) {
	var conditions []string
	var args []interface{}

	if len(languages) > 0 {
		placeholders := make([]string, len(languages))
		for i, lang := range languages {
			placeholders[i] = "?"
			args = append(args, lang)
		}
		conditions = append(conditions, fmt.Sprintf("language IN (%s)", strings.Join(placeholders, ",")))
	}

	if len(sourceIDs) > 0 {
		placeholders := make([]string, len(sourceIDs))
		for i, source := range sourceIDs {
			placeholders[i] = "?"
			args = append(args, source)
		}
		conditions = append(conditions, fmt.Sprintf("source_id IN (%s)", strings.Join(placeholders, ",")))
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

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent articles: %w", err)
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
		return nil, fmt.Errorf("error iterating article rows: %w", err)
	}

	return articles, nil
}
