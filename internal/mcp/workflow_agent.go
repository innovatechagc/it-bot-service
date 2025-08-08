package mcp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/company/bot-service/pkg/logger"
)

// WorkflowAgent implementa un agente que ejecuta workflows secuenciales
type workflowAgent struct {
	*baseAgent
	steps []WorkflowStep
}

// WorkflowStep representa un paso en un workflow
type WorkflowStep struct {
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	Config      map[string]interface{} `json:"config"`
	OnError     string                 `json:"on_error,omitempty"` // "continue", "stop", "retry"
	Timeout     time.Duration          `json:"timeout,omitempty"`
}

// NewWorkflowAgent crea un nuevo agente de workflow
func NewWorkflowAgent(config MCPConfig, logger logger.Logger) (Agent, error) {
	base := newBaseAgent(config, logger)
	base.capabilities = []string{"workflow", "sequence", "orchestration", "automation"}
	
	// Parsear pasos del workflow
	steps, err := parseWorkflowSteps(config.Config)
	if err != nil {
		return nil, fmt.Errorf("failed to parse workflow steps: %w", err)
	}
	
	return &workflowAgent{
		baseAgent: base,
		steps:     steps,
	}, nil
}

func (a *workflowAgent) Execute(ctx context.Context, task Task) (Result, error) {
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
	
	a.logger.Info("Workflow agent executing task", 
		"agent_id", a.id,
		"task_id", task.ID,
		"steps_count", len(a.steps))
	
	// Ejecutar pasos del workflow
	results := make([]map[string]interface{}, 0, len(a.steps))
	workflowData := make(map[string]interface{})
	
	// Inicializar datos del workflow con input de la tarea
	for k, v := range task.Input {
		workflowData[k] = v
	}
	
	for i, step := range a.steps {
		stepStart := time.Now()
		
		a.logger.Info("Executing workflow step", 
			"agent_id", a.id,
			"task_id", task.ID,
			"step", i+1,
			"step_type", step.Type)
		
		stepResult, err := a.executeStep(ctx, step, workflowData)
		stepDuration := time.Since(stepStart)
		
		stepInfo := map[string]interface{}{
			"step":        i + 1,
			"type":        step.Type,
			"description": step.Description,
			"duration":    stepDuration.Milliseconds(),
			"success":     err == nil,
		}
		
		if err != nil {
			stepInfo["error"] = err.Error()
			
			// Manejar error según configuración
			switch step.OnError {
			case "continue":
				a.logger.Warn("Step failed but continuing", 
					"agent_id", a.id,
					"step", i+1,
					"error", err)
				stepInfo["action"] = "continued"
			case "retry":
				// Implementar retry simple
				a.logger.Info("Retrying failed step", 
					"agent_id", a.id,
					"step", i+1)
				stepResult, err = a.executeStep(ctx, step, workflowData)
				if err != nil {
					stepInfo["retry_error"] = err.Error()
					stepInfo["action"] = "retry_failed"
				} else {
					stepInfo["action"] = "retry_success"
					stepInfo["success"] = true
				}
			default: // "stop"
				stepInfo["action"] = "stopped"
				results = append(results, stepInfo)
				
				duration := time.Since(start)
				a.updateMetrics(false, duration)
				
				return Result{
					TaskID:  task.ID,
					Success: false,
					Error:   fmt.Sprintf("workflow stopped at step %d: %v", i+1, err),
					Output: map[string]interface{}{
						"completed_steps": results,
						"failed_at_step":  i + 1,
						"workflow_data":   workflowData,
					},
					Duration: duration,
					Metadata: map[string]interface{}{
						"agent_id":     a.id,
						"agent_type":   a.agentType,
						"total_steps":  len(a.steps),
						"completed":    i,
					},
				}, err
			}
		}
		
		// Agregar resultado del paso a los datos del workflow
		if stepResult != nil {
			stepInfo["output"] = stepResult
			workflowData[fmt.Sprintf("step_%d_result", i+1)] = stepResult
		}
		
		results = append(results, stepInfo)
	}
	
	duration := time.Since(start)
	a.updateMetrics(true, duration)
	
	a.logger.Info("Workflow completed successfully", 
		"agent_id", a.id,
		"task_id", task.ID,
		"duration", duration,
		"steps_executed", len(results))
	
	return Result{
		TaskID:  task.ID,
		Success: true,
		Output: map[string]interface{}{
			"steps_executed": results,
			"workflow_data":  workflowData,
			"summary": map[string]interface{}{
				"total_steps":    len(a.steps),
				"completed":      len(results),
				"total_duration": duration.Milliseconds(),
			},
		},
		Duration: duration,
		Metadata: map[string]interface{}{
			"agent_id":   a.id,
			"agent_type": a.agentType,
		},
	}, nil
}

