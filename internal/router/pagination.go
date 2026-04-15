package router

import (
	"net/http"
	"strconv"
)

const (
	defaultPage     = 1
	defaultPageSize = 20
	maxPageSize     = 200
)

// PageParams holds parsed pagination parameters from a request.
type PageParams struct {
	Page int
	Size int
}

// Offset returns the zero-based row offset for the current page.
func (p PageParams) Offset() int {
	if p.Page < 1 {
		return 0
	}
	return (p.Page - 1) * p.Size
}

// ParsePageParams extracts "page" and "size" query parameters from r.
// Missing or invalid values fall back to defaults; size is capped at maxPageSize.
func ParsePageParams(r *http.Request) PageParams {
	page := queryInt(r, "page", defaultPage)
	size := queryInt(r, "size", defaultPageSize)

	if page < 1 {
		page = defaultPage
	}
	if size < 1 {
		size = defaultPageSize
	}
	if size > maxPageSize {
		size = maxPageSize
	}
	return PageParams{Page: page, Size: size}
}

func queryInt(r *http.Request, key string, fallback int) int {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return v
}
