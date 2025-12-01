package cmd

import (
	"fmt"
	"os"
	"strings"

	"auto-git/internal/config"
	"auto-git/internal/git"
	"auto-git/internal/ollama"
	"auto-git/internal/openai"
	"auto-git/internal/provider"
	"auto-git/internal/prompt"
	"auto-git/internal/ui"

	"github.com/spf13/cobra"
)

const (
	ProviderOllama      = "ollama"
	ProviderSiliconFlow = "siliconflow"
	ProviderOpenAI      = "openai"
)

// newProvider creates a new provider instance based on the provider type
func newProvider(providerType, endpoint, apiKey string) (provider.Provider, error) {
	providerType = strings.ToLower(strings.TrimSpace(providerType))

	switch providerType {
	case ProviderOllama:
		return ollama.NewClient(endpoint, apiKey), nil
	case ProviderSiliconFlow:
		return openai.NewClient(endpoint, apiKey, true), nil
	case ProviderOpenAI:
		return openai.NewClient(endpoint, apiKey, false), nil
	default:
		return nil, fmt.Errorf("unknown provider type: %s (supported: ollama, siliconflow, openai)", providerType)
	}
}

// getAPIKeyFromEnv retrieves the API key from environment variables based on provider type
func getAPIKeyFromEnv(providerType string) string {
	providerType = strings.ToLower(strings.TrimSpace(providerType))

	switch providerType {
	case ProviderOllama:
		return strings.TrimSpace(os.Getenv("OLLAMA_API_KEY"))
	case ProviderSiliconFlow:
		return strings.TrimSpace(os.Getenv("SILICON_KEY"))
	case ProviderOpenAI:
		return strings.TrimSpace(os.Getenv("OPENAI_API_KEY"))
	default:
		return ""
	}
}

var rootCmd = &cobra.Command{
	Use:   "auto-git",
	Short: "Auto-generate commit messages using LLM providers",
	Long:  `Auto-git scans your git repository for uncommitted changes and uses LLM providers (Ollama, SiliconFlow, OpenAI) to generate commit messages.`,
	Run:   run,
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
}

var setModelCmd = &cobra.Command{
	Use:   "set-model [model-name]",
	Short: "Set the default model",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}

		apiKey := getAPIKeyFromEnv(cfg.Provider)
		prov, err := newProvider(cfg.Provider, cfg.Endpoint, apiKey)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating provider: %v\n", err)
			os.Exit(1)
		}

		logAuthStatus(cfg.Provider, apiKey)

		spinner := ui.NewSpinner(fmt.Sprintf("Connecting to %s...", cfg.Provider))
		if err := prov.CheckConnection(); err != nil {
			spinner.Stop()
			fmt.Fprintf(os.Stderr, "Error connecting to %s: %v\n", cfg.Provider, err)
			os.Exit(1)
		}
		spinner.Stop()

		spinner = ui.NewSpinner("Fetching available models...")
		models, err := prov.ListModels()
		spinner.Stop()
		if err != nil {
			// If listing fails, allow manual entry
			fmt.Fprintf(os.Stderr, "Warning: Could not list models: %v\n", err)
			if len(args) == 0 {
				fmt.Fprintf(os.Stderr, "Please provide a model name: auto-git config set-model <model-name>\n")
				os.Exit(1)
			}
			selectedModel := args[0]
			if err := config.SetModel(selectedModel); err != nil {
				fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Model set to: %s\n", selectedModel)
			return
		}

		if len(models) == 0 {
			fmt.Fprintf(os.Stderr, "No models available. Please provide a model name manually.\n")
			if len(args) == 0 {
				os.Exit(1)
			}
			selectedModel := args[0]
			if err := config.SetModel(selectedModel); err != nil {
				fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Model set to: %s\n", selectedModel)
			return
		}

		var selectedModel string
		if len(args) == 1 {
			selectedModel = args[0]
			found := false
			for _, m := range models {
				if m.Name == selectedModel {
					found = true
					break
				}
			}
			if !found {
				fmt.Printf("Model '%s' not found. Please select a model:\n", selectedModel)
				selectedModel, err = ui.SelectModel(models, cfg.Model)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Error selecting model: %v\n", err)
					os.Exit(1)
				}
			}
		} else {
			fmt.Println("Select a model:")
			selectedModel, err = ui.SelectModel(models, cfg.Model)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error selecting model: %v\n", err)
				os.Exit(1)
			}
		}

		if err := config.SetModel(selectedModel); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Model set to: %s\n", selectedModel)
	},
}

var showConfigCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current configuration",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Provider: %s\n", cfg.Provider)
		if cfg.Endpoint != "" {
			fmt.Printf("Endpoint: %s\n", cfg.Endpoint)
		}
		fmt.Printf("Model: %s\n", cfg.Model)
	},
}

var setProviderCmd = &cobra.Command{
	Use:   "set-provider [provider]",
	Short: "Set the LLM provider (ollama, siliconflow, openai)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		providerType := strings.ToLower(strings.TrimSpace(args[0]))
		if providerType != ProviderOllama && providerType != ProviderSiliconFlow && providerType != ProviderOpenAI {
			fmt.Fprintf(os.Stderr, "Invalid provider: %s (supported: ollama, siliconflow, openai)\n", providerType)
			os.Exit(1)
		}

		if err := config.SetProvider(providerType); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Provider set to: %s\n", providerType)
	},
}

