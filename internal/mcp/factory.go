package mcp

import (
	"fmt"

	"github.com/company/bot-service/internal/adapters"
	"github.com/company/bot-service/pkg/logger"
)

// agentFactory implementa AgentFactory
type agentFactory struct {
	logger          logger.Logger
	adapterRegistry adapters.AdapterRegistry
	adapterFactory  adapters.AdapterFactory
}

// NewAgentFactory crea una nueva factory de agentes
func NewAgentFactory(logger logger.Logger) AgentFactory {
	// Crear registro y factory de adaptadores
	adapterRegistry := adapters.NewAdapterRegistry(logger)
	adapterFactory := adapters.NewAdapterFactory(logger)
	
	return &agentFactory{
		logger:          logger,
		adapterRegistry: adapterRegistry,
		adapterFactory:  adapterFactory,
	}
}

// CreateAgent crea un agente basado en la configuración
func (f *agentFactory) CreateAgent(config MCPConfig) (Agent, error) {
	switch config.Type {
	case "ai":
		return NewAIAgent(config, f.logger)
	case "http":
		return NewHTTPAgent(config, f.logger)
	case "workflow":
		return NewWorkflowAgent(config, f.logger)
	case "adapter":
		return NewAdapterAgent(config, f.adapterRegistry, f.adapterFactory, f.logger)
	case "mock":
		return NewMockAgent(config, f.logger)
	default:
		return nil, fmt.Errorf("unsupported agent type: %s", config.Type)
	}
}

// GetSupportedTypes devuelve los tipos de agentes soportados
func (f *agentFactory) GetSupportedTypes() []string {
	return []string{"ai", "http", "workflow", "adapter", "mock"}
}

// ValidateConfig valida la configuración de un agente
func (f *agentFactory) ValidateConfig(config MCPConfig) error {
	if config.Type == "" {
		return fmt.Errorf("agent type is required")
	}

	if config.Name == "" {
		return fmt.Errorf("agent name is required")
	}

	// Validar que el tipo sea soportado
	supportedTypes := f.GetSupportedTypes()
	typeSupported := false
	for _, supportedType := range supportedTypes {
		if config.Type == supportedType {
			typeSupported = true
			break
		}
	}

	if !typeSupported {
		return fmt.Errorf("unsupported agent type: %s, supported types: %v", config.Type, supportedTypes)
	}

	// Validaciones específicas por tipo
	switch config.Type {
	case "ai":
		return f.validateAIConfig(config)
	case "http":
		return f.validateHTTPConfig(config)
	case "workflow":
		return f.validateWorkflowConfig(config)
	case "adapter":
		return f.validateAdapterConfig(config)
	case "mock":
		return f.validateMockConfig(config)
	}

	return nil
}

// validateAIConfig valida configuración para agentes de IA
func (f *agentFactory) validateAIConfig(config MCPConfig) error {
	if config.Config == nil {
		return fmt.Errorf("AI agent requires config")
	}

	// Verificar que tenga al menos un proveedor de IA configurado
	if _, hasOpenAI := config.Config["openai_api_key"]; !hasOpenAI {
		if _, hasVertex := config.Config["vertex_project"]; !hasVertex {
			return fmt.Errorf("AI agent requires either openai_api_key or vertex_project")
		}
	}

	return nil
}

// validateHTTPConfig valida configuración para agentes HTTP
func (f *agentFactory) validateHTTPConfig(config MCPConfig) error {
	if config.Config == nil {
		return fmt.Errorf("HTTP agent requires config")
	}

	baseURL, exists := config.Config["base_url"]
	if !exists {
		return fmt.Errorf("HTTP agent requires base_url in config")
	}

	if _, ok := baseURL.(string); !ok {
		return fmt.Errorf("base_url must be a string")
	}

	return nil
}

// validateWorkflowConfig valida configuración para agentes de workflow
func (f *agentFactory) validateWorkflowConfig(config MCPConfig) error {
	if config.Config == nil {
		return fmt.Errorf("Workflow agent requires config")
	}

	steps, exists := config.Config["steps"]
	if !exists {
		return fmt.Errorf("Workflow agent requires steps in config")
	}

	if _, ok := steps.([]interface{}); !ok {
		return fmt.Errorf("steps must be an array")
	}

	return nil
}

// validateMockConfig valida configuración para agentes mock
func (f *agentFactory) validateMockConfig(config MCPConfig) error {
	// Los agentes mock no requieren configuración específica
	return nil
}

// validateAdapterConfig valida configuración para agentes de adaptador
func (f *agentFactory) validateAdapterConfig(config MCPConfig) error {
	// Los agentes de adaptador pueden funcionar sin configuración específica
	// ya que pueden crear adaptadores dinámicamente
	return nil
}