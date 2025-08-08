package repositories

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/company/bot-service/internal/domain"
	"github.com/google/uuid"
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

// MockConditionalRepository implementa ConditionalRepository para testing
type MockConditionalRepository struct {
	conditionals map[string]*domain.Conditional
}

func NewMockConditionalRepository() domain.ConditionalRepository {
	return &MockConditionalRepository{
		conditionals: make(map[string]*domain.Conditional),
	}
}

func (r *MockConditionalRepository) Create(ctx context.Context, conditional *domain.Conditional) error {
	if conditional.ID == "" {
		conditional.ID = uuid.New().String()
	}
	r.conditionals[conditional.ID] = conditional
	return nil
}

func (r *MockConditionalRepository) GetByID(ctx context.Context, id string) (*domain.Conditional, error) {
	if conditional, exists := r.conditionals[id]; exists {
		return conditional, nil
	}
	return nil, fmt.Errorf("conditional not found")
}

func (r *MockConditionalRepository) GetByBotID(ctx context.Context, botID string) ([]*domain.Conditional, error) {
	var result []*domain.Conditional
	for _, conditional := range r.conditionals {
		if conditional.BotID == botID {
			result = append(result, conditional)
		}
	}
	return result, nil
}

func (r *MockConditionalRepository) Update(ctx context.Context, conditional *domain.Conditional) error {
	if _, exists := r.conditionals[conditional.ID]; !exists {
		return fmt.Errorf("conditional not found")
	}
	r.conditionals[conditional.ID] = conditional
	return nil
}

func (r *MockConditionalRepository) Delete(ctx context.Context, id string) error {
	if _, exists := r.conditionals[id]; !exists {
		return fmt.Errorf("conditional not found")
	}
	delete(r.conditionals, id)
	return nil
}

func (r *MockConditionalRepository) GetByType(ctx context.Context, botID string, conditionalType domain.ConditionalType) ([]*domain.Conditional, error) {
	var result []*domain.Conditional
	for _, conditional := range r.conditionals {
		if conditional.BotID == botID && conditional.Type == conditionalType {
			result = append(result, conditional)
		}
	}
	return result, nil
}

func (r *MockConditionalRepository) Evaluate(ctx context.Context, id string, input map[string]interface{}) (bool, error) {
	conditional, exists := r.conditionals[id]
	if !exists {
		return false, fmt.Errorf("conditional not found")
	}
	
	// Simulación simple de evaluación
	if conditional.Expression == "user.age > 18" {
		if age, ok := input["user.age"].(float64); ok {
			return age > 18, nil
		}
	}
	
	return false, nil
}

// MockTriggerRepository implementa TriggerRepository para testing
type MockTriggerRepository struct {
	triggers map[string]*domain.Trigger
}

func NewMockTriggerRepository() domain.TriggerRepository {
	return &MockTriggerRepository{
		triggers: make(map[string]*domain.Trigger),
	}
}

func (r *MockTriggerRepository) Create(ctx context.Context, trigger *domain.Trigger) error {
	if trigger.ID == "" {
		trigger.ID = uuid.New().String()
	}
	r.triggers[trigger.ID] = trigger
	return nil
}

func (r *MockTriggerRepository) GetByID(ctx context.Context, id string) (*domain.Trigger, error) {
	if trigger, exists := r.triggers[id]; exists {
		return trigger, nil
	}
	return nil, fmt.Errorf("trigger not found")
}

func (r *MockTriggerRepository) GetByBotID(ctx context.Context, botID string) ([]*domain.Trigger, error) {
	var result []*domain.Trigger
	for _, trigger := range r.triggers {
		if trigger.BotID == botID {
			result = append(result, trigger)
		}
	}
	return result, nil
}

func (r *MockTriggerRepository) GetByEvent(ctx context.Context, botID string, event domain.TriggerEvent) ([]*domain.Trigger, error) {
	var result []*domain.Trigger
	for _, trigger := range r.triggers {
		if trigger.BotID == botID && trigger.Event == event {
			result = append(result, trigger)
		}
	}
	return result, nil
}

func (r *MockTriggerRepository) Update(ctx context.Context, trigger *domain.Trigger) error {
	if _, exists := r.triggers[trigger.ID]; !exists {
		return fmt.Errorf("trigger not found")
	}
	r.triggers[trigger.ID] = trigger
	return nil
}

