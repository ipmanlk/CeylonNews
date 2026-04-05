package dto

import (
	"errors"
	"fmt"
	"net/http"

	"ipmanlk/cnapi/pkg/httpx"
)

const (
	MinLimit     = 1
	MaxLimit     = 100
	DefaultLimit = 20
	MinOffset    = 0
)

var (
	ErrLimitTooSmall  = fmt.Errorf("limit must be at least %d", MinLimit)
	ErrLimitTooLarge  = fmt.Errorf("limit cannot exceed %d", MaxLimit)
	ErrOffsetNegative = errors.New("offset cannot be negative")
)

// PaginationRequest represents pagination parameters from HTTP request
type PaginationRequest struct {
	Limit  int
	Offset int
}

// Validate validates pagination parameters
func (p *PaginationRequest) Validate() error {
	if p.Limit < MinLimit {
		return ErrLimitTooSmall
	}
	if p.Limit > MaxLimit {
		return ErrLimitTooLarge
	}
	if p.Offset < MinOffset {
		return ErrOffsetNegative
	}
	return nil
}

// ParsePaginationRequest parses and validates pagination parameters from HTTP request
func ParsePaginationRequest(r *http.Request) (*PaginationRequest, error) {
	limit, err := httpx.ParseQueryInt(r, "limit", DefaultLimit)
	if err != nil {
		return nil, err
	}

	offset, err := httpx.ParseQueryInt(r, "offset", 0)
	if err != nil {
		return nil, err
	}

	req := &PaginationRequest{
		Limit:  limit,
		Offset: offset,
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}

	return req, nil
}
