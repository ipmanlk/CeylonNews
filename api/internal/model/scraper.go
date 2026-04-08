package model

import "time"

type ScrapedArticle struct {
	SourceID    string
	Title       string
	URL         string
	ContentText string
	ContentHTML string
	ImageURL    *string
	Categories  []string
	Language    Language
	PublishedAt time.Time
}

type Article struct {
	ID          int64     `json:"id" db:"id"`
	SourceID    string    `json:"source_id" db:"source_id"`
	Title       string    `json:"title" db:"title"`
	URL         string    `json:"url" db:"url"`
	ContentText string    `json:"content_text" db:"content_text"`
	ContentHTML string    `json:"content_html,omitempty" db:"content_html"`
	ImageURL    *string   `json:"image_url,omitempty" db:"image_url"`
	Language    string    `json:"language" db:"language"`
	PublishedAt time.Time `json:"published_at" db:"published_at"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type ArticleFilter struct {
	Languages   []string   `json:"languages,omitempty"`
	SourceIDs   []string   `json:"source_ids,omitempty"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	Limit       int        `json:"limit"`
	Offset      int        `json:"offset"`
	IncludeText bool       `json:"include_text,omitempty"`
}

type SearchFilter struct {
	Query     string     `json:"query"`
	Languages []string   `json:"languages,omitempty"`
	SourceIDs []string   `json:"source_ids,omitempty"`
	StartDate *time.Time `json:"start_date,omitempty"`
	EndDate   *time.Time `json:"end_date,omitempty"`
	Limit     int        `json:"limit"`
	Offset    int        `json:"offset"`
}

type SearchResult struct {
	Article
	RelevanceScore float64 `json:"relevance_score" db:"rank"`
}
