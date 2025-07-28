package repositories

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/company/bot-service/internal/domain"
)

// Mock implementations for development and testing

// MockBotRepository
type MockBotRepository struct {
	bots map[string]*domain.Bot
	mu   sync.RWMutex
}

func NewMockBotRepository() domain.BotRepository {
	return &MockBotRepository{
		bots: make(map[string]*domain.Bot),
	}
}

func (r *MockBotRepository) GetByID(ctx context.Context, id string) (*domain.Bot, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	bot, exists := r.bots[id]
	if !exists {
		return nil, fmt.Errorf("bot not found")
	}
	return bot, nil
}

func (r *MockBotRepository) GetByOwnerID(ctx context.Context, ownerID string) ([]*domain.Bot, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var bots []*domain.Bot
	for _, bot := range r.bots {
		if bot.OwnerID == ownerID {
			bots = append(bots, bot)
		}
	}
	return bots, nil
}

func (r *MockBotRepository) Create(ctx context.Context, bot *domain.Bot) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if bot.ID == "" {
		bot.ID = fmt.Sprintf("bot_%d", time.Now().UnixNano())
	}
	r.bots[bot.ID] = bot
	return nil
}

func (r *MockBotRepository) Update(ctx context.Context, bot *domain.Bot) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.bots[bot.ID]; !exists {
		return fmt.Errorf("bot not found")
	}
	r.bots[bot.ID] = bot
	return nil
}

func (r *MockBotRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	delete(r.bots, id)
	return nil
}

func (r *MockBotRepository) List(ctx context.Context, limit, offset int) ([]*domain.Bot, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var bots []*domain.Bot
	i := 0
	for _, bot := range r.bots {
		if i >= offset && len(bots) < limit {
			bots = append(bots, bot)
		}
		i++
	}
	return bots, nil
}

// MockBotFlowRepository
type MockBotFlowRepository struct {
	flows map[string]*domain.BotFlow
	mu    sync.RWMutex
}

func NewMockBotFlowRepository() domain.BotFlowRepository {
	return &MockBotFlowRepository{
		flows: make(map[string]*domain.BotFlow),
	}
}

func (r *MockBotFlowRepository) GetByID(ctx context.Context, id string) (*domain.BotFlow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	flow, exists := r.flows[id]
	if !exists {
		return nil, fmt.Errorf("flow not found")
	}
	return flow, nil
}

func (r *MockBotFlowRepository) GetByBotID(ctx context.Context, botID string) ([]*domain.BotFlow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var flows []*domain.BotFlow
	for _, flow := range r.flows {
		if flow.BotID == botID {
			flows = append(flows, flow)
		}
	}
	return flows, nil
}

func (r *MockBotFlowRepository) GetDefaultByBotID(ctx context.Context, botID string) (*domain.BotFlow, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	for _, flow := range r.flows {
		if flow.BotID == botID && flow.IsDefault {
			return flow, nil
		}
	}
	return nil, fmt.Errorf("default flow not found")
}

func (r *MockBotFlowRepository) Create(ctx context.Context, flow *domain.BotFlow) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if flow.ID == "" {
		flow.ID = fmt.Sprintf("flow_%d", time.Now().UnixNano())
	}
	r.flows[flow.ID] = flow
	return nil
}

func (r *MockBotFlowRepository) Update(ctx context.Context, flow *domain.BotFlow) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.flows[flow.ID]; !exists {
		return fmt.Errorf("flow not found")
	}
	r.flows[flow.ID] = flow
	return nil
}

func (r *MockBotFlowRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	delete(r.flows, id)
	return nil
}

// MockBotStepRepository
type MockBotStepRepository struct {
	steps map[string]*domain.BotStep
	mu    sync.RWMutex
}

func NewMockBotStepRepository() domain.BotStepRepository {
	return &MockBotStepRepository{
		steps: make(map[string]*domain.BotStep),
	}
}

func (r *MockBotStepRepository) GetByID(ctx context.Context, id string) (*domain.BotStep, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	step, exists := r.steps[id]
	if !exists {
		return nil, fmt.Errorf("step not found")
	}
	return step, nil
}

