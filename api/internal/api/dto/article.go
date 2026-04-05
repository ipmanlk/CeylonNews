package dto

import (
	"errors"
	"net/http"
	"time"

	"ipmanlk/cnapi/pkg/httpx"
)

var (
	ErrInvalidLanguage  = errors.New("language must be one of: en, si, ta")
	ErrInvalidDateRange = errors.New("start_date must be before end_date")
)

// ArticleFilterRequest represents article filtering parameters from HTTP request
type ArticleFilterRequest struct {
	Languages   []string
	SourceNames []string
	StartDate   *time.Time
	EndDate     *time.Time
}

// Validate validates article filter parameters
func (f *ArticleFilterRequest) Validate() error {
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

// ParseArticleFilterRequest parses and validates article filter parameters from HTTP request
func ParseArticleFilterRequest(r *http.Request) (*ArticleFilterRequest, error) {
	startDate, err := httpx.ParseQueryTime(r, "start_date")
	if err != nil {
		return nil, err
	}

	endDate, err := httpx.ParseQueryTime(r, "end_date")
	if err != nil {
		return nil, err
	}

	req := &ArticleFilterRequest{
		Languages:   httpx.ParseQueryStringsFromCSV(r, "languages"),
		SourceNames: httpx.ParseQueryStrings(r, "source_names"),
		StartDate:   startDate,
		EndDate:     endDate,
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	return req, nil
}

func isValidLanguage(lang string) bool {
	return lang == "en" || lang == "si" || lang == "ta"
}