func (r *MockTriggerRepository) Delete(ctx context.Context, id string) error {
	if _, exists := r.triggers[id]; !exists {
		return fmt.Errorf("trigger not found")
	}
	delete(r.triggers, id)
	return nil
}

func (r *MockTriggerRepository) GetEnabledByBotID(ctx context.Context, botID string) ([]*domain.Trigger, error) {
	var result []*domain.Trigger
	for _, trigger := range r.triggers {
		if trigger.BotID == botID && trigger.Enabled {
			result = append(result, trigger)
		}
	}
	return result, nil
}

func (r *MockTriggerRepository) Execute(ctx context.Context, id string, eventData map[string]interface{}) error {
	trigger, exists := r.triggers[id]
	if !exists {
		return fmt.Errorf("trigger not found")
	}
	
	// Simulación de ejecución
	trigger.UpdatedAt = time.Now()
	return nil
}

// MockTestCaseRepository implementa TestCaseRepository para testing
type MockTestCaseRepository struct {
	testCases map[string]*domain.TestCase
}

func NewMockTestCaseRepository() domain.TestCaseRepository {
	return &MockTestCaseRepository{
		testCases: make(map[string]*domain.TestCase),
	}
}

func (r *MockTestCaseRepository) Create(ctx context.Context, testCase *domain.TestCase) error {
	if testCase.ID == "" {
		testCase.ID = uuid.New().String()
	}
	r.testCases[testCase.ID] = testCase
	return nil
}

func (r *MockTestCaseRepository) GetByID(ctx context.Context, id string) (*domain.TestCase, error) {
	if testCase, exists := r.testCases[id]; exists {
		return testCase, nil
	}
	return nil, fmt.Errorf("test case not found")
}

func (r *MockTestCaseRepository) GetByBotID(ctx context.Context, botID string) ([]*domain.TestCase, error) {
	var result []*domain.TestCase
	for _, testCase := range r.testCases {
		if testCase.BotID == botID {
			result = append(result, testCase)
		}
	}
	return result, nil
}

func (r *MockTestCaseRepository) GetByStatus(ctx context.Context, botID string, status domain.TestStatus) ([]*domain.TestCase, error) {
	var result []*domain.TestCase
	for _, testCase := range r.testCases {
		if testCase.BotID == botID && testCase.Status == status {
			result = append(result, testCase)
		}
	}
	return result, nil
}

func (r *MockTestCaseRepository) Update(ctx context.Context, testCase *domain.TestCase) error {
	if _, exists := r.testCases[testCase.ID]; !exists {
		return fmt.Errorf("test case not found")
	}
	r.testCases[testCase.ID] = testCase
	return nil
}

func (r *MockTestCaseRepository) Delete(ctx context.Context, id string) error {
	if _, exists := r.testCases[id]; !exists {
		return fmt.Errorf("test case not found")
	}
	delete(r.testCases, id)
	return nil
}

func (r *MockTestCaseRepository) Execute(ctx context.Context, id string) (*domain.TestResult, error) {
	testCase, exists := r.testCases[id]
	if !exists {
		return nil, fmt.Errorf("test case not found")
	}
	
	// Simulación de ejecución
	result := &domain.TestResult{
		Success:        true,
		ActualResponse: "Simulated response",
		ExecutionTime:  100,
		ExecutedAt:     time.Now(),
	}
	
	testCase.Result = result
	testCase.Status = domain.TestStatusPassed
	return result, nil
}

func (r *MockTestCaseRepository) BulkExecute(ctx context.Context, ids []string) (map[string]*domain.TestResult, error) {
	results := make(map[string]*domain.TestResult)
	
	for _, id := range ids {
		result, err := r.Execute(ctx, id)
		if err != nil {
			results[id] = &domain.TestResult{
				Success:   false,
				Error:     err.Error(),
				ExecutedAt: time.Now(),
			}
		} else {
			results[id] = result
		}
	}
	
	return results, nil
}

// MockTestSuiteRepository implementa TestSuiteRepository para testing
type MockTestSuiteRepository struct {
	testSuites map[string]*domain.TestSuite
}

