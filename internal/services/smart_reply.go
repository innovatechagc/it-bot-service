package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/company/bot-service/internal/ai"
	"github.com/company/bot-service/internal/domain"
	"github.com/company/bot-service/internal/mcp"
	"github.com/company/bot-service/pkg/logger"
)

type smartReplyService struct {
	smartReplyRepo  domain.SmartReplyRepository
	aiClient        ai.AIClient
	mcpOrchestrator interface {
		mcp.MCPOrchestrator
		mcp.MCPDomainOrchestrator
	}
	logger          logger.Logger
}

func NewSmartReplyService(
	smartReplyRepo domain.SmartReplyRepository,
	aiClient ai.AIClient,
	mcpOrchestrator interface {
		mcp.MCPOrchestrator
		mcp.MCPDomainOrchestrator
	},
	logger logger.Logger,
) SmartReplyService {
	return &smartReplyService{
		smartReplyRepo:  smartReplyRepo,
		aiClient:        aiClient,
		mcpOrchestrator: mcpOrchestrator,
		logger:          logger,
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

	// Intentar usar MCP primero, fallback a AI client si falla
	smartReply, err := s.generateWithMCP(ctx, botID, fullPrompt, context)
	if err != nil {
		s.logger.Warn("MCP generation failed, falling back to AI client", "error", err)
		return s.generateWithAIClient(ctx, botID, fullPrompt, context)
	}

	return smartReply, nil
}

func (s *smartReplyService) generateWithMCP(ctx context.Context, botID, prompt string, context map[string]interface{}) (*domain.SmartReply, error) {
	// Crear tarea MCP para generación de texto
	task := &domain.MCPTask{
		ID:          fmt.Sprintf("smart-reply-%s-%d", botID, time.Now().UnixNano()),
		Type:        "text_generation",
		Description: "Generate smart reply for bot conversation",
		Input: map[string]interface{}{
			"prompt":      prompt,
			"temperature": 0.7,
			"max_tokens":  500,
			"system":      "You are a helpful customer service assistant. Provide clear, concise, and helpful responses.",
		},
		Priority:  5,
		Timeout:   30000, // 30 segundos
		Context:   context,
		Metadata: map[string]interface{}{
			"bot_id":     botID,
			"source":     "smart_reply_service",
			"task_type":  "conversation",
		},
		CreatedAt: time.Now(),
	}

	// Ejecutar tarea usando MCP
	result, err := s.mcpOrchestrator.ExecuteTaskDomain(ctx, task)
	if err != nil {
		return nil, fmt.Errorf("MCP task execution failed: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("MCP task failed: %s", result.Error)
	}

	// Extraer respuesta del resultado
	var responseText string
	var tokensUsed int
	var finishReason string

	if text, exists := result.Output["text"]; exists {
		if textStr, ok := text.(string); ok {
			responseText = textStr
		}
	}

	if tokens, exists := result.Output["tokens_used"]; exists {
		if tokensFloat, ok := tokens.(float64); ok {
			tokensUsed = int(tokensFloat)
		}
	}

	if reason, exists := result.Output["finish_reason"]; exists {
		if reasonStr, ok := reason.(string); ok {
			finishReason = reasonStr
		}
	}

	if responseText == "" {
		return nil, fmt.Errorf("no text response from MCP agent")
	}

	// Determinar intent basado en el prompt original
	intent := s.extractIntent(prompt)

	// Calcular confianza basada en el resultado MCP
	confidence := s.calculateMCPConfidence(result, finishReason, responseText)

	smartReply := &domain.SmartReply{
		BotID:      botID,
		Intent:     intent,
		Response:   responseText,
		Confidence: confidence,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	s.logger.Info("AI response generated via MCP", 
		"bot_id", botID,
		"intent", intent,
		"confidence", smartReply.Confidence,
		"tokens_used", tokensUsed,
		"agent_id", result.AgentID,
		"execution_time", result.ExecutionTime)

	return smartReply, nil
}

func (s *smartReplyService) generateWithAIClient(ctx context.Context, botID, prompt string, context map[string]interface{}) (*domain.SmartReply, error) {
	// Generar respuesta usando el cliente AI original como fallback
	response, err := s.aiClient.GenerateResponse(ctx, prompt, 
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

	s.logger.Info("AI response generated via fallback client", 
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

func (s *smartReplyService) calculateMCPConfidence(result *domain.MCPTaskResult, finishReason, responseText string) float64 {
	// Cálculo de confianza basado en el resultado MCP
	confidence := 0.7 // Base confidence
	
	// Ajustar basado en el éxito de la ejecución
	if result.Success {
		confidence += 0.1
	} else {
		confidence -= 0.3
	}
	
	// Ajustar basado en finish_reason si está disponible
	switch finishReason {
	case "stop":
		confidence += 0.2
	case "length":
		confidence += 0.1
	default:
		// No ajustar si no hay finish_reason o es desconocido
	}

	// Ajustar basado en tiempo de ejecución (respuestas más rápidas pueden ser menos confiables)
	if result.ExecutionTime < 100 { // menos de 100ms
		confidence -= 0.1
	} else if result.ExecutionTime > 5000 { // más de 5 segundos
		confidence -= 0.05
	}

	// Ajustar basado en longitud de respuesta
	if len(responseText) > 50 {
		confidence += 0.1
	}
	if len(responseText) < 10 {
		confidence -= 0.2
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