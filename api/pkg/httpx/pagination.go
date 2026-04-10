package httpx

import "net/http"

type Paginated[T any] struct {
	Data       []T   `json:"data"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

func NewPaginated[T any](data []T, total int64, page, perPage int) *Paginated[T] {
	totalPages := 0
	if perPage > 0 {
		totalPages = int((total + int64(perPage) - 1) / int64(perPage))
	}

	return &Paginated[T]{
		Data:       data,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

func TransformPaginated[T any, R any](
	result *Paginated[T],
	transform func(T) R,
) *Paginated[R] {
	transformed := make([]R, len(result.Data))
	for i, item := range result.Data {
		transformed[i] = transform(item)
	}

	return &Paginated[R]{
		Data:       transformed,
		Total:      result.Total,
		Page:       result.Page,
		PerPage:    result.PerPage,
		TotalPages: result.TotalPages,
		HasNext:    result.HasNext,
		HasPrev:    result.HasPrev,
	}
}

func RespondPaginated[T any](w http.ResponseWriter, result *Paginated[T]) {
	RespondJSON(w, http.StatusOK, result)
}
