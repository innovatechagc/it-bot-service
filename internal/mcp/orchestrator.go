package mcp

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/company/bot-service/internal/domain"
	"github.com/company/bot-service/pkg/logger"
)

// orchestrator implementa MCPOrchestrator
type orchestrator struct {
	agents        map[string]Agent
	factory       AgentFactory
	logger        logger.Logger
	mu            sync.RWMutex
	startTime     time.Time
	taskCounter   int64
	metrics       SystemMetrics
	agentMetrics  map[string]*domain.MCPAgentMetrics
}

// NewOrchestrator crea una nueva instancia del orquestador MCP
func NewOrchestrator(factory AgentFactory, logger logger.Logger) interface {
	MCPOrchestrator
	MCPDomainOrchestrator
} {
	return &orchestrator{
		agents:       make(map[string]Agent),
		factory:      factory,
		logger:       logger,
		startTime:    time.Now(),
		metrics:      SystemMetrics{},
		agentMetrics: make(map[string]*domain.MCPAgentMetrics),
	}
}

// InstantiateMCP crea e inicia un nuevo agente MCP
func (o *orchestrator) InstantiateMCP(ctx context.Context, config MCPConfig) (Agent, error) {
	o.mu.Lock()
	defer o.mu.Unlock()

	// Validar configuración
	if err := o.factory.ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid MCP config: %w", err)
	}

	// Crear agente
	agent, err := o.factory.CreateAgent(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	// Iniciar agente
	if err := agent.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start agent: %w", err)
	}

	// Registrar agente
	o.agents[agent.GetID()] = agent
	o.metrics.TotalAgents++
	o.metrics.ActiveAgents++

	o.logger.Info("MCP agent instantiated", 
		"agent_id", agent.GetID(),
		"type", agent.GetType(),
		"capabilities", agent.GetCapabilities())

	return agent, nil
}

// GetAgent obtiene un agente por ID
func (o *orchestrator) GetAgent(agentID string) (Agent, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	agent, exists := o.agents[agentID]
	if !exists {
		return nil, fmt.Errorf("agent not found: %s", agentID)
	}

	return agent, nil
}

// ListAgents devuelve todos los agentes registrados
func (o *orchestrator) ListAgents() []Agent {
	o.mu.RLock()
	defer o.mu.RUnlock()

	agents := make([]Agent, 0, len(o.agents))
	for _, agent := range o.agents {
		agents = append(agents, agent)
	}

	return agents
}

// TerminateAgent termina y elimina un agente
func (o *orchestrator) TerminateAgent(ctx context.Context, agentID string) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	agent, exists := o.agents[agentID]
	if !exists {
		return fmt.Errorf("agent not found: %s", agentID)
	}

	// Detener agente
	if err := agent.Stop(ctx); err != nil {
		o.logger.Error("Failed to stop agent gracefully", "agent_id", agentID, "error", err)
	}

	// Eliminar del registro
	delete(o.agents, agentID)
	o.metrics.ActiveAgents--

	o.logger.Info("MCP agent terminated", "agent_id", agentID)

	return nil
}

// ExecuteTask ejecuta una tarea en el agente más apropiado
func (o *orchestrator) ExecuteTask(ctx context.Context, task Task) (Result, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	// Buscar agente apropiado
	var selectedAgent Agent
	for _, agent := range o.agents {
		if agent.CanHandle(task.Type) && agent.IsHealthy() {
			state := agent.GetState()
			if state.Status == AgentStatusIdle {
				selectedAgent = agent
				break
			}
		}
	}

	if selectedAgent == nil {
		return Result{
			TaskID:  task.ID,
			Success: false,
			Error:   "no suitable agent available",
		}, fmt.Errorf("no suitable agent found for task type: %s", task.Type)
	}

	// Ejecutar tarea
	o.logger.Info("Executing task", 
		"task_id", task.ID,
		"task_type", task.Type,
		"agent_id", selectedAgent.GetID())

	start := time.Now()
	result, err := selectedAgent.Execute(ctx, task)
	duration := time.Since(start)

	// Actualizar métricas
	o.metrics.TotalTasks++
	if err != nil || !result.Success {
		o.metrics.FailedTasks++
	} else {
		o.metrics.CompletedTasks++
	}

	// Calcular tiempo promedio
	if o.metrics.TotalTasks > 0 {
		totalTime := o.metrics.AverageExecTime*time.Duration(o.metrics.TotalTasks-1) + duration
		o.metrics.AverageExecTime = totalTime / time.Duration(o.metrics.TotalTasks)
	} else {
		o.metrics.AverageExecTime = duration
	}

	if err != nil {
		o.logger.Error("Task execution failed", 
			"task_id", task.ID,
			"agent_id", selectedAgent.GetID(),
			"error", err)
		return result, err
	}

	o.logger.Info("Task completed", 
		"task_id", task.ID,
		"agent_id", selectedAgent.GetID(),
		"duration", duration,
		"success", result.Success)

	return result, nil
}