func NewMockTestSuiteRepository() domain.TestSuiteRepository {
	return &MockTestSuiteRepository{
		testSuites: make(map[string]*domain.TestSuite),
	}
}

func (r *MockTestSuiteRepository) Create(ctx context.Context, testSuite *domain.TestSuite) error {
	if testSuite.ID == "" {
		testSuite.ID = uuid.New().String()
	}
	r.testSuites[testSuite.ID] = testSuite
	return nil
}

func (r *MockTestSuiteRepository) GetByID(ctx context.Context, id string) (*domain.TestSuite, error) {
	if testSuite, exists := r.testSuites[id]; exists {
		return testSuite, nil
	}
	return nil, fmt.Errorf("test suite not found")
}

func (r *MockTestSuiteRepository) GetByBotID(ctx context.Context, botID string) ([]*domain.TestSuite, error) {
	var result []*domain.TestSuite
	for _, testSuite := range r.testSuites {
		if testSuite.BotID == botID {
			result = append(result, testSuite)
		}
	}
	return result, nil
}

func (r *MockTestSuiteRepository) GetByStatus(ctx context.Context, botID string, status domain.TestSuiteStatus) ([]*domain.TestSuite, error) {
	var result []*domain.TestSuite
	for _, testSuite := range r.testSuites {
		if testSuite.BotID == botID && testSuite.Status == status {
			result = append(result, testSuite)
		}
	}
	return result, nil
}

func (r *MockTestSuiteRepository) Update(ctx context.Context, testSuite *domain.TestSuite) error {
	if _, exists := r.testSuites[testSuite.ID]; !exists {
		return fmt.Errorf("test suite not found")
	}
	r.testSuites[testSuite.ID] = testSuite
	return nil
}

func (r *MockTestSuiteRepository) Delete(ctx context.Context, id string) error {
	if _, exists := r.testSuites[id]; !exists {
		return fmt.Errorf("test suite not found")
	}
	delete(r.testSuites, id)
	return nil
}

func (r *MockTestSuiteRepository) AddTestCaseToSuite(ctx context.Context, suiteID, testCaseID string) error {
	testSuite, exists := r.testSuites[suiteID]
	if !exists {
		return fmt.Errorf("test suite not found")
	}
	
	// Verificar que el test case no esté ya en el suite
	for _, existingID := range testSuite.TestCases {
		if existingID == testCaseID {
			return fmt.Errorf("test case already in suite")
		}
	}
	
	testSuite.TestCases = append(testSuite.TestCases, testCaseID)
	return nil
}

func (r *MockTestSuiteRepository) RemoveTestCaseFromSuite(ctx context.Context, suiteID, testCaseID string) error {
	testSuite, exists := r.testSuites[suiteID]
	if !exists {
		return fmt.Errorf("test suite not found")
	}
	
	// Remover el test case del suite
	for i, existingID := range testSuite.TestCases {
		if existingID == testCaseID {
			testSuite.TestCases = append(testSuite.TestCases[:i], testSuite.TestCases[i+1:]...)
			return nil
		}
	}
	
	return fmt.Errorf("test case not found in suite")
}

func (r *MockTestSuiteRepository) Execute(ctx context.Context, id string) (*domain.TestSuiteResult, error) {
	testSuite, exists := r.testSuites[id]
	if !exists {
		return nil, fmt.Errorf("test suite not found")
	}
	
	// Simulación de ejecución
	result := &domain.TestSuiteResult{
		TotalTests:    len(testSuite.TestCases),
		PassedTests:   len(testSuite.TestCases),
		FailedTests:   0,
		SkippedTests:  0,
		SuccessRate:   1.0,
		ExecutionTime: 500,
		StartedAt:     time.Now(),
		CompletedAt:   time.Now(),
		TestResults:   make(map[string]*domain.TestResult),
	}
	
	return result, nil
}

func (r *MockTestSuiteRepository) AddTestCase(ctx context.Context, suiteID, testCaseID string) error {
	return r.AddTestCaseToSuite(ctx, suiteID, testCaseID)
}

func (r *MockTestSuiteRepository) RemoveTestCase(ctx context.Context, suiteID, testCaseID string) error {
	return r.RemoveTestCaseFromSuite(ctx, suiteID, testCaseID)
}