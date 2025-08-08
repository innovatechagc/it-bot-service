package services

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/company/bot-service/internal/domain"
	"github.com/company/bot-service/pkg/logger"
)

// ConditionalService define las operaciones para manejar condiciones
type ConditionalService interface {
	GetConditional(ctx context.Context, id string) (*domain.Conditional, error)
	GetConditionalsByBot(ctx context.Context, botID string) ([]*domain.Conditional, error)
	CreateConditional(ctx context.Context, conditional *domain.Conditional) error
	UpdateConditional(ctx context.Context, conditional *domain.Conditional) error
	DeleteConditional(ctx context.Context, id string) error
	EvaluateConditional(ctx context.Context, id string, input map[string]interface{}) (bool, error)
	EvaluateExpression(ctx context.Context, expression string, input map[string]interface{}) (bool, error)
}

// TriggerService define las operaciones para manejar triggers
type TriggerService interface {
	GetTrigger(ctx context.Context, id string) (*domain.Trigger, error)
	GetTriggersByBot(ctx context.Context, botID string) ([]*domain.Trigger, error)
	GetTriggersByEvent(ctx context.Context, botID string, event domain.TriggerEvent) ([]*domain.Trigger, error)
	CreateTrigger(ctx context.Context, trigger *domain.Trigger) error
	UpdateTrigger(ctx context.Context, trigger *domain.Trigger) error
	DeleteTrigger(ctx context.Context, id string) error
	ExecuteTrigger(ctx context.Context, id string, eventData map[string]interface{}) error
	ProcessEvent(ctx context.Context, botID string, event domain.TriggerEvent, eventData map[string]interface{}) error
}

// conditionalService implementa ConditionalService
type conditionalService struct {
	conditionalRepo domain.ConditionalRepository
	logger          logger.Logger
}

// triggerService implementa TriggerService
type triggerService struct {
	triggerRepo domain.TriggerRepository
	conditionalSvc ConditionalService
	logger       logger.Logger
}

// NewConditionalService crea una nueva instancia de ConditionalService
func NewConditionalService(
	conditionalRepo domain.ConditionalRepository,
	logger logger.Logger,
) ConditionalService {
	return &conditionalService{
		conditionalRepo: conditionalRepo,
		logger:          logger,
	}
}

// NewTriggerService crea una nueva instancia de TriggerService
func NewTriggerService(
	triggerRepo domain.TriggerRepository,
	conditionalSvc ConditionalService,
	logger logger.Logger,
) TriggerService {
	return &triggerService{
		triggerRepo:     triggerRepo,
		conditionalSvc:  conditionalSvc,
		logger:          logger,
	}
}

// Implementación de ConditionalService
func (s *conditionalService) GetConditional(ctx context.Context, id string) (*domain.Conditional, error) {
	return s.conditionalRepo.GetByID(ctx, id)
}

func (s *conditionalService) GetConditionalsByBot(ctx context.Context, botID string) ([]*domain.Conditional, error) {
	return s.conditionalRepo.GetByBotID(ctx, botID)
}

func (s *conditionalService) CreateConditional(ctx context.Context, conditional *domain.Conditional) error {
	if conditional.ID == "" {
		conditional.ID = uuid.New().String()
	}
	conditional.CreatedAt = time.Now()
	conditional.UpdatedAt = time.Now()
	
	return s.conditionalRepo.Create(ctx, conditional)
}

func (s *conditionalService) UpdateConditional(ctx context.Context, conditional *domain.Conditional) error {
	conditional.UpdatedAt = time.Now()
	return s.conditionalRepo.Update(ctx, conditional)
}

func (s *conditionalService) DeleteConditional(ctx context.Context, id string) error {
	return s.conditionalRepo.Delete(ctx, id)
}

func (s *conditionalService) EvaluateConditional(ctx context.Context, id string, input map[string]interface{}) (bool, error) {
	return s.conditionalRepo.Evaluate(ctx, id, input)
}

func (s *conditionalService) EvaluateExpression(ctx context.Context, expression string, input map[string]interface{}) (bool, error) {
	// Implementación de evaluación de expresiones
	return s.evaluateExpression(expression, input)
}

