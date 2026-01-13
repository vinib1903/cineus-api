package httputil

import (
	"encoding/json"
	"net/http"
)

// Response é a estrutura padrão de resposta da API.
type Response struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo contém informações sobre um erro.
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// JSON envia uma resposta JSON com o status code especificado.
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := Response{
		Success: status >= 200 && status < 300,
		Data:    data,
	}

	json.NewEncoder(w).Encode(response)
}

// Error envia uma resposta de erro JSON.
func Error(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	response := Response{
		Success: false,
		Error: &ErrorInfo{
			Code:    code,
			Message: message,
		},
	}

	json.NewEncoder(w).Encode(response)
}

// Common error responses

// BadRequest envia um erro 400.
func BadRequest(w http.ResponseWriter, message string) {
	Error(w, http.StatusBadRequest, "BAD_REQUEST", message)
}

// Unauthorized envia um erro 401.
func Unauthorized(w http.ResponseWriter, message string) {
	Error(w, http.StatusUnauthorized, "UNAUTHORIZED", message)
}

// Forbidden envia um erro 403.
func Forbidden(w http.ResponseWriter, message string) {
	Error(w, http.StatusForbidden, "FORBIDDEN", message)
}

// NotFound envia um erro 404.
func NotFound(w http.ResponseWriter, message string) {
	Error(w, http.StatusNotFound, "NOT_FOUND", message)
}

// Conflict envia um erro 409.
func Conflict(w http.ResponseWriter, message string) {
	Error(w, http.StatusConflict, "CONFLICT", message)
}

// InternalServerError envia um erro 500.
func InternalServerError(w http.ResponseWriter, message string) {
	Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", message)
}
