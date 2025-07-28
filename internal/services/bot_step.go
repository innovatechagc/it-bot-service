package services

import (
	"context"
	"time"

	"github.com/company/bot-service/internal/domain"
	"github.com/company/bot-service/pkg/logger"
)

type botStepService struct {
	stepRepo domain.BotStepRepository
	logger   logger.Logger
}

func NewBotStepService(
	stepRepo domain.BotStepRepository,
	logger logger.Logger,
) BotStepService {
	return &botStepService{
		stepRepo: stepRepo,
		logger:   logger,
	}
}

func (s *botStepService) GetStep(ctx context.Context, id string) (*domain.BotStep, error) {
	return s.stepRepo.GetByID(ctx, id)
}

func (s *botStepService) GetStepsByFlow(ctx context.Context, flowID string) ([]*domain.BotStep, error) {
	return s.stepRepo.GetByFlowID(ctx, flowID)
}

func (s *botStepService) CreateStep(ctx context.Context, step *domain.BotStep) error {
	step.CreatedAt = time.Now()
	step.UpdatedAt = time.Now()
	return s.stepRepo.Create(ctx, step)
}

func (s *botStepService) UpdateStep(ctx context.Context, step *domain.BotStep) error {
	step.UpdatedAt = time.Now()
	return s.stepRepo.Update(ctx, step)
}

func (s *botStepService) DeleteStep(ctx context.Context, id string) error {
	return s.stepRepo.Delete(ctx, id)
}