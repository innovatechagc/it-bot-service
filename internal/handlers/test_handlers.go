package handlers

import (
	"net/http"

	"github.com/company/bot-service/internal/domain"
	"github.com/company/bot-service/internal/services"
	"github.com/company/bot-service/pkg/logger"
	"github.com/gin-gonic/gin"
)

// TestHandlers maneja las operaciones relacionadas con pruebas
type TestHandlers struct {
	conditionalService services.ConditionalService
	triggerService     services.TriggerService
	testService        services.TestService
	testSuiteService   services.TestSuiteService
	logger             logger.Logger
}

// NewTestHandlers crea una nueva instancia de TestHandlers
func NewTestHandlers(
	conditionalService services.ConditionalService,
	triggerService services.TriggerService,
	testService services.TestService,
	testSuiteService services.TestSuiteService,
	logger logger.Logger,
) *TestHandlers {
	return &TestHandlers{
		conditionalService: conditionalService,
		triggerService:     triggerService,
		testService:        testService,
		testSuiteService:   testSuiteService,
		logger:             logger,
	}
}

// RegisterRoutes registra las rutas de testing
func (h *TestHandlers) RegisterRoutes(router *gin.RouterGroup) {
	// Condicionales
	router.POST("/conditionals", h.CreateConditional)
	router.GET("/conditionals/:id", h.GetConditional)
	router.PUT("/conditionals/:id", h.UpdateConditional)
	router.DELETE("/conditionals/:id", h.DeleteConditional)
	router.GET("/conditionals/bot/:botId", h.GetConditionalsByBot)
	router.POST("/conditionals/:id/evaluate", h.EvaluateConditional)

	// Triggers
	router.POST("/triggers", h.CreateTrigger)
	router.GET("/triggers/:id", h.GetTrigger)
	router.PUT("/triggers/:id", h.UpdateTrigger)
	router.DELETE("/triggers/:id", h.DeleteTrigger)
	router.GET("/triggers/bot/:botId", h.GetTriggersByBot)
	router.POST("/triggers/:id/execute", h.ExecuteTrigger)

	// Casos de prueba
	router.POST("/test-cases", h.CreateTestCase)
	router.GET("/test-cases/:id", h.GetTestCase)
	router.PUT("/test-cases/:id", h.UpdateTestCase)
	router.DELETE("/test-cases/:id", h.DeleteTestCase)
	router.GET("/test-cases/bot/:botId", h.GetTestCasesByBot)
	router.POST("/test-cases/:id/execute", h.ExecuteTestCase)
	router.POST("/test-cases/bulk-execute", h.BulkExecuteTestCases)

	// Suites de prueba
	router.POST("/test-suites", h.CreateTestSuite)
	router.GET("/test-suites/:id", h.GetTestSuite)
	router.PUT("/test-suites/:id", h.UpdateTestSuite)
	router.DELETE("/test-suites/:id", h.DeleteTestSuite)
	router.GET("/test-suites/bot/:botId", h.GetTestSuitesByBot)
	router.POST("/test-suites/:id/execute", h.ExecuteTestSuite)
	router.POST("/test-suites/:id/test-cases", h.AddTestCaseToSuite)
	router.DELETE("/test-suites/:id/test-cases/:testCaseId", h.RemoveTestCaseFromSuite)
}

// CreateConditional crea un nuevo condicional
func (h *TestHandlers) CreateConditional(c *gin.Context) {
	var conditional domain.Conditional
	if err := c.ShouldBindJSON(&conditional); err != nil {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Datos inválidos",
			Data:    err.Error(),
		})
		return
	}

	err := h.conditionalService.CreateConditional(c.Request.Context(), &conditional)
	if err != nil {
		h.logger.Error("Error creating conditional", "error", err)
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Error al crear condicional",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Condicional creado exitosamente",
		Data:    conditional,
	})
}

