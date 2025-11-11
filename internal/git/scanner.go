package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
)

type ChangeType string

const (
	ChangeTypeAdded    ChangeType = "add"
	ChangeTypeModified ChangeType = "edit"
	ChangeTypeDeleted  ChangeType = "del"
	ChangeTypeRenamed  ChangeType = "rename"
)

type FileChange struct {
	Path      string
	Type      ChangeType
	Additions int
	Deletions int
}

type Changes struct {
	Staged   []FileChange
	Unstaged []FileChange
	Summary  string
}

func IsGitRepo(dir string) (bool, error) {
	gitDir := filepath.Join(dir, ".git")
	info, err := os.Stat(gitDir)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return info.IsDir(), nil
}

func GetChanges() (*Changes, error) {
	workDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	isRepo, err := IsGitRepo(workDir)
	if err != nil {
		return nil, fmt.Errorf("failed to check git repo: %w", err)
	}
	if !isRepo {
		return nil, fmt.Errorf("not a git repository")
	}

	staged, err := getStagedChanges()
	if err != nil {
		return nil, fmt.Errorf("failed to get staged changes: %w", err)
	}

	unstaged, err := getUnstagedChanges()
	if err != nil {
		return nil, fmt.Errorf("failed to get unstaged changes: %w", err)
	}

	if len(staged) == 0 && len(unstaged) == 0 {
		return nil, fmt.Errorf("no uncommitted changes found")
	}

	summary := buildSummary(staged, unstaged)

	return &Changes{
		Staged:   staged,
		Unstaged: unstaged,
		Summary:  summary,
	}, nil
}

func getStagedChanges() ([]FileChange, error) {
	cmd := exec.Command("git", "diff", "--cached", "--numstat")
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 0 {
			return []FileChange{}, nil
		}
		return nil, fmt.Errorf("failed to run git diff --cached: %w", err)
	}

	return parseDiffOutput(string(output), true)
}

func getUnstagedChanges() ([]FileChange, error) {
	cmd := exec.Command("git", "diff", "--numstat")
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 0 {
			return []FileChange{}, nil
		}
		return nil, fmt.Errorf("failed to run git diff: %w", err)
	}

	return parseDiffOutput(string(output), false)
}

func parseDiffOutput(output string, staged bool) ([]FileChange, error) {
	if output == "" {
		return []FileChange{}, nil
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	changes := make([]FileChange, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		var additions, deletions int
		fmt.Sscanf(parts[0], "%d", &additions)
		fmt.Sscanf(parts[1], "%d", &deletions)

		filePath := parts[2]
		if len(parts) > 3 {
			filePath = strings.Join(parts[2:], " ")
		}

		changeType := determineChangeType(additions, deletions)

		changes = append(changes, FileChange{
			Path:      filePath,
			Type:      changeType,
			Additions: additions,
			Deletions: deletions,
		})
	}

	return changes, nil
}

func determineChangeType(additions, deletions int) ChangeType {
	if additions > 0 && deletions == 0 {
		return ChangeTypeAdded
	}
	if additions == 0 && deletions > 0 {
		return ChangeTypeDeleted
	}
	return ChangeTypeModified
}

func buildSummary(staged, unstaged []FileChange) string {
	var parts []string
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	if len(staged) > 0 {
		parts = append(parts, fmt.Sprintf("%s: %d file(s)", yellow("Staged"), len(staged)))
		for _, change := range staged {
			addStr := green(fmt.Sprintf("+%d", change.Additions))
			delStr := red(fmt.Sprintf("-%d", change.Deletions))
			parts = append(parts, fmt.Sprintf("  %s %s %s", addStr, delStr, change.Path))
		}
	}

	if len(unstaged) > 0 {
		parts = append(parts, fmt.Sprintf("%s: %d file(s)", yellow("Unstaged"), len(unstaged)))
		for _, change := range unstaged {
			addStr := green(fmt.Sprintf("+%d", change.Additions))
			delStr := red(fmt.Sprintf("-%d", change.Deletions))
			parts = append(parts, fmt.Sprintf("  %s %s %s", addStr, delStr, change.Path))
		}
	}

	return strings.Join(parts, "\n")
}

func GetDiffContent() (string, error) {
	var stagedDiff, unstagedDiff string

	cmd := exec.Command("git", "diff", "--cached")
	output, err := cmd.Output()
	if err == nil {
		stagedDiff = string(output)
	}

	cmd = exec.Command("git", "diff")
	output, err = cmd.Output()
	if err == nil {
		unstagedDiff = string(output)
	}

	var parts []string
	if stagedDiff != "" {
		parts = append(parts, "=== STAGED CHANGES ===")
		parts = append(parts, stagedDiff)
	}
	if unstagedDiff != "" {
		parts = append(parts, "=== UNSTAGED CHANGES ===")
		parts = append(parts, unstagedDiff)
	}

	return strings.Join(parts, "\n\n"), nil
}

