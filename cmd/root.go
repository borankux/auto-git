package cmd

import (
	"fmt"
	"os"
	"strings"

	"auto-git/internal/config"
	"auto-git/internal/git"
	"auto-git/internal/ollama"
	"auto-git/internal/prompt"
	"auto-git/internal/ui"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "auto-git",
	Short: "Auto-generate commit messages using Ollama",
	Long:  `Auto-git scans your git repository for uncommitted changes and uses Ollama to generate commit messages.`,
	Run:   run,
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration",
}

var setModelCmd = &cobra.Command{
	Use:   "set-model [model-name]",
	Short: "Set the default Ollama model",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.LoadConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
			os.Exit(1)
		}

		client := ollama.NewClient("")

		spinner := ui.NewSpinner("Connecting to Ollama server...")
		if err := client.CheckConnection(); err != nil {
			spinner.Stop()
			fmt.Fprintf(os.Stderr, "Error connecting to Ollama: %v\n", err)
			os.Exit(1)
		}
		spinner.Stop()

		spinner = ui.NewSpinner("Fetching available models...")
		models, err := client.ListModels()
		spinner.Stop()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error listing models: %v\n", err)
			os.Exit(1)
		}

		if len(models) == 0 {
			fmt.Fprintf(os.Stderr, "No models available on Ollama server\n")
			os.Exit(1)
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
		fmt.Printf("Default model: %s\n", cfg.Model)
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

	client := ollama.NewClient("")

	spinner := ui.NewSpinner("Connecting to Ollama server...")
	if err := client.CheckConnection(); err != nil {
		spinner.Stop()
		fmt.Fprintf(os.Stderr, "Error connecting to Ollama: %v\n", err)
		os.Exit(1)
	}
	spinner.Stop()

	spinner = ui.NewSpinner("Fetching available models...")
	models, err := client.ListModels()
	spinner.Stop()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error listing models: %v\n", err)
		os.Exit(1)
	}

	if len(models) == 0 {
		fmt.Fprintf(os.Stderr, "No models available on Ollama server\n")
		os.Exit(1)
	}

	selectedModel := cfg.Model
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

	fmt.Printf("Using model: %s\n", selectedModel)

	systemPrompt, userPrompt := prompt.BuildFullPrompt(changes, diffContent)

	spinner = ui.NewSpinner("Generating commit message...")
	commitMessage, err := client.GenerateCommitMessage(selectedModel, systemPrompt, userPrompt)
	spinner.Stop()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error generating commit message: %v\n", err)
		os.Exit(1)
	}

	commitMessage = prompt.ExtractCommitMessage(commitMessage)

	if commitMessage == "" {
		fmt.Println("Generated commit message is empty. Please enter a commit message manually:")
		manualMessage, err := ui.EditCommitMessage("")
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		commitMessage = manualMessage
	} else {
		fmt.Printf("\nGenerated commit message:\n%s\n\n", commitMessage)
		fmt.Println("Edit the message if needed (or press Enter to use as-is):")

		editedMessage, err := ui.EditCommitMessage(commitMessage)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		commitMessage = editedMessage
	}

	if strings.TrimSpace(commitMessage) == "" {
		fmt.Fprintf(os.Stderr, "Commit message cannot be empty\n")
		os.Exit(1)
	}

	spinner = ui.NewSpinner(fmt.Sprintf("Committing and pushing: %s", commitMessage))
	if err := git.StageAndCommitAndPush(commitMessage); err != nil {
		spinner.Stop()
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	spinner.Stop()

	fmt.Println("Successfully committed and pushed!")
}
