package services

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/company/bot-service/internal/domain"
	"github.com/company/bot-service/pkg/logger"
)

// TestService define las operaciones para manejar casos de prueba
type TestService interface {
	GetTestCase(ctx context.Context, id string) (*domain.TestCase, error)
	GetTestCasesByBot(ctx context.Context, botID string) ([]*domain.TestCase, error)
	GetTestCasesByStatus(ctx context.Context, botID string, status domain.TestStatus) ([]*domain.TestCase, error)
	CreateTestCase(ctx context.Context, testCase *domain.TestCase) error
	UpdateTestCase(ctx context.Context, testCase *domain.TestCase) error
	DeleteTestCase(ctx context.Context, id string) error
	ExecuteTestCase(ctx context.Context, id string) (*domain.TestResult, error)
	BulkExecuteTestCases(ctx context.Context, ids []string) (map[string]*domain.TestResult, error)
}

// TestSuiteService define las operaciones para manejar suites de prueba
type TestSuiteService interface {
	GetTestSuite(ctx context.Context, id string) (*domain.TestSuite, error)
	GetTestSuitesByBot(ctx context.Context, botID string) ([]*domain.TestSuite, error)
	GetTestSuitesByStatus(ctx context.Context, botID string, status domain.TestSuiteStatus) ([]*domain.TestSuite, error)
	CreateTestSuite(ctx context.Context, testSuite *domain.TestSuite) error
	UpdateTestSuite(ctx context.Context, testSuite *domain.TestSuite) error
	DeleteTestSuite(ctx context.Context, id string) error
	ExecuteTestSuite(ctx context.Context, id string) (*domain.TestSuiteResult, error)
	AddTestCaseToSuite(ctx context.Context, suiteID, testCaseID string) error
	RemoveTestCaseFromSuite(ctx context.Context, suiteID, testCaseID string) error
}

// testService implementa TestService
type testService struct {
	testCaseRepo domain.TestCaseRepository
	botSvc       BotService
	conditionalSvc ConditionalService
	triggerSvc   TriggerService
	logger       logger.Logger
}

// testSuiteService implementa TestSuiteService
type testSuiteService struct {
	testSuiteRepo domain.TestSuiteRepository
	testSvc       TestService
	logger        logger.Logger
}

// NewTestService crea una nueva instancia de TestService
func NewTestService(
	testCaseRepo domain.TestCaseRepository,
	botSvc BotService,
	conditionalSvc ConditionalService,
	triggerSvc TriggerService,
	logger logger.Logger,
) TestService {
	return &testService{
		testCaseRepo:   testCaseRepo,
		botSvc:         botSvc,
		conditionalSvc: conditionalSvc,
		triggerSvc:     triggerSvc,
		logger:         logger,
	}
}

// NewTestSuiteService crea una nueva instancia de TestSuiteService
func NewTestSuiteService(
	testSuiteRepo domain.TestSuiteRepository,
	testSvc TestService,
	logger logger.Logger,
) TestSuiteService {
	return &testSuiteService{
		testSuiteRepo: testSuiteRepo,
		testSvc:       testSvc,
		logger:        logger,
	}
}

// Implementación de TestService
func (s *testService) GetTestCase(ctx context.Context, id string) (*domain.TestCase, error) {
	return s.testCaseRepo.GetByID(ctx, id)
}

func (s *testService) GetTestCasesByBot(ctx context.Context, botID string) ([]*domain.TestCase, error) {
	return s.testCaseRepo.GetByBotID(ctx, botID)
}

func (s *testService) GetTestCasesByStatus(ctx context.Context, botID string, status domain.TestStatus) ([]*domain.TestCase, error) {
	return s.testCaseRepo.GetByStatus(ctx, botID, status)
}

func (s *testService) CreateTestCase(ctx context.Context, testCase *domain.TestCase) error {
	if testCase.ID == "" {
		testCase.ID = uuid.New().String()
	}
	testCase.Status = domain.TestStatusPending
	testCase.CreatedAt = time.Now()
	testCase.UpdatedAt = time.Now()
	
	return s.testCaseRepo.Create(ctx, testCase)
}

func (s *testService) UpdateTestCase(ctx context.Context, testCase *domain.TestCase) error {
	testCase.UpdatedAt = time.Now()
	return s.testCaseRepo.Update(ctx, testCase)
}

func (s *testService) DeleteTestCase(ctx context.Context, id string) error {
	return s.testCaseRepo.Delete(ctx, id)
}

