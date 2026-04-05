package database

import (
	"database/sql"
	"ipmanlk/cnapi/internal/database/store"
	"ipmanlk/cnapi/internal/model"
)

// ArticlesStore defines the interface for article operations
type ArticlesStore interface {
	Create(scrapedArticle model.ScrapedArticle) (int64, error)
	Upsert(scrapedArticle model.ScrapedArticle) (int64, error)
	BulkCreate(scrapedArticles []model.ScrapedArticle) ([]int64, error)
	BulkUpsert(scrapedArticles []model.ScrapedArticle) ([]int64, error)
	GetByID(id int64) (*model.Article, error)
	GetByIDWithFilter(id int64, filter model.ArticleFilter) (*model.Article, error)
	GetByURL(url string) (*model.Article, error)
	List(filter model.ArticleFilter) ([]*model.Article, error)
	Count(filter model.ArticleFilter) (int64, error)
	ListPaginated(filter model.ArticleFilter) (*model.PaginatedResult[*model.Article], error)
	Update(article *model.Article) error
	Delete(id int64) error
	ExistsByURL(url string) (bool, error)
}

// SearchStore defines the interface for search operations
type SearchStore interface {
	Search(filter model.SearchFilter) (*model.PaginatedResult[*model.SearchResult], error)
	CountSearchResults(filter model.SearchFilter) (int64, error)
	GetAvailableSources() ([]string, error)
	GetAvailableLanguages() ([]string, error)
	GetSourcesByLanguage(language string) ([]string, error)
	GetRecentArticles(languages []string, sourceNames []string, limit int) ([]*model.Article, error)
}

// Store provides access to all database operations
type Store struct {
	Articles ArticlesStore
	Search   SearchStore
}

// NewStore creates a new database store with all required stores
func NewStore(db *sql.DB) *Store {
	return &Store{
		Articles: store.NewArticlesStore(db),
		Search:   store.NewSearchStore(db),
	}
}
