package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/company/bot-service/pkg/logger"
)

// HTTPAgent implementa un agente que hace llamadas HTTP
type httpAgent struct {
	*baseAgent
	client  *http.Client
	baseURL string
	headers map[string]string
	config  MCPConfig
}

// NewHTTPAgent crea un nuevo agente HTTP
func NewHTTPAgent(config MCPConfig, logger logger.Logger) (Agent, error) {
	base := newBaseAgent(config, logger)
	base.capabilities = []string{"http_request", "api_call", "webhook", "integration"}
	
	// Obtener configuración
	baseURL, _ := config.Config["base_url"].(string)
	
	// Configurar headers por defecto
	headers := map[string]string{
		"Content-Type": "application/json",
		"User-Agent":   fmt.Sprintf("it-bot-service-agent/%s", config.Version),
	}
	
	// Agregar headers personalizados
	if customHeaders, exists := config.Config["headers"]; exists {
		if headerMap, ok := customHeaders.(map[string]interface{}); ok {
			for k, v := range headerMap {
				if strVal, ok := v.(string); ok {
					headers[k] = strVal
				}
			}
		}
	}
	
	// Configurar timeout
	timeout := 30 * time.Second
	if config.Timeout > 0 {
		timeout = config.Timeout
	}
	
	return &httpAgent{
		baseAgent: base,
		client: &http.Client{
			Timeout: timeout,
		},
		baseURL: baseURL,
		headers: headers,
		config:  config,
	}, nil
}

func (a *httpAgent) Execute(ctx context.Context, task Task) (Result, error) {
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
	
	a.logger.Info("HTTP agent executing task", 
		"agent_id", a.id,
		"task_id", task.ID,
		"task_type", task.Type)
	
	// Construir request HTTP
	req, err := a.buildHTTPRequest(ctx, task)
	if err != nil {
		duration := time.Since(start)
		a.updateMetrics(false, duration)
		
		return Result{
			TaskID:   task.ID,
			Success:  false,
			Error:    fmt.Sprintf("failed to build request: %v", err),
			Duration: duration,
			Metadata: map[string]interface{}{
				"agent_id":   a.id,
				"agent_type": a.agentType,
			},
		}, err
	}
	
	// Ejecutar request
	resp, err := a.client.Do(req)
	if err != nil {
		duration := time.Since(start)
		a.updateMetrics(false, duration)
		
		return Result{
			TaskID:   task.ID,
			Success:  false,
			Error:    fmt.Sprintf("HTTP request failed: %v", err),
			Duration: duration,
			Metadata: map[string]interface{}{
				"agent_id":   a.id,
				"agent_type": a.agentType,
			},
		}, err
	}
	defer resp.Body.Close()
	
	// Leer respuesta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		duration := time.Since(start)
		a.updateMetrics(false, duration)
		
		return Result{
			TaskID:   task.ID,
			Success:  false,
			Error:    fmt.Sprintf("failed to read response: %v", err),
			Duration: duration,
			Metadata: map[string]interface{}{
				"agent_id":   a.id,
				"agent_type": a.agentType,
			},
		}, err
	}
	
	duration := time.Since(start)
	success := resp.StatusCode >= 200 && resp.StatusCode < 300
	a.updateMetrics(success, duration)
	
	// Parsear respuesta JSON si es posible
	var responseData interface{}
	if len(body) > 0 {
		if err := json.Unmarshal(body, &responseData); err != nil {
			// Si no es JSON válido, usar como string
			responseData = string(body)
		}
	}
	
	result := Result{
		TaskID:  task.ID,
		Success: success,
		Output: map[string]interface{}{
			"status_code": resp.StatusCode,
			"headers":     resp.Header,
			"body":        responseData,
			"url":         req.URL.String(),
			"method":      req.Method,
		},
		Duration: duration,
		Metadata: map[string]interface{}{
			"agent_id":     a.id,
			"agent_type":   a.agentType,
			"content_type": resp.Header.Get("Content-Type"),
		},
	}
	
	if !success {
		result.Error = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}
	
	a.logger.Info("HTTP agent task completed", 
		"agent_id", a.id,
		"task_id", task.ID,
		"duration", duration,
		"status_code", resp.StatusCode,
		"success", success)
	
	return result, nil
}

func (a *httpAgent) CanHandle(taskType string) bool {
	supportedTypes := []string{
		"http_request",
		"api_call",
		"webhook",
		"integration",
		"get",
		"post",
		"put",
		"patch",
		"delete",
	}
	
	for _, supported := range supportedTypes {
		if taskType == supported {
			return true
		}
	}
	
	return false
}

func (a *httpAgent) buildHTTPRequest(ctx context.Context, task Task) (*http.Request, error) {
	// Determinar método HTTP
	method := "POST"
	if methodVal, exists := task.Input["method"]; exists {
		if methodStr, ok := methodVal.(string); ok {
			method = methodStr
		}
	}
	
	// Determinar endpoint
	endpoint := ""
	if endpointVal, exists := task.Input["endpoint"]; exists {
		if endpointStr, ok := endpointVal.(string); ok {
			endpoint = endpointStr
		}
	}
	
	// Construir URL completa
	url := a.baseURL
	if endpoint != "" {
		if endpoint[0] != '/' && url[len(url)-1] != '/' {
			url += "/"
		}
		url += endpoint
	}
	
	// Preparar body
	var body io.Reader
	if bodyData, exists := task.Input["body"]; exists {
		if bodyBytes, err := json.Marshal(bodyData); err == nil {
			body = bytes.NewReader(bodyBytes)
		}
	}
	
	// Crear request
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	
	// Agregar headers por defecto
	for key, value := range a.headers {
		req.Header.Set(key, value)
	}
	
	// Agregar headers específicos de la tarea
	if headers, exists := task.Input["headers"]; exists {
		if headerMap, ok := headers.(map[string]interface{}); ok {
			for key, value := range headerMap {
				if strVal, ok := value.(string); ok {
					req.Header.Set(key, strVal)
				}
			}
		}
	}
	
	// Agregar query parameters
	if params, exists := task.Input["params"]; exists {
		if paramMap, ok := params.(map[string]interface{}); ok {
			q := req.URL.Query()
			for key, value := range paramMap {
				q.Add(key, fmt.Sprintf("%v", value))
			}
			req.URL.RawQuery = q.Encode()
		}
	}
	
	return req, nil
}