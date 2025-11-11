package prompt

import (
	"strings"

	"auto-git/internal/git"
)

func BuildSystemPrompt() string {
	return `You are an expert git commit message writer. Your task is to analyze git changes and generate concise, meaningful commit messages following the Conventional Commits specification.

Guidelines:
- Use conventional commit format: <type>(<scope>): <subject>
- Types: feat (new feature), fix (bug fix), core (core functionality), edit (edits/modifications), del (deletions), chore (maintenance), docs (documentation), style (formatting), refactor (code restructuring), perf (performance), test (tests), ci (CI/CD)
- Keep the subject line under 72 characters
- Use imperative mood ("add feature" not "added feature")
- Be specific and descriptive
- If multiple types apply, choose the most significant one

Analyze the changes and generate a single-line commit message.`
}

func BuildUserPrompt(changes *git.Changes, diffContent string) string {
	var parts []string

	parts = append(parts, "Analyze the following git changes and generate an appropriate commit message:")
	parts = append(parts, "")
	parts = append(parts, "=== CHANGE SUMMARY ===")
	parts = append(parts, changes.Summary)
	parts = append(parts, "")
	parts = append(parts, "=== DIFF CONTENT ===")
	parts = append(parts, diffContent)
	parts = append(parts, "")
	parts = append(parts, "Generate a commit message following the conventional commit format:")

	return strings.Join(parts, "\n")
}

func BuildFullPrompt(changes *git.Changes, diffContent string) (string, string) {
	systemPrompt := BuildSystemPrompt()
	userPrompt := BuildUserPrompt(changes, diffContent)
	return systemPrompt, userPrompt
}

func ExtractCommitMessage(response string) string {
	response = strings.TrimSpace(response)
	
	lines := strings.Split(response, "\n")
	if len(lines) == 0 {
		return ""
	}

	firstLine := strings.TrimSpace(lines[0])
	
	if firstLine == "" {
		return ""
	}

	if strings.HasPrefix(strings.ToLower(firstLine), "commit message:") {
		firstLine = strings.TrimPrefix(firstLine, "commit message:")
		firstLine = strings.TrimPrefix(firstLine, "Commit message:")
		firstLine = strings.TrimSpace(firstLine)
	}

	if strings.HasPrefix(firstLine, "```") {
		if len(lines) > 1 {
			firstLine = strings.TrimSpace(lines[1])
		} else {
			return ""
		}
	}

	if strings.HasSuffix(firstLine, "```") {
		firstLine = strings.TrimSuffix(firstLine, "```")
		firstLine = strings.TrimSpace(firstLine)
	}

	if len(firstLine) > 72 {
		firstLine = firstLine[:72]
	}

	return firstLine
}

func AnalyzeChangeTypes(changes *git.Changes) []string {
	typeCount := make(map[string]int)
	
	for _, change := range changes.Staged {
		typeCount[string(change.Type)]++
	}
	for _, change := range changes.Unstaged {
		typeCount[string(change.Type)]++
	}

	var types []string
	for t := range typeCount {
		types = append(types, t)
	}
	
	return types
}

func SuggestCommitType(changes *git.Changes) string {
	hasAdditions := false
	hasDeletions := false
	hasModifications := false

	for _, change := range changes.Staged {
		if change.Type == git.ChangeTypeAdded {
			hasAdditions = true
		} else if change.Type == git.ChangeTypeDeleted {
			hasDeletions = true
		} else if change.Type == git.ChangeTypeModified {
			hasModifications = true
		}
	}

	for _, change := range changes.Unstaged {
		if change.Type == git.ChangeTypeAdded {
			hasAdditions = true
		} else if change.Type == git.ChangeTypeDeleted {
			hasDeletions = true
		} else if change.Type == git.ChangeTypeModified {
			hasModifications = true
		}
	}

	if hasDeletions && !hasAdditions && !hasModifications {
		return "del"
	}
	if hasAdditions && !hasModifications {
		return "feat"
	}
	if hasModifications {
		return "fix"
	}
	return "chore"
}

