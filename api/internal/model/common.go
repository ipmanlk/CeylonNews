package model

import "ipmanlk/cnapi/pkg/httpx"

type Language string

const (
	LangEn Language = "en"
	LangSi Language = "si"
	LangTa Language = "ta"
)

type Paginated[T any] = httpx.Paginated[T]

func NewPaginated[T any](data []T, total int64, page, perPage int) *Paginated[T] {
	return httpx.NewPaginated(data, total, page, perPage)
}
