package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/company/bot-service/internal/domain"
	"github.com/company/bot-service/internal/services"
	"github.com/company/bot-service/pkg/logger"
	"github.com/gin-gonic/gin"
)

type BotHandler struct {
	botService         services.BotService
	flowService        services.BotFlowService
	stepService        services.BotStepService
	smartReplyService  services.SmartReplyService
	conversationService services.ConversationService
	logger             logger.Logger
}

func NewBotHandler(
	botService services.BotService,
	flowService services.BotFlowService,
	stepService services.BotStepService,
	smartReplyService services.SmartReplyService,
	conversationService services.ConversationService,
	logger logger.Logger,
) *BotHandler {
	return &BotHandler{
		botService:         botService,
		flowService:        flowService,
		stepService:        stepService,
		smartReplyService:  smartReplyService,
		conversationService: conversationService,
		logger:             logger,
	}
}

// Bot endpoints

// GetBots godoc
// @Summary Lista bots por usuario o tenant
// @Description Obtiene la lista de bots del usuario autenticado
// @Tags bots
// @Accept json
// @Produce json
// @Param owner_id query string false "ID del propietario"
// @Success 200 {object} domain.APIResponse
// @Router /bots [get]
func (h *BotHandler) GetBots(c *gin.Context) {
	ownerID := c.Query("owner_id")
	if ownerID == "" {
		// Obtener del JWT token (implementar middleware de auth)
		ownerID = c.GetString("user_id")
	}

	if ownerID == "" {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Owner ID is required",
		})
		return
	}

	bots, err := h.botService.GetBotsByOwner(c.Request.Context(), ownerID)
	if err != nil {
		h.logger.Error("Failed to get bots", "owner_id", ownerID, "error", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to retrieve bots",
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Bots retrieved successfully",
		Data:    bots,
	})
}

// GetBot godoc
// @Summary Detalle de un bot
// @Description Obtiene los detalles de un bot específico
// @Tags bots
// @Accept json
// @Produce json
// @Param id path string true "Bot ID"
// @Success 200 {object} domain.APIResponse
// @Router /bots/{id} [get]
func (h *BotHandler) GetBot(c *gin.Context) {
	id := c.Param("id")
	
	bot, err := h.botService.GetBot(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get bot", "bot_id", id, "error", err)
		c.JSON(http.StatusNotFound, domain.APIResponse{
			Code:    "NOT_FOUND",
			Message: "Bot not found",
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Bot retrieved successfully",
		Data:    bot,
	})
}

// CreateBot godoc
// @Summary Crear bot
// @Description Crea un nuevo bot conversacional
// @Tags bots
// @Accept json
// @Produce json
// @Param bot body domain.Bot true "Bot data"
// @Success 201 {object} domain.APIResponse
// @Router /bots [post]
func (h *BotHandler) CreateBot(c *gin.Context) {
	var bot domain.Bot
	if err := c.ShouldBindJSON(&bot); err != nil {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid bot data: " + err.Error(),
		})
		return
	}

	// Generar ID si no se proporciona
	if bot.ID == "" {
		bot.ID = generateUUID()
	}

	// Establecer propietario desde el token JWT
	if bot.OwnerID == "" {
		bot.OwnerID = c.GetString("user_id")
	}

	if err := h.botService.CreateBot(c.Request.Context(), &bot); err != nil {
		h.logger.Error("Failed to create bot", "bot", bot, "error", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to create bot",
		})
		return
	}

	c.JSON(http.StatusCreated, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Bot created successfully",
		Data:    bot,
	})
}

// UpdateBot godoc
// @Summary Editar bot
// @Description Actualiza un bot existente
// @Tags bots
// @Accept json
// @Produce json
// @Param id path string true "Bot ID"
// @Param bot body domain.Bot true "Bot data"
// @Success 200 {object} domain.APIResponse
// @Router /bots/{id} [patch]
func (h *BotHandler) UpdateBot(c *gin.Context) {
	id := c.Param("id")
	
	var bot domain.Bot
	if err := c.ShouldBindJSON(&bot); err != nil {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid bot data: " + err.Error(),
		})
		return
	}

	bot.ID = id
	if err := h.botService.UpdateBot(c.Request.Context(), &bot); err != nil {
		h.logger.Error("Failed to update bot", "bot_id", id, "error", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to update bot",
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Bot updated successfully",
		Data:    bot,
	})
}

