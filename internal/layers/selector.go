package layers

import (
	"context"

	"github.com/divanvisagie/ratatoskr/internal/config"
	"github.com/divanvisagie/ratatoskr/internal/logger"
	"github.com/divanvisagie/ratatoskr/pkg/types"
	openai "github.com/sashabaranov/go-openai"
)

type SelectionLayer struct {
	out          chan types.ResponseMessage
	capabilities *[]types.Capability
	cfg          config.Config
	logger       *logger.Logger
}

func NewSelectionLayer(cfg config.Config, caps *[]types.Capability) *SelectionLayer {
	layer := &SelectionLayer{
		out:          make(chan types.ResponseMessage),
		capabilities: caps,
		cfg:          cfg,
		logger:       logger.NewLogger("SelectionLayer"),
	}
	return layer
}

// Function that selects the most appropriate capability
func (s *SelectionLayer) selectCapability(msg types.RequestMessage) (*types.Capability, error) {
	client := openai.NewClient(s.cfg.OpenAIKey)

	tools := []openai.Tool{}
	for _, cap := range *s.capabilities {

		/*
			if any capabilities return 1 for a check that means they should override
			any AI selected capabilities	
		*/
		checkValue := cap.Check(msg)
		if checkValue == 1 {
			return &cap, nil
		}

		tools = append(tools, cap.Describe())
	}

	// Prompt that describes the task (selecting the capability)
	systemPrompt := "Select the best capability for the given message based on the descriptions provided below."

	// Create OpenAI function call request
	req := openai.ChatCompletionRequest{
		Model: openai.GPT4o, // Use GPT-4 or GPT-3.5 depending on your plan
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemPrompt,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: msg.Message, // The actual user message/request
			},
		},
		Tools: tools, // Pass the functions (capabilities) as part of the request
	}

	// Call the OpenAI API to select the appropriate function
	resp, err := client.CreateChatCompletion(context.Background(), req)
	if err != nil {
		s.logger.Error("Error calling OpenAI", err)
		return nil, err
	}

	if len(resp.Choices) == 0 || len(resp.Choices[0].Message.ToolCalls) == 0 {
		s.logger.Warn("No choices returned by OpenAI, defaulting to first capability")
		return &(*s.capabilities)[0], nil // Fallback to default if no choices are returned
	}

	// Parse the selected function (capability) from the response
	selectedFunction := resp.Choices[0].Message.ToolCalls[0].Function.Name

	s.logger.Info("Selected capability", selectedFunction)

	// Find and return the corresponding capability from the list
	for _, cap := range *s.capabilities {
		if cap.Describe().Function.Name == selectedFunction {
			return &cap, nil
		}
	}

	return nil, nil // Return nil if no matching capability is found
}

func (s *SelectionLayer) SendMessage(msg types.RequestMessage) {
	s.logger.Info("Received message", msg)

	// Select the appropriate capability using OpenAI's function-calling API
	cap, err := s.selectCapability(msg)
	if err != nil {
		s.logger.Error("Error selecting capability", err)
		response := types.ResponseMessage{
			ChatId:  msg.ChatId,
			Message: "I'm sorry, I'm having trouble processing your request",
		}
		s.out <- response
		return
	}

	if cap == nil {
		s.logger.Warn("No capability selected, defaulting to first capability")
		cap = &(*s.capabilities)[0] // Fallback to default if OpenAI didn't select one
	}

	// Now we do the work
	go types.ListenAndRespond(*cap, s.out)
	(*cap).SendMessage(msg)
}

func (s *SelectionLayer) ReceiveMessage() types.ResponseMessage {
	return <-s.out
}

func (s *SelectionLayer) GetUpdatesChan() chan types.ResponseMessage {
	return s.out
}
