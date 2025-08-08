package mcp

import (
	"context"
	"fmt"
	"time"

	"github.com/company/bot-service/internal/adapters"
	"github.com/company/bot-service/pkg/logger"
)

// adapterAgent implementa un agente que usa adaptadores para interoperabilidad
type adapterAgent struct {
	*baseAgent
	registry adapters.AdapterRegistry
	factory  adapters.AdapterFactory
}

// NewAdapterAgent crea un nuevo agente de adaptador
func NewAdapterAgent(config MCPConfig, registry adapters.AdapterRegistry, factory adapters.AdapterFactory, logger logger.Logger) (Agent, error) {
	base := newBaseAgent(config, logger)
	base.capabilities = []string{
		"http_request",
		"api_call",
		"webhook",
		"integration",
		"adapter_management",
		"interoperability",
	}
	
	return &adapterAgent{
		baseAgent: base,
		registry:  registry,
		factory:   factory,
	}, nil
}

func (a *adapterAgent) Execute(ctx context.Context, task Task) (Result, error) {
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
	
	a.logger.Info("Adapter agent executing task", 
		"agent_id", a.id,
		"task_id", task.ID,
		"task_type", task.Type)
	
	var result Result
	var err error
	
	switch task.Type {
	case "http_request", "api_call":
		result, err = a.executeHTTPTask(ctx, task)
	case "create_adapter":
		result, err = a.executeCreateAdapterTask(ctx, task)
	case "list_adapters":
		result, err = a.executeListAdaptersTask(ctx, task)
	case "adapter_health":
		result, err = a.executeAdapterHealthTask(ctx, task)
	default:
		// Intentar encontrar un adaptador que pueda manejar la tarea
		result, err = a.executeWithAdapter(ctx, task)
	}
	
	duration := time.Since(start)
	result.Duration = duration
	
	a.updateMetrics(result.Success, duration)
	
	a.logger.Info("Adapter agent task completed", 
		"agent_id", a.id,
		"task_id", task.ID,
		"duration", duration,
		"success", result.Success)
	
	return result, err
}

func (a *adapterAgent) executeHTTPTask(ctx context.Context, task Task) (Result, error) {
	// Buscar adaptador HTTP
	httpAdapters := a.registry.GetByType("http")
	if len(httpAdapters) == 0 {
		// Crear adaptador HTTP dinámicamente
		adapterConfig := map[string]interface{}{
			"name":    "dynamic-http-adapter",
			"version": "1.0",
		}
		
		adapter, err := a.factory.CreateAdapter("http", adapterConfig)
		if err != nil {
			return Result{
				TaskID:  task.ID,
				Success: false,
				Error:   fmt.Sprintf("failed to create HTTP adapter: %v", err),
			}, err
		}
		
		// Inicializar y registrar el adaptador
		if err := adapter.Initialize(ctx, adapterConfig); err != nil {
			return Result{
				TaskID:  task.ID,
				Success: false,
				Error:   fmt.Sprintf("failed to initialize HTTP adapter: %v", err),
			}, err
		}
		
		if err := adapter.Start(ctx); err != nil {
			return Result{
				TaskID:  task.ID,
				Success: false,
				Error:   fmt.Sprintf("failed to start HTTP adapter: %v", err),
			}, err
		}
		
		if err := a.registry.Register("dynamic-http-adapter", adapter); err != nil {
			return Result{
				TaskID:  task.ID,
				Success: false,
				Error:   fmt.Sprintf("failed to register HTTP adapter: %v", err),
			}, err
		}
		
		httpAdapters = []adapters.Adapter{adapter}
	}
	
	// Usar el primer adaptador HTTP disponible
	httpAdapter, ok := httpAdapters[0].(adapters.HTTPAdapter)
	if !ok {
		return Result{
			TaskID:  task.ID,
			Success: false,
			Error:   "HTTP adapter does not implement HTTPAdapter interface",
		}, fmt.Errorf("invalid HTTP adapter")
	}
	
	// Construir solicitud HTTP
	request := &adapters.HTTPRequest{
		Method:  "GET",
		Headers: make(map[string]string),
		Params:  make(map[string]interface{}),
	}
	
	// Extraer parámetros de la tarea
	if method, exists := task.Input["method"]; exists {
		if methodStr, ok := method.(string); ok {
			request.Method = methodStr
		}
	}
	
	if url, exists := task.Input["url"]; exists {
		if urlStr, ok := url.(string); ok {
			request.URL = urlStr
		}
	}
	
	if endpoint, exists := task.Input["endpoint"]; exists {
		if endpointStr, ok := endpoint.(string); ok {
			request.URL = endpointStr
		}
	}
	
	if headers, exists := task.Input["headers"]; exists {
		if headersMap, ok := headers.(map[string]interface{}); ok {
			for k, v := range headersMap {
				if strVal, ok := v.(string); ok {
					request.Headers[k] = strVal
				}
			}
		}
	}
	
	if params, exists := task.Input["params"]; exists {
		if paramsMap, ok := params.(map[string]interface{}); ok {
			request.Params = paramsMap
		}
	}
	
	if body, exists := task.Input["body"]; exists {
		request.Body = body
	}
	
	if timeout, exists := task.Input["timeout"]; exists {
		if timeoutFloat, ok := timeout.(float64); ok {
			request.Timeout = time.Duration(timeoutFloat) * time.Millisecond
		}
	}
	
	// Ejecutar solicitud
	response, err := httpAdapter.MakeRequest(ctx, request)
	if err != nil {
		return Result{
			TaskID:  task.ID,
			Success: false,
			Error:   fmt.Sprintf("HTTP request failed: %v", err),
			Output: map[string]interface{}{
				"adapter_name": httpAdapter.GetName(),
				"adapter_type": httpAdapter.GetType(),
			},
		}, err
	}
	
	return Result{
		TaskID:  task.ID,
		Success: response.Success,
		Error:   response.Error,
		Output: map[string]interface{}{
			"status_code":    response.StatusCode,
			"headers":        response.Headers,
			"body":           response.Body,
			"duration":       response.Duration.Milliseconds(),
			"adapter_name":   httpAdapter.GetName(),
			"adapter_type":   httpAdapter.GetType(),
			"adapter_healthy": httpAdapter.IsHealthy(),
		},
		Metadata: map[string]interface{}{
			"agent_id":     a.id,
			"agent_type":   a.agentType,
			"adapter_used": httpAdapter.GetName(),
		},
	}, nil
}

