package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"geminiBackend/pkg/logger"
)

// OllamaMessage представляет сообщение для Ollama API
type OllamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OllamaRequest представляет запрос к Ollama API
type OllamaRequest struct {
	Model    string          `json:"model"`
	Messages []OllamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
	Options  map[string]any  `json:"options,omitempty"`
}

// OllamaResponse представляет ответ от Ollama API
type OllamaResponse struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
	Done bool `json:"done"`
}

// LocalLLMClient представляет клиент для локальной LLM через Ollama
type LocalLLMClient struct {
	endpoint   string
	model      string
	maxChars   int
	httpClient *http.Client
}

// NewLocalLLMClient создает новый клиент для локальной LLM
func NewLocalLLMClient(endpoint, model string, maxChars int) *LocalLLMClient {
	// Если модель не указана, используем дефолтную
	if model == "" {
		model = "qwen2:1.5b"
	}
	return &LocalLLMClient{
		endpoint: endpoint,
		model:    model,
		maxChars: maxChars,
		httpClient: &http.Client{
			Timeout: 120 * time.Second, // увеличенный таймаут для CPU-инференса
		},
	}
}

// GenerateText генерирует текст через локальную LLM
func (c *LocalLLMClient) GenerateText(prompt string) (string, error) {
	// Проверка лимита на вход
	if len(prompt) > c.maxChars {
		return "", fmt.Errorf("prompt too long: %d chars (max %d)", len(prompt), c.maxChars)
	}

	// Системный промпт для OCR-коррекции
	systemPrompt := "Ты корректируешь текст после OCR-распознавания. Исправляй опечатки и ошибки распознавания, сохраняй форматирование и абзацы. Не добавляй новое содержание, только исправляй существующий текст."

	req := OllamaRequest{
		Model:  c.model,
		Stream: false,
		Messages: []OllamaMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: prompt},
		},
		Options: map[string]any{
			"temperature": 0.1,  // низкая температура для детерминированности
			"num_predict": 4096, // макс токенов ответа
		},
	}

	body, err := json.Marshal(req)
	if err != nil {
		logger.L.Error("failed to marshal local LLM request", "error", err.Error())
		return "", err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.endpoint+"/api/chat", bytes.NewReader(body))
	if err != nil {
		logger.L.Error("failed to create local LLM HTTP request", "error", err.Error())
		return "", err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		logger.L.Error("failed to call local LLM", "error", err.Error(), "endpoint", c.endpoint)
		return "", fmt.Errorf("local LLM unavailable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		logger.L.Error("local LLM returned error", "status", resp.StatusCode, "body", string(bodyBytes))
		return "", fmt.Errorf("local LLM error: status %d", resp.StatusCode)
	}

	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		logger.L.Error("failed to decode local LLM response", "error", err.Error())
		return "", err
	}

	return strings.TrimSpace(ollamaResp.Message.Content), nil
}

// GenerateTextChunked обрабатывает длинный текст по частям
func (c *LocalLLMClient) GenerateTextChunked(prompt string, chunkSize int) (string, error) {
	if chunkSize <= 0 {
		chunkSize = c.maxChars
	}

	// Если текст меньше лимита, обрабатываем целиком
	if len(prompt) <= chunkSize {
		return c.GenerateText(prompt)
	}

	// Разбиваем на чанки с учетом границ слов
	chunks := splitTextIntoChunks(prompt, chunkSize)
	logger.L.Info("processing text in chunks", "total_chars", len(prompt), "chunks", len(chunks))

	results := make([]string, len(chunks))
	for i, chunk := range chunks {
		result, err := c.GenerateText(chunk)
		if err != nil {
			logger.L.Error("failed to process chunk", "chunk_index", i, "error", err.Error())
			return "", fmt.Errorf("chunk %d failed: %w", i, err)
		}
		results[i] = result
		logger.L.Debug("processed chunk", "chunk_index", i, "input_len", len(chunk), "output_len", len(result))
	}

	return strings.Join(results, "\n\n"), nil
}

// splitTextIntoChunks разбивает текст на чанки по границам предложений/абзацев
func splitTextIntoChunks(text string, maxChars int) []string {
	if len(text) <= maxChars {
		return []string{text}
	}

	var chunks []string
	paragraphs := strings.Split(text, "\n\n")

	currentChunk := ""
	for _, para := range paragraphs {
		// Если параграф сам больше лимита, режем по предложениям
		if len(para) > maxChars {
			sentences := strings.Split(para, ". ")
			for _, sent := range sentences {
				if len(currentChunk)+len(sent)+2 > maxChars && currentChunk != "" {
					chunks = append(chunks, currentChunk)
					currentChunk = sent
				} else {
					if currentChunk != "" {
						currentChunk += ". "
					}
					currentChunk += sent
				}
			}
		} else {
			// Параграф помещается целиком
			if len(currentChunk)+len(para)+2 > maxChars && currentChunk != "" {
				chunks = append(chunks, currentChunk)
				currentChunk = para
			} else {
				if currentChunk != "" {
					currentChunk += "\n\n"
				}
				currentChunk += para
			}
		}
	}

	if currentChunk != "" {
		chunks = append(chunks, currentChunk)
	}

	return chunks
}
