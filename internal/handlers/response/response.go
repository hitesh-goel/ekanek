package response

import (
	"encoding/json"
	"net/http"
)

// APIError defines the structure of an error response from our APIs.
type APIError struct {
	Message    string `json:"message"`
	StatusCode int    `json:"code"`
}

func RespondWithError(w http.ResponseWriter, r *http.Request, msg string, statusCode int) {
	apiErr := &APIError{
		Message:    msg,
		StatusCode: statusCode,
	}
	RespondWithStatus(w, r, apiErr, statusCode)
}

func RespondWithStatus(w http.ResponseWriter, r *http.Request, data interface{}, statusCode int) {
	b, err := json.Marshal(data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, _ = w.Write(b)
	return
}