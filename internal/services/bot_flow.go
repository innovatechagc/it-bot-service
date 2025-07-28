package services

import (
	"context"
	"time"

	"github.com/company/bot-service/internal/domain"
	"github.com/company/bot-service/pkg/logger"
)

type botFlowService struct {
	flowRepo domain.BotFlowRepository
	stepRepo domain.BotStepRepository
	logger   logger.Logger
}

func NewBotFlowService(
	flowRepo domain.BotFlowRepository,
	stepRepo domain.BotStepRepository,
	logger logger.Logger,
) BotFlowService {
	return &botFlowService{
		flowRepo: flowRepo,
		stepRepo: stepRepo,
		logger:   logger,
	}
}

func (s *botFlowService) GetFlow(ctx context.Context, id string) (*domain.BotFlow, error) {
	return s.flowRepo.GetByID(ctx, id)
}

func (s *botFlowService) GetFlowsByBot(ctx context.Context, botID string) ([]*domain.BotFlow, error) {
	return s.flowRepo.GetByBotID(ctx, botID)
}

func (s *botFlowService) CreateFlow(ctx context.Context, flow *domain.BotFlow) error {
	flow.CreatedAt = time.Now()
	flow.UpdatedAt = time.Now()
	return s.flowRepo.Create(ctx, flow)
}

func (s *botFlowService) UpdateFlow(ctx context.Context, flow *domain.BotFlow) error {
	flow.UpdatedAt = time.Now()
	return s.flowRepo.Update(ctx, flow)
}

func (s *botFlowService) DeleteFlow(ctx context.Context, id string) error {
	// Eliminar todos los pasos del flujo primero
	steps, err := s.stepRepo.GetByFlowID(ctx, id)
	if err != nil {
		s.logger.Error("Failed to get steps for flow deletion", "flow_id", id, "error", err)
	} else {
		for _, step := range steps {
			if err := s.stepRepo.Delete(ctx, step.ID); err != nil {
				s.logger.Error("Failed to delete step", "step_id", step.ID, "error", err)
			}
		}
	}

	return s.flowRepo.Delete(ctx, id)
}