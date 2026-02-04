package coaching

import (
	"context"
	"fmt"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/anthropics/anthropic-sdk-go/packages/param"
)

const (
	DefaultModel       = anthropic.ModelClaudeSonnet4_20250514
	DefaultMaxTokens   = int64(2048)
	DefaultTemperature = 0.7
)

type ClaudeConfig struct {
	APIKey      string
	Model       string
	MaxTokens   int64
	Temperature float64
}

type ClaudeClient struct {
	client      anthropic.Client
	model       anthropic.Model
	maxTokens   int64
	temperature float64
}

func NewClaudeClient(cfg ClaudeConfig) (*ClaudeClient, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("anthropic API key is required")
	}

	client := anthropic.NewClient(option.WithAPIKey(cfg.APIKey))

	model := DefaultModel
	if cfg.Model != "" {
		model = anthropic.Model(cfg.Model)
	}

	maxTokens := DefaultMaxTokens
	if cfg.MaxTokens > 0 {
		maxTokens = cfg.MaxTokens
	}

	temperature := DefaultTemperature
	if cfg.Temperature > 0 {
		temperature = cfg.Temperature
	}

	return &ClaudeClient{
		client:      client,
		model:       model,
		maxTokens:   maxTokens,
		temperature: temperature,
	}, nil
}

func (c *ClaudeClient) Complete(ctx context.Context, system string, user string) (string, error) {
	params := anthropic.MessageNewParams{
		Model:     c.model,
		MaxTokens: c.maxTokens,
		Temperature: param.NewOpt(c.temperature),
		System: []anthropic.TextBlockParam{
			{Text: system},
		},
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(user)),
		},
	}

	response, err := c.client.Messages.New(ctx, params)
	if err != nil {
		return "", fmt.Errorf("claude API error: %w", err)
	}

	return extractText(response), nil
}

func extractText(msg *anthropic.Message) string {
	var parts []string
	for _, block := range msg.Content {
		if block.Type == "text" {
			parts = append(parts, block.Text)
		}
	}
	return strings.Join(parts, "")
}
