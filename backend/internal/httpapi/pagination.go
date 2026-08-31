package httpapi

import (
	"net/http"
	"strconv"
)

const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

type Pagination struct {
	Limit  int
	Offset int
	Page   int
}

// ParsePagination reads ?page=&page_size= and clamps page_size to avoid
// letting a client force an unbounded table scan on tables that will grow
// to millions of rows.
func ParsePagination(r *http.Request) Pagination {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	return Pagination{
		Limit:  pageSize,
		Offset: (page - 1) * pageSize,
		Page:   page,
	}
}

type PagedResult struct {
	Data       interface{} `json:"data"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalCount int64       `json:"total_count"`
}
