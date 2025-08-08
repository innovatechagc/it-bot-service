-- Script para crear tablas de condicionales, triggers y pruebas
-- Ejecutar después de init.sql

-- Tabla para condicionales
CREATE TABLE IF NOT EXISTS conditionals (
    id VARCHAR(255) PRIMARY KEY,
    bot_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    expression TEXT NOT NULL,
    type VARCHAR(50) NOT NULL,
    priority INTEGER DEFAULT 0,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Índices para condicionales
CREATE INDEX IF NOT EXISTS idx_conditionals_bot_id ON conditionals(bot_id);
CREATE INDEX IF NOT EXISTS idx_conditionals_type ON conditionals(type);
CREATE INDEX IF NOT EXISTS idx_conditionals_priority ON conditionals(priority);

-- Tabla para triggers
CREATE TABLE IF NOT EXISTS triggers (
    id VARCHAR(255) PRIMARY KEY,
    bot_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    event VARCHAR(100) NOT NULL,
    condition_id VARCHAR(255),
    action_type VARCHAR(100) NOT NULL,
    action_config JSONB,
    action_timeout BIGINT DEFAULT 5000,
    priority INTEGER DEFAULT 0,
    enabled BOOLEAN DEFAULT true,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (condition_id) REFERENCES conditionals(id) ON DELETE SET NULL
);

-- Índices para triggers
CREATE INDEX IF NOT EXISTS idx_triggers_bot_id ON triggers(bot_id);
CREATE INDEX IF NOT EXISTS idx_triggers_event ON triggers(event);
CREATE INDEX IF NOT EXISTS idx_triggers_enabled ON triggers(enabled);
CREATE INDEX IF NOT EXISTS idx_triggers_priority ON triggers(priority);

-- Tabla para casos de prueba
CREATE TABLE IF NOT EXISTS test_cases (
    id VARCHAR(255) PRIMARY KEY,
    bot_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    input_message TEXT NOT NULL,
    input_user_id VARCHAR(255) NOT NULL,
    input_context JSONB,
    input_metadata JSONB,
    expected_response TEXT,
    expected_next_step VARCHAR(255),
    expected_conditions JSONB,
    expected_triggers JSONB,
    expected_context JSONB,
    expected_timeout BIGINT DEFAULT 30000,
    conditions JSONB,
    triggers JSONB,
    status VARCHAR(50) DEFAULT 'pending',
    result_success BOOLEAN,
    result_actual_response TEXT,
    result_actual_next_step VARCHAR(255),
    result_executed_conditions JSONB,
    result_executed_triggers JSONB,
    result_actual_context JSONB,
    result_execution_time BIGINT,
    result_error TEXT,
    result_executed_at TIMESTAMP,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Índices para casos de prueba
CREATE INDEX IF NOT EXISTS idx_test_cases_bot_id ON test_cases(bot_id);
CREATE INDEX IF NOT EXISTS idx_test_cases_status ON test_cases(status);
CREATE INDEX IF NOT EXISTS idx_test_cases_created_at ON test_cases(created_at);

-- Tabla para suites de prueba
CREATE TABLE IF NOT EXISTS test_suites (
    id VARCHAR(255) PRIMARY KEY,
    bot_id VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    test_cases JSONB,
    status VARCHAR(50) DEFAULT 'pending',
    result_total_tests INTEGER,
    result_passed_tests INTEGER,
    result_failed_tests INTEGER,
    result_skipped_tests INTEGER,
    result_success_rate DECIMAL(5,2),
    result_execution_time BIGINT,
    result_started_at TIMESTAMP,
    result_completed_at TIMESTAMP,
    result_test_results JSONB,
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Índices para suites de prueba
CREATE INDEX IF NOT EXISTS idx_test_suites_bot_id ON test_suites(bot_id);
CREATE INDEX IF NOT EXISTS idx_test_suites_status ON test_suites(status);
CREATE INDEX IF NOT EXISTS idx_test_suites_created_at ON test_suites(created_at);

-- Tabla de relación entre suites y casos de prueba
CREATE TABLE IF NOT EXISTS test_suite_cases (
    suite_id VARCHAR(255) NOT NULL,
    test_case_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (suite_id, test_case_id),
    FOREIGN KEY (suite_id) REFERENCES test_suites(id) ON DELETE CASCADE,
    FOREIGN KEY (test_case_id) REFERENCES test_cases(id) ON DELETE CASCADE
);

-- Índices para la tabla de relación
CREATE INDEX IF NOT EXISTS idx_test_suite_cases_suite_id ON test_suite_cases(suite_id);
CREATE INDEX IF NOT EXISTS idx_test_suite_cases_test_case_id ON test_suite_cases(test_case_id);

-- Datos de ejemplo para desarrollo
INSERT INTO conditionals (id, bot_id, name, description, expression, type, priority) VALUES
('cond-001', 'bot-001', 'Usuario Nuevo', 'Verifica si el usuario es nuevo', '{{user_type}} == ''new''', 'simple', 1),
('cond-002', 'bot-001', 'Mensaje de Saludo', 'Verifica si el mensaje contiene saludos', '{{message}} contains ''hola'' || {{message}} contains ''buenos días''', 'complex', 2),
('cond-003', 'bot-001', 'Email Válido', 'Verifica si el email tiene formato válido', '{{email}} regex ''^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$''', 'regex', 3),
('cond-004', 'bot-001', 'Usuario Premium', 'Verifica si el usuario tiene suscripción premium', '{{subscription_type}} == ''premium'' && {{subscription_active}} == true', 'complex', 4);

INSERT INTO triggers (id, bot_id, name, description, event, condition_id, action_type, action_config, action_timeout, priority, enabled) VALUES
('trigger-001', 'bot-001', 'Bienvenida Usuario Nuevo', 'Envía mensaje de bienvenida a usuarios nuevos', 'message_received', 'cond-001', 'send_message', '{"message": "¡Bienvenido! Soy tu asistente virtual. ¿En qué puedo ayudarte?", "channel": "web"}', 5000, 1, true),
('trigger-002', 'bot-001', 'Respuesta a Saludos', 'Responde automáticamente a saludos', 'message_received', 'cond-002', 'send_message', '{"message": "¡Hola! ¿Cómo estás? ¿En qué puedo ayudarte hoy?", "channel": "web"}', 3000, 2, true),
('trigger-003', 'bot-001', 'Registro de Email', 'Registra email válido en la base de datos', 'message_received', 'cond-003', 'save_email', '{"table": "user_emails", "fields": ["user_id", "email", "created_at"]}', 10000, 3, true),
('trigger-004', 'bot-001', 'Funcionalidades Premium', 'Habilita funcionalidades premium para usuarios premium', 'message_received', 'cond-004', 'enable_premium_features', '{"features": ["advanced_ai", "priority_support", "custom_themes"]}', 2000, 4, true);

INSERT INTO test_cases (id, bot_id, name, description, input_message, input_user_id, input_context, expected_response, expected_conditions, expected_triggers, conditions, triggers, status) VALUES
('test-001', 'bot-001', 'Prueba Usuario Nuevo', 'Prueba el flujo de bienvenida para usuarios nuevos', 'Hola, soy nuevo aquí', 'user-001', '{"user_type": "new", "first_time": true}', '¡Bienvenido! Soy tu asistente virtual. ¿En qué puedo ayudarte?', '["cond-001"]', '["trigger-001"]', '["cond-001"]', '["trigger-001"]', 'pending'),
('test-002', 'bot-001', 'Prueba Saludo', 'Prueba la respuesta automática a saludos', '¡Hola! ¿Cómo estás?', 'user-002', '{"user_type": "existing"}', '¡Hola! ¿Cómo estás? ¿En qué puedo ayudarte hoy?', '["cond-002"]', '["trigger-002"]', '["cond-002"]', '["trigger-002"]', 'pending'),
('test-003', 'bot-001', 'Prueba Email Válido', 'Prueba el registro de email válido', 'Mi email es usuario@ejemplo.com', 'user-003', '{"email": "usuario@ejemplo.com"}', '', '["cond-003"]', '["trigger-003"]', '["cond-003"]', '["trigger-003"]', 'pending'),
('test-004', 'bot-001', 'Prueba Usuario Premium', 'Prueba la activación de funcionalidades premium', 'Quiero usar las funciones premium', 'user-004', '{"subscription_type": "premium", "subscription_active": true}', '', '["cond-004"]', '["trigger-004"]', '["cond-004"]', '["trigger-004"]', 'pending');

INSERT INTO test_suites (id, bot_id, name, description, test_cases, status) VALUES
('suite-001', 'bot-001', 'Suite de Pruebas Básicas', 'Suite de pruebas para funcionalidades básicas del bot', '["test-001", "test-002", "test-003", "test-004"]', 'pending');

INSERT INTO test_suite_cases (suite_id, test_case_id) VALUES
('suite-001', 'test-001'),
('suite-001', 'test-002'),
('suite-001', 'test-003'),
('suite-001', 'test-004');

-- Triggers de base de datos para actualizar updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Aplicar triggers a todas las tablas
CREATE TRIGGER update_conditionals_updated_at BEFORE UPDATE ON conditionals FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_triggers_updated_at BEFORE UPDATE ON triggers FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_test_cases_updated_at BEFORE UPDATE ON test_cases FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_test_suites_updated_at BEFORE UPDATE ON test_suites FOR EACH ROW EXECUTE FUNCTION update_updated_at_column(); 