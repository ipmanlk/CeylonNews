package service

import (
	"context"

	"ipmanlk/cnapi/internal/database"
	"ipmanlk/cnapi/internal/model"
)

type articleService struct {
	store database.ArticlesStore
}

func NewArticleService(store database.ArticlesStore) *articleService {
	return &articleService{store: store}
}

func (s *articleService) Create(ctx context.Context, article model.ScrapedArticle) (int64, error) {
	return s.store.Create(article)
}

func (s *articleService) BulkCreate(ctx context.Context, articles []model.ScrapedArticle) ([]int64, error) {
	return s.store.BulkCreate(articles)
}

func (s *articleService) BulkUpsert(ctx context.Context, articles []model.ScrapedArticle) ([]int64, error) {
	return s.store.BulkUpsert(articles)
}

func (s *articleService) GetByID(ctx context.Context, id int64) (*model.Article, error) {
	return s.store.GetByID(id)
}

func (s *articleService) GetByIDWithFilter(ctx context.Context, id int64, filter model.ArticleFilter) (*model.Article, error) {
	return s.store.GetByIDWithFilter(id, filter)
}

func (s *articleService) GetByURL(ctx context.Context, url string) (*model.Article, error) {
	return s.store.GetByURL(url)
}

func (s *articleService) List(ctx context.Context, filter model.ArticleFilter) ([]*model.Article, error) {
	return s.store.List(filter)
}

func (s *articleService) ListPaginated(ctx context.Context, filter model.ArticleFilter) (*model.Paginated[*model.Article], error) {
	return s.store.ListPaginated(filter)
}

func (s *articleService) Update(ctx context.Context, article *model.Article) error {
	return s.store.Update(article)
}

func (s *articleService) Delete(ctx context.Context, id int64) error {
	return s.store.Delete(id)
}

func (s *articleService) ExistsByURL(ctx context.Context, url string) (bool, error) {
	return s.store.ExistsByURL(url)
}
