package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/company/bot-service/internal/domain"
	"github.com/company/bot-service/internal/mcp"
	"github.com/company/bot-service/pkg/logger"
)

// TaskManager define las operaciones para gestión de tareas asíncronas
type TaskManager interface {
	// Gestión de tareas
	SubmitTask(ctx context.Context, task *domain.AsyncTask) error
	GetTask(ctx context.Context, taskID string) (*domain.AsyncTask, error)
	ListTasks(ctx context.Context, filters *TaskFilters) ([]*domain.AsyncTask, error)
	CancelTask(ctx context.Context, taskID string) error
	
	// Ejecución
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	
	// Monitoreo
	GetStats() *TaskStats
}

// TaskFilters define filtros para listar tareas
type TaskFilters struct {
	Status    *domain.TaskStatus `json:"status,omitempty"`
	Type      *string            `json:"type,omitempty"`
	UserID    *string            `json:"user_id,omitempty"`
	BotID     *string            `json:"bot_id,omitempty"`
	CreatedAt *TimeRange         `json:"created_at,omitempty"`
	Limit     int                `json:"limit,omitempty"`
	Offset    int                `json:"offset,omitempty"`
}

// TimeRange define un rango de tiempo
type TimeRange struct {
	From *time.Time `json:"from,omitempty"`
	To   *time.Time `json:"to,omitempty"`
}

// TaskStats define estadísticas del task manager
type TaskStats struct {
	TotalTasks     int64                        `json:"total_tasks"`
	PendingTasks   int64                        `json:"pending_tasks"`
	RunningTasks   int64                        `json:"running_tasks"`
	CompletedTasks int64                        `json:"completed_tasks"`
	FailedTasks    int64                        `json:"failed_tasks"`
	CancelledTasks int64                        `json:"cancelled_tasks"`
	TasksByType    map[string]int64             `json:"tasks_by_type"`
	AverageTime    time.Duration                `json:"average_execution_time"`
	WorkerStats    map[string]*WorkerStats      `json:"worker_stats"`
	LastUpdated    time.Time                    `json:"last_updated"`
}

// WorkerStats define estadísticas de un worker
type WorkerStats struct {
	ID            string        `json:"id"`
	Status        string        `json:"status"`
	TasksExecuted int64         `json:"tasks_executed"`
	LastTask      *string       `json:"last_task,omitempty"`
	LastActivity  time.Time     `json:"last_activity"`
	AverageTime   time.Duration `json:"average_execution_time"`
}

// taskManager implementa TaskManager
type taskManager struct {
	tasks           map[string]*domain.AsyncTask
	taskQueue       chan *domain.AsyncTask
	workers         []*taskWorker
	mcpOrchestrator interface {
		mcp.MCPOrchestrator
		mcp.MCPDomainOrchestrator
	}
	logger          logger.Logger
	mu              sync.RWMutex
	ctx             context.Context
	cancel          context.CancelFunc
	stats           *TaskStats
	workerCount     int
	maxQueueSize    int
}

// taskWorker representa un worker que ejecuta tareas
type taskWorker struct {
	id              string
	manager         *taskManager
	stats           *WorkerStats
	logger          logger.Logger
	mu              sync.RWMutex
}

// NewTaskManager crea un nuevo task manager
func NewTaskManager(
	mcpOrchestrator interface {
		mcp.MCPOrchestrator
		mcp.MCPDomainOrchestrator
	},
	logger logger.Logger,
	workerCount int,
	maxQueueSize int,
) TaskManager {
	if workerCount <= 0 {
		workerCount = 5
	}
	if maxQueueSize <= 0 {
		maxQueueSize = 1000
	}
	
	return &taskManager{
		tasks:           make(map[string]*domain.AsyncTask),
		taskQueue:       make(chan *domain.AsyncTask, maxQueueSize),
		workers:         make([]*taskWorker, 0, workerCount),
		mcpOrchestrator: mcpOrchestrator,
		logger:          logger,
		stats: &TaskStats{
			TasksByType: make(map[string]int64),
			WorkerStats: make(map[string]*WorkerStats),
		},
		workerCount:  workerCount,
		maxQueueSize: maxQueueSize,
	}
}

