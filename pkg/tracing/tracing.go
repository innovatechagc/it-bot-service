package tracing

import (
	"context"
)

// Config para configuración de tracing (simplificado)
type Config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	JaegerEndpoint string
	Enabled        bool
}

// InitTracing inicializa el tracing (versión simplificada sin OpenTelemetry)
func InitTracing(cfg Config) (func(context.Context) error, error) {
	// Por ahora, tracing está deshabilitado para evitar problemas de dependencias
	// En el futuro se puede implementar con OpenTelemetry cuando sea necesario
	return func(context.Context) error { return nil }, nil
}

// Trace representa un trace simplificado
type Trace struct {
	ID   string
	Name string
}

// StartTrace inicia un nuevo trace (stub)
func StartTrace(ctx context.Context, name string) (context.Context, *Trace) {
	trace := &Trace{
		ID:   "trace-" + name,
		Name: name,
	}
	return ctx, trace
}

// End termina el trace (stub)
func (t *Trace) End() {
	// No-op por ahora
}

// AddAttribute agrega un atributo al trace (stub)
func (t *Trace) AddAttribute(key, value string) {
	// No-op por ahora
}