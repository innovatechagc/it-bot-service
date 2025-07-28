package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/company/bot-service/internal/ai"
	"github.com/company/bot-service/internal/domain"
	"github.com/company/bot-service/internal/repositories"
	"github.com/company/bot-service/internal/services"
	"github.com/company/bot-service/pkg/logger"
)

func main() {
	logger := logger.NewLogger("info")
	
	// Initialize repositories
	botRepo := repositories.NewMockBotRepository()
	flowRepo := repositories.NewMockBotFlowRepository()
	stepRepo := repositories.NewMockBotStepRepository()
	smartReplyRepo := repositories.NewMockSmartReplyRepository()
	sessionRepo := repositories.NewMockConversationSessionRepository()
	
	// Initialize AI client
	aiClient := ai.NewMockAIClient([]string{
		"Hello! I'm here to help you with your questions.",
		"I understand what you're asking. Let me provide you with the information you need.",
		"Thank you for your message. Is there anything else I can help you with?",
	}, logger)
	
	// Initialize services
	conversationService := services.NewConversationService(sessionRepo, logger)
	smartReplyService := services.NewSmartReplyService(smartReplyRepo, aiClient, logger)
	botFlowService := services.NewBotFlowService(flowRepo, stepRepo, logger)
	botStepService := services.NewBotStepService(stepRepo, logger)
	botService := services.NewBotService(
		botRepo,
		flowRepo,
		stepRepo,
		sessionRepo,
		smartReplyRepo,
		conversationService,
		smartReplyService,
		logger,
	)
	
	ctx := context.Background()
	
	// Create sample bot
	bot := &domain.Bot{
		ID:      "bot-001",
		Name:    "Customer Support Bot",
		OwnerID: "owner-001",
		Channel: domain.ChannelWeb,
		Status:  domain.BotStatusActive,
		Config:  json.RawMessage(`{"welcome_message": "Hello! How can I help you today?"}`),
	}
	
	if err := botService.CreateBot(ctx, bot); err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}
	
	// Create sample flow
	flow := &domain.BotFlow{
		ID:         "flow-001",
		BotID:      "bot-001",
		Name:       "Welcome Flow",
		Trigger:    "hello",
		EntryPoint: "step-001",
		IsDefault:  true,
	}
	
	if err := botFlowService.CreateFlow(ctx, flow); err != nil {
		log.Fatalf("Failed to create flow: %v", err)
	}
	
	// Create sample steps
	step1Content := map[string]interface{}{
		"text": "Hello! Welcome to our support system. How can I help you today?",
		"type": "text",
		"options": []map[string]interface{}{
			{"id": "1", "text": "I have a question", "value": "question"},
			{"id": "2", "text": "I need technical support", "value": "support"},
			{"id": "3", "text": "I want to speak to a human", "value": "human"},
		},
	}
	step1ContentJSON, _ := json.Marshal(step1Content)
	
	step1 := &domain.BotStep{
		ID:         "step-001",
		FlowID:     "flow-001",
		Type:       domain.StepTypeMessage,
		Content:    step1ContentJSON,
		NextStepID: stringPtr("step-002"),
	}
	
	if err := botStepService.CreateStep(ctx, step1); err != nil {
		log.Fatalf("Failed to create step 1: %v", err)
	}
	
	// Create decision step
	step2Conditions := map[string]interface{}{
		"rules": []map[string]interface{}{
			{"condition": "question", "next_step": "step-003"},
			{"condition": "support", "next_step": "step-004"},
			{"condition": "human", "next_step": "step-005"},
		},
		"default": "step-006",
	}
	step2ConditionsJSON, _ := json.Marshal(step2Conditions)
	
	step2 := &domain.BotStep{
		ID:         "step-002",
		FlowID:     "flow-001",
		Type:       domain.StepTypeDecision,
		Content:    json.RawMessage(`{"text": "Processing your selection..."}`),
		Conditions: step2ConditionsJSON,
	}
	
	if err := botStepService.CreateStep(ctx, step2); err != nil {
		log.Fatalf("Failed to create step 2: %v", err)
	}
	
	// Create response steps
	step3Content := map[string]interface{}{
		"text": "I'd be happy to answer your question! Please go ahead and ask.",
		"type": "text",
	}
	step3ContentJSON, _ := json.Marshal(step3Content)
	
	step3 := &domain.BotStep{
		ID:      "step-003",
		FlowID:  "flow-001",
		Type:    domain.StepTypeAI,
		Content: step3ContentJSON,
	}
	
	if err := botStepService.CreateStep(ctx, step3); err != nil {
		log.Fatalf("Failed to create step 3: %v", err)
	}
	
	step4Content := map[string]interface{}{
		"text": "I'll help you with technical support. Can you describe the issue you're experiencing?",
		"type": "text",
	}
	step4ContentJSON, _ := json.Marshal(step4Content)
	
	step4 := &domain.BotStep{
		ID:      "step-004",
		FlowID:  "flow-001",
		Type:    domain.StepTypeInput,
		Content: step4ContentJSON,
	}
	
	if err := botStepService.CreateStep(ctx, step4); err != nil {
		log.Fatalf("Failed to create step 4: %v", err)
	}
	
	step5Content := map[string]interface{}{
		"text": "I'll connect you with a human agent. Please wait a moment...",
		"type": "text",
	}
	step5ContentJSON, _ := json.Marshal(step5Content)
	
	step5 := &domain.BotStep{
		ID:      "step-005",
		FlowID:  "flow-001",
		Type:    domain.StepTypeAPICall,
		Content: step5ContentJSON,
	}
	
	if err := botStepService.CreateStep(ctx, step5); err != nil {
		log.Fatalf("Failed to create step 5: %v", err)
	}
	
	// Create some smart replies
	smartReplies := []domain.SmartReply{
		{
			BotID:      "bot-001",
			Intent:     "greeting",
			Response:   "Hello! How can I help you today?",
			Confidence: 0.9,
		},
		{
			BotID:      "bot-001",
			Intent:     "goodbye",
			Response:   "Thank you for contacting us. Have a great day!",
			Confidence: 0.9,
		},
		{
			BotID:      "bot-001",
			Intent:     "help",
			Response:   "I'm here to help! You can ask me questions about our products and services.",
			Confidence: 0.8,
		},
	}
	
	if err := smartReplyService.TrainIntents(ctx, "bot-001", smartReplies); err != nil {
		log.Fatalf("Failed to train intents: %v", err)
	}
	
	fmt.Println("Sample data created successfully!")
	fmt.Println("Created:")
	fmt.Println("- 1 Bot (Customer Support Bot)")
	fmt.Println("- 1 Flow (Welcome Flow)")
	fmt.Println("- 5 Steps (Message, Decision, AI, Input, API Call)")
	fmt.Println("- 3 Smart Replies (Greeting, Goodbye, Help)")
	
	// Test message processing
	fmt.Println("\nTesting message processing...")
	
	message := &domain.IncomingMessage{
		ID:      "msg-001",
		BotID:   "bot-001",
		UserID:  "user-001",
		Content: "hello",
		Channel: domain.ChannelWeb,
		Metadata: map[string]interface{}{
			"source": "test",
		},
	}
	
	response, err := botService.ProcessIncomingMessage(ctx, message)
	if err != nil {
		log.Fatalf("Failed to process message: %v", err)
	}
	
	fmt.Printf("Bot Response: %s\n", response.Content)
	fmt.Printf("Response Type: %s\n", response.Type)
	if len(response.Options) > 0 {
		fmt.Println("Options:")
		for _, option := range response.Options {
			fmt.Printf("  - %s: %s\n", option.Text, option.Value)
		}
	}
}

func stringPtr(s string) *string {
	return &s
}