// DeleteBot godoc
// @Summary Eliminar o desactivar bot
// @Description Elimina o desactiva un bot
// @Tags bots
// @Accept json
// @Produce json
// @Param id path string true "Bot ID"
// @Success 200 {object} domain.APIResponse
// @Router /bots/{id} [delete]
func (h *BotHandler) DeleteBot(c *gin.Context) {
	id := c.Param("id")
	
	if err := h.botService.DeleteBot(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to delete bot", "bot_id", id, "error", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to delete bot",
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Bot deleted successfully",
	})
}

// Flow endpoints

// GetFlows godoc
// @Summary Lista flujos del bot
// @Description Obtiene todos los flujos de un bot específico
// @Tags flows
// @Accept json
// @Produce json
// @Param id path string true "Bot ID"
// @Success 200 {object} domain.APIResponse
// @Router /bots/{id}/flows [get]
func (h *BotHandler) GetFlows(c *gin.Context) {
	botID := c.Param("id")
	
	flows, err := h.flowService.GetFlowsByBot(c.Request.Context(), botID)
	if err != nil {
		h.logger.Error("Failed to get flows", "bot_id", botID, "error", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to retrieve flows",
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Flows retrieved successfully",
		Data:    flows,
	})
}

// CreateFlow godoc
// @Summary Crear flujo conversacional
// @Description Crea un nuevo flujo de conversación para un bot
// @Tags flows
// @Accept json
// @Produce json
// @Param id path string true "Bot ID"
// @Param flow body domain.BotFlow true "Flow data"
// @Success 201 {object} domain.APIResponse
// @Router /bots/{id}/flows [post]
func (h *BotHandler) CreateFlow(c *gin.Context) {
	botID := c.Param("id")
	
	var flow domain.BotFlow
	if err := c.ShouldBindJSON(&flow); err != nil {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid flow data: " + err.Error(),
		})
		return
	}

	flow.BotID = botID
	if flow.ID == "" {
		flow.ID = generateUUID()
	}

	if err := h.flowService.CreateFlow(c.Request.Context(), &flow); err != nil {
		h.logger.Error("Failed to create flow", "bot_id", botID, "error", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to create flow",
		})
		return
	}

	c.JSON(http.StatusCreated, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Flow created successfully",
		Data:    flow,
	})
}

// GetFlow godoc
// @Summary Obtener un flujo con sus pasos
// @Description Obtiene un flujo específico con todos sus pasos
// @Tags flows
// @Accept json
// @Produce json
// @Param id path string true "Flow ID"
// @Success 200 {object} domain.APIResponse
// @Router /flows/{id} [get]
func (h *BotHandler) GetFlow(c *gin.Context) {
	id := c.Param("id")
	
	flow, err := h.flowService.GetFlow(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get flow", "flow_id", id, "error", err)
		c.JSON(http.StatusNotFound, domain.APIResponse{
			Code:    "NOT_FOUND",
			Message: "Flow not found",
		})
		return
	}

	// Obtener pasos del flujo
	steps, err := h.stepService.GetStepsByFlow(c.Request.Context(), id)
	if err != nil {
		h.logger.Error("Failed to get flow steps", "flow_id", id, "error", err)
		steps = []*domain.BotStep{} // Continuar sin pasos si hay error
	}

	response := map[string]interface{}{
		"flow":  flow,
		"steps": steps,
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Flow retrieved successfully",
		Data:    response,
	})
}

// UpdateFlow godoc
// @Summary Editar un flujo
// @Description Actualiza un flujo existente
// @Tags flows
// @Accept json
// @Produce json
// @Param id path string true "Flow ID"
// @Param flow body domain.BotFlow true "Flow data"
// @Success 200 {object} domain.APIResponse
// @Router /flows/{id} [patch]
func (h *BotHandler) UpdateFlow(c *gin.Context) {
	id := c.Param("id")
	
	var flow domain.BotFlow
	if err := c.ShouldBindJSON(&flow); err != nil {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid flow data: " + err.Error(),
		})
		return
	}

	flow.ID = id
	if err := h.flowService.UpdateFlow(c.Request.Context(), &flow); err != nil {
		h.logger.Error("Failed to update flow", "flow_id", id, "error", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to update flow",
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Flow updated successfully",
		Data:    flow,
	})
}

