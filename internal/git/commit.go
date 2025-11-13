package git

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

const defaultRemote = "origin"

func getGitRoot() (string, error) {
	workDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	gitRoot, err := FindGitRoot(workDir)
	if err != nil {
		return "", err
	}

	return gitRoot, nil
}

func StageAll() error {
	gitRoot, err := getGitRoot()
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

	gitRoot, err := getGitRoot()
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
	gitRoot, err := getGitRoot()
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

func CommitAndPush(message string) (bool, error) {
	if err := Commit(message); err != nil {
		return false, err
	}

	pushed, err := pushIfRemoteExists()
	if err != nil {
		return false, fmt.Errorf("commit successful but push failed: %w", err)
	}

	return pushed, nil
}

func StageAndCommitAndPush(message string) (bool, error) {
	if err := StageAll(); err != nil {
		return false, fmt.Errorf("failed to stage changes: %w", err)
	}

	if err := Commit(message); err != nil {
		return false, err
	}

	pushed, err := pushIfRemoteExists()
	if err != nil {
		return false, fmt.Errorf("commit successful but push failed: %w", err)
	}

	return pushed, nil
}

func pushIfRemoteExists() (bool, error) {
	hasOrigin, err := hasRemote(defaultRemote)
	if err != nil {
		return false, err
	}
	if !hasOrigin {
		return false, nil
	}

	if err := Push(); err != nil {
		return false, err
	}
	return true, nil
}

func hasRemote(remoteName string) (bool, error) {
	gitRoot, err := getGitRoot()
	if err != nil {
		return false, err
	}

	cmd := exec.Command("git", "remote")
	cmd.Dir = gitRoot
	output, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to list git remotes: %w", err)
	}

	list := strings.TrimSpace(string(output))
	if list == "" {
		return false, nil
	}

	for _, remote := range strings.Split(list, "\n") {
		if strings.TrimSpace(remote) == remoteName {
			return true, nil
		}
	}

	return false, nil
}
