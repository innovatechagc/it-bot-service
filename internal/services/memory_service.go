package services

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/company/bot-service/internal/domain"
	"github.com/company/bot-service/pkg/logger"
)

// MemoryService define las operaciones para gestión de memoria persistente
type MemoryService interface {
	// Memoria a largo plazo
	StoreMemory(ctx context.Context, memory *domain.Memory) error
	GetMemory(ctx context.Context, userID, botID, key string) (*domain.Memory, error)
	GetUserMemories(ctx context.Context, userID, botID string) ([]*domain.Memory, error)
	UpdateMemory(ctx context.Context, memory *domain.Memory) error
	DeleteMemory(ctx context.Context, userID, botID, key string) error
	
	// Búsqueda semántica
	SearchMemories(ctx context.Context, userID, botID, query string, limit int) ([]*domain.Memory, error)
	
	// Gestión de contexto
	GetContextSummary(ctx context.Context, userID, botID string) (*domain.ContextSummary, error)
	UpdateContextSummary(ctx context.Context, summary *domain.ContextSummary) error
	
	// Limpieza
	CleanupExpiredMemories(ctx context.Context) error
	GetMemoryStats(ctx context.Context, userID, botID string) (*domain.MemoryStats, error)
}

// memoryService implementa MemoryService
type memoryService struct {
	memories      map[string]*domain.Memory // key: userID:botID:key
	summaries     map[string]*domain.ContextSummary // key: userID:botID
	mu            sync.RWMutex
	logger        logger.Logger
	maxMemories   int
	retentionDays int
}

// NewMemoryService crea un nuevo servicio de memoria
func NewMemoryService(logger logger.Logger, maxMemories, retentionDays int) MemoryService {
	if maxMemories <= 0 {
		maxMemories = 1000
	}
	if retentionDays <= 0 {
		retentionDays = 30
	}
	
	return &memoryService{
		memories:      make(map[string]*domain.Memory),
		summaries:     make(map[string]*domain.ContextSummary),
		logger:        logger,
		maxMemories:   maxMemories,
		retentionDays: retentionDays,
	}
}

// StoreMemory almacena una memoria a largo plazo
func (s *memoryService) StoreMemory(ctx context.Context, memory *domain.Memory) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// Generar clave única
	key := fmt.Sprintf("%s:%s:%s", memory.UserID, memory.BotID, memory.Key)
	
	// Verificar límite de memorias
	if len(s.memories) >= s.maxMemories {
		// Eliminar la memoria más antigua
		s.evictOldestMemory()
	}
	
	// Establecer timestamps
	now := time.Now()
	if memory.CreatedAt.IsZero() {
		memory.CreatedAt = now
	}
	memory.UpdatedAt = now
	
	// Calcular fecha de expiración
	if memory.ExpiresAt.IsZero() {
		memory.ExpiresAt = now.AddDate(0, 0, s.retentionDays)
	}
	
	// Almacenar memoria
	s.memories[key] = memory
	
	s.logger.Info("Memory stored", 
		"user_id", memory.UserID,
		"bot_id", memory.BotID,
		"key", memory.Key,
		"type", memory.Type,
		"importance", memory.Importance)
	
	return nil
}

// GetMemory obtiene una memoria específica
func (s *memoryService) GetMemory(ctx context.Context, userID, botID, key string) (*domain.Memory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	memoryKey := fmt.Sprintf("%s:%s:%s", userID, botID, key)
	memory, exists := s.memories[memoryKey]
	if !exists {
		return nil, fmt.Errorf("memory not found")
	}
	
	// Verificar si ha expirado
	if time.Now().After(memory.ExpiresAt) {
		return nil, fmt.Errorf("memory has expired")
	}
	
	// Crear copia para evitar modificaciones concurrentes
	memoryCopy := *memory
	return &memoryCopy, nil
}

// GetUserMemories obtiene todas las memorias de un usuario para un bot
func (s *memoryService) GetUserMemories(ctx context.Context, userID, botID string) ([]*domain.Memory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var result []*domain.Memory
	prefix := fmt.Sprintf("%s:%s:", userID, botID)
	
	for key, memory := range s.memories {
		if len(key) > len(prefix) && key[:len(prefix)] == prefix {
			// Verificar si ha expirado
			if time.Now().After(memory.ExpiresAt) {
				continue
			}
			
			// Crear copia
			memoryCopy := *memory
			result = append(result, &memoryCopy)
		}
	}
	
	return result, nil
}

// UpdateMemory actualiza una memoria existente
func (s *memoryService) UpdateMemory(ctx context.Context, memory *domain.Memory) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	key := fmt.Sprintf("%s:%s:%s", memory.UserID, memory.BotID, memory.Key)
	
	if _, exists := s.memories[key]; !exists {
		return fmt.Errorf("memory not found")
	}
	
	memory.UpdatedAt = time.Now()
	s.memories[key] = memory
	
	s.logger.Info("Memory updated", 
		"user_id", memory.UserID,
		"bot_id", memory.BotID,
		"key", memory.Key)
	
	return nil
}

// DeleteMemory elimina una memoria
func (s *memoryService) DeleteMemory(ctx context.Context, userID, botID, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	memoryKey := fmt.Sprintf("%s:%s:%s", userID, botID, key)
	
	if _, exists := s.memories[memoryKey]; !exists {
		return fmt.Errorf("memory not found")
	}
	
	delete(s.memories, memoryKey)
	
	s.logger.Info("Memory deleted", 
		"user_id", userID,
		"bot_id", botID,
		"key", key)
	
	return nil
}