func (a *workflowAgent) executeStep(ctx context.Context, step WorkflowStep, workflowData map[string]interface{}) (interface{}, error) {
	// Aplicar timeout del paso si está configurado
	stepCtx := ctx
	if step.Timeout > 0 {
		var cancel context.CancelFunc
		stepCtx, cancel = context.WithTimeout(ctx, step.Timeout)
		defer cancel()
	}
	
	switch step.Type {
	case "log":
		return a.executeLogStep(stepCtx, step, workflowData)
	case "delay":
		return a.executeDelayStep(stepCtx, step, workflowData)
	case "transform":
		return a.executeTransformStep(stepCtx, step, workflowData)
	case "condition":
		return a.executeConditionStep(stepCtx, step, workflowData)
	case "http_call":
		return a.executeHTTPCallStep(stepCtx, step, workflowData)
	case "set_variable":
		return a.executeSetVariableStep(stepCtx, step, workflowData)
	default:
		return nil, fmt.Errorf("unsupported step type: %s", step.Type)
	}
}

func (a *workflowAgent) executeLogStep(ctx context.Context, step WorkflowStep, workflowData map[string]interface{}) (interface{}, error) {
	message, ok := step.Config["message"].(string)
	if !ok {
		message = "Workflow step executed"
	}
	
	// Reemplazar variables en el mensaje
	message = a.replaceVariables(message, workflowData)
	
	a.logger.Info("Workflow log step", 
		"agent_id", a.id,
		"message", message)
	
	return map[string]interface{}{
		"message":   message,
		"timestamp": time.Now().Format(time.RFC3339),
	}, nil
}

