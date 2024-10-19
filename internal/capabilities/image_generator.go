package capabilities

import (
	"context"
	_ "embed"
	"io"
	"net/http"

	"github.com/divanvisagie/ratatoskr/internal/config"
	"github.com/divanvisagie/ratatoskr/internal/logger"
	"github.com/divanvisagie/ratatoskr/pkg/types"

	o "github.com/sashabaranov/go-openai"
)

type ImageGenerationCapability struct {
	cfg    *config.Config
	out    chan types.ResponseMessage
	done   chan bool
	logger *logger.Logger
}

// create constructor function
func NewImageGenerationCapability(cfg *config.Config) *ImageGenerationCapability {
	// load prompt string from file prompt.txt and include at build time

	instance := &ImageGenerationCapability{
		cfg:    cfg,
		out:    make(chan types.ResponseMessage),
		done:   make(chan bool),
		logger: logger.NewLogger("ImageGenerationCapability"),
	}

	go types.ListenAndRespond(instance, instance.out)

	return instance
}

func (i *ImageGenerationCapability) Tell(msg types.RequestMessage) {
	client := o.NewClient(i.cfg.OpenAIKey)

	response, err := client.CreateImage(context.Background(), o.ImageRequest{
		Model:          o.CreateImageModelDallE3,
		Size:           o.CreateImageSize1024x1024,
		User:           "async-openai",
		Prompt:         msg.Message,
		ResponseFormat: o.CreateImageResponseFormatURL,
	})
	if err != nil {
		i.out <- types.ResponseMessage{
			UserId:  msg.UserId,
			ChatId:  msg.ChatId,
			Message: "Failed to generate image with dalle",
		}
		return
	}

	// get image from response
	responseImageUrl := response.Data
	i.logger.Info("Generated image with dalle: %s", responseImageUrl)

	// download image form url and represent it as a byte array
	imageBytes, err := downloadImage(responseImageUrl[0].URL)
	if err != nil {
		i.out <- types.ResponseMessage{
			UserId:  msg.UserId,
			ChatId:  msg.ChatId,
			Message: "Failed to download image",
		}
		return
	}

	i.out <- types.ResponseMessage{
		UserId:  msg.UserId,
		ChatId:  msg.ChatId,
		Message: "I can't respond with images yet",
		Data:    imageBytes,
	}
}

func downloadImage(url string) ([]byte, error) {
	// download image from url
	client := http.Client{}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// read image into byte array
	imageBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return imageBytes, nil
}

func (i *ImageGenerationCapability) GetUpdatesChan() chan types.ResponseMessage {
	return i.out
}

func (i *ImageGenerationCapability) Describe() o.Tool {
	fd := o.FunctionDefinition{
		Name:        "ImageGenerationCapability",
		Description: "Can generate an image with dall-e if the user asks for it",
	}

	return o.Tool{
		Type:     o.ToolTypeFunction,
		Function: &fd,
	}
}

func (c *ImageGenerationCapability) Check(inputMessage types.RequestMessage) float64 {
	return 0.0
}
