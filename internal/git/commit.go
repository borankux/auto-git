package git

import (
	"fmt"
	"os/exec"
	"strings"
)

func StageAll() error {
	cmd := exec.Command("git", "add", "-A")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stage changes: %w", err)
	}
	return nil
}

func Commit(message string) error {
	if strings.TrimSpace(message) == "" {
		return fmt.Errorf("commit message cannot be empty")
	}

	cmd := exec.Command("git", "commit", "-m", message)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create commit: %w", err)
	}
	return nil
}

func Push() error {
	cmd := exec.Command("git", "push")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to push: %w", err)
	}
	return nil
}

func CommitAndPush(message string) error {
	if err := Commit(message); err != nil {
		return err
	}

	if err := Push(); err != nil {
		return fmt.Errorf("commit successful but push failed: %w", err)
	}

	return nil
}

func StageAndCommitAndPush(message string) error {
	if err := StageAll(); err != nil {
		return fmt.Errorf("failed to stage changes: %w", err)
	}

	if err := Commit(message); err != nil {
		return err
	}

	if err := Push(); err != nil {
		return fmt.Errorf("commit successful but push failed: %w", err)
	}

	return nil
}

