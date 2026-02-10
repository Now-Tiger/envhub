package utils

type ErrorResponse struct {
	Success    bool   `json:"success"`
	StatusCode uint16 `json:"status"`
	Message    string `json:"message"`
}