// DeleteFlow godoc
// @Summary Eliminar un flujo
// @Description Elimina un flujo y todos sus pasos
// @Tags flows
// @Accept json
// @Produce json
// @Param id path string true "Flow ID"
// @Success 200 {object} domain.APIResponse
// @Router /flows/{id} [delete]
func (h *BotHandler) DeleteFlow(c *gin.Context) {
	id := c.Param("id")
	
	if err := h.flowService.DeleteFlow(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to delete flow", "flow_id", id, "error", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to delete flow",
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Flow deleted successfully",
	})
}

// Step endpoints

// CreateStep godoc
// @Summary Agregar paso a un flujo
// @Description Crea un nuevo paso en un flujo específico
// @Tags steps
// @Accept json
// @Produce json
// @Param id path string true "Flow ID"
// @Param step body domain.BotStep true "Step data"
// @Success 201 {object} domain.APIResponse
// @Router /flows/{id}/steps [post]
func (h *BotHandler) CreateStep(c *gin.Context) {
	flowID := c.Param("id")
	
	var step domain.BotStep
	if err := c.ShouldBindJSON(&step); err != nil {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid step data: " + err.Error(),
		})
		return
	}

	step.FlowID = flowID
	if step.ID == "" {
		step.ID = generateUUID()
	}

	if err := h.stepService.CreateStep(c.Request.Context(), &step); err != nil {
		h.logger.Error("Failed to create step", "flow_id", flowID, "error", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to create step",
		})
		return
	}

	c.JSON(http.StatusCreated, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Step created successfully",
		Data:    step,
	})
}

// UpdateStep godoc
// @Summary Editar paso
// @Description Actualiza un paso existente
// @Tags steps
// @Accept json
// @Produce json
// @Param id path string true "Step ID"
// @Param step body domain.BotStep true "Step data"
// @Success 200 {object} domain.APIResponse
// @Router /steps/{id} [patch]
func (h *BotHandler) UpdateStep(c *gin.Context) {
	id := c.Param("id")
	
	var step domain.BotStep
	if err := c.ShouldBindJSON(&step); err != nil {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid step data: " + err.Error(),
		})
		return
	}

	step.ID = id
	if err := h.stepService.UpdateStep(c.Request.Context(), &step); err != nil {
		h.logger.Error("Failed to update step", "step_id", id, "error", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to update step",
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Step updated successfully",
		Data:    step,
	})
}

// DeleteStep godoc
// @Summary Eliminar paso
// @Description Elimina un paso específico
// @Tags steps
// @Accept json
// @Produce json
// @Param id path string true "Step ID"
// @Success 200 {object} domain.APIResponse
// @Router /steps/{id} [delete]
func (h *BotHandler) DeleteStep(c *gin.Context) {
	id := c.Param("id")
	
	if err := h.stepService.DeleteStep(c.Request.Context(), id); err != nil {
		h.logger.Error("Failed to delete step", "step_id", id, "error", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to delete step",
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Step deleted successfully",
	})
}

// Smart Reply endpoints

