package utils

import (
	"encoding/json"
	"net/http"
)

// RespondWithJSON отправляет JSON-ответ с заданным статусом.
func RespondWithJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(payload)
}

// Success: стандартный формат успешного ответа
func Success(w http.ResponseWriter, data interface{}) {
	RespondWithJSON(w, http.StatusOK, map[string]interface{}{
		"status": "success",
		"data":   data,
	})
}

// Error: стандартный формат ошибочного ответа
func Error(w http.ResponseWriter, httpStatus int, code, message string) {
	RespondWithJSON(w, httpStatus, map[string]interface{}{
		"status": "error",
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}
