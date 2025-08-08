package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/company/bot-service/pkg/logger"
)

// AIAgent implementa un agente que usa servicios de IA
type aiAgent struct {
	*baseAgent
	client    *http.Client
	apiKey    string
	model     string
	baseURL   string
	useMock   bool
	mockResponses []string
	mockIndex     int
}

// OpenAI API structures
type openAIRequest struct {
	Model       string    `json:"model"`
	Messages    []message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []choice `json:"choices"`
	Usage   usage    `json:"usage"`
}

type choice struct {
	Index        int     `json:"index"`
	Message      message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// NewAIAgent crea un nuevo agente de IA
func NewAIAgent(config MCPConfig, logger logger.Logger) (Agent, error) {
	base := newBaseAgent(config, logger)
	base.capabilities = []string{"text_generation", "conversation", "analysis", "summarization"}
	
	// Obtener configuración
	apiKey, _ := config.Config["openai_api_key"].(string)
	model, _ := config.Config["model"].(string)
	if model == "" {
		model = "gpt-3.5-turbo"
	}
	
	baseURL := "https://api.openai.com/v1"
	if customURL, exists := config.Config["base_url"].(string); exists {
		baseURL = customURL
	}
	
	// Determinar si usar mock
	useMock := apiKey == "" || apiKey == "sk-test-key"
	
	// Respuestas mock por defecto
	mockResponses := []string{
		"Hello! I'm an AI assistant ready to help you with your questions and tasks.",
		"I understand your request. Let me provide you with a comprehensive response based on the information provided.",
		"Thank you for your question. Here's my analysis and recommendations for your situation.",
		"Based on the context you've provided, I can offer the following insights and suggestions.",
		"I've processed your request and generated a response that should address your needs effectively.",
	}
	
	// Configurar timeout
	timeout := 30 * time.Second
	if config.Timeout > 0 {
		timeout = config.Timeout
	}
	
	return &aiAgent{
		baseAgent: base,
		client: &http.Client{
			Timeout: timeout,
		},
		apiKey:        apiKey,
		model:         model,
		baseURL:       baseURL,
		useMock:       useMock,
		mockResponses: mockResponses,
		mockIndex:     0,
	}, nil
}

func (a *aiAgent) Execute(ctx context.Context, task Task) (Result, error) {
	start := time.Now()
	
	// Actualizar estado
	a.mu.Lock()
	a.state.Status = AgentStatusBusy
	a.state.CurrentTask = &task
	a.mu.Unlock()
	
	defer func() {
		a.mu.Lock()
		a.state.Status = AgentStatusIdle
		a.state.CurrentTask = nil
		a.mu.Unlock()
	}()
	
	a.logger.Info("AI agent executing task", 
		"agent_id", a.id,
		"task_id", task.ID,
		"task_type", task.Type,
		"use_mock", a.useMock)
	
	var result Result
	var err error
	
	if a.useMock {
		result, err = a.executeMockTask(task)
	} else {
		result, err = a.executeRealTask(ctx, task)
	}
	
	duration := time.Since(start)
	result.Duration = duration
	
	a.updateMetrics(result.Success, duration)
	
	a.logger.Info("AI agent task completed", 
		"agent_id", a.id,
		"task_id", task.ID,
		"duration", duration,
		"success", result.Success)
	
	return result, err
}

func (a *aiAgent) executeMockTask(task Task) (Result, error) {
	// Simular procesamiento
	time.Sleep(200 * time.Millisecond)
	
	// Obtener respuesta mock
	response := a.mockResponses[a.mockIndex]
	a.mockIndex = (a.mockIndex + 1) % len(a.mockResponses)
	
	// Personalizar respuesta basada en el input
	if prompt, exists := task.Input["prompt"].(string); exists && prompt != "" {
		if strings.Contains(strings.ToLower(prompt), "email") {
			response = "Subject: Professional Response\n\nDear Customer,\n\nThank you for your inquiry. We appreciate your interest in our services and would be happy to provide you with the information you requested.\n\nBest regards,\nCustomer Service Team"
		} else if strings.Contains(strings.ToLower(prompt), "summary") {
			response = "Summary: Based on the provided information, the key points are: 1) Main topic identification, 2) Key insights extraction, 3) Actionable recommendations. This analysis provides a comprehensive overview of the subject matter."
		} else if strings.Contains(strings.ToLower(prompt), "analysis") {
			response = "Analysis Results: The data shows positive trends with several key indicators pointing toward successful outcomes. Recommendations include continued monitoring and strategic adjustments as needed."
		}
	}
	
	return Result{
		TaskID:  task.ID,
		Success: true,
		Output: map[string]interface{}{
			"text":         response,
			"model":        a.model + "-mock",
			"tokens_used":  len(response) / 4, // Rough token estimation
			"finish_reason": "stop",
		},
		Metadata: map[string]interface{}{
			"agent_id":   a.id,
			"agent_type": a.agentType,
			"mode":       "mock",
		},
	}, nil
}

func (a *aiAgent) executeRealTask(ctx context.Context, task Task) (Result, error) {
	// Extraer prompt de la tarea
	prompt, exists := task.Input["prompt"].(string)
	if !exists || prompt == "" {
		return Result{
			TaskID:  task.ID,
			Success: false,
			Error:   "prompt is required for AI tasks",
		}, fmt.Errorf("prompt is required")
	}
	
	// Configurar parámetros
	temperature := 0.7
	if temp, exists := task.Input["temperature"].(float64); exists {
		temperature = temp
	}
	
	maxTokens := 1000
	if tokens, exists := task.Input["max_tokens"].(float64); exists {
		maxTokens = int(tokens)
	}
	
	// Preparar mensajes
	messages := []message{
		{
			Role:    "user",
			Content: prompt,
		},
	}
	
	// Agregar contexto del sistema si existe
	if systemPrompt, exists := task.Input["system"].(string); exists && systemPrompt != "" {
		messages = append([]message{{
			Role:    "system",
			Content: systemPrompt,
		}}, messages...)
	}
	
	// Crear request
	reqBody := openAIRequest{
		Model:       a.model,
		Messages:    messages,
		Temperature: temperature,
		MaxTokens:   maxTokens,
	}
	
	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return Result{
			TaskID:  task.ID,
			Success: false,
			Error:   fmt.Sprintf("failed to marshal request: %v", err),
		}, err
	}
	
	// Crear HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", a.baseURL+"/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return Result{
			TaskID:  task.ID,
			Success: false,
			Error:   fmt.Sprintf("failed to create request: %v", err),
		}, err
	}
	
	// Agregar headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	
	// Ejecutar request
	resp, err := a.client.Do(req)
	if err != nil {
		return Result{
			TaskID:  task.ID,
			Success: false,
			Error:   fmt.Sprintf("API request failed: %v", err),
		}, err
	}
	defer resp.Body.Close()
	
	// Leer respuesta
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Result{
			TaskID:  task.ID,
			Success: false,
			Error:   fmt.Sprintf("failed to read response: %v", err),
		}, err
	}
	
	if resp.StatusCode != http.StatusOK {
		return Result{
			TaskID:  task.ID,
			Success: false,
			Error:   fmt.Sprintf("API error %d: %s", resp.StatusCode, string(body)),
		}, fmt.Errorf("API error: %d", resp.StatusCode)
	}
	
	// Parsear respuesta
	var openAIResp openAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return Result{
			TaskID:  task.ID,
			Success: false,
			Error:   fmt.Sprintf("failed to parse response: %v", err),
		}, err
	}
	
	if len(openAIResp.Choices) == 0 {
		return Result{
			TaskID:  task.ID,
			Success: false,
			Error:   "no choices in API response",
		}, fmt.Errorf("no choices in response")
	}
	
	// Extraer respuesta
	choice := openAIResp.Choices[0]
	
	return Result{
		TaskID:  task.ID,
		Success: true,
		Output: map[string]interface{}{
			"text":          choice.Message.Content,
			"model":         openAIResp.Model,
			"tokens_used":   openAIResp.Usage.TotalTokens,
			"finish_reason": choice.FinishReason,
		},
		Metadata: map[string]interface{}{
			"agent_id":         a.id,
			"agent_type":       a.agentType,
			"mode":             "real",
			"prompt_tokens":    openAIResp.Usage.PromptTokens,
			"completion_tokens": openAIResp.Usage.CompletionTokens,
		},
	}, nil
}

func (a *aiAgent) CanHandle(taskType string) bool {
	supportedTypes := []string{
		"text_generation",
		"conversation",
		"analysis",
		"summarization",
		"ai",
		"openai",
		"gpt",
		"chat",
	}
	
	for _, supported := range supportedTypes {
		if taskType == supported {
			return true
		}
	}
	
	return false
}