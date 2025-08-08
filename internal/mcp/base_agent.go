package mcp

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/company/bot-service/pkg/logger"
)

// baseAgent proporciona funcionalidad común para todos los agentes
type baseAgent struct {
	id           string
	agentType    string
	name         string
	version      string
	capabilities []string
	state        AgentState
	context      map[string]interface{}
	logger       logger.Logger
	mu           sync.RWMutex
	startTime    time.Time
}

// newBaseAgent crea una nueva instancia base de agente
func newBaseAgent(config MCPConfig, logger logger.Logger) *baseAgent {
	agentID := generateAgentID(config.Type, config.Name)
	
	return &baseAgent{
		id:           agentID,
		agentType:    config.Type,
		name:         config.Name,
		version:      config.Version,
		capabilities: config.Capabilities,
		state: AgentState{
			ID:           agentID,
			Status:       AgentStatusIdle,
			LastActivity: time.Now(),
			Metrics: AgentMetrics{
				TasksCompleted:  0,
				TasksFailed:     0,
				AverageExecTime: 0,
				TotalExecTime:   0,
				SuccessRate:     0.0,
			},
			Context: make(map[string]interface{}),
		},
		context:   make(map[string]interface{}),
		logger:    logger,
		startTime: time.Now(),
	}
}

// GetID devuelve el ID del agente
func (a *baseAgent) GetID() string {
	return a.id
}

// GetType devuelve el tipo del agente
func (a *baseAgent) GetType() string {
	return a.agentType
}

// GetCapabilities devuelve las capacidades del agente
func (a *baseAgent) GetCapabilities() []string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	
	capabilities := make([]string, len(a.capabilities))
	copy(capabilities, a.capabilities)
	return capabilities
}

// GetState devuelve el estado actual del agente
func (a *baseAgent) GetState() AgentState {
	a.mu.RLock()
	defer a.mu.RUnlock()
	
	// Crear una copia del estado para evitar modificaciones concurrentes
	state := a.state
	state.Context = make(map[string]interface{})
	for k, v := range a.state.Context {
		state.Context[k] = v
	}
	
	return state
}

// UpdateState actualiza el estado del agente
func (a *baseAgent) UpdateState(state AgentState) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	a.state = state
	a.state.LastActivity = time.Now()
	
	return nil
}

// SetContext establece el contexto del agente
func (a *baseAgent) SetContext(ctx map[string]interface{}) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	// Crear una copia del contexto
	a.context = make(map[string]interface{})
	for k, v := range ctx {
		a.context[k] = v
	}
	
	// También actualizar el contexto en el estado
	a.state.Context = make(map[string]interface{})
	for k, v := range ctx {
		a.state.Context[k] = v
	}
	
	a.state.LastActivity = time.Now()
	
	a.logger.Info("Context updated for agent", 
		"agent_id", a.id,
		"context_keys", getContextKeys(ctx))
	
	return nil
}

// GetContext devuelve el contexto actual del agente
func (a *baseAgent) GetContext() map[string]interface{} {
	a.mu.RLock()
	defer a.mu.RUnlock()
	
	// Crear una copia del contexto para evitar modificaciones concurrentes
	context := make(map[string]interface{})
	for k, v := range a.context {
		context[k] = v
	}
	
	return context
}

// Start inicia el agente
func (a *baseAgent) Start(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	a.state.Status = AgentStatusIdle
	a.state.LastActivity = time.Now()
	a.startTime = time.Now()
	
	a.logger.Info("Agent started", 
		"agent_id", a.id,
		"type", a.agentType,
		"name", a.name)
	
	return nil
}

// Stop detiene el agente
func (a *baseAgent) Stop(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	a.state.Status = AgentStatusTerminated
	a.state.LastActivity = time.Now()
	a.state.CurrentTask = nil
	
	a.logger.Info("Agent stopped", 
		"agent_id", a.id,
		"type", a.agentType,
		"uptime", time.Since(a.startTime))
	
	return nil
}

// IsHealthy verifica si el agente está saludable
func (a *baseAgent) IsHealthy() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	
	// Un agente está saludable si no está terminado y está en estado idle o busy
	if a.state.Status == AgentStatusTerminated {
		return false
	}
	
	// Considerar saludable si está en estado idle o busy (no terminado)
	if a.state.Status == AgentStatusIdle || a.state.Status == AgentStatusBusy {
		return true
	}
	
	return false
}

// updateMetrics actualiza las métricas del agente
func (a *baseAgent) updateMetrics(success bool, duration time.Duration) {
	a.mu.Lock()
	defer a.mu.Unlock()
	
	if success {
		a.state.Metrics.TasksCompleted++
	} else {
		a.state.Metrics.TasksFailed++
	}
	
	// Actualizar tiempo total de ejecución
	a.state.Metrics.TotalExecTime += duration
	
	// Calcular tiempo promedio de ejecución
	totalTasks := a.state.Metrics.TasksCompleted + a.state.Metrics.TasksFailed
	if totalTasks > 0 {
		a.state.Metrics.AverageExecTime = a.state.Metrics.TotalExecTime / time.Duration(totalTasks)
	}
	
	// Calcular tasa de éxito
	if totalTasks > 0 {
		a.state.Metrics.SuccessRate = float64(a.state.Metrics.TasksCompleted) / float64(totalTasks)
	}
	
	// Actualizar último error si la tarea falló
	if !success {
		a.state.Metrics.LastError = fmt.Sprintf("Task failed at %s", time.Now().Format(time.RFC3339))
	}
	
	a.state.LastActivity = time.Now()
}

// generateAgentID genera un ID único para el agente
func generateAgentID(agentType, name string) string {
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("%s-%s-%d", agentType, name, timestamp)
}

// getContextKeys obtiene las claves del contexto para logging
func getContextKeys(ctx map[string]interface{}) []string {
	keys := make([]string, 0, len(ctx))
	for k := range ctx {
		keys = append(keys, k)
	}
	return keys
}