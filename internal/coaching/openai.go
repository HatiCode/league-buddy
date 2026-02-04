package coaching

import (
	"context"
	"fmt"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/option"
	"github.com/openai/openai-go/v3/packages/param"
)

const (
	DefaultOpenAIModel       = openai.ChatModelGPT4o
	DefaultOpenAIMaxTokens   = int64(2048)
	DefaultOpenAITemperature = 0.7
)

type OpenAIConfig struct {
	APIKey      string
	Model       string
	MaxTokens   int64
	Temperature float64
}

type OpenAIClient struct {
	client      openai.Client
	model       openai.ChatModel
	maxTokens   int64
	temperature float64
}

func NewOpenAIClient(cfg OpenAIConfig) (*OpenAIClient, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("openai API key is required")
	}

	client := openai.NewClient(option.WithAPIKey(cfg.APIKey))

	model := DefaultOpenAIModel
	if cfg.Model != "" {
		model = openai.ChatModel(cfg.Model)
	}

	maxTokens := DefaultOpenAIMaxTokens
	if cfg.MaxTokens > 0 {
		maxTokens = cfg.MaxTokens
	}

	temperature := DefaultOpenAITemperature
	if cfg.Temperature > 0 {
		temperature = cfg.Temperature
	}

	return &OpenAIClient{
		client:      client,
		model:       model,
		maxTokens:   maxTokens,
		temperature: temperature,
	}, nil
}

func (c *OpenAIClient) Complete(ctx context.Context, system string, user string) (string, error) {
	params := openai.ChatCompletionNewParams{
		Model: c.model,
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.SystemMessage(system),
			openai.UserMessage(user),
		},
		MaxCompletionTokens: param.NewOpt(c.maxTokens),
		Temperature:         param.NewOpt(c.temperature),
	}

	response, err := c.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return "", fmt.Errorf("openai API error: %w", err)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("openai returned no choices")
	}

	return response.Choices[0].Message.Content, nil
}
