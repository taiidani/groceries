package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// ErrorResponse is the standard error body returned by all API endpoints.
type ErrorResponse struct {
	Error string `json:"error"`
}

// writeJSON serialises v as JSON and writes it to w with the given status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to encode JSON response", "error", err)
	}
}

// errorJSON writes a standard JSON error response.
func errorJSON(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, ErrorResponse{Error: msg})
}

// renderError is an alias for errorJSON used by catch-all route handlers that
// do not have access to the request object.
func renderError(w http.ResponseWriter, status int, err error) {
	errorJSON(w, status, err.Error())
}

// badRequest writes a 400 JSON error response.
func badRequest(w http.ResponseWriter, msg string) {
	errorJSON(w, http.StatusBadRequest, msg)
}

// notFound writes a 404 JSON error response.
func notFound(w http.ResponseWriter, resource string) {
	errorJSON(w, http.StatusNotFound, resource+" not found")
}

// conflict writes a 409 JSON error response.
func conflict(w http.ResponseWriter, msg string) {
	errorJSON(w, http.StatusConflict, msg)
}

// internalError logs err and writes a 500 JSON error response. The raw error
// is intentionally not forwarded to the client.
func internalError(w http.ResponseWriter, err error) {
	slog.Error("internal API error", "error", err)
	errorJSON(w, http.StatusInternalServerError, "an unexpected error occurred")
}