func (a *workflowAgent) executeDelayStep(ctx context.Context, step WorkflowStep, workflowData map[string]interface{}) (interface{}, error) {
	delayMs, ok := step.Config["delay_ms"].(float64)
	if !ok {
		delayMs = 1000 // 1 segundo por defecto
	}
	
	delay := time.Duration(delayMs) * time.Millisecond
	
	select {
	case <-time.After(delay):
		return map[string]interface{}{
			"delay_ms": delayMs,
			"message":  fmt.Sprintf("Delayed for %v", delay),
		}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (a *workflowAgent) executeTransformStep(ctx context.Context, step WorkflowStep, workflowData map[string]interface{}) (interface{}, error) {
	operation, ok := step.Config["operation"].(string)
	if !ok {
		return nil, fmt.Errorf("transform step requires 'operation' config")
	}
	
	switch operation {
	case "normalize":
		// Ejemplo de normalización simple
		result := make(map[string]interface{})
		for k, v := range workflowData {
			if str, ok := v.(string); ok {
				result[k] = strings.ToLower(strings.TrimSpace(str))
			} else {
				result[k] = v
			}
		}
		return result, nil
	case "uppercase":
		result := make(map[string]interface{})
		for k, v := range workflowData {
			if str, ok := v.(string); ok {
				result[k] = strings.ToUpper(str)
			} else {
				result[k] = v
			}
		}
		return result, nil
	default:
		return nil, fmt.Errorf("unsupported transform operation: %s", operation)
	}
}

func (a *workflowAgent) executeConditionStep(ctx context.Context, step WorkflowStep, workflowData map[string]interface{}) (interface{}, error) {
	condition, ok := step.Config["condition"].(string)
	if !ok {
		return nil, fmt.Errorf("condition step requires 'condition' config")
	}
	
	// Implementación simple de evaluación de condiciones
	// En un sistema real, esto sería más sofisticado
	result := a.evaluateCondition(condition, workflowData)
	
	return map[string]interface{}{
		"condition": condition,
		"result":    result,
		"evaluated": true,
	}, nil
}

func (a *workflowAgent) executeHTTPCallStep(ctx context.Context, step WorkflowStep, workflowData map[string]interface{}) (interface{}, error) {
	// Esta sería una llamada HTTP real en un sistema completo
	// Por ahora, simular la respuesta
	url, _ := step.Config["url"].(string)
	method, _ := step.Config["method"].(string)
	if method == "" {
		method = "GET"
	}
	
	return map[string]interface{}{
		"url":         url,
		"method":      method,
		"status_code": 200,
		"response":    "Simulated HTTP response",
		"simulated":   true,
	}, nil
}

func (a *workflowAgent) executeSetVariableStep(ctx context.Context, step WorkflowStep, workflowData map[string]interface{}) (interface{}, error) {
	varName, ok := step.Config["name"].(string)
	if !ok {
		return nil, fmt.Errorf("set_variable step requires 'name' config")
	}
	
	value := step.Config["value"]
	workflowData[varName] = value
	
	return map[string]interface{}{
		"variable": varName,
		"value":    value,
		"set":      true,
	}, nil
}

func (a *workflowAgent) replaceVariables(text string, data map[string]interface{}) string {
	// Implementación simple de reemplazo de variables
	// En un sistema real, esto sería más sofisticado
	result := text
	for k, v := range data {
		placeholder := fmt.Sprintf("{{%s}}", k)
		replacement := fmt.Sprintf("%v", v)
		result = strings.ReplaceAll(result, placeholder, replacement)
	}
	return result
}

func (a *workflowAgent) evaluateCondition(condition string, data map[string]interface{}) bool {
	// Implementación muy simple de evaluación de condiciones
	// En un sistema real, esto usaría un parser de expresiones
	return len(data) > 0 // Condición dummy
}

func (a *workflowAgent) CanHandle(taskType string) bool {
	supportedTypes := []string{
		"workflow",
		"sequence",
		"orchestration",
		"automation",
		"pipeline",
		"process",
	}
	
	for _, supported := range supportedTypes {
		if taskType == supported {
			return true
		}
	}
	
	return false
}

// parseWorkflowSteps parsea los pasos del workflow desde la configuración
func parseWorkflowSteps(config map[string]interface{}) ([]WorkflowStep, error) {
	stepsInterface, exists := config["steps"]
	if !exists {
		return nil, fmt.Errorf("workflow config must contain 'steps'")
	}
	
	stepsArray, ok := stepsInterface.([]interface{})
	if !ok {
		return nil, fmt.Errorf("steps must be an array")
	}
	
	steps := make([]WorkflowStep, 0, len(stepsArray))
	
	for i, stepInterface := range stepsArray {
		stepMap, ok := stepInterface.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("step %d must be an object", i)
		}
		
		step := WorkflowStep{
			Config: make(map[string]interface{}),
		}
		
		// Parsear campos del paso
		if stepType, exists := stepMap["type"]; exists {
			if typeStr, ok := stepType.(string); ok {
				step.Type = typeStr
			}
		}
		
		if description, exists := stepMap["description"]; exists {
			if descStr, ok := description.(string); ok {
				step.Description = descStr
			}
		}
		
		if onError, exists := stepMap["on_error"]; exists {
			if errorStr, ok := onError.(string); ok {
				step.OnError = errorStr
			}
		}
		
		if config, exists := stepMap["config"]; exists {
			if configMap, ok := config.(map[string]interface{}); ok {
				step.Config = configMap
			}
		}
		
		if step.Type == "" {
			return nil, fmt.Errorf("step %d must have a type", i)
		}
		
		steps = append(steps, step)
	}
	
	return steps, nil
}