// CoordinateAgents coordina múltiples agentes para una tarea compleja
func (o *orchestrator) CoordinateAgents(ctx context.Context, agents []Agent, task Task) (Result, error) {
	if len(agents) == 0 {
		return Result{
			TaskID:  task.ID,
			Success: false,
			Error:   "no agents provided",
		}, fmt.Errorf("no agents provided for coordination")
	}

	// Para tareas simples, usar el primer agente disponible
	if len(agents) == 1 {
		return agents[0].Execute(ctx, task)
	}

	// Para múltiples agentes, implementar coordinación básica
	// TODO: Implementar lógica de coordinación más sofisticada
	o.logger.Info("Coordinating multiple agents", 
		"task_id", task.ID,
		"agent_count", len(agents))

	// Por ahora, ejecutar en el primer agente disponible
	for _, agent := range agents {
		if agent.IsHealthy() && agent.CanHandle(task.Type) {
			return agent.Execute(ctx, task)
		}
	}

	return Result{
		TaskID:  task.ID,
		Success: false,
		Error:   "no healthy agent available",
	}, fmt.Errorf("no healthy agent available for coordination")
}

// PassContext pasa contexto a un agente específico
func (o *orchestrator) PassContext(ctx context.Context, agentID string, context map[string]interface{}) error {
	agent, err := o.GetAgent(agentID)
	if err != nil {
		return err
	}

	if err := agent.SetContext(context); err != nil {
		return fmt.Errorf("failed to set context for agent %s: %w", agentID, err)
	}

	o.logger.Info("Context passed to agent", 
		"agent_id", agentID,
		"context_keys", getMapKeys(context))

	return nil
}

// ShareContext comparte contexto entre agentes
func (o *orchestrator) ShareContext(ctx context.Context, fromAgentID, toAgentID string, keys []string) error {
	fromAgent, err := o.GetAgent(fromAgentID)
	if err != nil {
		return fmt.Errorf("source agent not found: %w", err)
	}

	toAgent, err := o.GetAgent(toAgentID)
	if err != nil {
		return fmt.Errorf("target agent not found: %w", err)
	}

	// Obtener contexto del agente origen
	fromContext := fromAgent.GetContext()
	
	// Crear contexto filtrado
	sharedContext := make(map[string]interface{})
	for _, key := range keys {
		if value, exists := fromContext[key]; exists {
			sharedContext[key] = value
		}
	}

	// Pasar contexto al agente destino
	if err := toAgent.SetContext(sharedContext); err != nil {
		return fmt.Errorf("failed to share context: %w", err)
	}

	o.logger.Info("Context shared between agents", 
		"from_agent", fromAgentID,
		"to_agent", toAgentID,
		"shared_keys", keys)

	return nil
}

// GetAgentMetrics obtiene métricas de un agente específico
func (o *orchestrator) GetAgentMetrics(agentID string) (AgentMetrics, error) {
	agent, err := o.GetAgent(agentID)
	if err != nil {
		return AgentMetrics{}, err
	}

	state := agent.GetState()
	return state.Metrics, nil
}