// Start inicia el task manager
func (tm *taskManager) Start(ctx context.Context) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	
	if tm.ctx != nil {
		return fmt.Errorf("task manager already started")
	}
	
	tm.ctx, tm.cancel = context.WithCancel(ctx)
	
	// Crear y iniciar workers
	for i := 0; i < tm.workerCount; i++ {
		worker := &taskWorker{
			id:      fmt.Sprintf("worker-%d", i+1),
			manager: tm,
			stats: &WorkerStats{
				ID:           fmt.Sprintf("worker-%d", i+1),
				Status:       "idle",
				LastActivity: time.Now(),
			},
			logger: tm.logger,
		}
		
		tm.workers = append(tm.workers, worker)
		tm.stats.WorkerStats[worker.id] = worker.stats
		
		go worker.run(tm.ctx)
	}
	
	tm.logger.Info("Task manager started", 
		"worker_count", tm.workerCount,
		"max_queue_size", tm.maxQueueSize)
	
	return nil
}

// Stop detiene el task manager
func (tm *taskManager) Stop(ctx context.Context) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	
	if tm.cancel != nil {
		tm.cancel()
		tm.cancel = nil
		tm.ctx = nil
	}
	
	// Cerrar canal de tareas
	close(tm.taskQueue)
	
	tm.logger.Info("Task manager stopped")
	return nil
}

// SubmitTask envía una tarea para ejecución asíncrona
func (tm *taskManager) SubmitTask(ctx context.Context, task *domain.AsyncTask) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	
	if tm.ctx == nil {
		return fmt.Errorf("task manager not started")
	}
	
	// Generar ID si no se proporciona
	if task.ID == "" {
		task.ID = fmt.Sprintf("task-%d", time.Now().UnixNano())
	}
	
	// Establecer timestamps
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()
	task.Status = domain.TaskStatusPending
	
	// Guardar tarea
	tm.tasks[task.ID] = task
	
	// Actualizar estadísticas
	tm.stats.TotalTasks++
	tm.stats.PendingTasks++
	tm.stats.TasksByType[task.Type]++
	
	// Enviar a la cola
	select {
	case tm.taskQueue <- task:
		tm.logger.Info("Task submitted", 
			"task_id", task.ID,
			"type", task.Type,
			"priority", task.Priority)
		return nil
	default:
		// Cola llena
		task.Status = domain.TaskStatusFailed
		task.Error = "task queue is full"
		task.CompletedAt = time.Now()
		tm.stats.PendingTasks--
		tm.stats.FailedTasks++
		return fmt.Errorf("task queue is full")
	}
}

// GetTask obtiene una tarea por ID
func (tm *taskManager) GetTask(ctx context.Context, taskID string) (*domain.AsyncTask, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	
	task, exists := tm.tasks[taskID]
	if !exists {
		return nil, fmt.Errorf("task not found: %s", taskID)
	}
	
	// Crear copia para evitar modificaciones concurrentes
	taskCopy := *task
	return &taskCopy, nil
}

// ListTasks lista tareas con filtros opcionales
func (tm *taskManager) ListTasks(ctx context.Context, filters *TaskFilters) ([]*domain.AsyncTask, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	
	var result []*domain.AsyncTask
	
	for _, task := range tm.tasks {
		// Aplicar filtros
		if filters != nil {
			if filters.Status != nil && task.Status != *filters.Status {
				continue
			}
			if filters.Type != nil && task.Type != *filters.Type {
				continue
			}
			if filters.UserID != nil && task.UserID != *filters.UserID {
				continue
			}
			if filters.BotID != nil && task.BotID != *filters.BotID {
				continue
			}
			if filters.CreatedAt != nil {
				if filters.CreatedAt.From != nil && task.CreatedAt.Before(*filters.CreatedAt.From) {
					continue
				}
				if filters.CreatedAt.To != nil && task.CreatedAt.After(*filters.CreatedAt.To) {
					continue
				}
			}
		}
		
		// Crear copia
		taskCopy := *task
		result = append(result, &taskCopy)
	}
	
	// Aplicar paginación
	if filters != nil {
		if filters.Offset > 0 && filters.Offset < len(result) {
			result = result[filters.Offset:]
		}
		if filters.Limit > 0 && filters.Limit < len(result) {
			result = result[:filters.Limit]
		}
	}
	
	return result, nil
}

// CancelTask cancela una tarea
func (tm *taskManager) CancelTask(ctx context.Context, taskID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	
	task, exists := tm.tasks[taskID]
	if !exists {
		return fmt.Errorf("task not found: %s", taskID)
	}
	
	if task.Status == domain.TaskStatusCompleted || task.Status == domain.TaskStatusFailed {
		return fmt.Errorf("cannot cancel completed task")
	}
	
	task.Status = domain.TaskStatusCancelled
	task.UpdatedAt = time.Now()
	task.CompletedAt = time.Now()
	task.Error = "task cancelled by user"
	
	// Actualizar estadísticas
	if task.Status == domain.TaskStatusPending {
		tm.stats.PendingTasks--
	} else if task.Status == domain.TaskStatusRunning {
		tm.stats.RunningTasks--
	}
	tm.stats.CancelledTasks++
	
	tm.logger.Info("Task cancelled", "task_id", taskID)
	return nil
}

