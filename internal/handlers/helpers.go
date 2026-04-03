package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"github.com/Now-Tiger/envhub/pkg/types"
)

// Global validator instance
var validate = validator.New()

// decodeJSON decodes JSON request body
func decodeJSON(r *http.Request, dest any) error {
	return json.NewDecoder(r.Body).Decode(dest)
}

// respondJSON writes JSON response
func respondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			fmt.Printf("error encoding response: %v\n", err)
		}
	}
}

// respondError writes error response
func respondError(w http.ResponseWriter, status int, message, details string) {
	errResp := types.ErrorResponse{
		Error: message,
	}
	if details != "" {
		errResp.Code = "DETAILS"
		errResp.Details = details
	}
	respondJSON(w, status, errResp)
}

// respondValidationError writes validation error response
func respondValidationError(w http.ResponseWriter, err error) {
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		respondError(w, http.StatusBadRequest, "validation error", err.Error())
		return
	}

	var errors []types.ValidationError
	for _, e := range validationErrors {
		errors = append(errors, types.ValidationError{
			Field:   e.Field(),
			Message: formatValidationError(e),
		})
	}

	resp := types.ValidationErrorResponse{
		Error:   "validation failed",
		Code:    "VALIDATION_ERROR",
		Details: errors,
	}
	respondJSON(w, http.StatusUnprocessableEntity, resp)
}

// formatValidationError formats validation error message
func formatValidationError(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "this field is required"
	case "max":
		return fmt.Sprintf("exceeds maximum length of %s", e.Param())
	case "min":
		return fmt.Sprintf("must be at least %s characters", e.Param())
	case "uuid":
		return "must be a valid UUID"
	case "hexcolor":
		return "must be a valid hex color"
	case "oneof":
		return fmt.Sprintf("must be one of: %s", e.Param())
	default:
		return fmt.Sprintf("invalid value for %s", e.Tag())
	}
}

// getIntQuery gets integer query parameter with default
func getIntQuery(r *http.Request, key string, defaultVal int) int {
	val := r.URL.Query().Get(key)
	if val == "" {
		return defaultVal
	}
	intVal, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return intVal
}

// getStringQuery gets string query parameter
func getStringQuery(r *http.Request, key string) string {
	return r.URL.Query().Get(key)
}

// getUUIDParam gets UUID from URL path parameter
func getUUIDParam(r *http.Request, param string) (uuid.UUID, error) {
	return uuid.Parse(r.URL.Query().Get(param))
}

// parseUUID parses UUID from string
func parseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

// writeSuccess writes success response
func writeSuccess(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"message": message})
}