// GetConditional obtiene un condicional por ID
func (h *TestHandlers) GetConditional(c *gin.Context) {
	id := c.Param("id")
	
	conditional, err := h.conditionalService.GetConditional(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, domain.APIResponse{
			Code:    "NOT_FOUND",
			Message: "Condicional no encontrado",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Condicional encontrado",
		Data:    conditional,
	})
}

// UpdateConditional actualiza un condicional
func (h *TestHandlers) UpdateConditional(c *gin.Context) {
	id := c.Param("id")
	var conditional domain.Conditional
	if err := c.ShouldBindJSON(&conditional); err != nil {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Datos inválidos",
			Data:    err.Error(),
		})
		return
	}

	conditional.ID = id
	err := h.conditionalService.UpdateConditional(c.Request.Context(), &conditional)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Error al actualizar condicional",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Condicional actualizado exitosamente",
		Data:    conditional,
	})
}

// DeleteConditional elimina un condicional
func (h *TestHandlers) DeleteConditional(c *gin.Context) {
	id := c.Param("id")
	
	err := h.conditionalService.DeleteConditional(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Error al eliminar condicional",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Condicional eliminado exitosamente",
		Data:    nil,
	})
}

// GetConditionalsByBot obtiene condicionales por bot
func (h *TestHandlers) GetConditionalsByBot(c *gin.Context) {
	botID := c.Param("botId")
	
	conditionals, err := h.conditionalService.GetConditionalsByBot(c.Request.Context(), botID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Error al obtener condicionales",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Condicionales obtenidos exitosamente",
		Data:    conditionals,
	})
}

// EvaluateConditional evalúa un condicional
func (h *TestHandlers) EvaluateConditional(c *gin.Context) {
	id := c.Param("id")
	var request struct {
		Context map[string]interface{} `json:"context"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Datos inválidos",
			Data:    err.Error(),
		})
		return
	}

	result, err := h.conditionalService.EvaluateConditional(c.Request.Context(), id, request.Context)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Error al evaluar condicional",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Condicional evaluado exitosamente",
		Data:    result,
	})
}

// CreateTrigger crea un nuevo trigger
func (h *TestHandlers) CreateTrigger(c *gin.Context) {
	var trigger domain.Trigger
	if err := c.ShouldBindJSON(&trigger); err != nil {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Datos inválidos",
			Data:    err.Error(),
		})
		return
	}

	err := h.triggerService.CreateTrigger(c.Request.Context(), &trigger)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Error al crear trigger",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Trigger creado exitosamente",
		Data:    trigger,
	})
}

// GetTrigger obtiene un trigger por ID
func (h *TestHandlers) GetTrigger(c *gin.Context) {
	id := c.Param("id")
	
	trigger, err := h.triggerService.GetTrigger(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, domain.APIResponse{
			Code:    "NOT_FOUND",
			Message: "Trigger no encontrado",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Trigger encontrado",
		Data:    trigger,
	})
}

// UpdateTrigger actualiza un trigger
func (h *TestHandlers) UpdateTrigger(c *gin.Context) {
	id := c.Param("id")
	var trigger domain.Trigger
	if err := c.ShouldBindJSON(&trigger); err != nil {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Datos inválidos",
			Data:    err.Error(),
		})
		return
	}

	trigger.ID = id
	err := h.triggerService.UpdateTrigger(c.Request.Context(), &trigger)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Error al actualizar trigger",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Trigger actualizado exitosamente",
		Data:    trigger,
	})
}

// DeleteTrigger elimina un trigger
func (h *TestHandlers) DeleteTrigger(c *gin.Context) {
	id := c.Param("id")
	
	err := h.triggerService.DeleteTrigger(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Error al eliminar trigger",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Trigger eliminado exitosamente",
		Data:    nil,
	})
}

// GetTriggersByBot obtiene triggers por bot
func (h *TestHandlers) GetTriggersByBot(c *gin.Context) {
	botID := c.Param("botId")
	
	triggers, err := h.triggerService.GetTriggersByBot(c.Request.Context(), botID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Error al obtener triggers",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Triggers obtenidos exitosamente",
		Data:    triggers,
	})
}

// ExecuteTrigger ejecuta un trigger
func (h *TestHandlers) ExecuteTrigger(c *gin.Context) {
	id := c.Param("id")
	var request struct {
		Context map[string]interface{} `json:"context"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Datos inválidos",
			Data:    err.Error(),
		})
		return
	}

	err := h.triggerService.ExecuteTrigger(c.Request.Context(), id, request.Context)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Error al ejecutar trigger",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Trigger ejecutado exitosamente",
		Data:    map[string]interface{}{"trigger_id": id, "status": "executed"},
	})
}

