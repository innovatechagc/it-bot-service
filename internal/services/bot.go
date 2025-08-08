package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/company/bot-service/internal/domain"
	"github.com/company/bot-service/internal/mcp"
	"github.com/company/bot-service/pkg/logger"
)

// BotService define las operaciones de negocio para bots
type BotService interface {
	GetBot(ctx context.Context, id string) (*domain.Bot, error)
	GetBotsByOwner(ctx context.Context, ownerID string) ([]*domain.Bot, error)
	CreateBot(ctx context.Context, bot *domain.Bot) error
	UpdateBot(ctx context.Context, bot *domain.Bot) error
	DeleteBot(ctx context.Context, id string) error
	ProcessIncomingMessage(ctx context.Context, message *domain.IncomingMessage) (*domain.BotResponse, error)
}

// BotFlowService define las operaciones de negocio para flujos de bot
type BotFlowService interface {
	GetFlow(ctx context.Context, id string) (*domain.BotFlow, error)
	GetFlowsByBot(ctx context.Context, botID string) ([]*domain.BotFlow, error)
	CreateFlow(ctx context.Context, flow *domain.BotFlow) error
	UpdateFlow(ctx context.Context, flow *domain.BotFlow) error
	DeleteFlow(ctx context.Context, id string) error
}

// BotStepService define las operaciones de negocio para pasos de flujo
type BotStepService interface {
	GetStep(ctx context.Context, id string) (*domain.BotStep, error)
	GetStepsByFlow(ctx context.Context, flowID string) ([]*domain.BotStep, error)
	CreateStep(ctx context.Context, step *domain.BotStep) error
	UpdateStep(ctx context.Context, step *domain.BotStep) error
	DeleteStep(ctx context.Context, id string) error
}

// SmartReplyService define las operaciones de negocio para respuestas inteligentes
type SmartReplyService interface {
	GetSmartReply(ctx context.Context, id string) (*domain.SmartReply, error)
	GetSmartRepliesByBot(ctx context.Context, botID string) ([]*domain.SmartReply, error)
	CreateSmartReply(ctx context.Context, reply *domain.SmartReply) error
	UpdateSmartReply(ctx context.Context, reply *domain.SmartReply) error
	DeleteSmartReply(ctx context.Context, id string) error
	GenerateAIResponse(ctx context.Context, botID, prompt string, context map[string]interface{}) (*domain.SmartReply, error)
	TrainIntents(ctx context.Context, botID string, intents []domain.SmartReply) error
}

// ConversationService define las operaciones para manejo de conversaciones
type ConversationService interface {
	GetSession(ctx context.Context, userID, botID string) (*domain.ConversationSession, error)
	CreateSession(ctx context.Context, session *domain.ConversationSession) error
	UpdateSession(ctx context.Context, session *domain.ConversationSession) error
	DeleteSession(ctx context.Context, id string) error
	CleanupExpiredSessions(ctx context.Context) error
}

// Implementaciones
type botService struct {
	botRepo         domain.BotRepository
	flowRepo        domain.BotFlowRepository
	stepRepo        domain.BotStepRepository
	sessionRepo     domain.ConversationSessionRepository
	smartReplyRepo  domain.SmartReplyRepository
	conversationSvc ConversationService
	smartReplySvc   SmartReplyService
	mcpOrchestrator interface {
		mcp.MCPOrchestrator
		mcp.MCPDomainOrchestrator
	}
	logger          logger.Logger
}

func NewBotService(
	botRepo domain.BotRepository,
	flowRepo domain.BotFlowRepository,
	stepRepo domain.BotStepRepository,
	sessionRepo domain.ConversationSessionRepository,
	smartReplyRepo domain.SmartReplyRepository,
	conversationSvc ConversationService,
	smartReplySvc SmartReplyService,
	mcpOrchestrator interface {
		mcp.MCPOrchestrator
		mcp.MCPDomainOrchestrator
	},
	logger logger.Logger,
) BotService {
	return &botService{
		botRepo:         botRepo,
		flowRepo:        flowRepo,
		stepRepo:        stepRepo,
		sessionRepo:     sessionRepo,
		smartReplyRepo:  smartReplyRepo,
		conversationSvc: conversationSvc,
		smartReplySvc:   smartReplySvc,
		mcpOrchestrator: mcpOrchestrator,
		logger:          logger,
	}
}

