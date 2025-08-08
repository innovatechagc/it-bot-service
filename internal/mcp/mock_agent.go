package mcp

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/company/bot-service/pkg/logger"
)

// MockAgent implementa un agente mock para testing y desarrollo
type mockAgent struct {
	*baseAgent
	responses []string
	config    MCPConfig
}

// NewMockAgent crea un nuevo agente mock
func NewMockAgent(config MCPConfig, logger logger.Logger) (Agent, error) {
	base := newBaseAgent(config, logger)
	base.capabilities = []string{"mock", "test", "development", "simulation"}
	
	// Respuestas por defecto
	responses := []string{
		"Mock agent executed task successfully",
		"Task completed with mock data",
		"Simulated execution completed",
		"Mock response generated",
	}
	
	// Usar respuestas personalizadas si están configuradas
	if customResponses, exists := config.Config["responses"]; exists {
		if responseList, ok := customResponses.([]interface{}); ok {
			responses = make([]string, 0, len(responseList))
			for _, resp := range responseList {
				if respStr, ok := resp.(string); ok {
					responses = append(responses, respStr)
				}
			}
		}
	}
	
	return &mockAgent{
		baseAgent: base,
		responses: responses,
		config:    config,
	}, nil
}

func (a *mockAgent) Execute(ctx context.Context, task Task) (Result, error) {
	start := time.Now()
	
	// Actualizar estado
	a.mu.Lock()
	a.state.Status = AgentStatusBusy
	a.state.CurrentTask = &task
	a.mu.Unlock()
	
	defer func() {
		a.mu.Lock()
		a.state.Status = AgentStatusIdle
		a.state.CurrentTask = nil
		a.mu.Unlock()
	}()
	
	a.logger.Info("Mock agent executing task", 
		"agent_id", a.id,
		"task_id", task.ID,
		"task_type", task.Type)
	
	// Simular tiempo de procesamiento
	processingTime := a.getProcessingTime()
	time.Sleep(processingTime)
	
	// Simular posible fallo (5% de probabilidad)
	if a.shouldSimulateFailure() {
		duration := time.Since(start)
		a.updateMetrics(false, duration)
		
		return Result{
			TaskID:   task.ID,
			Success:  false,
			Error:    "Simulated failure for testing purposes",
			Duration: duration,
			Metadata: map[string]interface{}{
				"agent_id":   a.id,
				"agent_type": a.agentType,
				"simulated":  true,
			},
		}, fmt.Errorf("simulated failure")
	}
	
	// Generar respuesta mock
	response := a.generateMockResponse(task)
	
	duration := time.Since(start)
	a.updateMetrics(true, duration)
	
	result := Result{
		TaskID:  task.ID,
		Success: true,
		Output: map[string]interface{}{
			"response":    response,
			"task_type":   task.Type,
			"input_keys":  getInputKeys(task.Input),
			"context_keys": getMapKeys(a.context),
			"timestamp":   time.Now().Unix(),
		},
		Duration: duration,
		Metadata: map[string]interface{}{
			"agent_id":   a.id,
			"agent_type": a.agentType,
			"simulated":  true,
		},
	}
	
	a.logger.Info("Mock agent task completed", 
		"agent_id", a.id,
		"task_id", task.ID,
		"duration", duration,
		"response_length", len(response))
	
	return result, nil
}

func (a *mockAgent) CanHandle(taskType string) bool {
	// El agente mock puede manejar cualquier tipo de tarea
	return true
}

func (a *mockAgent) getProcessingTime() time.Duration {
	// Simular tiempo de procesamiento variable (100ms - 2s)
	minMs := 100
	maxMs := 2000
	
	// Usar configuración personalizada si está disponible
	if minTime, exists := a.config.Config["min_processing_time_ms"]; exists {
		if minInt, ok := minTime.(float64); ok {
			minMs = int(minInt)
		}
	}
	
	if maxTime, exists := a.config.Config["max_processing_time_ms"]; exists {
		if maxInt, ok := maxTime.(float64); ok {
			maxMs = int(maxInt)
		}
	}
	
	randomMs := minMs + rand.Intn(maxMs-minMs)
	return time.Duration(randomMs) * time.Millisecond
}

func (a *mockAgent) shouldSimulateFailure() bool {
	// 5% de probabilidad de fallo por defecto
	failureRate := 0.05
	
	if customRate, exists := a.config.Config["failure_rate"]; exists {
		if rate, ok := customRate.(float64); ok {
			failureRate = rate
		}
	}
	
	return rand.Float64() < failureRate
}

func (a *mockAgent) generateMockResponse(task Task) string {
	// Seleccionar respuesta aleatoria
	if len(a.responses) == 0 {
		return "Mock agent response"
	}
	
	baseResponse := a.responses[rand.Intn(len(a.responses))]
	
	// Personalizar respuesta basada en la tarea
	return fmt.Sprintf("%s for task '%s' (ID: %s)", baseResponse, task.Type, task.ID)
}

// Funciones auxiliares
func getInputKeys(input map[string]interface{}) []string {
	keys := make([]string, 0, len(input))
	for k := range input {
		keys = append(keys, k)
	}
	return keys
}