// GetStats obtiene estadísticas del task manager
func (tm *taskManager) GetStats() *TaskStats {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	
	// Crear copia de las estadísticas
	stats := *tm.stats
	stats.LastUpdated = time.Now()
	
	// Copiar mapas
	stats.TasksByType = make(map[string]int64)
	for k, v := range tm.stats.TasksByType {
		stats.TasksByType[k] = v
	}
	
	stats.WorkerStats = make(map[string]*WorkerStats)
	for k, v := range tm.stats.WorkerStats {
		workerStats := *v
		stats.WorkerStats[k] = &workerStats
	}
	
	return &stats
}

// run ejecuta el loop principal del worker
func (w *taskWorker) run(ctx context.Context) {
	w.logger.Info("Task worker started", "worker_id", w.id)
	
	for {
		select {
		case <-ctx.Done():
			w.logger.Info("Task worker stopped", "worker_id", w.id)
			return
		case task, ok := <-w.manager.taskQueue:
			if !ok {
				w.logger.Info("Task worker stopped - queue closed", "worker_id", w.id)
				return
			}
			
			w.executeTask(ctx, task)
		}
	}
}

// executeTask ejecuta una tarea
func (w *taskWorker) executeTask(ctx context.Context, task *domain.AsyncTask) {
	start := time.Now()
	
	w.mu.Lock()
	w.stats.Status = "busy"
	w.stats.LastTask = &task.ID
	w.stats.LastActivity = time.Now()
	w.mu.Unlock()
	
	defer func() {
		w.mu.Lock()
		w.stats.Status = "idle"
		w.stats.TasksExecuted++
		duration := time.Since(start)
		if w.stats.TasksExecuted == 1 {
			w.stats.AverageTime = duration
		} else {
			w.stats.AverageTime = (w.stats.AverageTime + duration) / 2
		}
		w.mu.Unlock()
	}()
	
	w.logger.Info("Executing task", 
		"worker_id", w.id,
		"task_id", task.ID,
		"type", task.Type)
	
	// Actualizar estado de la tarea
	w.manager.mu.Lock()
	task.Status = domain.TaskStatusRunning
	task.UpdatedAt = time.Now()
	task.StartedAt = time.Now()
	w.manager.stats.PendingTasks--
	w.manager.stats.RunningTasks++
	w.manager.mu.Unlock()
	
	// Crear tarea MCP
	mcpTask := &domain.MCPTask{
		ID:          task.ID,
		Type:        task.Type,
		Description: task.Description,
		Input:       task.Input,
		Priority:    task.Priority,
		Timeout:     task.Timeout,
		Context:     task.Context,
		Metadata:    task.Metadata,
		CreatedAt:   task.CreatedAt,
	}
	
	// Ejecutar usando MCP
	result, err := w.manager.mcpOrchestrator.ExecuteTaskDomain(ctx, mcpTask)
	
	duration := time.Since(start)
	
	// Actualizar tarea con resultado
	w.manager.mu.Lock()
	task.UpdatedAt = time.Now()
	task.CompletedAt = time.Now()
	task.ExecutionTime = duration.Milliseconds()
	w.manager.stats.RunningTasks--
	
	if err != nil || !result.Success {
		task.Status = domain.TaskStatusFailed
		if err != nil {
			task.Error = err.Error()
		} else {
			task.Error = result.Error
		}
		task.Result = map[string]interface{}{
			"success": false,
			"error":   task.Error,
		}
		w.manager.stats.FailedTasks++
		
		w.logger.Error("Task execution failed", 
			"worker_id", w.id,
			"task_id", task.ID,
			"duration", duration,
			"error", task.Error)
	} else {
		task.Status = domain.TaskStatusCompleted
		task.Result = map[string]interface{}{
			"success":        true,
			"output":         result.Output,
			"agent_id":       result.AgentID,
			"execution_time": result.ExecutionTime,
		}
		w.manager.stats.CompletedTasks++
		
		w.logger.Info("Task execution completed", 
			"worker_id", w.id,
			"task_id", task.ID,
			"duration", duration,
			"agent_id", result.AgentID)
	}
	
	// Actualizar tiempo promedio
	totalCompleted := w.manager.stats.CompletedTasks + w.manager.stats.FailedTasks
	if totalCompleted == 1 {
		w.manager.stats.AverageTime = duration
	} else {
		w.manager.stats.AverageTime = (w.manager.stats.AverageTime + duration) / 2
	}
	
	w.manager.mu.Unlock()
}