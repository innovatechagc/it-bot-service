package domain

import (
	"time"
	"encoding/json"
)

// User representa un usuario del sistema
type User struct {
	ID        string    `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Name      string    `json:"name" db:"name"`
	Roles     []string  `json:"roles" db:"roles"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// AuditLog representa un registro de auditoría
type AuditLog struct {
	ID        string                 `json:"id" db:"id"`
	UserID    string                 `json:"user_id" db:"user_id"`
	Action    string                 `json:"action" db:"action"`
	Resource  string                 `json:"resource" db:"resource"`
	Details   map[string]interface{} `json:"details" db:"details"`
	IPAddress string                 `json:"ip_address" db:"ip_address"`
	UserAgent string                 `json:"user_agent" db:"user_agent"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
}

// Bot representa un bot conversacional
type Bot struct {
	ID        string          `json:"id" db:"id"`
	Name      string          `json:"name" db:"name"`
	OwnerID   string          `json:"owner_id" db:"owner_id"`
	Channel   ChannelType     `json:"channel" db:"channel"`
	Status    BotStatus       `json:"status" db:"status"`
	Config    json.RawMessage `json:"config" db:"config"`
	CreatedAt time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt time.Time       `json:"updated_at" db:"updated_at"`
}

// BotFlow representa un flujo de conversación
type BotFlow struct {
	ID         string    `json:"id" db:"id"`
	BotID      string    `json:"bot_id" db:"bot_id"`
	Name       string    `json:"name" db:"name"`
	Trigger    string    `json:"trigger" db:"trigger"`
	EntryPoint string    `json:"entry_point" db:"entry_point"`
	IsDefault  bool      `json:"is_default" db:"is_default"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// BotStep representa un paso en un flujo de conversación
type BotStep struct {
	ID           string          `json:"id" db:"id"`
	FlowID       string          `json:"flow_id" db:"flow_id"`
	Type         StepType        `json:"type" db:"type"`
	Content      json.RawMessage `json:"content" db:"content"`
	NextStepID   *string         `json:"next_step_id" db:"next_step_id"`
	Conditions   json.RawMessage `json:"conditions" db:"conditions"`
	CreatedAt    time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at" db:"updated_at"`
}

// SmartReply representa una respuesta inteligente basada en IA
type SmartReply struct {
	ID         string    `json:"id" db:"id"`
	BotID      string    `json:"bot_id" db:"bot_id"`
	Intent     string    `json:"intent" db:"intent"`
	Response   string    `json:"response" db:"response"`
	Confidence float64   `json:"confidence" db:"confidence"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// IncomingMessage representa un mensaje entrante
type IncomingMessage struct {
	ID        string                 `json:"id"`
	BotID     string                 `json:"bot_id"`
	UserID    string                 `json:"user_id"`
	Content   string                 `json:"content"`
	Channel   ChannelType            `json:"channel"`
	Metadata  map[string]interface{} `json:"metadata"`
	Timestamp time.Time              `json:"timestamp"`
}

// BotResponse representa la respuesta del bot
type BotResponse struct {
	Content    string                 `json:"content"`
	Type       ResponseType           `json:"type"`
	Options    []ResponseOption       `json:"options,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	NextStepID *string                `json:"next_step_id,omitempty"`
}

// ResponseOption representa una opción de respuesta
type ResponseOption struct {
	ID    string `json:"id"`
	Text  string `json:"text"`
	Value string `json:"value"`
}

// ConversationSession representa una sesión de conversación activa
type ConversationSession struct {
	ID            string                 `json:"id"`
	BotID         string                 `json:"bot_id"`
	UserID        string                 `json:"user_id"`
	CurrentFlowID string                 `json:"current_flow_id"`
	CurrentStepID string                 `json:"current_step_id"`
	Context       map[string]interface{} `json:"context"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	ExpiresAt     time.Time              `json:"expires_at"`
}

// Enums
type ChannelType string

const (
	ChannelWeb      ChannelType = "web"
	ChannelWhatsApp ChannelType = "whatsapp"
	ChannelTelegram ChannelType = "telegram"
	ChannelSlack    ChannelType = "slack"
)

type BotStatus string

const (
	BotStatusActive   BotStatus = "active"
	BotStatusDisabled BotStatus = "disabled"
	BotStatusDraft    BotStatus = "draft"
)

type StepType string

const (
	StepTypeMessage  StepType = "message"
	StepTypeDecision StepType = "decision"
	StepTypeInput    StepType = "input"
	StepTypeAPICall  StepType = "api_call"
	StepTypeAI       StepType = "ai"
)

type ResponseType string

const (
	ResponseTypeText    ResponseType = "text"
	ResponseTypeButtons ResponseType = "buttons"
	ResponseTypeCards   ResponseType = "cards"
	ResponseTypeImage   ResponseType = "image"
)

// APIResponse estructura estándar para respuestas de API
type APIResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// HealthStatus representa el estado de salud del servicio
type HealthStatus struct {
	Status    string                 `json:"status"`
	Timestamp time.Time              `json:"timestamp"`
	Uptime    string                 `json:"uptime"`
	Service   string                 `json:"service"`
	Version   string                 `json:"version"`
	Checks    map[string]interface{} `json:"checks,omitempty"`
}

// MCP (Model Context Protocol) Entities

// MCPAgent representa un agente MCP
type MCPAgent struct {
	ID           string                 `json:"id"`
	Type         string                 `json:"type"`
	Name         string                 `json:"name"`
	Version      string                 `json:"version"`
	Config       map[string]interface{} `json:"config"`
	Capabilities []string               `json:"capabilities"`
	Status       MCPAgentStatus         `json:"status"`
	Timeout      int64                  `json:"timeout"` // en milliseconds
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// MCPTask representa una tarea para ejecutar en un agente MCP
type MCPTask struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Input       map[string]interface{} `json:"input"`
	Priority    int                    `json:"priority"`
	Timeout     int64                  `json:"timeout"` // en milliseconds
	Context     map[string]interface{} `json:"context,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

// MCPTaskResult representa el resultado de la ejecución de una tarea MCP
type MCPTaskResult struct {
	TaskID        string                 `json:"task_id"`
	AgentID       string                 `json:"agent_id"`
	Success       bool                   `json:"success"`
	Output        map[string]interface{} `json:"output"`
	Error         string                 `json:"error,omitempty"`
	ExecutionTime int64                  `json:"execution_time"` // en milliseconds
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	CompletedAt   time.Time              `json:"completed_at"`
}

// MCPAgentMetrics representa métricas de rendimiento de un agente MCP
type MCPAgentMetrics struct {
	AgentID             string    `json:"agent_id"`
	TasksExecuted       int64     `json:"tasks_executed"`
	TasksSuccessful     int64     `json:"tasks_successful"`
	TasksFailed         int64     `json:"tasks_failed"`
	ErrorCount          int64     `json:"error_count"`
	AverageResponseTime int64     `json:"average_response_time"` // en milliseconds
	LastExecution       time.Time `json:"last_execution"`
	LastError           time.Time `json:"last_error"`
	SuccessRate         float64   `json:"success_rate"`
}

// MCPSystemMetrics representa métricas del sistema MCP
type MCPSystemMetrics struct {
	TotalAgents         int       `json:"total_agents"`
	ActiveAgents        int       `json:"active_agents"`
	TotalTasks          int64     `json:"total_tasks"`
	CompletedTasks      int64     `json:"completed_tasks"`
	FailedTasks         int64     `json:"failed_tasks"`
	AverageResponseTime int64     `json:"average_response_time"` // en milliseconds
	SystemUptime        int64     `json:"system_uptime"`         // en seconds
	LastUpdated         time.Time `json:"last_updated"`
}

// MCPAgentStatus representa los posibles estados de un agente MCP
type MCPAgentStatus string

const (
	MCPAgentStatusIdle       MCPAgentStatus = "idle"
	MCPAgentStatusBusy       MCPAgentStatus = "busy"
	MCPAgentStatusError      MCPAgentStatus = "error"
	MCPAgentStatusTerminated MCPAgentStatus = "terminated"
)

// MCPTaskType representa los tipos de tareas MCP soportadas
type MCPTaskType string

const (
	MCPTaskTypeTextGeneration MCPTaskType = "text_generation"
	MCPTaskTypeConversation   MCPTaskType = "conversation"
	MCPTaskTypeAnalysis       MCPTaskType = "analysis"
	MCPTaskTypeSummarization  MCPTaskType = "summarization"
	MCPTaskTypeHTTPRequest    MCPTaskType = "http_request"
	MCPTaskTypeAPICall        MCPTaskType = "api_call"
	MCPTaskTypeWebhook        MCPTaskType = "webhook"
	MCPTaskTypeIntegration    MCPTaskType = "integration"
	MCPTaskTypeWorkflow       MCPTaskType = "workflow"
	MCPTaskTypeSequence       MCPTaskType = "sequence"
	MCPTaskTypeOrchestration  MCPTaskType = "orchestration"
	MCPTaskTypeAutomation     MCPTaskType = "automation"
)

// MCPAgentType representa los tipos de agentes MCP soportados
type MCPAgentType string

const (
	MCPAgentTypeAI       MCPAgentType = "ai"
	MCPAgentTypeHTTP     MCPAgentType = "http"
	MCPAgentTypeWorkflow MCPAgentType = "workflow"
	MCPAgentTypeMock     MCPAgentType = "mock"
)

// Async Task Entities

// AsyncTask representa una tarea asíncrona
type AsyncTask struct {
	ID            string                 `json:"id"`
	Type          string                 `json:"type"`
	Description   string                 `json:"description"`
	UserID        string                 `json:"user_id"`
	BotID         string                 `json:"bot_id"`
	Input         map[string]interface{} `json:"input"`
	Context       map[string]interface{} `json:"context,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
	Priority      int                    `json:"priority"`
	Timeout       int64                  `json:"timeout"` // en milliseconds
	Status        TaskStatus             `json:"status"`
	Result        map[string]interface{} `json:"result,omitempty"`
	Error         string                 `json:"error,omitempty"`
	ExecutionTime int64                  `json:"execution_time,omitempty"` // en milliseconds
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
	StartedAt     time.Time              `json:"started_at,omitempty"`
	CompletedAt   time.Time              `json:"completed_at,omitempty"`
}

// TaskStatus representa los posibles estados de una tarea asíncrona
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
)

// Memory Management Entities

// Memory representa una memoria persistente a largo plazo
type Memory struct {
	ID         string                 `json:"id"`
	UserID     string                 `json:"user_id"`
	BotID      string                 `json:"bot_id"`
	Key        string                 `json:"key"`
	Type       MemoryType             `json:"type"`
	Content    map[string]interface{} `json:"content"`
	Tags       []string               `json:"tags"`
	Importance int                    `json:"importance"` // 1-10, donde 10 es más importante
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
	ExpiresAt  time.Time              `json:"expires_at"`
}

// ContextSummary representa un resumen del contexto de conversación
type ContextSummary struct {
	UserID    string                 `json:"user_id"`
	BotID     string                 `json:"bot_id"`
	Summary   string                 `json:"summary"`
	KeyPoints []string               `json:"key_points"`
	Entities  map[string]interface{} `json:"entities"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
}

// MemoryStats representa estadísticas de memoria para un usuario
type MemoryStats struct {
	UserID               string            `json:"user_id"`
	BotID                string            `json:"bot_id"`
	TotalMemories        int               `json:"total_memories"`
	MemoriesByType       map[string]int    `json:"memories_by_type"`
	MemoriesByImportance map[int]int       `json:"memories_by_importance"`
	OldestMemory         time.Time         `json:"oldest_memory"`
	NewestMemory         time.Time         `json:"newest_memory"`
	LastUpdated          time.Time         `json:"last_updated"`
}

// MemoryType representa los tipos de memoria soportados
type MemoryType string

const (
	MemoryTypePersonal     MemoryType = "personal"     // Información personal del usuario
	MemoryTypePreference   MemoryType = "preference"   // Preferencias del usuario
	MemoryTypeConversation MemoryType = "conversation" // Contexto de conversaciones
	MemoryTypeFact         MemoryType = "fact"         // Hechos importantes
	MemoryTypeGoal         MemoryType = "goal"         // Objetivos del usuario
	MemoryTypeHistory      MemoryType = "history"      // Historial de interacciones
	MemoryTypeCustom       MemoryType = "custom"       // Memoria personalizada
)

// Conditional representa una condición evaluable
type Conditional struct {
	ID          string                 `json:"id"`
	BotID       string                 `json:"bot_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Expression  string                 `json:"expression"`
	Type        ConditionalType        `json:"type"`
	Priority    int                    `json:"priority"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// ConditionalType representa los tipos de condiciones
type ConditionalType string

const (
	ConditionalTypeSimple    ConditionalType = "simple"    // Condición básica
	ConditionalTypeComplex   ConditionalType = "complex"   // Condición compleja
	ConditionalTypeRegex     ConditionalType = "regex"     // Expresión regular
	ConditionalTypeAI        ConditionalType = "ai"        // Evaluación con IA
	ConditionalTypeExternal  ConditionalType = "external"  // Condición externa
)

// Trigger representa un disparador de eventos
type Trigger struct {
	ID          string                 `json:"id"`
	BotID       string                 `json:"bot_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Event       TriggerEvent           `json:"event"`
	Condition   string                 `json:"condition"` // ID de la condición
	Action      TriggerAction          `json:"action"`
	Priority    int                    `json:"priority"`
	Enabled     bool                   `json:"enabled"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// TriggerEvent representa los tipos de eventos que pueden disparar triggers
type TriggerEvent string

const (
	TriggerEventMessageReceived TriggerEvent = "message_received"
	TriggerEventUserJoined      TriggerEvent = "user_joined"
	TriggerEventUserLeft        TriggerEvent = "user_left"
	TriggerEventTimeout         TriggerEvent = "timeout"
	TriggerEventError           TriggerEvent = "error"
	TriggerEventCustom          TriggerEvent = "custom"
)

// TriggerAction representa las acciones que puede ejecutar un trigger
type TriggerAction struct {
	Type    string                 `json:"type"`
	Config  map[string]interface{} `json:"config"`
	Timeout int64                  `json:"timeout"` // en milliseconds
}

// TestCase representa un caso de prueba
type TestCase struct {
	ID          string                 `json:"id"`
	BotID       string                 `json:"bot_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Input       TestInput              `json:"input"`
	Expected    TestExpected           `json:"expected"`
	Conditions  []string               `json:"conditions"` // IDs de condiciones
	Triggers    []string               `json:"triggers"`   // IDs de triggers
	Status      TestStatus             `json:"status"`
	Result      *TestResult            `json:"result,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// TestInput representa la entrada de un caso de prueba
type TestInput struct {
	Message   string                 `json:"message"`
	UserID    string                 `json:"user_id"`
	Context   map[string]interface{} `json:"context,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// TestExpected representa el resultado esperado de un caso de prueba
type TestExpected struct {
	Response    string                 `json:"response"`
	NextStep    string                 `json:"next_step,omitempty"`
	Conditions  []string               `json:"conditions,omitempty"`
	Triggers    []string               `json:"triggers,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Timeout     int64                  `json:"timeout"` // en milliseconds
}

// TestResult representa el resultado de ejecutar un caso de prueba
type TestResult struct {
	Success       bool                   `json:"success"`
	ActualResponse string                `json:"actual_response"`
	ActualNextStep string                `json:"actual_next_step,omitempty"`
	ExecutedConditions []string          `json:"executed_conditions,omitempty"`
	ExecutedTriggers   []string          `json:"executed_triggers,omitempty"`
	ActualContext      map[string]interface{} `json:"actual_context,omitempty"`
	ExecutionTime      int64             `json:"execution_time"` // en milliseconds
	Error             string             `json:"error,omitempty"`
	ExecutedAt        time.Time          `json:"executed_at"`
}

// TestStatus representa el estado de un caso de prueba
type TestStatus string

const (
	TestStatusPending   TestStatus = "pending"
	TestStatusRunning   TestStatus = "running"
	TestStatusPassed    TestStatus = "passed"
	TestStatusFailed    TestStatus = "failed"
	TestStatusSkipped   TestStatus = "skipped"
)

// TestSuite representa una suite de pruebas
type TestSuite struct {
	ID          string                 `json:"id"`
	BotID       string                 `json:"bot_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	TestCases   []string               `json:"test_cases"` // IDs de casos de prueba
	Status      TestSuiteStatus        `json:"status"`
	Result      *TestSuiteResult       `json:"result,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// TestSuiteStatus representa el estado de una suite de pruebas
type TestSuiteStatus string

const (
	TestSuiteStatusPending TestSuiteStatus = "pending"
	TestSuiteStatusRunning TestSuiteStatus = "running"
	TestSuiteStatusPassed  TestSuiteStatus = "passed"
	TestSuiteStatusFailed  TestSuiteStatus = "failed"
	TestSuiteStatusPartial TestSuiteStatus = "partial"
)

// TestSuiteResult representa el resultado de una suite de pruebas
type TestSuiteResult struct {
	TotalTests     int                    `json:"total_tests"`
	PassedTests    int                    `json:"passed_tests"`
	FailedTests    int                    `json:"failed_tests"`
	SkippedTests   int                    `json:"skipped_tests"`
	SuccessRate    float64                `json:"success_rate"`
	ExecutionTime  int64                  `json:"execution_time"` // en milliseconds
	StartedAt      time.Time              `json:"started_at"`
	CompletedAt    time.Time              `json:"completed_at"`
	TestResults    map[string]*TestResult `json:"test_results,omitempty"`
}