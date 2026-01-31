package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"skillshare/internal/config"
	"skillshare/internal/ui"
)

// pushOptions holds parsed push command options
type pushOptions struct {
	dryRun  bool
	message string
}

// parsePushArgs parses push command arguments
func parsePushArgs(args []string) *pushOptions {
	opts := &pushOptions{message: "Update skills"}

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch arg {
		case "--dry-run", "-n":
			opts.dryRun = true
		case "-m", "--message":
			if i+1 < len(args) {
				i++
				opts.message = args[i]
			}
		default:
			if strings.HasPrefix(arg, "-m=") {
				opts.message = strings.TrimPrefix(arg, "-m=")
			} else if strings.HasPrefix(arg, "--message=") {
				opts.message = strings.TrimPrefix(arg, "--message=")
			}
		}
	}

	return opts
}

// checkGitRepo verifies source is a git repo with remote
func checkGitRepo(sourcePath string) error {
	gitDir := sourcePath + "/.git"
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		ui.Error("Source is not a git repository")
		ui.Info("  Run: cd %s && git init", sourcePath)
		return fmt.Errorf("not a git repository")
	}

	cmd := exec.Command("git", "remote")
	cmd.Dir = sourcePath
	output, err := cmd.Output()
	if err != nil || strings.TrimSpace(string(output)) == "" {
		ui.Error("No git remote configured")
		ui.Info("  Run: cd %s && git remote add origin <url>", sourcePath)
		ui.Info("  Or:  skillshare init --remote <url>")
		return fmt.Errorf("no remote configured")
	}

	return nil
}

// getGitChanges returns uncommitted changes
func getGitChanges(sourcePath string) (string, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = sourcePath
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to check git status: %w", err)
	}
	return strings.TrimSpace(string(output)), nil
}

// stageAndCommit stages all changes and commits
func stageAndCommit(sourcePath, message string) error {
	ui.Info("Staging changes...")
	cmd := exec.Command("git", "add", "-A")
	cmd.Dir = sourcePath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stage changes: %w", err)
	}

	ui.Info("Committing...")
	cmd = exec.Command("git", "commit", "-m", message)
	cmd.Dir = sourcePath
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to commit: %w", err)
	}

	return nil
}

// gitPush pushes to remote
func gitPush(sourcePath string) error {
	ui.Info("Pushing to remote...")
	cmd := exec.Command("git", "push")
	cmd.Dir = sourcePath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Println()
		ui.Error("Push failed - remote may have newer changes")
		ui.Info("  Run: skillshare pull --remote")
		ui.Info("  Then: skillshare push")
		return fmt.Errorf("push failed")
	}
	return nil
}

func cmdPush(args []string) error {
	opts := parsePushArgs(args)

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config not found: run 'skillshare init' first")
	}

	ui.Header("Pushing to remote")

	if err := checkGitRepo(cfg.Source); err != nil {
		return nil // Error already displayed
	}

	changes, err := getGitChanges(cfg.Source)
	if err != nil {
		return err
	}
	hasChanges := changes != ""

	if opts.dryRun {
		ui.Warning("[dry-run] No changes will be made")
		fmt.Println()
		if hasChanges {
			lines := strings.Split(changes, "\n")
			ui.Info("Would stage %d file(s):", len(lines))
			for _, line := range lines {
				ui.Info("  %s", line)
			}
			ui.Info("Would commit with message: %s", opts.message)
		} else {
			ui.Info("No changes to commit")
		}
		ui.Info("Would push to remote")
		return nil
	}

	if hasChanges {
		if err := stageAndCommit(cfg.Source, opts.message); err != nil {
			return err
		}
	} else {
		ui.Info("No changes to commit")
	}

	if err := gitPush(cfg.Source); err != nil {
		return nil // Error already displayed
	}

	fmt.Println()
	ui.Success("Pushed to remote")
	return nil
}
