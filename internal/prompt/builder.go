package prompt

import (
	"strings"

	"auto-git/internal/git"
)

func BuildSystemPrompt() string {
	return `You are an expert git commit message writer. Your task is to analyze git changes and generate concise, meaningful commit messages following the Conventional Commits specification.

Guidelines:
- Use conventional commit format: <type>(<scope>): <subject> or <emoji> <type>(<scope>): <subject>
- Types (STRICT - use exactly these): feat (new feature), fix (bug fix), core (core functionality), edit (edits/modifications), del (deletions), chore (maintenance), docs (documentation), style (formatting), refactor (code restructuring), perf (performance), test (tests), ci (CI/CD)
- Use emojis when appropriate (e.g., ✨ for feat, 🐛 for fix, 🗑️ for del, 📝 for docs, ♻️ for refactor, ⚡ for perf, 🎨 for style, 🔧 for chore)
- Keep messages compact but descriptive - prioritize clarity over strict length limits
- Use imperative mood ("add feature" not "added feature")
- Be specific and descriptive
- If multiple types apply, choose the most significant one
- Output exactly one line containing only the commit message (no explanations, code fences, or prefixes such as "Commit message:")
- Type must be lowercase and match one of the valid types exactly
`
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
	parts = append(parts, "Requirements:")
	parts = append(parts, "- Respond with exactly one line containing only the commit message.")
	parts = append(parts, "- Use the format <emoji> <type>(<optional scope>): <subject> or <type>(<scope>): <subject> (emojis are optional but encouraged).")
	parts = append(parts, "- Type MUST be one of: feat, fix, core, edit, del, chore, docs, style, refactor, perf, test, ci (lowercase, exact match).")
	parts = append(parts, "- Keep messages compact but descriptive - no strict length limit, prioritize clarity.")
	parts = append(parts, "- Write in imperative mood.")
	parts = append(parts, "- Do NOT include explanations, bullet lists, code fences, or backticks.")
	parts = append(parts, "- If unsure, default the type to chore.")
	parts = append(parts, "")
	parts = append(parts, "Return only the commit message text:")

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

	// Validate and normalize commit type
	firstLine = validateAndNormalizeCommitType(firstLine)

	return firstLine
}

// Valid commit types (must be lowercase)
var validCommitTypes = map[string]bool{
	"feat":     true,
	"fix":      true,
	"core":     true,
	"edit":     true,
	"del":      true,
	"chore":    true,
	"docs":     true,
	"style":    true,
	"refactor": true,
	"perf":     true,
	"test":     true,
	"ci":       true,
}

func validateAndNormalizeCommitType(message string) string {
	// Pattern: [emoji] type(scope): subject or type(scope): subject or type: subject
	// Extract the type part
	parts := strings.Fields(message)
	if len(parts) == 0 {
		return message
	}

	// Find the type - it's either the first part (if no emoji) or second part (if emoji present)
	typeIndex := 0
	// Check if first part is likely an emoji (contains non-ASCII or is a single character)
	if len(parts) > 1 && (len([]rune(parts[0])) == 1 || !isASCII(parts[0])) {
		typeIndex = 1
	}

	if typeIndex >= len(parts) {
		return message
	}

	typePart := parts[typeIndex]
	
	// Extract type from "type(scope):" or "type:"
	typeName := ""
	if strings.Contains(typePart, "(") {
		// Format: type(scope):
		idx := strings.Index(typePart, "(")
		typeName = strings.ToLower(typePart[:idx])
	} else if strings.Contains(typePart, ":") {
		// Format: type:
		idx := strings.Index(typePart, ":")
		typeName = strings.ToLower(typePart[:idx])
	} else {
		// No colon found, might be just the type word
		typeName = strings.ToLower(typePart)
	}

	// Validate type
	if !validCommitTypes[typeName] {
		// If type is invalid, try to fix common issues or default to chore
		// Check if it's a known type with wrong case
		for validType := range validCommitTypes {
			if strings.EqualFold(typeName, validType) {
				// Replace with correct lowercase type
				if typeIndex == 0 {
					parts[0] = strings.Replace(parts[0], typePart, validType, 1)
				} else {
					parts[typeIndex] = strings.Replace(parts[typeIndex], typePart, validType, 1)
				}
				return strings.Join(parts, " ")
			}
		}
		// If still not found, prepend "chore: " if message doesn't start with a valid type
		if !strings.HasPrefix(strings.ToLower(message), "chore") &&
			!strings.HasPrefix(strings.ToLower(message), "feat") &&
			!strings.HasPrefix(strings.ToLower(message), "fix") {
			return "chore: " + message
		}
	} else {
		// Type is valid, ensure it's lowercase in the message
		if typeName != typePart {
			correctedPart := strings.Replace(parts[typeIndex], typePart, typeName, 1)
			parts[typeIndex] = correctedPart
			return strings.Join(parts, " ")
		}
	}

	return message
}

func isASCII(s string) bool {
	for _, r := range s {
		if r > 127 {
			return false
		}
	}
	return true
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
