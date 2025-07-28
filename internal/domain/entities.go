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