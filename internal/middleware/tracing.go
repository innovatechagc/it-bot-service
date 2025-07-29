package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

// Tracing middleware simplificado (sin OpenTelemetry por ahora)
func Tracing() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Agregar información básica de tracing al contexto
		start := time.Now()
		
		// Generar un ID de trace simple
		traceID := generateTraceID()
		c.Set("trace_id", traceID)
		c.Header("X-Trace-ID", traceID)
		
		// Procesar request
		c.Next()
		
		// Log básico de tracing (se puede expandir)
		duration := time.Since(start)
		c.Set("request_duration", duration)
		
		// En el futuro se puede integrar con sistemas de tracing reales
	}
}

// generateTraceID genera un ID de trace simple
func generateTraceID() string {
	return "trace-" + time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString genera una cadena aleatoria simple
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}