// evaluateExpression evalúa una expresión condicional
func (s *conditionalService) evaluateExpression(expression string, input map[string]interface{}) (bool, error) {
	// Implementación básica de evaluación de expresiones
	// Se puede expandir para soportar expresiones más complejas
	
	// Reemplazar variables con valores
	evaluatedExpr := s.replaceVariables(expression, input)
	
	// Evaluar expresiones simples
	if strings.Contains(evaluatedExpr, "==") {
		parts := strings.Split(evaluatedExpr, "==")
		if len(parts) == 2 {
			return strings.TrimSpace(parts[0]) == strings.TrimSpace(parts[1]), nil
		}
	}
	
	if strings.Contains(evaluatedExpr, "!=") {
		parts := strings.Split(evaluatedExpr, "!=")
		if len(parts) == 2 {
			return strings.TrimSpace(parts[0]) != strings.TrimSpace(parts[1]), nil
		}
	}
	
	if strings.Contains(evaluatedExpr, "contains") {
		// Formato: "text contains keyword"
		parts := strings.Split(evaluatedExpr, "contains")
		if len(parts) == 2 {
			text := strings.TrimSpace(parts[0])
			keyword := strings.TrimSpace(parts[1])
			return strings.Contains(strings.ToLower(text), strings.ToLower(keyword)), nil
		}
	}
	
	if strings.Contains(evaluatedExpr, "regex") {
		// Formato: "text regex pattern"
		parts := strings.Split(evaluatedExpr, "regex")
		if len(parts) == 2 {
			text := strings.TrimSpace(parts[0])
			pattern := strings.TrimSpace(parts[1])
			matched, err := regexp.MatchString(pattern, text)
			return matched, err
		}
	}
	
	// Evaluación booleana simple
	switch strings.ToLower(evaluatedExpr) {
	case "true", "1", "yes":
		return true, nil
	case "false", "0", "no":
		return false, nil
	default:
		// Si no coincide con ningún patrón, devolver false
		return false, nil
	}
}

// replaceVariables reemplaza variables en la expresión con valores del input
func (s *conditionalService) replaceVariables(expression string, input map[string]interface{}) string {
	result := expression
	for key, value := range input {
		placeholder := fmt.Sprintf("{{%s}}", key)
		replacement := fmt.Sprintf("%v", value)
		result = strings.ReplaceAll(result, placeholder, replacement)
	}
	return result
}

// Implementación de TriggerService
func (s *triggerService) GetTrigger(ctx context.Context, id string) (*domain.Trigger, error) {
	return s.triggerRepo.GetByID(ctx, id)
}

func (s *triggerService) GetTriggersByBot(ctx context.Context, botID string) ([]*domain.Trigger, error) {
	return s.triggerRepo.GetByBotID(ctx, botID)
}

func (s *triggerService) GetTriggersByEvent(ctx context.Context, botID string, event domain.TriggerEvent) ([]*domain.Trigger, error) {
	return s.triggerRepo.GetByEvent(ctx, botID, event)
}

func (s *triggerService) CreateTrigger(ctx context.Context, trigger *domain.Trigger) error {
	if trigger.ID == "" {
		trigger.ID = uuid.New().String()
	}
	trigger.CreatedAt = time.Now()
	trigger.UpdatedAt = time.Now()
	
	return s.triggerRepo.Create(ctx, trigger)
}

func (s *triggerService) UpdateTrigger(ctx context.Context, trigger *domain.Trigger) error {
	trigger.UpdatedAt = time.Now()
	return s.triggerRepo.Update(ctx, trigger)
}

func (s *triggerService) DeleteTrigger(ctx context.Context, id string) error {
	return s.triggerRepo.Delete(ctx, id)
}

func (s *triggerService) ExecuteTrigger(ctx context.Context, id string, eventData map[string]interface{}) error {
	return s.triggerRepo.Execute(ctx, id, eventData)
}

func (s *triggerService) ProcessEvent(ctx context.Context, botID string, event domain.TriggerEvent, eventData map[string]interface{}) error {
	// Obtener triggers habilitados para el evento
	triggers, err := s.triggerRepo.GetEnabledByBotID(ctx, botID)
	if err != nil {
		return fmt.Errorf("failed to get triggers: %w", err)
	}
	
	// Filtrar triggers por evento
	var matchingTriggers []*domain.Trigger
	for _, trigger := range triggers {
		if trigger.Event == event {
			matchingTriggers = append(matchingTriggers, trigger)
		}
	}
	
	// Ejecutar triggers en orden de prioridad
	for _, trigger := range matchingTriggers {
		// Evaluar condición si existe
		if trigger.Condition != "" {
			conditionMet, err := s.conditionalSvc.EvaluateConditional(ctx, trigger.Condition, eventData)
			if err != nil {
				s.logger.Error("Failed to evaluate trigger condition", "trigger_id", trigger.ID, "error", err)
				continue
			}
			
			if !conditionMet {
				continue
			}
		}
		
		// Ejecutar trigger
		if err := s.ExecuteTrigger(ctx, trigger.ID, eventData); err != nil {
			s.logger.Error("Failed to execute trigger", "trigger_id", trigger.ID, "error", err)
		}
	}
	
	return nil
} 