// CreateTestCase crea un nuevo caso de prueba
func (h *TestHandlers) CreateTestCase(c *gin.Context) {
	var testCase domain.TestCase
	if err := c.ShouldBindJSON(&testCase); err != nil {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Datos inválidos",
			Data:    err.Error(),
		})
		return
	}

	err := h.testService.CreateTestCase(c.Request.Context(), &testCase)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Error al crear caso de prueba",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Caso de prueba creado exitosamente",
		Data:    testCase,
	})
}

// GetTestCase obtiene un caso de prueba por ID
func (h *TestHandlers) GetTestCase(c *gin.Context) {
	id := c.Param("id")
	
	testCase, err := h.testService.GetTestCase(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, domain.APIResponse{
			Code:    "NOT_FOUND",
			Message: "Caso de prueba no encontrado",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Caso de prueba encontrado",
		Data:    testCase,
	})
}

// UpdateTestCase actualiza un caso de prueba
func (h *TestHandlers) UpdateTestCase(c *gin.Context) {
	id := c.Param("id")
	var testCase domain.TestCase
	if err := c.ShouldBindJSON(&testCase); err != nil {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Datos inválidos",
			Data:    err.Error(),
		})
		return
	}

	testCase.ID = id
	err := h.testService.UpdateTestCase(c.Request.Context(), &testCase)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Error al actualizar caso de prueba",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Caso de prueba actualizado exitosamente",
		Data:    testCase,
	})
}

// DeleteTestCase elimina un caso de prueba
func (h *TestHandlers) DeleteTestCase(c *gin.Context) {
	id := c.Param("id")
	
	err := h.testService.DeleteTestCase(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Error al eliminar caso de prueba",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Caso de prueba eliminado exitosamente",
		Data:    nil,
	})
}

// GetTestCasesByBot obtiene casos de prueba por bot
func (h *TestHandlers) GetTestCasesByBot(c *gin.Context) {
	botID := c.Param("botId")
	
	testCases, err := h.testService.GetTestCasesByBot(c.Request.Context(), botID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Error al obtener casos de prueba",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Casos de prueba obtenidos exitosamente",
		Data:    testCases,
	})
}

// ExecuteTestCase ejecuta un caso de prueba
func (h *TestHandlers) ExecuteTestCase(c *gin.Context) {
	id := c.Param("id")
	
	result, err := h.testService.ExecuteTestCase(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Error al ejecutar caso de prueba",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Caso de prueba ejecutado exitosamente",
		Data:    result,
	})
}

