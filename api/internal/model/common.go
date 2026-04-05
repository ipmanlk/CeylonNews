package model

type Language string

const (
	LangEn Language = "en"
	LangSi Language = "si"
	LangTa Language = "ta"
)

type PaginatedResult[T any] struct {
	Data       []T   `json:"data"`
	Total      int64 `json:"total"`
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	TotalPages int   `json:"total_pages"`
	HasNext    bool  `json:"has_next"`
	HasPrev    bool  `json:"has_prev"`
}

func NewPaginatedResult[T any](data []T, total int64, page, perPage int) *PaginatedResult[T] {
	totalPages := int((total + int64(perPage) - 1) / int64(perPage))

	return &PaginatedResult[T]{
		Data:       data,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}
