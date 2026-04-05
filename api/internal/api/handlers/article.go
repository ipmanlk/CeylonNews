package handlers

import (
	"context"
	"log/slog"
	"net/http"

	"ipmanlk/cnapi/internal/api/dto"
	"ipmanlk/cnapi/internal/model"
	"ipmanlk/cnapi/pkg/httpx"
)

type ArticleService interface {
	GetByIDWithFilter(ctx context.Context, id int64, filter model.ArticleFilter) (*model.Article, error)
	ListPaginated(ctx context.Context, filter model.ArticleFilter) (*model.PaginatedResult[*model.Article], error)
}

type ArticleListResponse struct {
	ID          int64   `json:"id"`
	SourceName  string  `json:"source_name"`
	Title       string  `json:"title"`
	URL         string  `json:"url"`
	ImageURL    *string `json:"image_url,omitempty"`
	Language    string  `json:"language"`
	PublishedAt string  `json:"published_at"`
}

type ArticleResponse struct {
	ID          int64   `json:"id"`
	SourceName  string  `json:"source_name"`
	Title       string  `json:"title"`
	URL         string  `json:"url"`
	ContentHTML string  `json:"content_html,omitempty"`
	ContentText *string `json:"content_text,omitempty"`
	ImageURL    *string `json:"image_url,omitempty"`
	Language    string  `json:"language"`
	PublishedAt string  `json:"published_at"`
}

type ArticleHandler struct {
	articleService ArticleService
}

func NewArticleHandler(articleService ArticleService) *ArticleHandler {
	return &ArticleHandler{
		articleService: articleService,
	}
}

func toArticleListResponse(article *model.Article) ArticleListResponse {
	return ArticleListResponse{
		ID:          article.ID,
		SourceName:  article.SourceName,
		Title:       article.Title,
		URL:         article.URL,
		ImageURL:    article.ImageURL,
		Language:    article.Language,
		PublishedAt: article.PublishedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func (h *ArticleHandler) List(w http.ResponseWriter, r *http.Request) {
	pagination, err := dto.ParsePaginationRequest(r)
	if err != nil {
		httpx.RespondBadRequest(w, err.Error())
		return
	}

	filterParams, err := dto.ParseArticleFilterRequest(r)
	if err != nil {
		httpx.RespondBadRequest(w, err.Error())
		return
	}

	filter := model.ArticleFilter{
		Languages:   filterParams.Languages,
		SourceNames: filterParams.SourceNames,
		StartDate:   filterParams.StartDate,
		EndDate:     filterParams.EndDate,
		Limit:       pagination.Limit,
		Offset:      pagination.Offset,
	}

	result, err := h.articleService.ListPaginated(r.Context(), filter)
	if err != nil {
		slog.Error("failed to list articles", "error", err)
		httpx.RespondInternalError(w, "failed to retrieve articles")
		return
	}

	response := httpx.TransformPaginated(result, toArticleListResponse)
	httpx.RespondPaginated(w, response)
}

func (h *ArticleHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := httpx.ParsePathInt64(r, "id")
	if err != nil {
		httpx.RespondBadRequest(w, err.Error())
		return
	}

	includeText := r.URL.Query().Get("include_text") == "true"
	filter := model.ArticleFilter{IncludeText: includeText}

	article, err := h.articleService.GetByIDWithFilter(r.Context(), id, filter)
	if err != nil {
		slog.Error("failed to get article", "id", id, "error", err)
		httpx.RespondInternalError(w, "failed to retrieve article")
		return
	}

	if article == nil {
		httpx.RespondNotFound(w, "article not found")
		return
	}

	response := ArticleResponse{
		ID:          article.ID,
		SourceName:  article.SourceName,
		Title:       article.Title,
		URL:         article.URL,
		ContentHTML: article.ContentHTML,
		ImageURL:    article.ImageURL,
		Language:    article.Language,
		PublishedAt: article.PublishedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if includeText {
		response.ContentText = &article.ContentText
	}

	httpx.RespondJSON(w, http.StatusOK, response)
}