func (a *adapterAgent) executeCreateAdapterTask(ctx context.Context, task Task) (Result, error) {
	// Extraer configuración del adaptador
	adapterType, ok := task.Input["adapter_type"].(string)
	if !ok {
		return Result{
			TaskID:  task.ID,
			Success: false,
			Error:   "adapter_type is required",
		}, fmt.Errorf("adapter_type is required")
	}
	
	adapterName, ok := task.Input["adapter_name"].(string)
	if !ok {
		adapterName = fmt.Sprintf("%s-adapter-%d", adapterType, time.Now().UnixNano())
	}
	
	config, ok := task.Input["config"].(map[string]interface{})
	if !ok {
		config = make(map[string]interface{})
	}
	
	config["name"] = adapterName
	
	// Validar configuración
	if err := a.factory.ValidateConfig(adapterType, config); err != nil {
		return Result{
			TaskID:  task.ID,
			Success: false,
			Error:   fmt.Sprintf("invalid adapter config: %v", err),
		}, err
	}
	
	// Crear adaptador
	adapter, err := a.factory.CreateAdapter(adapterType, config)
	if err != nil {
		return Result{
			TaskID:  task.ID,
			Success: false,
			Error:   fmt.Sprintf("failed to create adapter: %v", err),
		}, err
	}
	
	// Inicializar adaptador
	if err := adapter.Initialize(ctx, config); err != nil {
		return Result{
			TaskID:  task.ID,
			Success: false,
			Error:   fmt.Sprintf("failed to initialize adapter: %v", err),
		}, err
	}
	
	// Iniciar adaptador
	if err := adapter.Start(ctx); err != nil {
		return Result{
			TaskID:  task.ID,
			Success: false,
			Error:   fmt.Sprintf("failed to start adapter: %v", err),
		}, err
	}
	
	// Registrar adaptador
	if err := a.registry.Register(adapterName, adapter); err != nil {
		return Result{
			TaskID:  task.ID,
			Success: false,
			Error:   fmt.Sprintf("failed to register adapter: %v", err),
		}, err
	}
	
	return Result{
		TaskID:  task.ID,
		Success: true,
		Output: map[string]interface{}{
			"adapter_name":    adapter.GetName(),
			"adapter_type":    adapter.GetType(),
			"adapter_version": adapter.GetVersion(),
			"capabilities":    adapter.GetCapabilities(),
			"healthy":         adapter.IsHealthy(),
		},
		Metadata: map[string]interface{}{
			"agent_id":   a.id,
			"agent_type": a.agentType,
			"operation":  "create_adapter",
		},
	}, nil
}

func (a *adapterAgent) executeListAdaptersTask(ctx context.Context, task Task) (Result, error) {
	adapters := a.registry.List()
	
	adapterList := make([]map[string]interface{}, 0, len(adapters))
	for _, adapter := range adapters {
		adapterList = append(adapterList, map[string]interface{}{
			"name":         adapter.GetName(),
			"type":         adapter.GetType(),
			"version":      adapter.GetVersion(),
			"capabilities": adapter.GetCapabilities(),
			"healthy":      adapter.IsHealthy(),
		})
	}
	
	return Result{
		TaskID:  task.ID,
		Success: true,
		Output: map[string]interface{}{
			"adapters": adapterList,
			"count":    len(adapterList),
		},
		Metadata: map[string]interface{}{
			"agent_id":   a.id,
			"agent_type": a.agentType,
			"operation":  "list_adapters",
		},
	}, nil
}