var setEndpointCmd = &cobra.Command{
	Use:   "set-endpoint [endpoint]",
	Short: "Set the API endpoint URL",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		endpoint := strings.TrimSpace(args[0])
		if err := config.SetEndpoint(endpoint); err != nil {
			fmt.Fprintf(os.Stderr, "Error saving config: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Endpoint set to: %s\n", endpoint)
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	configCmd.AddCommand(setModelCmd)
	configCmd.AddCommand(setProviderCmd)
	configCmd.AddCommand(setEndpointCmd)
	configCmd.AddCommand(showConfigCmd)
	rootCmd.AddCommand(configCmd)
}

func run(cmd *cobra.Command, args []string) {
	fmt.Println("Scanning git repository for changes...")

	changes, err := git.GetChanges()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Changes detected:")
	fmt.Println(changes.Summary)
	fmt.Println()

	diffContent, err := git.GetDiffContent()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting diff: %v\n", err)
		os.Exit(1)
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	apiKey := getAPIKeyFromEnv(cfg.Provider)
	prov, err := newProvider(cfg.Provider, cfg.Endpoint, apiKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating provider: %v\n", err)
		os.Exit(1)
	}

	logAuthStatus(cfg.Provider, apiKey)

	spinner := ui.NewSpinner(fmt.Sprintf("Connecting to %s...", cfg.Provider))
	if err := prov.CheckConnection(); err != nil {
		spinner.Stop()
		fmt.Fprintf(os.Stderr, "Error connecting to %s: %v\n", cfg.Provider, err)
		os.Exit(1)
	}
	spinner.Stop()

	selectedModel := cfg.Model

	// Try to list models and validate the selected model
	spinner = ui.NewSpinner("Fetching available models...")
	models, err := prov.ListModels()
	spinner.Stop()
	if err == nil && len(models) > 0 {
		found := false
		for _, m := range models {
			if m.Name == selectedModel {
				found = true
				break
			}
		}

		if !found {
			fmt.Printf("Model '%s' not found. Please select a model:\n", selectedModel)
			selected, err := ui.SelectModel(models, models[0].Name)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error selecting model: %v\n", err)
				os.Exit(1)
			}
			selectedModel = selected
			if err := config.SetModel(selectedModel); err != nil {
				fmt.Fprintf(os.Stderr, "Warning: failed to save model preference: %v\n", err)
			}
		}
	} else if err != nil {
		// If listing fails, continue with configured model
		fmt.Printf("Warning: Could not list models: %v. Using configured model: %s\n", err, selectedModel)
	}

	fmt.Printf("Using provider: %s, model: %s\n", cfg.Provider, selectedModel)

	systemPrompt, userPrompt := prompt.BuildFullPrompt(changes, diffContent)

	spinner = ui.NewSpinner("Generating commit message...")
	commitMessage, err := prov.GenerateCommitMessage(selectedModel, systemPrompt, userPrompt)
	spinner.Stop()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating commit message: %v\n", err)
		os.Exit(1)
	}

	commitMessage = prompt.ExtractCommitMessage(commitMessage)

	if strings.TrimSpace(commitMessage) == "" {
		fmt.Println("Generated commit message is empty. Please enter a commit message manually:")
		manualMessage, err := ui.EditCommitMessage("")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		commitMessage = manualMessage
		if strings.TrimSpace(commitMessage) == "" {
			fmt.Fprintf(os.Stderr, "Commit message cannot be empty\n")
			os.Exit(1)
		}
	} else {
		// Server responded with non-empty value - automate, don't pause
		fmt.Printf("\nGenerated commit message:\n%s\n\n", commitMessage)
		fmt.Println("Proceeding with commit and push...")
	}

	spinner = ui.NewSpinner(fmt.Sprintf("Recording git changes: %s", commitMessage))
	pushed, err := git.StageAndCommitAndPush(commitMessage)
	if err != nil {
		spinner.Stop()
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	spinner.Stop()

	if pushed {
		fmt.Println("Successfully committed and pushed!")
	} else {
		fmt.Println("Committed locally; remote 'origin' not configured, skipping push.")
	}
}

func logAuthStatus(providerType, apiKey string) {
	if apiKey == "" {
		var envVar string
		switch providerType {
		case ProviderOllama:
			envVar = "OLLAMA_API_KEY"
		case ProviderSiliconFlow:
			envVar = "SILICON_KEY"
		case ProviderOpenAI:
			envVar = "OPENAI_API_KEY"
		}
		fmt.Printf("Connecting to %s without %s (requests may be unauthenticated).\n", providerType, envVar)
		return
	}

	var envVar string
	switch providerType {
	case ProviderOllama:
		envVar = "OLLAMA_API_KEY"
	case ProviderSiliconFlow:
		envVar = "SILICON_KEY"
	case ProviderOpenAI:
		envVar = "OPENAI_API_KEY"
	}
	fmt.Printf("Using %s for authentication (%s)\n", envVar, maskAPIKey(apiKey))
}

func maskAPIKey(key string) string {
	const visible = 4
	if len(key) <= visible {
		return key
	}
	return strings.Repeat("*", len(key)-visible) + key[len(key)-visible:]
}
