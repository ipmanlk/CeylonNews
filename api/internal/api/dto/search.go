package dto

import (
	"errors"
	"net/http"
	"time"

	"ipmanlk/cnapi/pkg/httpx"
)

var (
	ErrQueryRequired = errors.New("query parameter 'query' is required")
	ErrQueryTooShort = errors.New("query must be at least 2 characters")
)

// SearchFilterRequest represents search filtering parameters from HTTP request
type SearchFilterRequest struct {
	Query     string
	Languages []string
	SourceIDs []string
	StartDate *time.Time
	EndDate   *time.Time
}

func (f *SearchFilterRequest) Validate() error {
	if f.Query == "" {
		return ErrQueryRequired
	}

	if len(f.Query) < 2 {
		return ErrQueryTooShort
	}

	for _, lang := range f.Languages {
		if !isValidLanguage(lang) {
			return ErrInvalidLanguage
		}
	}

	if f.StartDate != nil && f.EndDate != nil {
		if f.StartDate.After(*f.EndDate) {
			return ErrInvalidDateRange
		}
	}

	return nil
}

// ParseSearchFilterRequest parses and validates search filter parameters from HTTP request
func ParseSearchFilterRequest(r *http.Request) (*SearchFilterRequest, error) {
	query := httpx.ParseQueryString(r, "query", "")

	startDate, err := httpx.ParseQueryTime(r, "start_date")
	if err != nil {
		return nil, err
	}

	endDate, err := httpx.ParseQueryTime(r, "end_date")
	if err != nil {
		return nil, err
	}

	req := &SearchFilterRequest{
		Query:     query,
		Languages: httpx.ParseQueryStringsFromCSV(r, "languages"),
		SourceIDs: httpx.ParseQueryStrings(r, "source_ids"),
		StartDate: startDate,
		EndDate:   endDate,
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	return req, nil
}

type SourceResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
