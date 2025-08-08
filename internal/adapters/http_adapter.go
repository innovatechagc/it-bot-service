package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/company/bot-service/pkg/logger"
)

// httpAdapter implementa HTTPAdapter
type httpAdapter struct {
	name           string
	version        string
	client         *http.Client
	defaultHeaders map[string]string
	timeout        time.Duration
	logger         logger.Logger
	mu             sync.RWMutex
	healthy        bool
}

// NewHTTPAdapter crea un nuevo adaptador HTTP
func NewHTTPAdapter(name, version string, logger logger.Logger) HTTPAdapter {
	return &httpAdapter{
		name:    name,
		version: version,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		defaultHeaders: make(map[string]string),
		timeout:        30 * time.Second,
		logger:         logger,
		healthy:        false,
	}
}

// GetName devuelve el nombre del adaptador
func (a *httpAdapter) GetName() string {
	return a.name
}

// GetType devuelve el tipo del adaptador
func (a *httpAdapter) GetType() string {
	return "http"
}

// GetVersion devuelve la versión del adaptador
func (a *httpAdapter) GetVersion() string {
	return a.version
}

// Initialize inicializa el adaptador con la configuración
func (a *httpAdapter) Initialize(ctx context.Context, config map[string]interface{}) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	// Configurar timeout si se proporciona
	if timeoutVal, exists := config["timeout"]; exists {
		if timeoutStr, ok := timeoutVal.(string); ok {
			if timeout, err := time.ParseDuration(timeoutStr); err == nil {
				a.timeout = timeout
				a.client.Timeout = timeout
			}
		} else if timeoutFloat, ok := timeoutVal.(float64); ok {
			a.timeout = time.Duration(timeoutFloat) * time.Second
			a.client.Timeout = a.timeout
		}
	}

	// Configurar headers por defecto
	if headersVal, exists := config["default_headers"]; exists {
		if headersMap, ok := headersVal.(map[string]interface{}); ok {
			for k, v := range headersMap {
				if strVal, ok := v.(string); ok {
					a.defaultHeaders[k] = strVal
				}
			}
		}
	}

	// Configurar User-Agent por defecto
	if _, exists := a.defaultHeaders["User-Agent"]; !exists {
		a.defaultHeaders["User-Agent"] = fmt.Sprintf("bot-service-http-adapter/%s", a.version)
	}

	a.logger.Info("HTTP adapter initialized", 
		"name", a.name,
		"timeout", a.timeout,
		"default_headers", len(a.defaultHeaders))

	return nil
}

// Start inicia el adaptador
func (a *httpAdapter) Start(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.healthy = true
	a.logger.Info("HTTP adapter started", "name", a.name)
	return nil
}

// Stop detiene el adaptador
func (a *httpAdapter) Stop(ctx context.Context) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.healthy = false
	a.logger.Info("HTTP adapter stopped", "name", a.name)
	return nil
}

// IsHealthy verifica si el adaptador está saludable
func (a *httpAdapter) IsHealthy() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.healthy
}

// GetCapabilities devuelve las capacidades del adaptador
func (a *httpAdapter) GetCapabilities() []string {
	return []string{
		"http_request",
		"rest_api",
		"webhook",
		"json",
		"xml",
		"form_data",
		"file_upload",
	}
}

// CanHandle verifica si el adaptador puede manejar una operación
func (a *httpAdapter) CanHandle(operation string) bool {
	capabilities := a.GetCapabilities()
	for _, cap := range capabilities {
		if cap == operation {
			return true
		}
	}

	// También puede manejar métodos HTTP específicos
	httpMethods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	operation = strings.ToUpper(operation)
	for _, method := range httpMethods {
		if method == operation {
			return true
		}
	}

	return false
}

