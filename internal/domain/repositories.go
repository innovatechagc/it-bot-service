package domain

import "context"

// UserRepository define las operaciones de persistencia para usuarios
type UserRepository interface {
	GetByID(ctx context.Context, id string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*User, error)
}

// AuditRepository define las operaciones de persistencia para auditoría
type AuditRepository interface {
	Create(ctx context.Context, log *AuditLog) error
	GetByUserID(ctx context.Context, userID string, limit, offset int) ([]*AuditLog, error)
	GetByAction(ctx context.Context, action string, limit, offset int) ([]*AuditLog, error)
}

// HealthRepository define las operaciones para health checks
type HealthRepository interface {
	CheckDatabase(ctx context.Context) error
	CheckExternalServices(ctx context.Context) map[string]error
}

// BotRepository define las operaciones de persistencia para bots
type BotRepository interface {
	GetByID(ctx context.Context, id string) (*Bot, error)
	GetByOwnerID(ctx context.Context, ownerID string) ([]*Bot, error)
	Create(ctx context.Context, bot *Bot) error
	Update(ctx context.Context, bot *Bot) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]*Bot, error)
}

// BotFlowRepository define las operaciones de persistencia para flujos de bot
type BotFlowRepository interface {
	GetByID(ctx context.Context, id string) (*BotFlow, error)
	GetByBotID(ctx context.Context, botID string) ([]*BotFlow, error)
	GetDefaultByBotID(ctx context.Context, botID string) (*BotFlow, error)
	Create(ctx context.Context, flow *BotFlow) error
	Update(ctx context.Context, flow *BotFlow) error
	Delete(ctx context.Context, id string) error
}

// BotStepRepository define las operaciones de persistencia para pasos de flujo
type BotStepRepository interface {
	GetByID(ctx context.Context, id string) (*BotStep, error)
	GetByFlowID(ctx context.Context, flowID string) ([]*BotStep, error)
	Create(ctx context.Context, step *BotStep) error
	Update(ctx context.Context, step *BotStep) error
	Delete(ctx context.Context, id string) error
}

// SmartReplyRepository define las operaciones de persistencia para respuestas inteligentes
type SmartReplyRepository interface {
	GetByID(ctx context.Context, id string) (*SmartReply, error)
	GetByBotID(ctx context.Context, botID string) ([]*SmartReply, error)
	GetByIntent(ctx context.Context, botID, intent string) (*SmartReply, error)
	Create(ctx context.Context, reply *SmartReply) error
	Update(ctx context.Context, reply *SmartReply) error
	Delete(ctx context.Context, id string) error
}

// ConversationSessionRepository define las operaciones para sesiones de conversación
type ConversationSessionRepository interface {
	GetByID(ctx context.Context, id string) (*ConversationSession, error)
	GetByUserAndBot(ctx context.Context, userID, botID string) (*ConversationSession, error)
	Create(ctx context.Context, session *ConversationSession) error
	Update(ctx context.Context, session *ConversationSession) error
	Delete(ctx context.Context, id string) error
	DeleteExpired(ctx context.Context) error
}

// ConditionalRepository define las operaciones de persistencia para condiciones
type ConditionalRepository interface {
	GetByID(ctx context.Context, id string) (*Conditional, error)
	GetByBotID(ctx context.Context, botID string) ([]*Conditional, error)
	GetByType(ctx context.Context, botID string, conditionalType ConditionalType) ([]*Conditional, error)
	Create(ctx context.Context, conditional *Conditional) error
	Update(ctx context.Context, conditional *Conditional) error
	Delete(ctx context.Context, id string) error
	Evaluate(ctx context.Context, id string, input map[string]interface{}) (bool, error)
}

// TriggerRepository define las operaciones de persistencia para triggers
type TriggerRepository interface {
	GetByID(ctx context.Context, id string) (*Trigger, error)
	GetByBotID(ctx context.Context, botID string) ([]*Trigger, error)
	GetByEvent(ctx context.Context, botID string, event TriggerEvent) ([]*Trigger, error)
	GetEnabledByBotID(ctx context.Context, botID string) ([]*Trigger, error)
	Create(ctx context.Context, trigger *Trigger) error
	Update(ctx context.Context, trigger *Trigger) error
	Delete(ctx context.Context, id string) error
	Execute(ctx context.Context, id string, eventData map[string]interface{}) error
}

// TestCaseRepository define las operaciones de persistencia para casos de prueba
type TestCaseRepository interface {
	GetByID(ctx context.Context, id string) (*TestCase, error)
	GetByBotID(ctx context.Context, botID string) ([]*TestCase, error)
	GetByStatus(ctx context.Context, botID string, status TestStatus) ([]*TestCase, error)
	Create(ctx context.Context, testCase *TestCase) error
	Update(ctx context.Context, testCase *TestCase) error
	Delete(ctx context.Context, id string) error
	Execute(ctx context.Context, id string) (*TestResult, error)
	BulkExecute(ctx context.Context, ids []string) (map[string]*TestResult, error)
}

// TestSuiteRepository define las operaciones de persistencia para suites de prueba
type TestSuiteRepository interface {
	GetByID(ctx context.Context, id string) (*TestSuite, error)
	GetByBotID(ctx context.Context, botID string) ([]*TestSuite, error)
	GetByStatus(ctx context.Context, botID string, status TestSuiteStatus) ([]*TestSuite, error)
	Create(ctx context.Context, testSuite *TestSuite) error
	Update(ctx context.Context, testSuite *TestSuite) error
	Delete(ctx context.Context, id string) error
	Execute(ctx context.Context, id string) (*TestSuiteResult, error)
	AddTestCase(ctx context.Context, suiteID, testCaseID string) error
	RemoveTestCase(ctx context.Context, suiteID, testCaseID string) error
}