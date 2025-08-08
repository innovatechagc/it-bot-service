package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/company/bot-service/internal/domain"
	"github.com/company/bot-service/internal/mcp"
	"github.com/company/bot-service/pkg/logger"
	"github.com/gin-gonic/gin"
)

type MCPHandler struct {
	orchestrator interface {
		mcp.MCPOrchestrator
		mcp.MCPDomainOrchestrator
	}
	logger logger.Logger
}

func NewMCPHandler(orchestrator interface {
	mcp.MCPOrchestrator
	mcp.MCPDomainOrchestrator
}, logger logger.Logger) *MCPHandler {
	return &MCPHandler{
		orchestrator: orchestrator,
		logger:       logger,
	}
}

// CreateAgent godoc
// @Summary Crear agente MCP
// @Description Instancia un nuevo agente MCP con la configuración especificada
// @Tags mcp
// @Accept json
// @Produce json
// @Param agent body mcp.MCPConfig true "Configuración del agente"
// @Success 201 {object} domain.APIResponse
// @Router /mcp/agents [post]
func (h *MCPHandler) CreateAgent(c *gin.Context) {
	var config mcp.MCPConfig
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid agent configuration: " + err.Error(),
		})
		return
	}

	// Establecer valores por defecto
	if config.Version == "" {
		config.Version = "1.0"
	}
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	agent, err := h.orchestrator.InstantiateMCP(c.Request.Context(), config)
	if err != nil {
		h.logger.Error("Failed to create MCP agent", "error", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to create agent: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Agent created successfully",
		Data: map[string]interface{}{
			"agent_id":     agent.GetID(),
			"type":         agent.GetType(),
			"capabilities": agent.GetCapabilities(),
			"state":        agent.GetState(),
		},
	})
}

// ListAgents godoc
// @Summary Listar agentes MCP
// @Description Obtiene la lista de todos los agentes MCP activos
// @Tags mcp
// @Accept json
// @Produce json
// @Success 200 {object} domain.APIResponse
// @Router /mcp/agents [get]
func (h *MCPHandler) ListAgents(c *gin.Context) {
	agents := h.orchestrator.ListAgents()
	
	agentList := make([]map[string]interface{}, 0, len(agents))
	for _, agent := range agents {
		agentList = append(agentList, map[string]interface{}{
			"agent_id":     agent.GetID(),
			"type":         agent.GetType(),
			"capabilities": agent.GetCapabilities(),
			"state":        agent.GetState(),
			"healthy":      agent.IsHealthy(),
		})
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Agents retrieved successfully",
		Data: map[string]interface{}{
			"agents": agentList,
			"count":  len(agentList),
		},
	})
}

// GetAgent godoc
// @Summary Obtener agente MCP
// @Description Obtiene los detalles de un agente MCP específico
// @Tags mcp
// @Accept json
// @Produce json
// @Param id path string true "Agent ID"
// @Success 200 {object} domain.APIResponse
// @Router /mcp/agents/{id} [get]
func (h *MCPHandler) GetAgent(c *gin.Context) {
	agentID := c.Param("id")
	
	agent, err := h.orchestrator.GetAgent(agentID)
	if err != nil {
		c.JSON(http.StatusNotFound, domain.APIResponse{
			Code:    "NOT_FOUND",
			Message: "Agent not found",
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Agent retrieved successfully",
		Data: map[string]interface{}{
			"agent_id":     agent.GetID(),
			"type":         agent.GetType(),
			"capabilities": agent.GetCapabilities(),
			"state":        agent.GetState(),
			"context":      agent.GetContext(),
			"healthy":      agent.IsHealthy(),
		},
	})
}

// TerminateAgent godoc
// @Summary Terminar agente MCP
// @Description Termina y elimina un agente MCP específico
// @Tags mcp
// @Accept json
// @Produce json
// @Param id path string true "Agent ID"
// @Success 200 {object} domain.APIResponse
// @Router /mcp/agents/{id} [delete]
func (h *MCPHandler) TerminateAgent(c *gin.Context) {
	agentID := c.Param("id")
	
	if err := h.orchestrator.TerminateAgent(c.Request.Context(), agentID); err != nil {
		h.logger.Error("Failed to terminate agent", "agent_id", agentID, "error", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to terminate agent",
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Agent terminated successfully",
	})
}

// ExecuteTask godoc
// @Summary Ejecutar tarea en agente MCP
// @Description Ejecuta una tarea específica usando el sistema de orquestación MCP
// @Tags mcp
// @Accept json
// @Produce json
// @Param task body domain.MCPTask true "Tarea a ejecutar"
// @Success 200 {object} domain.APIResponse
// @Router /mcp/tasks [post]
func (h *MCPHandler) ExecuteTask(c *gin.Context) {
	var task domain.MCPTask
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid task data: " + err.Error(),
		})
		return
	}

	// Generar ID si no se proporciona
	if task.ID == "" {
		task.ID = generateTaskID()
	}

	// Establecer prioridad por defecto
	if task.Priority == 0 {
		task.Priority = 5
	}

	// Establecer timeout por defecto
	if task.Timeout == 0 {
		task.Timeout = 30000 // 30 segundos en milliseconds
	}

	// Establecer timestamp
	task.CreatedAt = time.Now()

	result, err := h.orchestrator.ExecuteTaskDomain(c.Request.Context(), &task)
	if err != nil {
		h.logger.Error("Task execution failed", "task_id", task.ID, "error", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Task execution failed: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Task executed successfully",
		Data:    result,
	})
}

