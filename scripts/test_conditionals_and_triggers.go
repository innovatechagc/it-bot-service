package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"it-bot-service/internal/domain"
)

// Ejemplo de uso de condicionales y triggers
func main() {
	ctx := context.Background()

	fmt.Println("=== Ejemplo de Condicionales y Triggers ===")

	// 1. Crear condicionales de ejemplo
	conditionals := createExampleConditionals()
	fmt.Println("\n1. Condicionales creados:")
	for _, conditional := range conditionals {
		fmt.Printf("   - %s: %s\n", conditional.Name, conditional.Description)
	}

	// 2. Crear triggers de ejemplo
	triggers := createExampleTriggers()
	fmt.Println("\n2. Triggers creados:")
	for _, trigger := range triggers {
		fmt.Printf("   - %s: %s\n", trigger.Name, trigger.Description)
	}

	// 3. Crear casos de prueba de ejemplo
	testCases := createExampleTestCases()
	fmt.Println("\n3. Casos de prueba creados:")
	for _, testCase := range testCases {
		fmt.Printf("   - %s: %s\n", testCase.Name, testCase.Description)
	}

	// 4. Crear suite de pruebas de ejemplo
	testSuite := createExampleTestSuite()
	fmt.Println("\n4. Suite de pruebas creada:")
	fmt.Printf("   - %s: %s\n", testSuite.Name, testSuite.Description)

	// 5. Ejemplo de evaluación de condicionales
	fmt.Println("\n5. Ejemplo de evaluación de condicionales:")
	evaluateConditionalsExample()

	// 6. Ejemplo de ejecución de triggers
	fmt.Println("\n6. Ejemplo de ejecución de triggers:")
	executeTriggersExample()

	// 7. Ejemplo de ejecución de casos de prueba
	fmt.Println("\n7. Ejemplo de ejecución de casos de prueba:")
	executeTestCasesExample()

	fmt.Println("\n=== Fin del ejemplo ===")
}

