package service

type AIService struct {
	// placeholder for provider client
}

func NewAIService() *AIService { return &AIService{} }

func (s *AIService) Ask(prompt string) (string, error) {
	// stub
	return "stub-response", nil
}