// GetSystemMetrics obtiene métricas del sistema
func (o *orchestrator) GetSystemMetrics() (SystemMetrics, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	// Actualizar métricas del sistema
	o.metrics.SystemUptime = time.Since(o.startTime)
	
	// Contar agentes activos
	activeCount := 0
	for _, agent := range o.agents {
		if agent.IsHealthy() {
			activeCount++
		}
	}
	o.metrics.ActiveAgents = activeCount

	return o.metrics, nil
}

// Start inicia el orquestador
func (o *orchestrator) Start(ctx context.Context) error {
	o.logger.Info("Starting MCP orchestrator")
	o.startTime = time.Now()
	return nil
}

// Stop detiene el orquestador y todos los agentes
func (o *orchestrator) Stop(ctx context.Context) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	o.logger.Info("Stopping MCP orchestrator")

	// Detener todos los agentes
	for agentID, agent := range o.agents {
		if err := agent.Stop(ctx); err != nil {
			o.logger.Error("Failed to stop agent", "agent_id", agentID, "error", err)
		}
	}

	// Limpiar registro de agentes
	o.agents = make(map[string]Agent)
	o.metrics.ActiveAgents = 0

	o.logger.Info("MCP orchestrator stopped")
	return nil
}

// Función auxiliar para obtener las claves de un mapa
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// ExecuteTaskDomain ejecuta una tarea usando las estructuras de dominio
func (o *orchestrator) ExecuteTaskDomain(ctx context.Context, task *domain.MCPTask) (*domain.MCPTaskResult, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	o.logger.Info("Executing MCP domain task", "task_id", task.ID, "type", task.Type)

	// Buscar agente apropiado
	var selectedAgent Agent
	for _, agent := range o.agents {
		if agent.CanHandle(task.Type) && agent.IsHealthy() {
			state := agent.GetState()
			if state.Status == AgentStatusIdle {
				selectedAgent = agent
				break
			}
		}
	}

	if selectedAgent == nil {
		return &domain.MCPTaskResult{
			TaskID:        task.ID,
			Success:       false,
			Error:         "no suitable agent available",
			ExecutionTime: 0,
			CompletedAt:   time.Now(),
		}, fmt.Errorf("no suitable agent found for task type: %s", task.Type)
	}

	// Pasar contexto al agente si es necesario
	if task.Context != nil {
		if err := selectedAgent.SetContext(task.Context); err != nil {
			o.logger.Error("Failed to pass context to agent", "agent_id", selectedAgent.GetID(), "error", err)
		}
	}

	// Convertir tarea de dominio a tarea interna
	internalTask := Task{
		ID:          task.ID,
		Type:        task.Type,
		Description: task.Description,
		Input:       task.Input,
		Priority:    task.Priority,
		Metadata:    task.Metadata,
	}

	// Ejecutar tarea con timeout
	taskCtx := ctx
	if task.Timeout > 0 {
		var cancel context.CancelFunc
		taskCtx, cancel = context.WithTimeout(ctx, time.Duration(task.Timeout)*time.Millisecond)
		defer cancel()
	}

	start := time.Now()
	result, err := selectedAgent.Execute(taskCtx, internalTask)
	executionTime := time.Since(start).Milliseconds()

	// Inicializar métricas del agente si no existen
	agentID := selectedAgent.GetID()
	if _, exists := o.agentMetrics[agentID]; !exists {
		o.agentMetrics[agentID] = &domain.MCPAgentMetrics{
			AgentID:             agentID,
			TasksExecuted:       0,
			TasksSuccessful:     0,
			TasksFailed:         0,
			ErrorCount:          0,
			AverageResponseTime: 0,
			SuccessRate:         0.0,
		}
	}

	// Actualizar métricas del agente
	metrics := o.agentMetrics[agentID]
	metrics.TasksExecuted++
	metrics.LastExecution = time.Now()

	if err != nil || !result.Success {
		metrics.TasksFailed++
		metrics.ErrorCount++
		metrics.LastError = time.Now()
		
		o.logger.Error("Task execution failed", "task_id", task.ID, "agent_id", agentID, "error", err)
		
		return &domain.MCPTaskResult{
			TaskID:        task.ID,
			AgentID:       agentID,
			Success:       false,
			Error:         result.Error,
			ExecutionTime: executionTime,
			Metadata:      result.Metadata,
			CompletedAt:   time.Now(),
		}, err
	}

	// Actualizar métricas de éxito
	metrics.TasksSuccessful++
	if metrics.TasksExecuted > 0 {
		metrics.SuccessRate = float64(metrics.TasksSuccessful) / float64(metrics.TasksExecuted)
	}
	
	// Calcular tiempo promedio de respuesta
	if metrics.TasksExecuted == 1 {
		metrics.AverageResponseTime = executionTime
	} else {
		metrics.AverageResponseTime = (metrics.AverageResponseTime + executionTime) / 2
	}

	o.logger.Info("Task executed successfully", 
		"task_id", task.ID, 
		"agent_id", agentID, 
		"duration", executionTime)

	return &domain.MCPTaskResult{
		TaskID:        task.ID,
		AgentID:       agentID,
		Success:       true,
		Output:        result.Output,
		ExecutionTime: executionTime,
		Metadata:      result.Metadata,
		CompletedAt:   time.Now(),
	}, nil
}

