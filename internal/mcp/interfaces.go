package mcp

import (
	"context"
	"time"
	
	"github.com/company/bot-service/internal/domain"
)

// MCPConfig representa la configuración para instanciar un MCP
type MCPConfig struct {
	Type        string                 `json:"type"`        // Tipo de MCP (openai, claude, custom, etc.)
	Name        string                 `json:"name"`        // Nombre del agente
	Version     string                 `json:"version"`     // Versión del MCP
	Config      map[string]interface{} `json:"config"`      // Configuración específica
	Capabilities []string              `json:"capabilities"` // Capacidades del agente
	Timeout     time.Duration          `json:"timeout"`     // Timeout para operaciones
}

// Task representa una tarea que debe ejecutar un agente
type Task struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`        // Tipo de tarea
	Description string                 `json:"description"` // Descripción de la tarea
	Input       map[string]interface{} `json:"input"`       // Datos de entrada
	Priority    int                    `json:"priority"`    // Prioridad (1-10)
	Deadline    *time.Time             `json:"deadline"`    // Deadline opcional
	Metadata    map[string]interface{} `json:"metadata"`    // Metadata adicional
}

// Result representa el resultado de la ejecución de una tarea
type Result struct {
	TaskID      string                 `json:"task_id"`
	Success     bool                   `json:"success"`
	Output      map[string]interface{} `json:"output"`
	Error       string                 `json:"error,omitempty"`
	Duration    time.Duration          `json:"duration"`
	Metadata    map[string]interface{} `json:"metadata"`
	NextActions []string               `json:"next_actions,omitempty"`
}

// AgentState representa el estado actual de un agente
type AgentState struct {
	ID           string                 `json:"id"`
	Status       AgentStatus            `json:"status"`
	CurrentTask  *Task                  `json:"current_task,omitempty"`
	LastActivity time.Time              `json:"last_activity"`
	Metrics      AgentMetrics           `json:"metrics"`
	Context      map[string]interface{} `json:"context"`
}

// AgentStatus representa los posibles estados de un agente
type AgentStatus string

const (
	AgentStatusIdle       AgentStatus = "idle"
	AgentStatusBusy       AgentStatus = "busy"
	AgentStatusError      AgentStatus = "error"
	AgentStatusTerminated AgentStatus = "terminated"
)

// AgentMetrics representa métricas de rendimiento de un agente
type AgentMetrics struct {
	TasksCompleted   int           `json:"tasks_completed"`
	TasksFailed      int           `json:"tasks_failed"`
	AverageExecTime  time.Duration `json:"average_exec_time"`
	TotalExecTime    time.Duration `json:"total_exec_time"`
	LastError        string        `json:"last_error,omitempty"`
	SuccessRate      float64       `json:"success_rate"`
}

// Agent interface define las operaciones básicas de un agente MCP
type Agent interface {
	// Información del agente
	GetID() string
	GetType() string
	GetCapabilities() []string
	
	// Ejecución de tareas
	Execute(ctx context.Context, task Task) (Result, error)
	CanHandle(taskType string) bool
	
	// Estado del agente
	GetState() AgentState
	UpdateState(state AgentState) error
	
	// Gestión del contexto
	SetContext(ctx map[string]interface{}) error
	GetContext() map[string]interface{}
	
	// Ciclo de vida
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	IsHealthy() bool
}

// MCPOrchestrator interface define las operaciones de orquestación
type MCPOrchestrator interface {
	// Gestión de agentes
	InstantiateMCP(ctx context.Context, config MCPConfig) (Agent, error)
	GetAgent(agentID string) (Agent, error)
	ListAgents() []Agent
	TerminateAgent(ctx context.Context, agentID string) error
	
	// Coordinación de tareas
	ExecuteTask(ctx context.Context, task Task) (Result, error)
	CoordinateAgents(ctx context.Context, agents []Agent, task Task) (Result, error)
	
	// Gestión de contexto
	PassContext(ctx context.Context, agentID string, context map[string]interface{}) error
	ShareContext(ctx context.Context, fromAgentID, toAgentID string, keys []string) error
	
	// Monitoreo
	GetAgentMetrics(agentID string) (AgentMetrics, error)
	GetSystemMetrics() (SystemMetrics, error)
	
	// Ciclo de vida
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

// SystemMetrics representa métricas del sistema de orquestación
type SystemMetrics struct {
	TotalAgents      int           `json:"total_agents"`
	ActiveAgents     int           `json:"active_agents"`
	TotalTasks       int           `json:"total_tasks"`
	CompletedTasks   int           `json:"completed_tasks"`
	FailedTasks      int           `json:"failed_tasks"`
	AverageExecTime  time.Duration `json:"average_exec_time"`
	SystemUptime     time.Duration `json:"system_uptime"`
	MemoryUsage      int64         `json:"memory_usage"`
	CPUUsage         float64       `json:"cpu_usage"`
}

// AgentFactory interface para crear diferentes tipos de agentes
type AgentFactory interface {
	CreateAgent(config MCPConfig) (Agent, error)
	GetSupportedTypes() []string
	ValidateConfig(config MCPConfig) error
}

// MCPDomainOrchestrator interface adicional para trabajar con estructuras de dominio
type MCPDomainOrchestrator interface {
	// Métodos que trabajan con estructuras de dominio
	ExecuteTaskDomain(ctx context.Context, task *domain.MCPTask) (*domain.MCPTaskResult, error)
	GetAgentMetricsDomain(agentID string) (*domain.MCPAgentMetrics, error)
	GetSystemMetricsDomain() (*domain.MCPSystemMetrics, error)
	CreateAgentFromDomain(ctx context.Context, agentConfig *domain.MCPAgent) (Agent, error)
	GetSupportedAgentTypes() []string
}