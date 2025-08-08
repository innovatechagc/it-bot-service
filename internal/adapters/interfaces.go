package adapters

import (
	"context"
	"time"
)

// Adapter define la interfaz base para todos los adaptadores
type Adapter interface {
	// Información del adaptador
	GetName() string
	GetType() string
	GetVersion() string
	
	// Ciclo de vida
	Initialize(ctx context.Context, config map[string]interface{}) error
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	IsHealthy() bool
	
	// Capacidades
	GetCapabilities() []string
	CanHandle(operation string) bool
}

// HTTPAdapter define operaciones para adaptadores HTTP
type HTTPAdapter interface {
	Adapter
	MakeRequest(ctx context.Context, request *HTTPRequest) (*HTTPResponse, error)
	SetDefaultHeaders(headers map[string]string)
	SetTimeout(timeout time.Duration)
}

// DatabaseAdapter define operaciones para adaptadores de base de datos
type DatabaseAdapter interface {
	Adapter
	Query(ctx context.Context, query string, params ...interface{}) (*QueryResult, error)
	Execute(ctx context.Context, command string, params ...interface{}) (*ExecuteResult, error)
	BeginTransaction(ctx context.Context) (Transaction, error)
}

// MessageQueueAdapter define operaciones para adaptadores de cola de mensajes
type MessageQueueAdapter interface {
	Adapter
	Publish(ctx context.Context, topic string, message *Message) error
	Subscribe(ctx context.Context, topic string, handler MessageHandler) error
	Unsubscribe(ctx context.Context, topic string) error
}

// WebhookAdapter define operaciones para adaptadores de webhooks
type WebhookAdapter interface {
	Adapter
	RegisterWebhook(ctx context.Context, webhook *WebhookConfig) error
	UnregisterWebhook(ctx context.Context, webhookID string) error
	ProcessWebhook(ctx context.Context, payload *WebhookPayload) (*WebhookResponse, error)
}

// Estructuras de datos

// HTTPRequest representa una solicitud HTTP
type HTTPRequest struct {
	Method  string                 `json:"method"`
	URL     string                 `json:"url"`
	Headers map[string]string      `json:"headers"`
	Body    interface{}            `json:"body"`
	Timeout time.Duration          `json:"timeout"`
	Params  map[string]interface{} `json:"params"`
}

// HTTPResponse representa una respuesta HTTP
type HTTPResponse struct {
	StatusCode int                    `json:"status_code"`
	Headers    map[string]string      `json:"headers"`
	Body       interface{}            `json:"body"`
	Duration   time.Duration          `json:"duration"`
	Success    bool                   `json:"success"`
	Error      string                 `json:"error,omitempty"`
}

// QueryResult representa el resultado de una consulta de base de datos
type QueryResult struct {
	Rows     []map[string]interface{} `json:"rows"`
	Count    int                      `json:"count"`
	Duration time.Duration            `json:"duration"`
	Success  bool                     `json:"success"`
	Error    string                   `json:"error,omitempty"`
}

// ExecuteResult representa el resultado de un comando de base de datos
type ExecuteResult struct {
	RowsAffected int64         `json:"rows_affected"`
	LastInsertID int64         `json:"last_insert_id"`
	Duration     time.Duration `json:"duration"`
	Success      bool          `json:"success"`
	Error        string        `json:"error,omitempty"`
}

// Transaction representa una transacción de base de datos
type Transaction interface {
	Query(ctx context.Context, query string, params ...interface{}) (*QueryResult, error)
	Execute(ctx context.Context, command string, params ...interface{}) (*ExecuteResult, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

// Message representa un mensaje en una cola
type Message struct {
	ID        string                 `json:"id"`
	Topic     string                 `json:"topic"`
	Payload   interface{}            `json:"payload"`
	Headers   map[string]string      `json:"headers"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// MessageHandler define el manejador de mensajes
type MessageHandler func(ctx context.Context, message *Message) error

// WebhookConfig representa la configuración de un webhook
type WebhookConfig struct {
	ID          string            `json:"id"`
	URL         string            `json:"url"`
	Events      []string          `json:"events"`
	Secret      string            `json:"secret"`
	Headers     map[string]string `json:"headers"`
	Timeout     time.Duration     `json:"timeout"`
	RetryPolicy *RetryPolicy      `json:"retry_policy"`
}

// WebhookPayload representa el payload de un webhook
type WebhookPayload struct {
	Event     string                 `json:"event"`
	Data      interface{}            `json:"data"`
	Headers   map[string]string      `json:"headers"`
	Signature string                 `json:"signature"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// WebhookResponse representa la respuesta de un webhook
type WebhookResponse struct {
	Success   bool                   `json:"success"`
	Message   string                 `json:"message"`
	Data      interface{}            `json:"data"`
	Processed bool                   `json:"processed"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// RetryPolicy define la política de reintentos
type RetryPolicy struct {
	MaxRetries int           `json:"max_retries"`
	Delay      time.Duration `json:"delay"`
	Backoff    string        `json:"backoff"` // "linear", "exponential"
}

// AdapterRegistry define el registro de adaptadores
type AdapterRegistry interface {
	Register(name string, adapter Adapter) error
	Unregister(name string) error
	Get(name string) (Adapter, error)
	List() []Adapter
	GetByType(adapterType string) []Adapter
	GetByCapability(capability string) []Adapter
}

// AdapterFactory define la factory de adaptadores
type AdapterFactory interface {
	CreateAdapter(adapterType string, config map[string]interface{}) (Adapter, error)
	GetSupportedTypes() []string
	ValidateConfig(adapterType string, config map[string]interface{}) error
}