// PassContext godoc
// @Summary Pasar contexto a agente
// @Description Pasa contexto específico a un agente MCP
// @Tags mcp
// @Accept json
// @Produce json
// @Param id path string true "Agent ID"
// @Param context body map[string]interface{} true "Contexto a pasar"
// @Success 200 {object} domain.APIResponse
// @Router /mcp/agents/{id}/context [post]
func (h *MCPHandler) PassContext(c *gin.Context) {
	agentID := c.Param("id")
	
	var context map[string]interface{}
	if err := c.ShouldBindJSON(&context); err != nil {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid context data: " + err.Error(),
		})
		return
	}

	if err := h.orchestrator.PassContext(c.Request.Context(), agentID, context); err != nil {
		h.logger.Error("Failed to pass context", "agent_id", agentID, "error", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to pass context: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Context passed successfully",
	})
}

// GetAgentMetrics godoc
// @Summary Obtener métricas de agente
// @Description Obtiene las métricas de rendimiento de un agente específico
// @Tags mcp
// @Accept json
// @Produce json
// @Param id path string true "Agent ID"
// @Success 200 {object} domain.APIResponse
// @Router /mcp/agents/{id}/metrics [get]
func (h *MCPHandler) GetAgentMetrics(c *gin.Context) {
	agentID := c.Param("id")
	
	metrics, err := h.orchestrator.GetAgentMetricsDomain(agentID)
	if err != nil {
		c.JSON(http.StatusNotFound, domain.APIResponse{
			Code:    "NOT_FOUND",
			Message: "Agent not found",
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Agent metrics retrieved successfully",
		Data:    metrics,
	})
}

// GetSystemMetrics godoc
// @Summary Obtener métricas del sistema MCP
// @Description Obtiene las métricas generales del sistema de orquestación MCP
// @Tags mcp
// @Accept json
// @Produce json
// @Success 200 {object} domain.APIResponse
// @Router /mcp/metrics [get]
func (h *MCPHandler) GetSystemMetrics(c *gin.Context) {
	metrics, err := h.orchestrator.GetSystemMetricsDomain()
	if err != nil {
		h.logger.Error("Failed to get system metrics", "error", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to get system metrics",
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "System metrics retrieved successfully",
		Data:    metrics,
	})
}

// GetSupportedAgentTypes godoc
// @Summary Obtener tipos de agentes soportados
// @Description Obtiene la lista de tipos de agentes MCP soportados
// @Tags mcp
// @Accept json
// @Produce json
// @Success 200 {object} domain.APIResponse
// @Router /mcp/agent-types [get]
func (h *MCPHandler) GetSupportedAgentTypes(c *gin.Context) {
	types := h.orchestrator.GetSupportedAgentTypes()
	
	// Crear información detallada para cada tipo
	supportedTypes := make([]map[string]interface{}, 0, len(types))
	
	for _, agentType := range types {
		var description string
		var capabilities []string
		var configRequired map[string]string
		
		switch agentType {
		case "ai":
			description = "Agent that uses AI services for text generation and analysis"
			capabilities = []string{"text_generation", "conversation", "analysis", "summarization"}
			configRequired = map[string]string{
				"openai_api_key": "OpenAI API key (optional, uses mock if not provided)",
				"model":          "AI model to use (e.g., gpt-3.5-turbo)",
			}
		case "http":
			description = "Agent that makes HTTP requests to external APIs"
			capabilities = []string{"http_request", "api_call", "webhook", "integration"}
			configRequired = map[string]string{
				"base_url": "Base URL for HTTP requests",
				"headers":  "Default headers for requests (optional)",
			}
		case "workflow":
			description = "Agent that executes sequential workflow steps"
			capabilities = []string{"workflow", "sequence", "orchestration", "automation"}
			configRequired = map[string]string{
				"steps": "Array of workflow steps to execute",
			}
		case "mock":
			description = "Mock agent for testing and development"
			capabilities = []string{"mock", "test", "development", "simulation"}
			configRequired = map[string]string{}
		default:
			description = "Custom agent type"
			capabilities = []string{"custom"}
			configRequired = map[string]string{}
		}
		
		supportedTypes = append(supportedTypes, map[string]interface{}{
			"type":            agentType,
			"description":     description,
			"capabilities":    capabilities,
			"config_required": configRequired,
		})
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Supported agent types retrieved successfully",
		Data: map[string]interface{}{
			"types": supportedTypes,
			"count": len(supportedTypes),
		},
	})
}

// Función auxiliar para generar IDs de tarea
func generateTaskID() string {
	return fmt.Sprintf("task-%d", time.Now().UnixNano())
}

// SetupMCPRoutes configura todas las rutas relacionadas con MCP
func SetupMCPRoutes(router *gin.RouterGroup, handler *MCPHandler) {
	// MCP Agent Management
	router.GET("/mcp/agents", handler.ListAgents)
	router.POST("/mcp/agents", handler.CreateAgent)
	router.GET("/mcp/agents/:id", handler.GetAgent)
	router.DELETE("/mcp/agents/:id", handler.TerminateAgent)
	
	// Agent Context Management
	router.POST("/mcp/agents/:id/context", handler.PassContext)
	
	// Task Execution
	router.POST("/mcp/tasks", handler.ExecuteTask)
	
	// Metrics and Monitoring
	router.GET("/mcp/agents/:id/metrics", handler.GetAgentMetrics)
	router.GET("/mcp/metrics", handler.GetSystemMetrics)
	
	// Agent Types Information
	router.GET("/mcp/agent-types", handler.GetSupportedAgentTypes)
}