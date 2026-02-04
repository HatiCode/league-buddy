package coaching

import (
	"testing"

	"github.com/anthropics/anthropic-sdk-go"
)

func TestNewClaudeClientRequiresAPIKey(t *testing.T) {
	_, err := NewClaudeClient(ClaudeConfig{})
	if err == nil {
		t.Fatal("expected error for empty API key")
	}
}

func TestNewClaudeClientDefaults(t *testing.T) {
	client, err := NewClaudeClient(ClaudeConfig{APIKey: "test-key"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.model != DefaultModel {
		t.Errorf("model = %q, want %q", client.model, DefaultModel)
	}
	if client.maxTokens != DefaultMaxTokens {
		t.Errorf("maxTokens = %d, want %d", client.maxTokens, DefaultMaxTokens)
	}
	if client.temperature != DefaultTemperature {
		t.Errorf("temperature = %f, want %f", client.temperature, DefaultTemperature)
	}
}

func TestNewClaudeClientCustomConfig(t *testing.T) {
	client, err := NewClaudeClient(ClaudeConfig{
		APIKey:      "test-key",
		Model:       "claude-opus-4-5-20251101",
		MaxTokens:   4096,
		Temperature: 0.3,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.model != anthropic.ModelClaudeOpus4_5_20251101 {
		t.Errorf("model = %q, want %q", client.model, anthropic.ModelClaudeOpus4_5_20251101)
	}
	if client.maxTokens != 4096 {
		t.Errorf("maxTokens = %d, want 4096", client.maxTokens)
	}
	if client.temperature != 0.3 {
		t.Errorf("temperature = %f, want 0.3", client.temperature)
	}
}

func TestExtractText(t *testing.T) {
	msg := &anthropic.Message{
		Content: []anthropic.ContentBlockUnion{
			{Type: "text", Text: "Hello "},
			{Type: "text", Text: "world"},
		},
	}

	result := extractText(msg)
	if result != "Hello world" {
		t.Errorf("extractText = %q, want %q", result, "Hello world")
	}
}

func TestExtractTextEmpty(t *testing.T) {
	msg := &anthropic.Message{
		Content: []anthropic.ContentBlockUnion{},
	}

	result := extractText(msg)
	if result != "" {
		t.Errorf("extractText = %q, want empty", result)
	}
}
