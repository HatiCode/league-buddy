package coaching

import "context"

// LLMClient is a provider-agnostic interface for LLM text completion.
type LLMClient interface {
	Complete(ctx context.Context, system string, user string) (string, error)
}
