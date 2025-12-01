package provider

// Model represents a language model available from a provider
type Model struct {
	Name       string `json:"name"`
	ModifiedAt string `json:"modified_at,omitempty"`
	Size       int64  `json:"size,omitempty"`
}

// Provider defines the interface that all LLM providers must implement
type Provider interface {
	// GenerateCommitMessage generates a commit message using the specified model and prompts
	GenerateCommitMessage(model string, systemPrompt, userPrompt string) (string, error)

	// ListModels returns a list of available models from the provider
	ListModels() ([]Model, error)

	// CheckConnection verifies that the provider is accessible
	CheckConnection() error
}
