package services

import (
	"context"
	"fmt"
	"time"

	"github.com/company/bot-service/internal/domain"
	"github.com/company/bot-service/pkg/logger"
)

type conversationService struct {
	sessionRepo domain.ConversationSessionRepository
	logger      logger.Logger
}

func NewConversationService(
	sessionRepo domain.ConversationSessionRepository,
	logger logger.Logger,
) ConversationService {
	return &conversationService{
		sessionRepo: sessionRepo,
		logger:      logger,
	}
}

func (s *conversationService) GetSession(ctx context.Context, userID, botID string) (*domain.ConversationSession, error) {
	session, err := s.sessionRepo.GetByUserAndBot(ctx, userID, botID)
	if err != nil {
		return nil, err
	}

	// Verificar si la sesión ha expirado
	if session.ExpiresAt.Before(time.Now()) {
		// Eliminar sesión expirada
		if err := s.sessionRepo.Delete(ctx, session.ID); err != nil {
			s.logger.Error("Failed to delete expired session", "session_id", session.ID, "error", err)
		}
		return nil, fmt.Errorf("session expired")
	}

	return session, nil
}

func (s *conversationService) CreateSession(ctx context.Context, session *domain.ConversationSession) error {
	session.CreatedAt = time.Now()
	session.UpdatedAt = time.Now()
	if session.ExpiresAt.IsZero() {
		session.ExpiresAt = time.Now().Add(24 * time.Hour) // Default 24 hours
	}
	return s.sessionRepo.Create(ctx, session)
}

func (s *conversationService) UpdateSession(ctx context.Context, session *domain.ConversationSession) error {
	session.UpdatedAt = time.Now()
	// Extender expiración en cada actualización
	session.ExpiresAt = time.Now().Add(24 * time.Hour)
	return s.sessionRepo.Update(ctx, session)
}

func (s *conversationService) DeleteSession(ctx context.Context, id string) error {
	return s.sessionRepo.Delete(ctx, id)
}

func (s *conversationService) CleanupExpiredSessions(ctx context.Context) error {
	err := s.sessionRepo.DeleteExpired(ctx)
	if err != nil {
		s.logger.Error("Failed to cleanup expired sessions", "error", err)
		return err
	}

	s.logger.Info("Expired sessions cleaned up successfully")
	return nil
}