// BulkExecuteTestCases ejecuta múltiples casos de prueba
func (h *TestHandlers) BulkExecuteTestCases(c *gin.Context) {
	var request struct {
		TestCaseIDs []string `json:"test_case_ids"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Datos inválidos",
			Data:    err.Error(),
		})
		return
	}

	results, err := h.testService.BulkExecuteTestCases(c.Request.Context(), request.TestCaseIDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Error al ejecutar casos de prueba",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Casos de prueba ejecutados exitosamente",
		Data:    results,
	})
}

// CreateTestSuite crea un nuevo suite de prueba
func (h *TestHandlers) CreateTestSuite(c *gin.Context) {
	var testSuite domain.TestSuite
	if err := c.ShouldBindJSON(&testSuite); err != nil {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Datos inválidos",
			Data:    err.Error(),
		})
		return
	}

	err := h.testSuiteService.CreateTestSuite(c.Request.Context(), &testSuite)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Error al crear suite de prueba",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Suite de prueba creado exitosamente",
		Data:    testSuite,
	})
}

// GetTestSuite obtiene un suite de prueba por ID
func (h *TestHandlers) GetTestSuite(c *gin.Context) {
	id := c.Param("id")
	
	testSuite, err := h.testSuiteService.GetTestSuite(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, domain.APIResponse{
			Code:    "NOT_FOUND",
			Message: "Suite de prueba no encontrado",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Suite de prueba encontrado",
		Data:    testSuite,
	})
}

// UpdateTestSuite actualiza un suite de prueba
func (h *TestHandlers) UpdateTestSuite(c *gin.Context) {
	id := c.Param("id")
	var testSuite domain.TestSuite
	if err := c.ShouldBindJSON(&testSuite); err != nil {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Datos inválidos",
			Data:    err.Error(),
		})
		return
	}

	testSuite.ID = id
	err := h.testSuiteService.UpdateTestSuite(c.Request.Context(), &testSuite)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Error al actualizar suite de prueba",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Suite de prueba actualizado exitosamente",
		Data:    testSuite,
	})
}

// DeleteTestSuite elimina un suite de prueba
func (h *TestHandlers) DeleteTestSuite(c *gin.Context) {
	id := c.Param("id")
	
	err := h.testSuiteService.DeleteTestSuite(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Error al eliminar suite de prueba",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Suite de prueba eliminado exitosamente",
		Data:    nil,
	})
}

// GetTestSuitesByBot obtiene suites de prueba por bot
func (h *TestHandlers) GetTestSuitesByBot(c *gin.Context) {
	botID := c.Param("botId")
	
	testSuites, err := h.testSuiteService.GetTestSuitesByBot(c.Request.Context(), botID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Error al obtener suites de prueba",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Suites de prueba obtenidos exitosamente",
		Data:    testSuites,
	})
}

// ExecuteTestSuite ejecuta un suite de prueba
func (h *TestHandlers) ExecuteTestSuite(c *gin.Context) {
	id := c.Param("id")
	
	result, err := h.testSuiteService.ExecuteTestSuite(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Error al ejecutar suite de prueba",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Suite de prueba ejecutado exitosamente",
		Data:    result,
	})
}

// AddTestCaseToSuite agrega un caso de prueba a un suite
func (h *TestHandlers) AddTestCaseToSuite(c *gin.Context) {
	suiteID := c.Param("id")
	var request struct {
		TestCaseID string `json:"test_case_id"`
	}
	
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, domain.APIResponse{
			Code:    "INVALID_REQUEST",
			Message: "Datos inválidos",
			Data:    err.Error(),
		})
		return
	}

	err := h.testSuiteService.AddTestCaseToSuite(c.Request.Context(), suiteID, request.TestCaseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Error al agregar caso de prueba al suite",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Caso de prueba agregado al suite exitosamente",
		Data:    nil,
	})
}

// RemoveTestCaseFromSuite remueve un caso de prueba de un suite
func (h *TestHandlers) RemoveTestCaseFromSuite(c *gin.Context) {
	suiteID := c.Param("id")
	testCaseID := c.Param("testCaseId")
	
	err := h.testSuiteService.RemoveTestCaseFromSuite(c.Request.Context(), suiteID, testCaseID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, domain.APIResponse{
			Code:    "INTERNAL_ERROR",
			Message: "Error al remover caso de prueba del suite",
			Data:    err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, domain.APIResponse{
		Code:    "SUCCESS",
		Message: "Caso de prueba removido del suite exitosamente",
		Data:    nil,
	})
} 