func (r *MockBotStepRepository) GetByFlowID(ctx context.Context, flowID string) ([]*domain.BotStep, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var steps []*domain.BotStep
	for _, step := range r.steps {
		if step.FlowID == flowID {
			steps = append(steps, step)
		}
	}
	return steps, nil
}

func (r *MockBotStepRepository) Create(ctx context.Context, step *domain.BotStep) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if step.ID == "" {
		step.ID = fmt.Sprintf("step_%d", time.Now().UnixNano())
	}
	r.steps[step.ID] = step
	return nil
}

func (r *MockBotStepRepository) Update(ctx context.Context, step *domain.BotStep) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.steps[step.ID]; !exists {
		return fmt.Errorf("step not found")
	}
	r.steps[step.ID] = step
	return nil
}

func (r *MockBotStepRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	delete(r.steps, id)
	return nil
}

// MockSmartReplyRepository
type MockSmartReplyRepository struct {
	replies map[string]*domain.SmartReply
	mu      sync.RWMutex
}

func NewMockSmartReplyRepository() domain.SmartReplyRepository {
	return &MockSmartReplyRepository{
		replies: make(map[string]*domain.SmartReply),
	}
}

func (r *MockSmartReplyRepository) GetByID(ctx context.Context, id string) (*domain.SmartReply, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	reply, exists := r.replies[id]
	if !exists {
		return nil, fmt.Errorf("smart reply not found")
	}
	return reply, nil
}

func (r *MockSmartReplyRepository) GetByBotID(ctx context.Context, botID string) ([]*domain.SmartReply, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var replies []*domain.SmartReply
	for _, reply := range r.replies {
		if reply.BotID == botID {
			replies = append(replies, reply)
		}
	}
	return replies, nil
}

func (r *MockSmartReplyRepository) GetByIntent(ctx context.Context, botID, intent string) (*domain.SmartReply, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	for _, reply := range r.replies {
		if reply.BotID == botID && reply.Intent == intent {
			return reply, nil
		}
	}
	return nil, fmt.Errorf("smart reply not found for intent")
}

func (r *MockSmartReplyRepository) Create(ctx context.Context, reply *domain.SmartReply) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if reply.ID == "" {
		reply.ID = fmt.Sprintf("reply_%d", time.Now().UnixNano())
	}
	r.replies[reply.ID] = reply
	return nil
}

func (r *MockSmartReplyRepository) Update(ctx context.Context, reply *domain.SmartReply) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.replies[reply.ID]; !exists {
		return fmt.Errorf("smart reply not found")
	}
	r.replies[reply.ID] = reply
	return nil
}

func (r *MockSmartReplyRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	delete(r.replies, id)
	return nil
}

// MockConversationSessionRepository
type MockConversationSessionRepository struct {
	sessions map[string]*domain.ConversationSession
	mu       sync.RWMutex
}

func NewMockConversationSessionRepository() domain.ConversationSessionRepository {
	return &MockConversationSessionRepository{
		sessions: make(map[string]*domain.ConversationSession),
	}
}

func (r *MockConversationSessionRepository) GetByID(ctx context.Context, id string) (*domain.ConversationSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	session, exists := r.sessions[id]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}
	return session, nil
}

func (r *MockConversationSessionRepository) GetByUserAndBot(ctx context.Context, userID, botID string) (*domain.ConversationSession, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	for _, session := range r.sessions {
		if session.UserID == userID && session.BotID == botID {
			return session, nil
		}
	}
	return nil, fmt.Errorf("session not found")
}

func (r *MockConversationSessionRepository) Create(ctx context.Context, session *domain.ConversationSession) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if session.ID == "" {
		session.ID = fmt.Sprintf("session_%d", time.Now().UnixNano())
	}
	r.sessions[session.ID] = session
	return nil
}

func (r *MockConversationSessionRepository) Update(ctx context.Context, session *domain.ConversationSession) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	if _, exists := r.sessions[session.ID]; !exists {
		return fmt.Errorf("session not found")
	}
	r.sessions[session.ID] = session
	return nil
}

func (r *MockConversationSessionRepository) Delete(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	delete(r.sessions, id)
	return nil
}

func (r *MockConversationSessionRepository) DeleteExpired(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	now := time.Now()
	for id, session := range r.sessions {
		if session.ExpiresAt.Before(now) {
			delete(r.sessions, id)
		}
	}
	return nil
}