// SearchMemories busca memorias por contenido (implementación simple)
func (s *memoryService) SearchMemories(ctx context.Context, userID, botID, query string, limit int) ([]*domain.Memory, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if limit <= 0 {
		limit = 10
	}
	
	var result []*domain.Memory
	prefix := fmt.Sprintf("%s:%s:", userID, botID)
	
	for key, memory := range s.memories {
		if len(result) >= limit {
			break
		}
		
		if len(key) > len(prefix) && key[:len(prefix)] == prefix {
			// Verificar si ha expirado
			if time.Now().After(memory.ExpiresAt) {
				continue
			}
			
			// Búsqueda simple por contenido
			if s.matchesQuery(memory, query) {
				memoryCopy := *memory
				result = append(result, &memoryCopy)
			}
		}
	}
	
	return result, nil
}

// GetContextSummary obtiene el resumen de contexto para un usuario y bot
func (s *memoryService) GetContextSummary(ctx context.Context, userID, botID string) (*domain.ContextSummary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	key := fmt.Sprintf("%s:%s", userID, botID)
	summary, exists := s.summaries[key]
	if !exists {
		// Crear resumen vacío
		return &domain.ContextSummary{
			UserID:    userID,
			BotID:     botID,
			Summary:   "",
			KeyPoints: []string{},
			Entities:  make(map[string]interface{}),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}, nil
	}
	
	// Crear copia
	summaryCopy := *summary
	return &summaryCopy, nil
}

// UpdateContextSummary actualiza el resumen de contexto
func (s *memoryService) UpdateContextSummary(ctx context.Context, summary *domain.ContextSummary) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	key := fmt.Sprintf("%s:%s", summary.UserID, summary.BotID)
	summary.UpdatedAt = time.Now()
	
	if summary.CreatedAt.IsZero() {
		summary.CreatedAt = time.Now()
	}
	
	s.summaries[key] = summary
	
	s.logger.Info("Context summary updated", 
		"user_id", summary.UserID,
		"bot_id", summary.BotID,
		"key_points", len(summary.KeyPoints))
	
	return nil
}

// CleanupExpiredMemories limpia memorias expiradas
func (s *memoryService) CleanupExpiredMemories(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	now := time.Now()
	var expiredKeys []string
	
	for key, memory := range s.memories {
		if now.After(memory.ExpiresAt) {
			expiredKeys = append(expiredKeys, key)
		}
	}
	
	for _, key := range expiredKeys {
		delete(s.memories, key)
	}
	
	if len(expiredKeys) > 0 {
		s.logger.Info("Expired memories cleaned up", "count", len(expiredKeys))
	}
	
	return nil
}

// GetMemoryStats obtiene estadísticas de memoria para un usuario y bot
func (s *memoryService) GetMemoryStats(ctx context.Context, userID, botID string) (*domain.MemoryStats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	stats := &domain.MemoryStats{
		UserID:        userID,
		BotID:         botID,
		TotalMemories: 0,
		MemoriesByType: make(map[string]int),
		MemoriesByImportance: make(map[int]int),
		OldestMemory:  time.Now(),
		NewestMemory:  time.Time{},
		LastUpdated:   time.Now(),
	}
	
	prefix := fmt.Sprintf("%s:%s:", userID, botID)
	
	for key, memory := range s.memories {
		if len(key) > len(prefix) && key[:len(prefix)] == prefix {
			// Verificar si ha expirado
			if time.Now().After(memory.ExpiresAt) {
				continue
			}
			
			stats.TotalMemories++
			stats.MemoriesByType[string(memory.Type)]++
			stats.MemoriesByImportance[memory.Importance]++
			
			if memory.CreatedAt.Before(stats.OldestMemory) {
				stats.OldestMemory = memory.CreatedAt
			}
			
			if memory.CreatedAt.After(stats.NewestMemory) {
				stats.NewestMemory = memory.CreatedAt
			}
		}
	}
	
	return stats, nil
}

// evictOldestMemory elimina la memoria más antigua para hacer espacio
func (s *memoryService) evictOldestMemory() {
	var oldestKey string
	var oldestTime time.Time
	
	for key, memory := range s.memories {
		if oldestKey == "" || memory.CreatedAt.Before(oldestTime) {
			oldestKey = key
			oldestTime = memory.CreatedAt
		}
	}
	
	if oldestKey != "" {
		delete(s.memories, oldestKey)
		s.logger.Info("Evicted oldest memory", "key", oldestKey)
	}
}

// matchesQuery verifica si una memoria coincide con la consulta
func (s *memoryService) matchesQuery(memory *domain.Memory, query string) bool {
	// Implementación simple de búsqueda por texto
	query = strings.ToLower(query)
	
	// Buscar en el contenido
	if content, ok := memory.Content["text"].(string); ok {
		if strings.Contains(strings.ToLower(content), query) {
			return true
		}
	}
	
	// Buscar en las etiquetas
	for _, tag := range memory.Tags {
		if strings.Contains(strings.ToLower(tag), query) {
			return true
		}
	}
	
	// Buscar en la clave
	if strings.Contains(strings.ToLower(memory.Key), query) {
		return true
	}
	
	return false
}