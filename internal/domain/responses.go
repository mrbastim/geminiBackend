package domain

// AITextRequest запрос на генерацию текста
type AITextRequest struct {
	Prompt string `json:"prompt"`
	Model  string `json:"model"`
}

// AITextResponse данные успешного ответа генерации текста
type AITextResponse struct {
	Text string `json:"text"`
}

// LoginSuccessResponse успешный ответ логина (обёртка стандартного формата)
type LoginSuccessResponse struct {
	Status string        `json:"status"`
	Data   LoginResponse `json:"data"`
}

// OptionsSuccessResponse успешный ответ options
type OptionsSuccessResponse struct {
	Status string            `json:"status"`
	Data   map[string]string `json:"data"`
}

// AITextSuccessResponse успешный ответ генерации текста (обёртка)
type AITextSuccessResponse struct {
	Status string         `json:"status"`
	Data   AITextResponse `json:"data"`
}

// ErrorDetails подробности ошибки
type ErrorDetails struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorResponse стандартный ошибочный ответ
type ErrorResponse struct {
	Status string       `json:"status"`
	Error  ErrorDetails `json:"error"`
}
