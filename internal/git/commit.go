package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func StageAll() error {
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	gitRoot, err := FindGitRoot(workDir)
	if err != nil {
		return err
	}

	cmd := exec.Command("git", "add", "-A")
	cmd.Dir = gitRoot
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stage changes: %w", err)
	}
	return nil
}

func Commit(message string) error {
	if strings.TrimSpace(message) == "" {
		return fmt.Errorf("commit message cannot be empty")
	}

	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	gitRoot, err := FindGitRoot(workDir)
	if err != nil {
		return err
	}

	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Dir = gitRoot
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to create commit: %w", err)
	}
	return nil
}

func Push() error {
	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get working directory: %w", err)
	}

	gitRoot, err := FindGitRoot(workDir)
	if err != nil {
		return err
	}

	cmd := exec.Command("git", "push")
	cmd.Dir = gitRoot
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

