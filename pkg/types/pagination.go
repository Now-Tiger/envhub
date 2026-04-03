package types

// PaginationParams represents pagination request parameters
type PaginationParams struct {
	Page     int `form:"page,default=1"`
	PageSize int `form:"page_size,default=20"`
}

// PaginationResponse represents pagination response metadata
type PaginationResponse struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"page_size"`
	TotalItems int64 `json:"total_items"`
	TotalPages int   `json:"total_pages"`
}
