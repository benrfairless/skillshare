package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"skillshare/internal/config"
	"skillshare/internal/install"
	"skillshare/internal/ui"
)

func cmdUninstall(args []string) error {
	var skillName string
	var force, dryRun bool

	// Parse arguments
	i := 0
	for i < len(args) {
		arg := args[i]
		switch {
		case arg == "--force" || arg == "-f":
			force = true
		case arg == "--dry-run" || arg == "-n":
			dryRun = true
		case arg == "--help" || arg == "-h":
			printUninstallHelp()
			return nil
		case strings.HasPrefix(arg, "-"):
			return fmt.Errorf("unknown option: %s", arg)
		default:
			if skillName != "" {
				return fmt.Errorf("unexpected argument: %s", arg)
			}
			skillName = arg
		}
		i++
	}

	if skillName == "" {
		printUninstallHelp()
		return fmt.Errorf("skill name is required")
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Normalize _ prefix for tracked repos
	// Check if name already has prefix, or if the path (with or without prefix) is a git repo
	if !strings.HasPrefix(skillName, "_") {
		// Try with _ prefix first
		prefixedPath := filepath.Join(cfg.Source, "_"+skillName)
		if install.IsGitRepo(prefixedPath) {
			skillName = "_" + skillName
		}
	}

	// Check if skill exists
	skillPath := filepath.Join(cfg.Source, skillName)
	info, err := os.Stat(skillPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("skill '%s' not found in source", skillName)
		}
		return fmt.Errorf("cannot access skill: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("'%s' is not a directory", skillName)
	}

	// Detect if this is a tracked repo
	isTrackedRepo := install.IsGitRepo(skillPath)

	// Display info
	if isTrackedRepo {
		ui.Header("Uninstalling tracked repository")
	} else {
		ui.Header("Uninstalling skill")
	}
	fmt.Println(strings.Repeat("-", 45))
	ui.Info("Name: %s", skillName)
	ui.Info("Path: %s", skillPath)
	if isTrackedRepo {
		ui.Info("Type: tracked repository")
	}

	// Show metadata if available (for regular skills)
	if !isTrackedRepo {
		if meta, err := install.ReadMeta(skillPath); err == nil && meta != nil {
			ui.Info("Source: %s", meta.Source)
			ui.Info("Installed: %s", meta.InstalledAt.Format("2006-01-02 15:04"))
		}
	}
	fmt.Println()

	// For tracked repos, check for uncommitted changes
	if isTrackedRepo && !dryRun {
		isDirty, dirtyErr := isRepoDirty(skillPath)
		if dirtyErr != nil {
			ui.Warning("Could not check git status: %v", dirtyErr)
		} else if isDirty {
			if !force {
				ui.Error("Repository has uncommitted changes!")
				ui.Info("Use --force to uninstall anyway, or commit/stash your changes first")
				return fmt.Errorf("uncommitted changes detected, use --force to override")
			}
			ui.Warning("Repository has uncommitted changes (proceeding with --force)")
		}
	}

	if dryRun {
		ui.Warning("[dry-run] would remove %s", skillPath)
		if isTrackedRepo {
			ui.Warning("[dry-run] would remove %s from .gitignore", skillName)
		}
		return nil
	}

	// Confirm unless --force
	if !force {
		prompt := "Are you sure you want to uninstall this skill?"
		if isTrackedRepo {
			prompt = "Are you sure you want to uninstall this tracked repository?"
		}
		fmt.Printf("%s [y/N]: ", prompt)
		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return err
		}
		input = strings.TrimSpace(strings.ToLower(input))
		if input != "y" && input != "yes" {
			ui.Info("Cancelled")
			return nil
		}
	}

	// For tracked repos, clean up .gitignore
	if isTrackedRepo {
		removed, gitignoreErr := install.RemoveFromGitIgnore(cfg.Source, skillName)
		if gitignoreErr != nil {
			ui.Warning("Could not update .gitignore: %v", gitignoreErr)
		} else if removed {
			ui.Info("Removed %s from .gitignore", skillName)
		}
	}

	// Remove the skill/repo
	if err := os.RemoveAll(skillPath); err != nil {
		return fmt.Errorf("failed to remove: %w", err)
	}

	if isTrackedRepo {
		ui.Success("Uninstalled tracked repository: %s", skillName)
	} else {
		ui.Success("Uninstalled: %s", skillName)
	}
	fmt.Println()
	ui.Info("Run 'skillshare sync' to update all targets")

	return nil
}

// isRepoDirty checks if a git repository has uncommitted changes
func isRepoDirty(repoPath string) (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return len(strings.TrimSpace(string(output))) > 0, nil
}

func printUninstallHelp() {
	fmt.Println(`Usage: skillshare uninstall <name> [options]

Remove a skill or tracked repository from the source directory.

For tracked repositories (_repo-name):
  - Checks for uncommitted changes (requires --force to override)
  - Automatically removes the entry from .gitignore
  - The _ prefix is optional (automatically detected)

Options:
  --force, -f     Skip confirmation and ignore uncommitted changes
  --dry-run, -n   Preview without making changes
  --help, -h      Show this help

Examples:
  skillshare uninstall my-skill              # Remove a skill
  skillshare uninstall my-skill --force      # Skip confirmation
  skillshare uninstall _team-repo            # Remove tracked repository
  skillshare uninstall team-repo             # _ prefix is optional
  skillshare uninstall _team-repo --force    # Force remove with uncommitted changes`)
}