// MakeRequest realiza una solicitud HTTP
func (a *httpAdapter) MakeRequest(ctx context.Context, request *HTTPRequest) (*HTTPResponse, error) {
	if !a.IsHealthy() {
		return nil, fmt.Errorf("HTTP adapter is not healthy")
	}

	start := time.Now()

	// Preparar el cuerpo de la solicitud
	var body io.Reader
	var contentType string

	if request.Body != nil {
		switch v := request.Body.(type) {
		case string:
			body = strings.NewReader(v)
			contentType = "text/plain"
		case []byte:
			body = bytes.NewReader(v)
			contentType = "application/octet-stream"
		case map[string]interface{}:
			jsonBody, err := json.Marshal(v)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal JSON body: %w", err)
			}
			body = bytes.NewReader(jsonBody)
			contentType = "application/json"
		default:
			jsonBody, err := json.Marshal(v)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal body: %w", err)
			}
			body = bytes.NewReader(jsonBody)
			contentType = "application/json"
		}
	}

	// Crear la solicitud HTTP
	httpReq, err := http.NewRequestWithContext(ctx, request.Method, request.URL, body)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Agregar headers por defecto
	a.mu.RLock()
	for k, v := range a.defaultHeaders {
		httpReq.Header.Set(k, v)
	}
	a.mu.RUnlock()

	// Agregar Content-Type si se determinó automáticamente
	if contentType != "" && httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", contentType)
	}

	// Agregar headers específicos de la solicitud
	for k, v := range request.Headers {
		httpReq.Header.Set(k, v)
	}

	// Agregar parámetros de consulta
	if len(request.Params) > 0 {
		q := httpReq.URL.Query()
		for k, v := range request.Params {
			q.Add(k, fmt.Sprintf("%v", v))
		}
		httpReq.URL.RawQuery = q.Encode()
	}

	// Configurar timeout específico si se proporciona
	client := a.client
	if request.Timeout > 0 {
		client = &http.Client{
			Timeout: request.Timeout,
		}
	}

	// Realizar la solicitud
	a.logger.Info("Making HTTP request", 
		"method", request.Method,
		"url", request.URL,
		"headers", len(request.Headers))

	resp, err := client.Do(httpReq)
	if err != nil {
		duration := time.Since(start)
		a.logger.Error("HTTP request failed", 
			"method", request.Method,
			"url", request.URL,
			"duration", duration,
			"error", err)

		return &HTTPResponse{
			StatusCode: 0,
			Success:    false,
			Error:      err.Error(),
			Duration:   duration,
		}, err
	}
	defer resp.Body.Close()

	// Leer el cuerpo de la respuesta
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		duration := time.Since(start)
		return &HTTPResponse{
			StatusCode: resp.StatusCode,
			Success:    false,
			Error:      fmt.Sprintf("failed to read response body: %v", err),
			Duration:   duration,
		}, err
	}

	duration := time.Since(start)
	success := resp.StatusCode >= 200 && resp.StatusCode < 300

	// Convertir headers de respuesta
	headers := make(map[string]string)
	for k, v := range resp.Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}

	// Intentar parsear el cuerpo como JSON si es apropiado
	var parsedBody interface{}
	contentType = resp.Header.Get("Content-Type")
	if strings.Contains(contentType, "application/json") && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, &parsedBody); err != nil {
			// Si no se puede parsear como JSON, usar como string
			parsedBody = string(respBody)
		}
	} else {
		parsedBody = string(respBody)
	}

	response := &HTTPResponse{
		StatusCode: resp.StatusCode,
		Headers:    headers,
		Body:       parsedBody,
		Duration:   duration,
		Success:    success,
	}

	if !success {
		response.Error = fmt.Sprintf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	a.logger.Info("HTTP request completed", 
		"method", request.Method,
		"url", request.URL,
		"status_code", resp.StatusCode,
		"duration", duration,
		"success", success)

	return response, nil
}

// SetDefaultHeaders establece headers por defecto
func (a *httpAdapter) SetDefaultHeaders(headers map[string]string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	for k, v := range headers {
		a.defaultHeaders[k] = v
	}

	a.logger.Info("Default headers updated", 
		"name", a.name,
		"headers_count", len(a.defaultHeaders))
}

// SetTimeout establece el timeout por defecto
func (a *httpAdapter) SetTimeout(timeout time.Duration) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.timeout = timeout
	a.client.Timeout = timeout

	a.logger.Info("Timeout updated", 
		"name", a.name,
		"timeout", timeout)
}