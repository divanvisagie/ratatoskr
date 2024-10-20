package capabilities

import (
	"context"
	"io"
	"strings"

	"github.com/divanvisagie/ratatoskr/internal/config"
	"github.com/divanvisagie/ratatoskr/internal/logger"
	"github.com/divanvisagie/ratatoskr/pkg/types"
	"github.com/sashabaranov/go-openai"
)

type SayCapability struct {
	logger *logger.Logger
	out    chan types.ResponseMessage
	cfg    *config.Config
}

func NewSayCapability(cfg *config.Config) *SayCapability {
	sc := &SayCapability{
		logger: logger.NewLogger("SayCapability"),
		out:    make(chan types.ResponseMessage),
		cfg:    cfg,
	}

	go types.ListenAndRespond(sc, sc.out)

	return sc
}

func (sc *SayCapability) Tell(msg types.RequestMessage) {
	ctx := context.Background()
	// get the index of /say and get everything after it
	start := strings.Index(msg.Message, "/say")
	message := msg.Message[start+5:]

	client := openai.NewClient(sc.cfg.OpenAIKey)
	req := openai.CreateSpeechRequest{
		Input:          message,
		Voice:          openai.VoiceFable,
		ResponseFormat: openai.SpeechResponseFormatMp3,
		Model:          openai.TTSModel1,
	}

	sc.out <- types.ResponseMessage{
		ChatId:   msg.ChatId,
		DataType: types.BUSY,
	}

	res, err := client.CreateSpeech(ctx, req) //openai.RawResponse
	if err != nil {
		sc.logger.Error("Failed to create speech", err)
		message = "Failed to create speech"
	}

	// get the response in bytes from the raw response
	data, err := io.ReadAll(res.ReadCloser)
	if err != nil {
		sc.logger.Error("Failed to read response", err)
		message = "Failed to read response"
	}

	sc.out <- types.ResponseMessage{
		UserId:   msg.UserId,
		ChatId:   msg.ChatId,
		Message:  message,
		Data:     data,
		DataType: types.MP3,
	}
}

func (sc *SayCapability) GetUpdatesChan() chan types.ResponseMessage {
	return sc.out
}

func (sc *SayCapability) Check(msg types.RequestMessage) float64 {
	if strings.Contains(msg.Message, "/say") {
		return 1.0
	}
	return 0.0
}

func (sc *SayCapability) Describe() openai.Tool {
	fd := openai.FunctionDefinition{
		Name:        "Say",
		Description: "Command for when the user starts the message with /say, everything after /say should be turned into text to speech",
	}

	return openai.Tool{
		Type:     openai.ToolTypeFunction,
		Function: &fd,
	}
}
