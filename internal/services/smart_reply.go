package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/company/bot-service/internal/ai"
	"github.com/company/bot-service/internal/domain"
	"github.com/company/bot-service/pkg/logger"
)

type smartReplyService struct {
	smartReplyRepo domain.SmartReplyRepository
	aiClient       ai.AIClient
	logger         logger.Logger
}

func NewSmartReplyService(
	smartReplyRepo domain.SmartReplyRepository,
	aiClient ai.AIClient,
	logger logger.Logger,
) SmartReplyService {
	return &smartReplyService{
		smartReplyRepo: smartReplyRepo,
		aiClient:       aiClient,
		logger:         logger,
	}
}

func (s *smartReplyService) GetSmartReply(ctx context.Context, id string) (*domain.SmartReply, error) {
	return s.smartReplyRepo.GetByID(ctx, id)
}

func (s *smartReplyService) GetSmartRepliesByBot(ctx context.Context, botID string) ([]*domain.SmartReply, error) {
	return s.smartReplyRepo.GetByBotID(ctx, botID)
}

func (s *smartReplyService) CreateSmartReply(ctx context.Context, reply *domain.SmartReply) error {
	reply.CreatedAt = time.Now()
	reply.UpdatedAt = time.Now()
	return s.smartReplyRepo.Create(ctx, reply)
}

func (s *smartReplyService) UpdateSmartReply(ctx context.Context, reply *domain.SmartReply) error {
	reply.UpdatedAt = time.Now()
	return s.smartReplyRepo.Update(ctx, reply)
}

func (s *smartReplyService) DeleteSmartReply(ctx context.Context, id string) error {
	return s.smartReplyRepo.Delete(ctx, id)
}

func (s *smartReplyService) GenerateAIResponse(ctx context.Context, botID, prompt string, context map[string]interface{}) (*domain.SmartReply, error) {
	// Construir prompt con contexto
	fullPrompt := s.buildPromptWithContext(prompt, context)

	// Generar respuesta usando IA
	response, err := s.aiClient.GenerateResponse(ctx, fullPrompt, 
		ai.WithMaxTokens(500),
		ai.WithTemperature(0.7),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate AI response: %w", err)
	}

	// Determinar intent basado en el prompt
	intent := s.extractIntent(prompt)

	smartReply := &domain.SmartReply{
		BotID:      botID,
		Intent:     intent,
		Response:   response.Content,
		Confidence: s.calculateConfidence(response),
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	s.logger.Info("AI response generated", 
		"bot_id", botID,
		"intent", intent,
		"confidence", smartReply.Confidence,
		"tokens_used", response.TokensUsed)

	return smartReply, nil
}

func (s *smartReplyService) TrainIntents(ctx context.Context, botID string, intents []domain.SmartReply) error {
	// Guardar intents entrenados
	for _, intent := range intents {
		intent.BotID = botID
		intent.CreatedAt = time.Now()
		intent.UpdatedAt = time.Now()
		
		if err := s.smartReplyRepo.Create(ctx, &intent); err != nil {
			s.logger.Error("Failed to save trained intent", 
				"bot_id", botID,
				"intent", intent.Intent,
				"error", err)
			return fmt.Errorf("failed to save intent %s: %w", intent.Intent, err)
		}
	}

	s.logger.Info("Intents trained successfully", 
		"bot_id", botID,
		"count", len(intents))

	return nil
}

func (s *smartReplyService) buildPromptWithContext(prompt string, context map[string]interface{}) string {
	var contextStr strings.Builder
	
	contextStr.WriteString("Context:\n")
	for key, value := range context {
		contextStr.WriteString(fmt.Sprintf("- %s: %v\n", key, value))
	}
	
	contextStr.WriteString("\nUser message: ")
	contextStr.WriteString(prompt)
	contextStr.WriteString("\n\nPlease provide a helpful and contextually appropriate response.")

	return contextStr.String()
}

func (s *smartReplyService) extractIntent(prompt string) string {
	prompt = strings.ToLower(prompt)
	
	// Mapeo simple de palabras clave a intents
	intentKeywords := map[string][]string{
		"greeting":     {"hello", "hi", "hey", "good morning", "good afternoon"},
		"goodbye":      {"bye", "goodbye", "see you", "farewell"},
		"help":         {"help", "assist", "support", "how to"},
		"information":  {"what", "how", "when", "where", "why", "info"},
		"complaint":    {"problem", "issue", "wrong", "error", "complaint"},
		"compliment":   {"good", "great", "excellent", "amazing", "wonderful"},
		"question":     {"?", "question", "ask"},
	}

	for intent, keywords := range intentKeywords {
		for _, keyword := range keywords {
			if strings.Contains(prompt, keyword) {
				return intent
			}
		}
	}

	return "general"
}

func (s *smartReplyService) calculateConfidence(response *ai.Response) float64 {
	// Cálculo simple de confianza basado en la respuesta
	confidence := 0.7 // Base confidence
	
	// Ajustar basado en finish_reason
	switch response.FinishReason {
	case "stop":
		confidence += 0.2
	case "length":
		confidence += 0.1
	default:
		confidence -= 0.1
	}

	// Ajustar basado en longitud de respuesta
	if len(response.Content) > 50 {
		confidence += 0.1
	}

	// Asegurar que esté en rango [0, 1]
	if confidence > 1.0 {
		confidence = 1.0
	}
	if confidence < 0.0 {
		confidence = 0.0
	}

	return confidence
}