func (s *testService) ExecuteTestCase(ctx context.Context, id string) (*domain.TestResult, error) {
	// Obtener el caso de prueba
	testCase, err := s.testCaseRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get test case: %w", err)
	}
	
	// Actualizar estado a running
	testCase.Status = domain.TestStatusRunning
	testCase.UpdatedAt = time.Now()
	if err := s.testCaseRepo.Update(ctx, testCase); err != nil {
		return nil, fmt.Errorf("failed to update test case status: %w", err)
	}
	
	// Ejecutar el caso de prueba
	result, err := s.executeTestCase(ctx, testCase)
	if err != nil {
		// Actualizar estado a failed
		testCase.Status = domain.TestStatusFailed
		testCase.Result = &domain.TestResult{
			Success:    false,
			Error:      err.Error(),
			ExecutedAt: time.Now(),
		}
		s.testCaseRepo.Update(ctx, testCase)
		return nil, err
	}
	
	// Actualizar resultado
	testCase.Result = result
	if result.Success {
		testCase.Status = domain.TestStatusPassed
	} else {
		testCase.Status = domain.TestStatusFailed
	}
	testCase.UpdatedAt = time.Now()
	
	if err := s.testCaseRepo.Update(ctx, testCase); err != nil {
		return nil, fmt.Errorf("failed to update test case result: %w", err)
	}
	
	return result, nil
}

func (s *testService) BulkExecuteTestCases(ctx context.Context, ids []string) (map[string]*domain.TestResult, error) {
	return s.testCaseRepo.BulkExecute(ctx, ids)
}

// executeTestCase ejecuta un caso de prueba específico
func (s *testService) executeTestCase(ctx context.Context, testCase *domain.TestCase) (*domain.TestResult, error) {
	startTime := time.Now()
	
	// Crear mensaje de entrada
	message := &domain.IncomingMessage{
		ID:        uuid.New().String(),
		BotID:     testCase.BotID,
		UserID:    testCase.Input.UserID,
		Content:   testCase.Input.Message,
		Channel:   domain.ChannelWeb, // Por defecto para pruebas
		Metadata:  testCase.Input.Metadata,
		Timestamp: time.Now(),
	}
	
	// Procesar mensaje con el bot
	response, err := s.botSvc.ProcessIncomingMessage(ctx, message)
	if err != nil {
		return &domain.TestResult{
			Success:    false,
			Error:      err.Error(),
			ExecutedAt: time.Now(),
		}, nil
	}
	
	// Evaluar condiciones
	executedConditions := []string{}
	for _, conditionID := range testCase.Conditions {
		conditionMet, err := s.conditionalSvc.EvaluateConditional(ctx, conditionID, testCase.Input.Context)
		if err != nil {
			s.logger.Error("Failed to evaluate condition", "condition_id", conditionID, "error", err)
			continue
		}
		if conditionMet {
			executedConditions = append(executedConditions, conditionID)
		}
	}
	
	// Evaluar triggers
	executedTriggers := []string{}
	for _, triggerID := range testCase.Triggers {
		if err := s.triggerSvc.ExecuteTrigger(ctx, triggerID, testCase.Input.Context); err != nil {
			s.logger.Error("Failed to execute trigger", "trigger_id", triggerID, "error", err)
		} else {
			executedTriggers = append(executedTriggers, triggerID)
		}
	}
	
	// Verificar resultados esperados
	success := s.verifyExpectedResults(testCase, response, executedConditions, executedTriggers)
	
	executionTime := time.Since(startTime).Milliseconds()
	
	return &domain.TestResult{
		Success:            success,
		ActualResponse:     response.Content,
		ActualNextStep:     "",
		ExecutedConditions: executedConditions,
		ExecutedTriggers:   executedTriggers,
		ActualContext:      testCase.Input.Context,
		ExecutionTime:      executionTime,
		ExecutedAt:         time.Now(),
	}, nil
}

