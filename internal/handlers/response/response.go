package response

import (
	"encoding/json"
	"net/http"
)

type APIResponse struct {
	Message    string      `json:"message"`
	Data       interface{} `json:"data,omitempty"`
	StatusCode int         `json:"status_code"`
}

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
	respondWithStatus(w, r, apiErr, statusCode)
}

func RespondWithSuccess(w http.ResponseWriter, r *http.Request, msg string, data interface{}, statusCode int) {
	apiResp := &APIResponse{
		Message:    msg,
		Data:       data,
		StatusCode: statusCode,
	}
	respondWithStatus(w, r, apiResp, statusCode)
}

func respondWithStatus(w http.ResponseWriter, r *http.Request, data interface{}, statusCode int) {
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
