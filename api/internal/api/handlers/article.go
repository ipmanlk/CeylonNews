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
	ListPaginated(ctx context.Context, filter model.ArticleFilter) (*model.Paginated[*model.Article], error)
}

type ArticleListResponse struct {
	ID          int64   `json:"id"`
	SourceID    string  `json:"source_id"`
	SourceName  string  `json:"source_name"`
	Title       string  `json:"title"`
	URL         string  `json:"url"`
	ImageURL    *string `json:"image_url,omitempty"`
	Language    string  `json:"language"`
	PublishedAt string  `json:"published_at"`
}

type ArticleResponse struct {
	ID          int64   `json:"id"`
	SourceID    string  `json:"source_id"`
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
	sourceResolver SourceResolver
}

func NewArticleHandler(articleService ArticleService, sourceResolver SourceResolver) *ArticleHandler {
	return &ArticleHandler{
		articleService: articleService,
		sourceResolver: sourceResolver,
	}
}

func (h *ArticleHandler) toArticleListResponse(article *model.Article) ArticleListResponse {
	sourceName, _ := h.sourceResolver.GetSourceNameByID(article.SourceID)
	return ArticleListResponse{
		ID:          article.ID,
		SourceID:    article.SourceID,
		SourceName:  sourceName,
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

	// TODO: Remove this compatibility layer when mobile app is updated to use source_ids
	sourceIDs := filterParams.SourceIDs
	if len(sourceIDs) == 0 {
		sourceNames := r.URL.Query()["source_names"]
		if len(sourceNames) > 0 {
			for _, name := range sourceNames {
				if id, ok := h.sourceResolver.GetSourceIDByName(name); ok {
					sourceIDs = append(sourceIDs, id)
				}
			}
		}
	}

	filter := model.ArticleFilter{
		Languages: filterParams.Languages,
		SourceIDs: sourceIDs,
		StartDate: filterParams.StartDate,
		EndDate:   filterParams.EndDate,
		Limit:     pagination.Limit,
		Offset:    pagination.Offset,
	}

	result, err := h.articleService.ListPaginated(r.Context(), filter)
	if err != nil {
		slog.Error("failed to list articles", "error", err)
		httpx.RespondInternalError(w, "failed to retrieve articles")
		return
	}

	response := httpx.TransformPaginated(result, h.toArticleListResponse)
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

	sourceName, _ := h.sourceResolver.GetSourceNameByID(article.SourceID)
	response := ArticleResponse{
		ID:          article.ID,
		SourceID:    article.SourceID,
		SourceName:  sourceName,
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
