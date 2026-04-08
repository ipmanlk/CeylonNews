package handlers

import (
	"context"
	"log/slog"
	"net/http"

	"ipmanlk/cnapi/internal/api/dto"
	"ipmanlk/cnapi/internal/model"
	"ipmanlk/cnapi/pkg/httpx"
)

type SearchService interface {
	Search(ctx context.Context, filter model.SearchFilter) (*model.Paginated[*model.SearchResult], error)
	GetAvailableSources() ([]string, error)
	GetAvailableLanguages() ([]string, error)
	GetSourcesByLanguage(language string) ([]string, error)
	GetRecentArticles(languages []string, sourceIDs []string, limit int) ([]*model.Article, error)
}

type SearchResultResponse struct {
	ID             int64   `json:"id"`
	SourceID       string  `json:"source_id"`
	SourceName     string  `json:"source_name"`
	Title          string  `json:"title"`
	URL            string  `json:"url"`
	ImageURL       *string `json:"image_url,omitempty"`
	Language       string  `json:"language"`
	PublishedAt    string  `json:"published_at"`
	RelevanceScore float64 `json:"relevance_score"`
}

type SearchHandler struct {
	searchService  SearchService
	sourceResolver SourceResolver
}

func NewSearchHandler(searchService SearchService, sourceResolver SourceResolver) *SearchHandler {
	return &SearchHandler{
		searchService:  searchService,
		sourceResolver: sourceResolver,
	}
}

func (h *SearchHandler) toSearchResultResponse(result *model.SearchResult) SearchResultResponse {
	sourceName, _ := h.sourceResolver.GetSourceNameByID(result.SourceID)
	return SearchResultResponse{
		ID:             result.ID,
		SourceID:       result.SourceID,
		SourceName:     sourceName,
		Title:          result.Title,
		URL:            result.URL,
		ImageURL:       result.ImageURL,
		Language:       result.Language,
		PublishedAt:    result.PublishedAt.Format("2006-01-02T15:04:05Z07:00"),
		RelevanceScore: result.RelevanceScore,
	}
}

func (h *SearchHandler) Search(w http.ResponseWriter, r *http.Request) {
	pagination, err := dto.ParsePaginationRequest(r)
	if err != nil {
		httpx.RespondBadRequest(w, err.Error())
		return
	}

	searchParams, err := dto.ParseSearchFilterRequest(r)
	if err != nil {
		httpx.RespondBadRequest(w, err.Error())
		return
	}

	if searchParams.Query == "" {
		httpx.RespondBadRequest(w, "query parameter 'query' is required")
		return
	}

	// TODO: Remove this compatibility layer when mobile app is updated to use source_ids
	sourceIDs := searchParams.SourceIDs
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

	filter := model.SearchFilter{
		Query:     searchParams.Query,
		Languages: searchParams.Languages,
		SourceIDs: sourceIDs,
		StartDate: searchParams.StartDate,
		EndDate:   searchParams.EndDate,
		Limit:     pagination.Limit,
		Offset:    pagination.Offset,
	}

	paginatedResult, err := h.searchService.Search(r.Context(), filter)
	if err != nil {
		slog.Error("failed to search articles", "query", searchParams.Query, "error", err)
		httpx.RespondInternalError(w, "failed to search articles")
		return
	}

	response := httpx.TransformPaginated(paginatedResult, h.toSearchResultResponse)
	httpx.RespondPaginated(w, response)
}

func (h *SearchHandler) GetAvailableSources(w http.ResponseWriter, r *http.Request) {
	sources, err := h.searchService.GetAvailableSources()
	if err != nil {
		slog.Error("failed to get available sources", "error", err)
		httpx.RespondInternalError(w, "failed to retrieve sources")
		return
	}

	httpx.RespondJSON(w, http.StatusOK, map[string]any{
		"sources": sources,
	})
}

func (h *SearchHandler) GetAvailableLanguages(w http.ResponseWriter, r *http.Request) {
	languages, err := h.searchService.GetAvailableLanguages()
	if err != nil {
		slog.Error("failed to get available languages", "error", err)
		httpx.RespondInternalError(w, "failed to retrieve languages")
		return
	}

	httpx.RespondJSON(w, http.StatusOK, map[string]any{
		"languages": languages,
	})
}

func (h *SearchHandler) GetSourcesByLanguage(w http.ResponseWriter, r *http.Request) {
	language := httpx.ParseQueryString(r, "language", "")
	if language == "" {
		httpx.RespondBadRequest(w, "query parameter 'language' is required")
		return
	}

	sources, err := h.searchService.GetSourcesByLanguage(language)
	if err != nil {
		slog.Error("failed to get sources by language", "language", language, "error", err)
		httpx.RespondInternalError(w, "failed to retrieve sources")
		return
	}

	httpx.RespondJSON(w, http.StatusOK, map[string]any{
		"language": language,
		"sources":  sources,
	})
}

func (h *SearchHandler) GetRecentArticles(w http.ResponseWriter, r *http.Request) {
	languages := httpx.ParseQueryStringsFromCSV(r, "languages")
	sourceIDs := httpx.ParseQueryStrings(r, "source_ids")

	// TODO: Remove this compatibility layer when mobile app is updated to use source_ids
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

	limit, err := httpx.ParseQueryInt(r, "limit", 20)
	if err != nil {
		httpx.RespondBadRequest(w, err.Error())
		return
	}

	articles, err := h.searchService.GetRecentArticles(languages, sourceIDs, limit)
	if err != nil {
		slog.Error("failed to get recent articles", "error", err)
		httpx.RespondInternalError(w, "failed to retrieve recent articles")
		return
	}

	searchResponses := make([]SearchResultResponse, len(articles))
	for i, article := range articles {
		sourceName, _ := h.sourceResolver.GetSourceNameByID(article.SourceID)
		searchResponses[i] = SearchResultResponse{
			ID:             article.ID,
			SourceID:       article.SourceID,
			SourceName:     sourceName,
			Title:          article.Title,
			URL:            article.URL,
			ImageURL:       article.ImageURL,
			Language:       article.Language,
			PublishedAt:    article.PublishedAt.Format("2006-01-02T15:04:05Z07:00"),
			RelevanceScore: 0,
		}
	}

	httpx.RespondJSON(w, http.StatusOK, map[string]any{
		"articles": searchResponses,
		"count":    len(searchResponses),
	})
}
