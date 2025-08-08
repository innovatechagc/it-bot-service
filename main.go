package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/company/bot-service/internal/ai"
	"github.com/company/bot-service/internal/config"
	"github.com/company/bot-service/internal/handlers"
	"github.com/company/bot-service/internal/mcp"
	"github.com/company/bot-service/internal/middleware"
	"github.com/company/bot-service/internal/repositories"
	"github.com/company/bot-service/internal/services"
	"github.com/company/bot-service/pkg/logger"
	"github.com/gin-gonic/gin"
)

// @title Microservice Template API
// @version 1.0
// @description Template para microservicios en Go
// @host localhost:8080
// @BasePath /api/v1
func main() {
	// Cargar configuración
	cfg := config.Load()
	
	// Inicializar logger
	logger := logger.NewLogger(cfg.LogLevel)
	
	// Inicializar cliente de Vault (comentado para testing)
	// vaultClient, err := vault.NewClient(cfg.VaultConfig)
	// if err != nil {
	// 	logger.Fatal("Failed to initialize Vault client", err)
	// }
	
	// Inicializar cliente de IA
	aiClient := ai.NewMockAIClient([]string{
		"Hello! How can I help you today?",
		"I understand your question. Let me help you with that.",
		"Thank you for your message. Is there anything else I can assist you with?",
	}, logger)
	
	// Inicializar sistema MCP
	agentFactory := mcp.NewAgentFactory(logger)
	mcpOrchestrator := mcp.NewOrchestrator(agentFactory, logger)
	
	// Iniciar orquestador MCP
	if err := mcpOrchestrator.Start(context.Background()); err != nil {
		logger.Fatal("Failed to start MCP orchestrator", err)
	}
	
	// Inicializar repositorios (usando mocks para desarrollo)
	botRepo := repositories.NewMockBotRepository()
	flowRepo := repositories.NewMockBotFlowRepository()
	stepRepo := repositories.NewMockBotStepRepository()
	smartReplyRepo := repositories.NewMockSmartReplyRepository()
	sessionRepo := repositories.NewMockConversationSessionRepository()
	
	// Inicializar repositorios de testing
	conditionalRepo := repositories.NewMockConditionalRepository()
	triggerRepo := repositories.NewMockTriggerRepository()
	testCaseRepo := repositories.NewMockTestCaseRepository()
	testSuiteRepo := repositories.NewMockTestSuiteRepository()
	
	// Inicializar servicios
	healthService := services.NewHealthService()
	conversationService := services.NewConversationService(sessionRepo, logger)
	smartReplyService := services.NewSmartReplyService(smartReplyRepo, aiClient, mcpOrchestrator, logger)
	botFlowService := services.NewBotFlowService(flowRepo, stepRepo, logger)
	botStepService := services.NewBotStepService(stepRepo, logger)
	taskManager := services.NewTaskManager(mcpOrchestrator, logger, 5, 1000)
	botService := services.NewBotService(
		botRepo,
		flowRepo,
		stepRepo,
		sessionRepo,
		smartReplyRepo,
		conversationService,
		smartReplyService,
		mcpOrchestrator,
		logger,
	)
	
	// Inicializar servicios de testing
	conditionalService := services.NewConditionalService(conditionalRepo, logger)
	triggerService := services.NewTriggerService(triggerRepo, conditionalService, logger)
	testService := services.NewTestService(testCaseRepo, botService, conditionalService, triggerService, logger)
	testSuiteService := services.NewTestSuiteService(testSuiteRepo, testService, logger)
	
	// Inicializar handlers
	botHandler := handlers.NewBotHandler(
		botService,
		botFlowService,
		botStepService,
		smartReplyService,
		conversationService,
		logger,
	)
	
	mcpHandler := handlers.NewMCPHandler(mcpOrchestrator, logger)
	taskHandler := handlers.NewTaskHandler(taskManager, logger)
	testHandler := handlers.NewTestHandlers(
		conditionalService,
		triggerService,
		testService,
		testSuiteService,
		logger,
	)
	
	// Iniciar task manager
	if err := taskManager.Start(context.Background()); err != nil {
		logger.Fatal("Failed to start task manager", err)
	}
	
	// Configurar Gin
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.Logger(logger))
	router.Use(middleware.CORS())
	router.Use(middleware.Metrics())
	
	// Rutas
	handlers.SetupRoutes(router, healthService, botHandler, mcpHandler, taskHandler, testHandler, logger)
	
	// Servidor HTTP
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}
	
	// Iniciar servidor en goroutine
	go func() {
		logger.Info("Starting bot-service on port " + cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", err)
		}
	}()
	
	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	
	logger.Info("Shutting down server...")
	
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	
	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", err)
	}
	
	logger.Info("Server exited")
}