// GetAgentMetricsDomain obtiene métricas de un agente usando estructuras de dominio
func (o *orchestrator) GetAgentMetricsDomain(agentID string) (*domain.MCPAgentMetrics, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	// Verificar que el agente existe
	if _, exists := o.agents[agentID]; !exists {
		return nil, fmt.Errorf("agent not found: %s", agentID)
	}

	// Obtener métricas del agente
	if metrics, exists := o.agentMetrics[agentID]; exists {
		return metrics, nil
	}

	// Si no hay métricas, crear métricas vacías
	return &domain.MCPAgentMetrics{
		AgentID:             agentID,
		TasksExecuted:       0,
		TasksSuccessful:     0,
		TasksFailed:         0,
		ErrorCount:          0,
		AverageResponseTime: 0,
		SuccessRate:         0.0,
	}, nil
}

// GetSystemMetricsDomain obtiene métricas del sistema usando estructuras de dominio
func (o *orchestrator) GetSystemMetricsDomain() (*domain.MCPSystemMetrics, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	// Contar agentes activos
	activeCount := 0
	for _, agent := range o.agents {
		if agent.IsHealthy() {
			activeCount++
		}
	}

	// Calcular métricas agregadas
	var totalTasks, completedTasks, failedTasks int64
	var totalResponseTime int64
	agentCount := 0

	for _, metrics := range o.agentMetrics {
		totalTasks += metrics.TasksExecuted
		completedTasks += metrics.TasksSuccessful
		failedTasks += metrics.TasksFailed
		totalResponseTime += metrics.AverageResponseTime
		agentCount++
	}

	var averageResponseTime int64
	if agentCount > 0 {
		averageResponseTime = totalResponseTime / int64(agentCount)
	}

	return &domain.MCPSystemMetrics{
		TotalAgents:         len(o.agents),
		ActiveAgents:        activeCount,
		TotalTasks:          totalTasks,
		CompletedTasks:      completedTasks,
		FailedTasks:         failedTasks,
		AverageResponseTime: averageResponseTime,
		SystemUptime:        int64(time.Since(o.startTime).Seconds()),
		LastUpdated:         time.Now(),
	}, nil
}

// CreateAgentFromDomain crea un agente usando la estructura de dominio
func (o *orchestrator) CreateAgentFromDomain(ctx context.Context, agentConfig *domain.MCPAgent) (Agent, error) {
	// Convertir configuración de dominio a configuración interna
	config := MCPConfig{
		Type:         agentConfig.Type,
		Name:         agentConfig.Name,
		Version:      agentConfig.Version,
		Config:       agentConfig.Config,
		Capabilities: agentConfig.Capabilities,
		Timeout:      time.Duration(agentConfig.Timeout) * time.Millisecond,
	}

	return o.InstantiateMCP(ctx, config)
}

// GetSupportedAgentTypes obtiene los tipos de agentes soportados
func (o *orchestrator) GetSupportedAgentTypes() []string {
	return o.factory.GetSupportedTypes()
}