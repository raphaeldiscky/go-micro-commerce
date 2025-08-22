// Package pageutils provides utility functions for generating pagination links.
package pageutils

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/raphaeldiscky/go-micro-template/pkg/dto"
)

const (
	linkFormat = "%v%v?%v"
)

// NewLinks creates a new Links instance for pagination.
func NewLinks(r *http.Request, page, size, totalPage int) *dto.Links {
	host := r.Host
	path := r.URL.Path

	if r.TLS != nil {
		host = fmt.Sprintf("https://%v", host)
	} else {
		host = fmt.Sprintf("http://%v", host)
	}

	// Preserve existing queries
	queries := r.URL.Query()
	queries.Set("size", strconv.Itoa(size))

	// Self link
	selfQueries := cloneQuery(queries)
	selfQueries.Set("page", strconv.Itoa(page))
	selfLink := fmt.Sprintf(linkFormat, host, path, selfQueries.Encode())

	// First link
	firstQueries := cloneQuery(queries)
	firstQueries.Set("page", "1")
	firstLink := fmt.Sprintf(linkFormat, host, path, firstQueries.Encode())

	// Last link
	lastQueries := cloneQuery(queries)
	if totalPage > 0 {
		lastQueries.Set("page", strconv.Itoa(totalPage))
	} else {
		lastQueries.Set("page", "1")
	}

	lastLink := fmt.Sprintf(linkFormat, host, path, lastQueries.Encode())

	// Prev link
	prevLink := createPrevLink(queries, host, path, page)

	// Next link
	nextLink := createNextLink(queries, host, path, page, totalPage)

	return &dto.Links{
		Self:  selfLink,
		First: firstLink,
		Prev:  prevLink,
		Next:  nextLink,
		Last:  lastLink,
	}
}

// createNextLink creates the next link for pagination.
func createNextLink(queries url.Values, host, path string, page, totalPage int) string {
	if page >= totalPage { // no next page
		return ""
	}

	q := cloneQuery(queries)
	q.Set("page", strconv.Itoa(page+1))

	return fmt.Sprintf(linkFormat, host, path, q.Encode())
}

// createPrevLink creates the previous link for pagination.
func createPrevLink(queries url.Values, host, path string, page int) string {
	if page <= 1 { // no previous page
		return ""
	}

	q := cloneQuery(queries)
	q.Set("page", strconv.Itoa(page-1))

	return fmt.Sprintf(linkFormat, host, path, q.Encode())
}

// cloneQuery makes a copy of url.Values so mutations don’t leak across links.
func cloneQuery(q url.Values) url.Values {
	clone := url.Values{}
	for k, v := range q {
		clone[k] = append([]string(nil), v...)
	}

	return clone
}