func (a *adapterAgent) executeAdapterHealthTask(ctx context.Context, task Task) (Result, error) {
	adapters := a.registry.List()
	
	healthStatus := make(map[string]interface{})
	healthyCount := 0
	
	for _, adapter := range adapters {
		healthy := adapter.IsHealthy()
		if healthy {
			healthyCount++
		}
		
		healthStatus[adapter.GetName()] = map[string]interface{}{
			"healthy":      healthy,
			"type":         adapter.GetType(),
			"version":      adapter.GetVersion(),
			"capabilities": adapter.GetCapabilities(),
		}
	}
	
	return Result{
		TaskID:  task.ID,
		Success: true,
		Output: map[string]interface{}{
			"adapters":      healthStatus,
			"total_count":   len(adapters),
			"healthy_count": healthyCount,
			"overall_health": float64(healthyCount) / float64(len(adapters)),
		},
		Metadata: map[string]interface{}{
			"agent_id":   a.id,
			"agent_type": a.agentType,
			"operation":  "adapter_health",
		},
	}, nil
}

func (a *adapterAgent) executeWithAdapter(ctx context.Context, task Task) (Result, error) {
	// Buscar adaptadores que puedan manejar la tarea
	capableAdapters := a.registry.GetByCapability(task.Type)
	
	if len(capableAdapters) == 0 {
		return Result{
			TaskID:  task.ID,
			Success: false,
			Error:   fmt.Sprintf("no adapter found capable of handling task type: %s", task.Type),
		}, fmt.Errorf("no capable adapter found")
	}
	
	// Usar el primer adaptador capaz y saludable
	var selectedAdapter adapters.Adapter
	for _, adapter := range capableAdapters {
		if adapter.IsHealthy() {
			selectedAdapter = adapter
			break
		}
	}
	
	if selectedAdapter == nil {
		return Result{
			TaskID:  task.ID,
			Success: false,
			Error:   "no healthy adapter available",
		}, fmt.Errorf("no healthy adapter available")
	}
	
	// Por ahora, solo manejar adaptadores HTTP
	if httpAdapter, ok := selectedAdapter.(adapters.HTTPAdapter); ok {
		return a.executeHTTPWithAdapter(ctx, task, httpAdapter)
	}
	
	return Result{
		TaskID:  task.ID,
		Success: false,
		Error:   fmt.Sprintf("adapter type %s not supported for generic execution", selectedAdapter.GetType()),
	}, fmt.Errorf("unsupported adapter type")
}

func (a *adapterAgent) executeHTTPWithAdapter(ctx context.Context, task Task, httpAdapter adapters.HTTPAdapter) (Result, error) {
	// Reutilizar la lógica de executeHTTPTask pero con el adaptador específico
	request := &adapters.HTTPRequest{
		Method:  "GET",
		Headers: make(map[string]string),
		Params:  make(map[string]interface{}),
	}
	
	// Extraer parámetros (código similar a executeHTTPTask)
	if method, exists := task.Input["method"]; exists {
		if methodStr, ok := method.(string); ok {
			request.Method = methodStr
		}
	}
	
	if url, exists := task.Input["url"]; exists {
		if urlStr, ok := url.(string); ok {
			request.URL = urlStr
		}
	}
	
	// ... resto de la configuración de la solicitud
	
	response, err := httpAdapter.MakeRequest(ctx, request)
	if err != nil {
		return Result{
			TaskID:  task.ID,
			Success: false,
			Error:   fmt.Sprintf("HTTP request failed: %v", err),
		}, err
	}
	
	return Result{
		TaskID:  task.ID,
		Success: response.Success,
		Error:   response.Error,
		Output: map[string]interface{}{
			"status_code":  response.StatusCode,
			"headers":      response.Headers,
			"body":         response.Body,
			"duration":     response.Duration.Milliseconds(),
			"adapter_name": httpAdapter.GetName(),
		},
	}, nil
}

func (a *adapterAgent) CanHandle(taskType string) bool {
	supportedTypes := []string{
		"http_request",
		"api_call",
		"webhook",
		"integration",
		"create_adapter",
		"list_adapters",
		"adapter_health",
	}
	
	for _, supported := range supportedTypes {
		if taskType == supported {
			return true
		}
	}
	
	// También verificar si algún adaptador registrado puede manejar la tarea
	capableAdapters := a.registry.GetByCapability(taskType)
	return len(capableAdapters) > 0
}