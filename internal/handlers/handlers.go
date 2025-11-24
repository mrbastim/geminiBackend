package handlers

import (
	"geminiBackend/config"
	"geminiBackend/pkg/utils"
	"net/http"
)

// Обработка получения опций от сервера
func GetOptions(w http.ResponseWriter, r *http.Request, config *config.Config) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Логика обработки GET-запроса и отправки опций клиенту
	payload := map[string]interface{}{
		"Port":   config.Port,
		"apiKey": config.ApiGemini,
	}
	utils.RespondWithJSON(w, http.StatusOK, payload)
}

// Получение ответа от Gemini API и отправка его клиенту
func PostResponse(w http.ResponseWriter, r *http.Request) {

}