// verifyExpectedResults verifica si los resultados coinciden con lo esperado
func (s *testService) verifyExpectedResults(testCase *domain.TestCase, response *domain.BotResponse, executedConditions, executedTriggers []string) bool {
	// Verificar respuesta esperada
	if testCase.Expected.Response != "" && response.Content != testCase.Expected.Response {
		return false
	}
	
	// Verificar condiciones esperadas
	if len(testCase.Expected.Conditions) > 0 {
		for _, expectedCondition := range testCase.Expected.Conditions {
			found := false
			for _, executedCondition := range executedConditions {
				if executedCondition == expectedCondition {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}
	
	// Verificar triggers esperados
	if len(testCase.Expected.Triggers) > 0 {
		for _, expectedTrigger := range testCase.Expected.Triggers {
			found := false
			for _, executedTrigger := range executedTriggers {
				if executedTrigger == expectedTrigger {
					found = true
					break
				}
			}
			if !found {
				return false
			}
		}
	}
	
	return true
}

// Implementación de TestSuiteService
func (s *testSuiteService) GetTestSuite(ctx context.Context, id string) (*domain.TestSuite, error) {
	return s.testSuiteRepo.GetByID(ctx, id)
}

func (s *testSuiteService) GetTestSuitesByBot(ctx context.Context, botID string) ([]*domain.TestSuite, error) {
	return s.testSuiteRepo.GetByBotID(ctx, botID)
}

func (s *testSuiteService) GetTestSuitesByStatus(ctx context.Context, botID string, status domain.TestSuiteStatus) ([]*domain.TestSuite, error) {
	return s.testSuiteRepo.GetByStatus(ctx, botID, status)
}

func (s *testSuiteService) CreateTestSuite(ctx context.Context, testSuite *domain.TestSuite) error {
	if testSuite.ID == "" {
		testSuite.ID = uuid.New().String()
	}
	testSuite.Status = domain.TestSuiteStatusPending
	testSuite.CreatedAt = time.Now()
	testSuite.UpdatedAt = time.Now()
	
	return s.testSuiteRepo.Create(ctx, testSuite)
}

func (s *testSuiteService) UpdateTestSuite(ctx context.Context, testSuite *domain.TestSuite) error {
	testSuite.UpdatedAt = time.Now()
	return s.testSuiteRepo.Update(ctx, testSuite)
}

func (s *testSuiteService) DeleteTestSuite(ctx context.Context, id string) error {
	return s.testSuiteRepo.Delete(ctx, id)
}

func (s *testSuiteService) ExecuteTestSuite(ctx context.Context, id string) (*domain.TestSuiteResult, error) {
	// Obtener la suite de pruebas
	testSuite, err := s.testSuiteRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get test suite: %w", err)
	}
	
	// Actualizar estado a running
	testSuite.Status = domain.TestSuiteStatusRunning
	testSuite.UpdatedAt = time.Now()
	if err := s.testSuiteRepo.Update(ctx, testSuite); err != nil {
		return nil, fmt.Errorf("failed to update test suite status: %w", err)
	}
	
	// Ejecutar todos los casos de prueba
	startTime := time.Now()
	testResults := make(map[string]*domain.TestResult)
	
	totalTests := len(testSuite.TestCases)
	passedTests := 0
	failedTests := 0
	skippedTests := 0
	
	for _, testCaseID := range testSuite.TestCases {
		result, err := s.testSvc.ExecuteTestCase(ctx, testCaseID)
		if err != nil {
			s.logger.Error("Failed to execute test case", "test_case_id", testCaseID, "error", err)
			failedTests++
			testResults[testCaseID] = &domain.TestResult{
				Success:    false,
				Error:      err.Error(),
				ExecutedAt: time.Now(),
			}
			continue
		}
		
		testResults[testCaseID] = result
		if result.Success {
			passedTests++
		} else {
			failedTests++
		}
	}
	
	executionTime := time.Since(startTime).Milliseconds()
	successRate := 0.0
	if totalTests > 0 {
		successRate = float64(passedTests) / float64(totalTests) * 100
	}
	
	// Determinar estado final
	var finalStatus domain.TestSuiteStatus
	if failedTests == 0 {
		finalStatus = domain.TestSuiteStatusPassed
	} else if passedTests == 0 {
		finalStatus = domain.TestSuiteStatusFailed
	} else {
		finalStatus = domain.TestSuiteStatusPartial
	}
	
	result := &domain.TestSuiteResult{
		TotalTests:     totalTests,
		PassedTests:    passedTests,
		FailedTests:    failedTests,
		SkippedTests:   skippedTests,
		SuccessRate:    successRate,
		ExecutionTime:  executionTime,
		StartedAt:      startTime,
		CompletedAt:    time.Now(),
		TestResults:    testResults,
	}
	
	// Actualizar suite con resultados
	testSuite.Status = finalStatus
	testSuite.Result = result
	testSuite.UpdatedAt = time.Now()
	
	if err := s.testSuiteRepo.Update(ctx, testSuite); err != nil {
		return nil, fmt.Errorf("failed to update test suite result: %w", err)
	}
	
	return result, nil
}

func (s *testSuiteService) AddTestCaseToSuite(ctx context.Context, suiteID, testCaseID string) error {
	return s.testSuiteRepo.AddTestCase(ctx, suiteID, testCaseID)
}

func (s *testSuiteService) RemoveTestCaseFromSuite(ctx context.Context, suiteID, testCaseID string) error {
	return s.testSuiteRepo.RemoveTestCase(ctx, suiteID, testCaseID)
} 