package dto

// PageMetaData represents metadata about the current page of results.
type PageMetaData struct {
	Page      int64  `json:"page"`
	Size      int64  `json:"size"`
	TotalItem int64  `json:"total_item"`
	TotalPage int64  `json:"total_page"`
	Links     *Links `json:"links"`
}

// Links represents pagination links.
type Links struct {
	Self  string `json:"self"`
	First string `json:"first"`
	Prev  string `json:"prev"`
	Next  string `json:"next"`
	Last  string `json:"last"`
}
