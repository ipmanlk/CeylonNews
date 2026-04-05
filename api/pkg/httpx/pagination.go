package httpx

import (
	"net/http"

	"ipmanlk/cnapi/internal/model"
)

func TransformPaginated[T any, R any](
	result *model.PaginatedResult[T],
	transform func(T) R,
) *model.PaginatedResult[R] {
	transformed := make([]R, len(result.Data))
	for i, item := range result.Data {
		transformed[i] = transform(item)
	}

	return &model.PaginatedResult[R]{
		Data:       transformed,
		Total:      result.Total,
		Page:       result.Page,
		PerPage:    result.PerPage,
		TotalPages: result.TotalPages,
		HasNext:    result.HasNext,
		HasPrev:    result.HasPrev,
	}
}

func RespondPaginated[T any](w http.ResponseWriter, result *model.PaginatedResult[T]) {
	RespondJSON(w, http.StatusOK, result)
}