// SmartReply godoc
// @Summary Consulta rápida a IA
// @Description Genera una respuesta usando IA para un prompt específico
// @Tags smart-replies
// @Accept json
// @Produce json
// @Param id path string true "Bot ID"
// @Param request body map[string]interface{} true "Smart reply request"
// @Success 200 {object} domain.APIResponse
// @Router /bots/{id}/smart-reply [post]
func (h *BotHandler) SmartReply(c *gin.Context) {
	botID := c.Param("id")
	
	var request struct {
		Prompt  string                 `json:"prompt" binding:"required"`
		Context map[string]interface{} `json:"context"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid request data: " + err.Error(),
		})
		return
	}

	if request.Context == nil {
		request.Context = make(map[string]interface{})
	}

	reply, err := h.smartReplyService.GenerateAIResponse(c.Request.Context(), botID, request.Prompt, request.Context)
	if err != nil {
		h.logger.Error("Failed to generate smart reply", "bot_id", botID, "error", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to generate smart reply",
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Smart reply generated successfully",
		Data:    reply,
	})
}

// TrainIntents godoc
// @Summary Entrenar respuestas automáticas
// @Description Entrena el bot con intents y respuestas predefinidas
// @Tags smart-replies
// @Accept json
// @Produce json
// @Param id path string true "Bot ID"
// @Param intents body []domain.SmartReply true "Training intents"
// @Success 200 {object} domain.APIResponse
// @Router /bots/{id}/intents/train [post]
func (h *BotHandler) TrainIntents(c *gin.Context) {
	botID := c.Param("id")
	
	var intents []domain.SmartReply
	if err := c.ShouldBindJSON(&intents); err != nil {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid intents data: " + err.Error(),
		})
		return
	}

	if err := h.smartReplyService.TrainIntents(c.Request.Context(), botID, intents); err != nil {
		h.logger.Error("Failed to train intents", "bot_id", botID, "error", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to train intents",
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Intents trained successfully",
		Data: map[string]interface{}{
			"trained_count": len(intents),
		},
	})
}

// GetIntents godoc
// @Summary Listar intents configurados
// @Description Obtiene todos los intents configurados para un bot
// @Tags smart-replies
// @Accept json
// @Produce json
// @Param id path string true "Bot ID"
// @Success 200 {object} domain.APIResponse
// @Router /bots/{id}/intents [get]
func (h *BotHandler) GetIntents(c *gin.Context) {
	botID := c.Param("id")
	
	intents, err := h.smartReplyService.GetSmartRepliesByBot(c.Request.Context(), botID)
	if err != nil {
		h.logger.Error("Failed to get intents", "bot_id", botID, "error", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to retrieve intents",
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Intents retrieved successfully",
		Data:    intents,
	})
}

// ProcessIncomingMessage godoc
// @Summary Procesar mensaje entrante
// @Description Recibe un mensaje entrante desde messaging-service y responde según flujo
// @Tags messaging
// @Accept json
// @Produce json
// @Param message body domain.IncomingMessage true "Incoming message"
// @Success 200 {object} domain.APIResponse
// @Router /incoming [post]
func (h *BotHandler) ProcessIncomingMessage(c *gin.Context) {
	var message domain.IncomingMessage
	if err := c.ShouldBindJSON(&message); err != nil {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid message data: " + err.Error(),
		})
		return
	}

	// Generar ID si no se proporciona
	if message.ID == "" {
		message.ID = generateUUID()
	}

	response, err := h.botService.ProcessIncomingMessage(c.Request.Context(), &message)
	if err != nil {
		h.logger.Error("Failed to process incoming message", 
			"message_id", message.ID,
			"bot_id", message.BotID,
			"error", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to process message",
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Message processed successfully",
		Data:    response,
	})
}

// Utility functions

func generateUUID() string {
	// Implementación simple de UUID - en producción usar una librería como google/uuid
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// SetupBotRoutes configura todas las rutas relacionadas with bots
func SetupBotRoutes(router *gin.RouterGroup, handler *BotHandler) {
	// Bot routes
	router.GET("/bots", handler.GetBots)
	router.GET("/bots/:id", handler.GetBot)
	router.POST("/bots", handler.CreateBot)
	router.PATCH("/bots/:id", handler.UpdateBot)
	router.DELETE("/bots/:id", handler.DeleteBot)

	// Flow routes
	router.GET("/bots/:id/flows", handler.GetFlows)
	router.POST("/bots/:id/flows", handler.CreateFlow)
	router.GET("/flows/:id", handler.GetFlow)
	router.PATCH("/flows/:id", handler.UpdateFlow)
	router.DELETE("/flows/:id", handler.DeleteFlow)

	// Step routes
	router.POST("/flows/:id/steps", handler.CreateStep)
	router.PATCH("/steps/:id", handler.UpdateStep)
	router.DELETE("/steps/:id", handler.DeleteStep)

	// Smart Reply routes
	router.POST("/bots/:id/smart-reply", handler.SmartReply)
	router.POST("/bots/:id/intents/train", handler.TrainIntents)
	router.GET("/bots/:id/intents", handler.GetIntents)

	// Incoming message processing
	router.POST("/incoming", handler.ProcessIncomingMessage)
}