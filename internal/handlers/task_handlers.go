package handlers

import (
	"net/http"
	"strconv"

	"github.com/company/bot-service/internal/domain"
	"github.com/company/bot-service/internal/services"
	"github.com/company/bot-service/pkg/logger"
	"github.com/gin-gonic/gin"
)

// TaskHandler maneja las operaciones relacionadas con tareas asíncronas
type TaskHandler struct {
	taskManager services.TaskManager
	logger      logger.Logger
}

// NewTaskHandler crea un nuevo handler de tareas
func NewTaskHandler(taskManager services.TaskManager, logger logger.Logger) *TaskHandler {
	return &TaskHandler{
		taskManager: taskManager,
		logger:      logger,
	}
}

// SubmitTask godoc
// @Summary Enviar tarea asíncrona
// @Description Envía una tarea para ejecución asíncrona
// @Tags tasks
// @Accept json
// @Produce json
// @Param task body domain.AsyncTask true "Tarea a ejecutar"
// @Success 202 {object} domain.APIResponse
// @Router /tasks [post]
func (h *TaskHandler) SubmitTask(c *gin.Context) {
	var task domain.AsyncTask
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Invalid task data: " + err.Error(),
		})
		return
	}

	// Establecer valores por defecto
	if task.Priority == 0 {
		task.Priority = 5
	}
	if task.Timeout == 0 {
		task.Timeout = 30000 // 30 segundos
	}

	if err := h.taskManager.SubmitTask(c.Request.Context(), &task); err != nil {
		h.logger.Error("Failed to submit task", "error", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to submit task: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusAccepted, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Task submitted successfully",
		Data: map[string]interface{}{
			"task_id": task.ID,
			"status":  task.Status,
		},
	})
}

// GetTask godoc
// @Summary Obtener tarea
// @Description Obtiene el estado y resultado de una tarea asíncrona
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Success 200 {object} domain.APIResponse
// @Router /tasks/{id} [get]
func (h *TaskHandler) GetTask(c *gin.Context) {
	taskID := c.Param("id")

	task, err := h.taskManager.GetTask(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, domain.APIResponse{
			Code:    "NOT_FOUND",
			Message: "Task not found",
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Task retrieved successfully",
		Data:    task,
	})
}

// ListTasks godoc
// @Summary Listar tareas
// @Description Lista tareas asíncronas con filtros opcionales
// @Tags tasks
// @Accept json
// @Produce json
// @Param status query string false "Filter by status"
// @Param type query string false "Filter by type"
// @Param user_id query string false "Filter by user ID"
// @Param bot_id query string false "Filter by bot ID"
// @Param limit query int false "Limit results"
// @Param offset query int false "Offset results"
// @Success 200 {object} domain.APIResponse
// @Router /tasks [get]
func (h *TaskHandler) ListTasks(c *gin.Context) {
	filters := &services.TaskFilters{}

	// Parsear filtros de query parameters
	if status := c.Query("status"); status != "" {
		taskStatus := domain.TaskStatus(status)
		filters.Status = &taskStatus
	}

	if taskType := c.Query("type"); taskType != "" {
		filters.Type = &taskType
	}

	if userID := c.Query("user_id"); userID != "" {
		filters.UserID = &userID
	}

	if botID := c.Query("bot_id"); botID != "" {
		filters.BotID = &botID
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filters.Limit = limit
		}
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil {
			filters.Offset = offset
		}
	}

	tasks, err := h.taskManager.ListTasks(c.Request.Context(), filters)
	if err != nil {
		h.logger.Error("Failed to list tasks", "error", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to list tasks",
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Tasks retrieved successfully",
		Data: map[string]interface{}{
			"tasks": tasks,
			"count": len(tasks),
		},
	})
}

// CancelTask godoc
// @Summary Cancelar tarea
// @Description Cancela una tarea asíncrona
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path string true "Task ID"
// @Success 200 {object} domain.APIResponse
// @Router /tasks/{id}/cancel [post]
func (h *TaskHandler) CancelTask(c *gin.Context) {
	taskID := c.Param("id")

	if err := h.taskManager.CancelTask(c.Request.Context(), taskID); err != nil {
		h.logger.Error("Failed to cancel task", "task_id", taskID, "error", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Failed to cancel task: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Task cancelled successfully",
	})
}

// GetTaskStats godoc
// @Summary Obtener estadísticas de tareas
// @Description Obtiene estadísticas del sistema de tareas asíncronas
// @Tags tasks
// @Accept json
// @Produce json
// @Success 200 {object} domain.APIResponse
// @Router /tasks/stats [get]
func (h *TaskHandler) GetTaskStats(c *gin.Context) {
	stats := h.taskManager.GetStats()

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Task statistics retrieved successfully",
		Data:    stats,
	})
}

// SetupTaskRoutes configura las rutas relacionadas con tareas asíncronas
func SetupTaskRoutes(router *gin.RouterGroup, handler *TaskHandler) {
	// Task Management
	router.POST("/tasks", handler.SubmitTask)
	router.GET("/tasks", handler.ListTasks)
	router.GET("/tasks/:id", handler.GetTask)
	router.POST("/tasks/:id/cancel", handler.CancelTask)
	
	// Statistics
	router.GET("/tasks/stats", handler.GetTaskStats)
}