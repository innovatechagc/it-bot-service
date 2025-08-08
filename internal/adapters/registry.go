package adapters

import (
	"fmt"
	"sync"

	"github.com/company/bot-service/pkg/logger"
)

// adapterRegistry implementa AdapterRegistry
type adapterRegistry struct {
	adapters map[string]Adapter
	mu       sync.RWMutex
	logger   logger.Logger
}

// NewAdapterRegistry crea un nuevo registro de adaptadores
func NewAdapterRegistry(logger logger.Logger) AdapterRegistry {
	return &adapterRegistry{
		adapters: make(map[string]Adapter),
		logger:   logger,
	}
}

// Register registra un adaptador
func (r *adapterRegistry) Register(name string, adapter Adapter) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.adapters[name]; exists {
		return fmt.Errorf("adapter with name '%s' already registered", name)
	}

	r.adapters[name] = adapter
	r.logger.Info("Adapter registered", 
		"name", name,
		"type", adapter.GetType(),
		"version", adapter.GetVersion())

	return nil
}

// Unregister desregistra un adaptador
func (r *adapterRegistry) Unregister(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	adapter, exists := r.adapters[name]
	if !exists {
		return fmt.Errorf("adapter with name '%s' not found", name)
	}

	delete(r.adapters, name)
	r.logger.Info("Adapter unregistered", 
		"name", name,
		"type", adapter.GetType())

	return nil
}

// Get obtiene un adaptador por nombre
func (r *adapterRegistry) Get(name string) (Adapter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	adapter, exists := r.adapters[name]
	if !exists {
		return nil, fmt.Errorf("adapter with name '%s' not found", name)
	}

	return adapter, nil
}

// List devuelve todos los adaptadores registrados
func (r *adapterRegistry) List() []Adapter {
	r.mu.RLock()
	defer r.mu.RUnlock()

	adapters := make([]Adapter, 0, len(r.adapters))
	for _, adapter := range r.adapters {
		adapters = append(adapters, adapter)
	}

	return adapters
}

// GetByType devuelve adaptadores por tipo
func (r *adapterRegistry) GetByType(adapterType string) []Adapter {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []Adapter
	for _, adapter := range r.adapters {
		if adapter.GetType() == adapterType {
			result = append(result, adapter)
		}
	}

	return result
}

// GetByCapability devuelve adaptadores que tienen una capacidad específica
func (r *adapterRegistry) GetByCapability(capability string) []Adapter {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []Adapter
	for _, adapter := range r.adapters {
		if adapter.CanHandle(capability) {
			result = append(result, adapter)
		}
	}

	return result
}

// adapterFactory implementa AdapterFactory
type adapterFactory struct {
	logger logger.Logger
}

// NewAdapterFactory crea una nueva factory de adaptadores
func NewAdapterFactory(logger logger.Logger) AdapterFactory {
	return &adapterFactory{
		logger: logger,
	}
}

// CreateAdapter crea un adaptador basado en el tipo
func (f *adapterFactory) CreateAdapter(adapterType string, config map[string]interface{}) (Adapter, error) {
	switch adapterType {
	case "http":
		return f.createHTTPAdapter(config)
	case "webhook":
		return f.createWebhookAdapter(config)
	default:
		return nil, fmt.Errorf("unsupported adapter type: %s", adapterType)
	}
}

// GetSupportedTypes devuelve los tipos de adaptadores soportados
func (f *adapterFactory) GetSupportedTypes() []string {
	return []string{
		"http",
		"webhook",
		"database",
		"message_queue",
	}
}

// ValidateConfig valida la configuración de un adaptador
func (f *adapterFactory) ValidateConfig(adapterType string, config map[string]interface{}) error {
	switch adapterType {
	case "http":
		return f.validateHTTPConfig(config)
	case "webhook":
		return f.validateWebhookConfig(config)
	case "database":
		return f.validateDatabaseConfig(config)
	case "message_queue":
		return f.validateMessageQueueConfig(config)
	default:
		return fmt.Errorf("unsupported adapter type: %s", adapterType)
	}
}

// createHTTPAdapter crea un adaptador HTTP
func (f *adapterFactory) createHTTPAdapter(config map[string]interface{}) (Adapter, error) {
	name, _ := config["name"].(string)
	if name == "" {
		name = "default-http-adapter"
	}

	version, _ := config["version"].(string)
	if version == "" {
		version = "1.0"
	}

	adapter := NewHTTPAdapter(name, version, f.logger)
	return adapter, nil
}

// createWebhookAdapter crea un adaptador de webhook (placeholder)
func (f *adapterFactory) createWebhookAdapter(config map[string]interface{}) (Adapter, error) {
	// TODO: Implementar adaptador de webhook
	return nil, fmt.Errorf("webhook adapter not implemented yet")
}

// validateHTTPConfig valida la configuración HTTP
func (f *adapterFactory) validateHTTPConfig(config map[string]interface{}) error {
	// La configuración HTTP es opcional, solo validar tipos si están presentes
	if timeout, exists := config["timeout"]; exists {
		switch timeout.(type) {
		case string, float64, int:
			// Tipos válidos
		default:
			return fmt.Errorf("timeout must be a string, number, or duration")
		}
	}

	if headers, exists := config["default_headers"]; exists {
		if _, ok := headers.(map[string]interface{}); !ok {
			return fmt.Errorf("default_headers must be an object")
		}
	}

	return nil
}

// validateWebhookConfig valida la configuración de webhook
func (f *adapterFactory) validateWebhookConfig(config map[string]interface{}) error {
	// TODO: Implementar validación de webhook
	return fmt.Errorf("webhook validation not implemented yet")
}

// validateDatabaseConfig valida la configuración de base de datos
func (f *adapterFactory) validateDatabaseConfig(config map[string]interface{}) error {
	// TODO: Implementar validación de base de datos
	return fmt.Errorf("database validation not implemented yet")
}

// validateMessageQueueConfig valida la configuración de cola de mensajes
func (f *adapterFactory) validateMessageQueueConfig(config map[string]interface{}) error {
	// TODO: Implementar validación de cola de mensajes
	return fmt.Errorf("message queue validation not implemented yet")
}