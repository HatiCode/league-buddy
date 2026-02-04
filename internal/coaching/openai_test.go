package coaching

import (
	"testing"

	"github.com/openai/openai-go/v3"
)

func TestNewOpenAIClientRequiresAPIKey(t *testing.T) {
	_, err := NewOpenAIClient(OpenAIConfig{})
	if err == nil {
		t.Fatal("expected error for empty API key")
	}
}

func TestNewOpenAIClientDefaults(t *testing.T) {
	client, err := NewOpenAIClient(OpenAIConfig{APIKey: "test-key"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.model != DefaultOpenAIModel {
		t.Errorf("model = %q, want %q", client.model, DefaultOpenAIModel)
	}
	if client.maxTokens != DefaultOpenAIMaxTokens {
		t.Errorf("maxTokens = %d, want %d", client.maxTokens, DefaultOpenAIMaxTokens)
	}
	if client.temperature != DefaultOpenAITemperature {
		t.Errorf("temperature = %f, want %f", client.temperature, DefaultOpenAITemperature)
	}
}

func TestNewOpenAIClientCustomConfig(t *testing.T) {
	client, err := NewOpenAIClient(OpenAIConfig{
		APIKey:      "test-key",
		Model:       "gpt-4o-mini",
		MaxTokens:   4096,
		Temperature: 0.3,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.model != openai.ChatModelGPT4oMini {
		t.Errorf("model = %q, want %q", client.model, openai.ChatModelGPT4oMini)
	}
	if client.maxTokens != 4096 {
		t.Errorf("maxTokens = %d, want 4096", client.maxTokens)
	}
	if client.temperature != 0.3 {
		t.Errorf("temperature = %f, want 0.3", client.temperature)
	}
}