func (s *botService) GetBot(ctx context.Context, id string) (*domain.Bot, error) {
	return s.botRepo.GetByID(ctx, id)
}

func (s *botService) GetBotsByOwner(ctx context.Context, ownerID string) ([]*domain.Bot, error) {
	return s.botRepo.GetByOwnerID(ctx, ownerID)
}

func (s *botService) CreateBot(ctx context.Context, bot *domain.Bot) error {
	bot.CreatedAt = time.Now()
	bot.UpdatedAt = time.Now()
	return s.botRepo.Create(ctx, bot)
}

func (s *botService) UpdateBot(ctx context.Context, bot *domain.Bot) error {
	bot.UpdatedAt = time.Now()
	return s.botRepo.Update(ctx, bot)
}

func (s *botService) DeleteBot(ctx context.Context, id string) error {
	return s.botRepo.Delete(ctx, id)
}

func (s *botService) ProcessIncomingMessage(ctx context.Context, message *domain.IncomingMessage) (*domain.BotResponse, error) {
	// Obtener o crear sesión de conversación
	session, err := s.conversationSvc.GetSession(ctx, message.UserID, message.BotID)
	if err != nil {
		// Crear nueva sesión
		session = &domain.ConversationSession{
			BotID:     message.BotID,
			UserID:    message.UserID,
			Context:   make(map[string]interface{}),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
	}

	// Obtener bot
	bot, err := s.botRepo.GetByID(ctx, message.BotID)
	if err != nil {
		return nil, fmt.Errorf("bot not found: %w", err)
	}

	if bot.Status != domain.BotStatusActive {
		return &domain.BotResponse{
			Content: "Bot is currently unavailable",
			Type:    domain.ResponseTypeText,
		}, nil
	}

	// Determinar flujo a ejecutar
	var flow *domain.BotFlow
	if session.CurrentFlowID != "" {
		flow, err = s.flowRepo.GetByID(ctx, session.CurrentFlowID)
		if err != nil {
			s.logger.Warn("Current flow not found, using default", "flow_id", session.CurrentFlowID)
			flow = nil
		}
	}

	if flow == nil {
		// Buscar flujo por trigger o usar default
		flows, err := s.flowRepo.GetByBotID(ctx, message.BotID)
		if err != nil {
			return nil, fmt.Errorf("failed to get flows: %w", err)
		}

		for _, f := range flows {
			if f.Trigger != "" && f.Trigger == message.Content {
				flow = f
				break
			}
		}

		if flow == nil {
			// Usar flujo por defecto
			flow, err = s.flowRepo.GetDefaultByBotID(ctx, message.BotID)
			if err != nil {
				return nil, fmt.Errorf("no default flow found: %w", err)
			}
		}
	}

	// Ejecutar paso actual o inicial
	var currentStep *domain.BotStep
	if session.CurrentStepID != "" {
		currentStep, err = s.stepRepo.GetByID(ctx, session.CurrentStepID)
		if err != nil {
			s.logger.Warn("Current step not found, using entry point", "step_id", session.CurrentStepID)
			currentStep = nil
		}
	}

	if currentStep == nil {
		// Usar punto de entrada del flujo
		currentStep, err = s.stepRepo.GetByID(ctx, flow.EntryPoint)
		if err != nil {
			return nil, fmt.Errorf("entry point step not found: %w", err)
		}
	}

	// Procesar paso
	response, nextStepID, err := s.processStep(ctx, currentStep, message, session)
	if err != nil {
		return nil, fmt.Errorf("failed to process step: %w", err)
	}

	// Actualizar sesión
	session.CurrentFlowID = flow.ID
	session.CurrentStepID = ""
	if nextStepID != nil {
		session.CurrentStepID = *nextStepID
	}
	session.UpdatedAt = time.Now()
	session.Context["last_message"] = message.Content
	session.Context["last_response"] = response.Content

	if err := s.conversationSvc.UpdateSession(ctx, session); err != nil {
		s.logger.Error("Failed to update session", "error", err)
	}

	return response, nil
}

func (s *botService) processStep(ctx context.Context, step *domain.BotStep, message *domain.IncomingMessage, session *domain.ConversationSession) (*domain.BotResponse, *string, error) {
	switch step.Type {
	case domain.StepTypeMessage:
		return s.processMessageStep(ctx, step, message, session)
	case domain.StepTypeDecision:
		return s.processDecisionStep(ctx, step, message, session)
	case domain.StepTypeInput:
		return s.processInputStep(ctx, step, message, session)
	case domain.StepTypeAPICall:
		return s.processAPICallStep(ctx, step, message, session)
	case domain.StepTypeAI:
		return s.processAIStep(ctx, step, message, session)
	default:
		return &domain.BotResponse{
			Content: "Unknown step type",
			Type:    domain.ResponseTypeText,
		}, step.NextStepID, nil
	}
}

func (s *botService) processMessageStep(ctx context.Context, step *domain.BotStep, message *domain.IncomingMessage, session *domain.ConversationSession) (*domain.BotResponse, *string, error) {
	// Parsear contenido del paso
	var content struct {
		Text    string                   `json:"text"`
		Type    domain.ResponseType      `json:"type"`
		Options []domain.ResponseOption  `json:"options,omitempty"`
	}
	
	if err := json.Unmarshal(step.Content, &content); err != nil {
		return nil, nil, fmt.Errorf("failed to parse step content: %w", err)
	}

	response := &domain.BotResponse{
		Content:    content.Text,
		Type:       content.Type,
		Options:    content.Options,
		NextStepID: step.NextStepID,
	}

	return response, step.NextStepID, nil
}

func (s *botService) processDecisionStep(ctx context.Context, step *domain.BotStep, message *domain.IncomingMessage, session *domain.ConversationSession) (*domain.BotResponse, *string, error) {
	// Evaluar condiciones y determinar siguiente paso
	var conditions struct {
		Rules []struct {
			Condition string `json:"condition"`
			NextStep  string `json:"next_step"`
		} `json:"rules"`
		Default string `json:"default"`
	}

	if err := json.Unmarshal(step.Conditions, &conditions); err != nil {
		return nil, nil, fmt.Errorf("failed to parse conditions: %w", err)
	}

	// Evaluación simple de condiciones (se puede expandir)
	for _, rule := range conditions.Rules {
		if s.evaluateCondition(rule.Condition, message.Content, session.Context) {
			return &domain.BotResponse{
				Content: "Condition matched, proceeding...",
				Type:    domain.ResponseTypeText,
			}, &rule.NextStep, nil
		}
	}

	// Usar paso por defecto
	return &domain.BotResponse{
		Content: "Proceeding with default path...",
		Type:    domain.ResponseTypeText,
	}, &conditions.Default, nil
}

func (s *botService) processInputStep(ctx context.Context, step *domain.BotStep, message *domain.IncomingMessage, session *domain.ConversationSession) (*domain.BotResponse, *string, error) {
	// Guardar input del usuario en el contexto
	var content struct {
		Prompt   string `json:"prompt"`
		Variable string `json:"variable"`
	}

	if err := json.Unmarshal(step.Content, &content); err != nil {
		return nil, nil, fmt.Errorf("failed to parse step content: %w", err)
	}

	// Guardar respuesta del usuario
	session.Context[content.Variable] = message.Content

	response := &domain.BotResponse{
		Content: fmt.Sprintf("Thank you! I've saved your response: %s", message.Content),
		Type:    domain.ResponseTypeText,
	}

	return response, step.NextStepID, nil
}

func (s *botService) processAPICallStep(ctx context.Context, step *domain.BotStep, message *domain.IncomingMessage, session *domain.ConversationSession) (*domain.BotResponse, *string, error) {
	// Parsear configuración del paso API
	var content struct {
		AgentType string                 `json:"agent_type"`
		Config    map[string]interface{} `json:"config"`
		Task      map[string]interface{} `json:"task"`
	}
	
	if err := json.Unmarshal(step.Content, &content); err != nil {
		return nil, nil, fmt.Errorf("failed to parse API step content: %w", err)
	}

	// Configurar agente MCP si es necesario
	agentConfig := mcp.MCPConfig{
		Type:         content.AgentType,
		Name:         fmt.Sprintf("api-agent-%s", step.ID),
		Version:      "1.0",
		Config:       content.Config,
		Capabilities: []string{"http_request", "api_call"},
		Timeout:      30 * time.Second,
	}

	// Instanciar agente MCP
	agent, err := s.mcpOrchestrator.InstantiateMCP(ctx, agentConfig)
	if err != nil {
		s.logger.Error("Failed to instantiate MCP agent", "error", err)
		return &domain.BotResponse{
			Content: "Unable to process API request at this time",
			Type:    domain.ResponseTypeText,
		}, step.NextStepID, nil
	}

	// Pasar contexto al agente
	agentContext := make(map[string]interface{})
	for k, v := range session.Context {
		agentContext[k] = v
	}
	agentContext["user_message"] = message.Content
	agentContext["user_id"] = message.UserID
	agentContext["bot_id"] = message.BotID

	if err := s.mcpOrchestrator.PassContext(ctx, agent.GetID(), agentContext); err != nil {
		s.logger.Error("Failed to pass context to agent", "error", err)
	}

	// Crear tarea para el agente
	task := mcp.Task{
		ID:          fmt.Sprintf("task-%s-%d", step.ID, time.Now().UnixNano()),
		Type:        content.AgentType,
		Description: fmt.Sprintf("API call for step %s", step.ID),
		Input:       content.Task,
		Priority:    5,
		Metadata: map[string]interface{}{
			"step_id":    step.ID,
			"message_id": message.ID,
		},
	}

	// Ejecutar tarea
	result, err := s.mcpOrchestrator.ExecuteTask(ctx, task)
	if err != nil {
		s.logger.Error("MCP task execution failed", "error", err)
		return &domain.BotResponse{
			Content: "API request failed. Please try again later.",
			Type:    domain.ResponseTypeText,
		}, step.NextStepID, nil
	}

	// Procesar resultado
	var responseContent string
	if result.Success {
		if output, exists := result.Output["response"]; exists {
			responseContent = fmt.Sprintf("API call successful: %v", output)
		} else {
			responseContent = "API call completed successfully"
		}
		
		// Guardar resultado en el contexto de la sesión
		session.Context["api_result"] = result.Output
	} else {
		responseContent = fmt.Sprintf("API call failed: %s", result.Error)
	}

	// Terminar agente después del uso
	if err := s.mcpOrchestrator.TerminateAgent(ctx, agent.GetID()); err != nil {
		s.logger.Error("Failed to terminate agent", "agent_id", agent.GetID(), "error", err)
	}

	response := &domain.BotResponse{
		Content: responseContent,
		Type:    domain.ResponseTypeText,
		Metadata: map[string]interface{}{
			"task_id":     result.TaskID,
			"duration":    result.Duration.String(),
			"agent_type":  content.AgentType,
		},
	}

	return response, step.NextStepID, nil
}

func (s *botService) processAIStep(ctx context.Context, step *domain.BotStep, message *domain.IncomingMessage, session *domain.ConversationSession) (*domain.BotResponse, *string, error) {
	// Generar respuesta usando IA
	smartReply, err := s.smartReplySvc.GenerateAIResponse(ctx, message.BotID, message.Content, session.Context)
	if err != nil {
		s.logger.Error("Failed to generate AI response", "error", err)
		return &domain.BotResponse{
			Content: "I'm having trouble understanding. Could you please rephrase?",
			Type:    domain.ResponseTypeText,
		}, step.NextStepID, nil
	}

	response := &domain.BotResponse{
		Content: smartReply.Response,
		Type:    domain.ResponseTypeText,
		Metadata: map[string]interface{}{
			"confidence": smartReply.Confidence,
			"intent":     smartReply.Intent,
		},
	}

	return response, step.NextStepID, nil
}

func (s *botService) evaluateCondition(condition, userInput string, context map[string]interface{}) bool {
	// Implementación simple de evaluación de condiciones
	// Se puede expandir para soportar expresiones más complejas
	switch condition {
	case "contains_yes":
		return contains(userInput, []string{"yes", "sí", "si", "ok", "okay"})
	case "contains_no":
		return contains(userInput, []string{"no", "nope", "not"})
	default:
		return userInput == condition
	}
}

func contains(text string, keywords []string) bool {
	text = strings.ToLower(text)
	for _, keyword := range keywords {
		if strings.Contains(text, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}