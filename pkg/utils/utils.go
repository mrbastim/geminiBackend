package utils

import (
	"encoding/json"
	"net/http"
)

type successEnvelope struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data"`
}

type errorEnvelope struct {
	Status string       `json:"status"`
	Error  errorPayload `json:"error"`
}

type errorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// RespondSuccess отправляет успешный ответ в стандартном формате.
func RespondSuccess(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(successEnvelope{Status: "success", Data: data})
}

// RespondError отправляет ошибку в стандартном формате.
func RespondError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(errorEnvelope{Status: "error", Error: errorPayload{Code: code, Message: message}})
}