// createExampleConditionals crea condicionales de ejemplo
func createExampleConditionals() []*domain.Conditional {
	conditionals := []*domain.Conditional{
		{
			ID:          "cond-001",
			BotID:       "bot-001",
			Name:        "Usuario Nuevo",
			Description: "Verifica si el usuario es nuevo",
			Expression:  "{{user_type}} == 'new'",
			Type:        domain.ConditionalTypeSimple,
			Priority:    1,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "cond-002",
			BotID:       "bot-001",
			Name:        "Mensaje de Saludo",
			Description: "Verifica si el mensaje contiene saludos",
			Expression:  "{{message}} contains 'hola' || {{message}} contains 'buenos días'",
			Type:        domain.ConditionalTypeComplex,
			Priority:    2,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "cond-003",
			BotID:       "bot-001",
			Name:        "Email Válido",
			Description: "Verifica si el email tiene formato válido",
			Expression:  "{{email}} regex `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}$`",
			Type:        domain.ConditionalTypeRegex,
			Priority:    3,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "cond-004",
			BotID:       "bot-001",
			Name:        "Usuario Premium",
			Description: "Verifica si el usuario tiene suscripción premium",
			Expression:  "{{subscription_type}} == 'premium' && {{subscription_active}} == true",
			Type:        domain.ConditionalTypeComplex,
			Priority:    4,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	return conditionals
}

// createExampleTriggers crea triggers de ejemplo
func createExampleTriggers() []*domain.Trigger {
	triggers := []*domain.Trigger{
		{
			ID:          "trigger-001",
			BotID:       "bot-001",
			Name:        "Bienvenida Usuario Nuevo",
			Description: "Envía mensaje de bienvenida a usuarios nuevos",
			Event:        domain.TriggerEventMessageReceived,
			Condition:    "cond-001",
			Action: domain.TriggerAction{
				Type: "send_message",
				Config: map[string]interface{}{
					"message": "¡Bienvenido! Soy tu asistente virtual. ¿En qué puedo ayudarte?",
					"channel": "web",
				},
				Timeout: 5000,
			},
			Priority:  1,
			Enabled:   true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "trigger-002",
			BotID:       "bot-001",
			Name:        "Respuesta a Saludos",
			Description: "Responde automáticamente a saludos",
			Event:        domain.TriggerEventMessageReceived,
			Condition:    "cond-002",
			Action: domain.TriggerAction{
				Type: "send_message",
				Config: map[string]interface{}{
					"message": "¡Hola! ¿Cómo estás? ¿En qué puedo ayudarte hoy?",
					"channel": "web",
				},
				Timeout: 3000,
			},
			Priority:  2,
			Enabled:   true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "trigger-003",
			BotID:       "bot-001",
			Name:        "Registro de Email",
			Description: "Registra email válido en la base de datos",
			Event:        domain.TriggerEventMessageReceived,
			Condition:    "cond-003",
			Action: domain.TriggerAction{
				Type: "save_email",
				Config: map[string]interface{}{
					"table": "user_emails",
					"fields": []string{"user_id", "email", "created_at"},
				},
				Timeout: 10000,
			},
			Priority:  3,
			Enabled:   true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "trigger-004",
			BotID:       "bot-001",
			Name:        "Funcionalidades Premium",
			Description: "Habilita funcionalidades premium para usuarios premium",
			Event:        domain.TriggerEventMessageReceived,
			Condition:    "cond-004",
			Action: domain.TriggerAction{
				Type: "enable_premium_features",
				Config: map[string]interface{}{
					"features": []string{"advanced_ai", "priority_support", "custom_themes"},
				},
				Timeout: 2000,
			},
			Priority:  4,
			Enabled:   true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	return triggers
}

// createExampleTestCases crea casos de prueba de ejemplo
func createExampleTestCases() []*domain.TestCase {
	testCases := []*domain.TestCase{
		{
			ID:          "test-001",
			BotID:       "bot-001",
			Name:        "Prueba Usuario Nuevo",
			Description: "Prueba el flujo de bienvenida para usuarios nuevos",
			Input: domain.TestInput{
				Message: "Hola, soy nuevo aquí",
				UserID:  "user-001",
				Context: map[string]interface{}{
					"user_type": "new",
					"first_time": true,
				},
			},
			Expected: domain.TestExpected{
				Response: "¡Bienvenido! Soy tu asistente virtual. ¿En qué puedo ayudarte?",
				Conditions: []string{"cond-001"},
				Triggers:   []string{"trigger-001"},
			},
			Conditions: []string{"cond-001"},
			Triggers:   []string{"trigger-001"},
			Status:     domain.TestStatusPending,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
		{
			ID:          "test-002",
			BotID:       "bot-001",
			Name:        "Prueba Saludo",
			Description: "Prueba la respuesta automática a saludos",
			Input: domain.TestInput{
				Message: "¡Hola! ¿Cómo estás?",
				UserID:  "user-002",
				Context: map[string]interface{}{
					"user_type": "existing",
				},
			},
			Expected: domain.TestExpected{
				Response: "¡Hola! ¿Cómo estás? ¿En qué puedo ayudarte hoy?",
				Conditions: []string{"cond-002"},
				Triggers:   []string{"trigger-002"},
			},
			Conditions: []string{"cond-002"},
			Triggers:   []string{"trigger-002"},
			Status:     domain.TestStatusPending,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
		{
			ID:          "test-003",
			BotID:       "bot-001",
			Name:        "Prueba Email Válido",
			Description: "Prueba el registro de email válido",
			Input: domain.TestInput{
				Message: "Mi email es usuario@ejemplo.com",
				UserID:  "user-003",
				Context: map[string]interface{}{
					"email": "usuario@ejemplo.com",
				},
			},
			Expected: domain.TestExpected{
				Conditions: []string{"cond-003"},
				Triggers:   []string{"trigger-003"},
			},
			Conditions: []string{"cond-003"},
			Triggers:   []string{"trigger-003"},
			Status:     domain.TestStatusPending,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
		{
			ID:          "test-004",
			BotID:       "bot-001",
			Name:        "Prueba Usuario Premium",
			Description: "Prueba la activación de funcionalidades premium",
			Input: domain.TestInput{
				Message: "Quiero usar las funciones premium",
				UserID:  "user-004",
				Context: map[string]interface{}{
					"subscription_type":  "premium",
					"subscription_active": true,
				},
			},
			Expected: domain.TestExpected{
				Conditions: []string{"cond-004"},
				Triggers:   []string{"trigger-004"},
			},
			Conditions: []string{"cond-004"},
			Triggers:   []string{"trigger-004"},
			Status:     domain.TestStatusPending,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		},
	}

	return testCases
}

// createExampleTestSuite crea una suite de pruebas de ejemplo
func createExampleTestSuite() *domain.TestSuite {
	return &domain.TestSuite{
		ID:          "suite-001",
		BotID:       "bot-001",
		Name:        "Suite de Pruebas Básicas",
		Description: "Suite de pruebas para funcionalidades básicas del bot",
		TestCases:   []string{"test-001", "test-002", "test-003", "test-004"},
		Status:      domain.TestSuiteStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// evaluateConditionalsExample muestra ejemplos de evaluación de condicionales
func evaluateConditionalsExample() {
	conditionals := createExampleConditionals()
	
	// Ejemplo 1: Usuario nuevo
	input1 := map[string]interface{}{
		"user_type": "new",
		"first_time": true,
	}
	
	fmt.Printf("   Evaluando 'Usuario Nuevo' con input: %v\n", input1)
	for _, conditional := range conditionals {
		if conditional.Name == "Usuario Nuevo" {
			fmt.Printf("   Resultado: %s = %v\n", conditional.Expression, true) // Simulado
			break
		}
	}
	
	// Ejemplo 2: Mensaje de saludo
	input2 := map[string]interface{}{
		"message": "¡Hola! ¿Cómo estás?",
	}
	
	fmt.Printf("   Evaluando 'Mensaje de Saludo' con input: %v\n", input2)
	for _, conditional := range conditionals {
		if conditional.Name == "Mensaje de Saludo" {
			fmt.Printf("   Resultado: %s = %v\n", conditional.Expression, true) // Simulado
			break
		}
	}
}

// executeTriggersExample muestra ejemplos de ejecución de triggers
func executeTriggersExample() {
	triggers := createExampleTriggers()
	
	// Ejemplo 1: Trigger de bienvenida
	fmt.Printf("   Ejecutando trigger 'Bienvenida Usuario Nuevo'\n")
	for _, trigger := range triggers {
		if trigger.Name == "Bienvenida Usuario Nuevo" {
			fmt.Printf("   Acción: %s\n", trigger.Action.Type)
			fmt.Printf("   Config: %v\n", trigger.Action.Config)
			break
		}
	}
	
	// Ejemplo 2: Trigger de respuesta a saludos
	fmt.Printf("   Ejecutando trigger 'Respuesta a Saludos'\n")
	for _, trigger := range triggers {
		if trigger.Name == "Respuesta a Saludos" {
			fmt.Printf("   Acción: %s\n", trigger.Action.Type)
			fmt.Printf("   Config: %v\n", trigger.Action.Config)
			break
		}
	}
}

// executeTestCasesExample muestra ejemplos de ejecución de casos de prueba
func executeTestCasesExample() {
	testCases := createExampleTestCases()
	
	for _, testCase := range testCases {
		fmt.Printf("   Ejecutando caso de prueba: %s\n", testCase.Name)
		fmt.Printf("   Input: %s\n", testCase.Input.Message)
		fmt.Printf("   Condiciones esperadas: %v\n", testCase.Expected.Conditions)
		fmt.Printf("   Triggers esperados: %v\n", testCase.Expected.Triggers)
		fmt.Printf("   Respuesta esperada: %s\n", testCase.Expected.Response)
		fmt.Println()
	}
}

// Función auxiliar para imprimir JSON de manera formateada
func printJSON(v interface{}) {
	jsonData, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return
	}
	fmt.Println(string(jsonData))
} 