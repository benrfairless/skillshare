package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"skillshare/internal/config"
	"skillshare/internal/install"
	"skillshare/internal/ui"
)

func cmdUpdate(args []string) error {
	var name string
	var updateAll bool
	var dryRun bool
	var force bool

	// Parse arguments
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--all" || arg == "-a":
			updateAll = true
		case arg == "--dry-run" || arg == "-n":
			dryRun = true
		case arg == "--force" || arg == "-f":
			force = true
		case arg == "--help" || arg == "-h":
			printUpdateHelp()
			return nil
		case strings.HasPrefix(arg, "-"):
			return fmt.Errorf("unknown option: %s", arg)
		default:
			if name != "" {
				return fmt.Errorf("unexpected argument: %s", arg)
			}
			name = arg
		}
	}

	if name == "" && !updateAll {
		printUpdateHelp()
		return fmt.Errorf("specify a skill or repo name, or use --all")
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if updateAll {
		return updateAllTrackedRepos(cfg, dryRun, force)
	}

	// Determine if it's a tracked repo or regular skill
	return updateSkillOrRepo(cfg, name, dryRun, force)
}

func updateAllTrackedRepos(cfg *config.Config, dryRun, force bool) error {
	repos, err := install.GetTrackedRepos(cfg.Source)
	if err != nil {
		return fmt.Errorf("failed to get tracked repos: %w", err)
	}

	if len(repos) == 0 {
		ui.Info("No tracked repositories found")
		ui.Info("Use 'skillshare install <repo> --track' to add a tracked repository")
		return nil
	}

	ui.Header("Updating tracked repositories")
	fmt.Println(strings.Repeat("-", 45))

	hasError := false
	updated := 0
	skipped := 0

	for _, repo := range repos {
		repoPath := filepath.Join(cfg.Source, repo)

		// Check for uncommitted changes
		if isDirty, _ := isRepoDirty(repoPath); isDirty {
			if !force {
				ui.Warning("  %s: has uncommitted changes (use --force to override)", repo)
				skipped++
				continue
			}
			ui.Warning("  %s: has uncommitted changes, updating anyway (--force)", repo)
		}

		if dryRun {
			ui.Info("[dry-run] Would update %s", repo)
			continue
		}

		ui.Info("Updating %s...", repo)
		if err := gitPullRepo(repoPath); err != nil {
			ui.Error("  %s: %v", repo, err)
			hasError = true
			continue
		}

		ui.Success("  %s: updated", repo)
		updated++
	}

	if !dryRun && updated > 0 {
		fmt.Println()
		ui.Info("Run 'skillshare sync' to distribute updated skills to targets")
	}

	if skipped > 0 {
		fmt.Println()
		ui.Info("Skipped %d repo(s) with uncommitted changes", skipped)
	}

	if hasError {
		return fmt.Errorf("some repositories failed to update")
	}

	return nil
}

func updateSkillOrRepo(cfg *config.Config, name string, dryRun, force bool) error {
	// Try tracked repo first (with _ prefix)
	repoName := name
	if !strings.HasPrefix(repoName, "_") {
		repoName = "_" + name
	}
	repoPath := filepath.Join(cfg.Source, repoName)

	if install.IsGitRepo(repoPath) {
		return updateTrackedRepo(cfg, repoName, dryRun, force)
	}

	// Try as regular skill
	skillPath := filepath.Join(cfg.Source, name)
	if _, err := install.ReadMeta(skillPath); err == nil {
		return updateRegularSkill(cfg, name, dryRun, force)
	}

	// Check if it's a nested path that exists
	if install.IsGitRepo(skillPath) {
		return updateTrackedRepo(cfg, name, dryRun, force)
	}

	return fmt.Errorf("'%s' not found as tracked repo or skill with metadata", name)
}

func updateTrackedRepo(cfg *config.Config, repoName string, dryRun, force bool) error {
	repoPath := filepath.Join(cfg.Source, repoName)

	ui.Header("Updating tracked repository")
	fmt.Println(strings.Repeat("-", 45))
	ui.Info("Repository: %s", repoName)
	ui.Info("Path: %s", repoPath)
	fmt.Println()

	// Check for uncommitted changes
	if isDirty, _ := isRepoDirty(repoPath); isDirty {
		if !force {
			ui.Warning("Repository has uncommitted changes:")
			showDirtyFiles(repoPath)
			fmt.Println()
			ui.Error("Use --force to update anyway (may cause merge conflicts)")
			return fmt.Errorf("uncommitted changes in repository")
		}
		ui.Warning("Repository has uncommitted changes, updating anyway (--force)")
	}

	if dryRun {
		ui.Warning("[dry-run] Would run: git pull")
		return nil
	}

	ui.Info("Running git pull...")
	if err := gitPullRepo(repoPath); err != nil {
		return fmt.Errorf("git pull failed: %w", err)
	}

	ui.Success("Updated successfully")
	fmt.Println()
	ui.Info("Run 'skillshare sync' to distribute updated skills to targets")

	return nil
}

func updateRegularSkill(cfg *config.Config, skillName string, dryRun, force bool) error {
	skillPath := filepath.Join(cfg.Source, skillName)

	// Read metadata to get source
	meta, err := install.ReadMeta(skillPath)
	if err != nil {
		return fmt.Errorf("cannot read metadata for '%s': %w", skillName, err)
	}
	if meta == nil || meta.Source == "" {
		return fmt.Errorf("skill '%s' has no source metadata, cannot update", skillName)
	}

	ui.Header("Updating skill")
	fmt.Println(strings.Repeat("-", 45))
	ui.Info("Skill: %s", skillName)
	ui.Info("Source: %s", meta.Source)
	ui.Info("Path: %s", skillPath)
	fmt.Println()

	if dryRun {
		ui.Warning("[dry-run] Would reinstall from: %s", meta.Source)
		return nil
	}

	// Parse source and reinstall
	source, err := install.ParseSource(meta.Source)
	if err != nil {
		return fmt.Errorf("invalid source in metadata: %w", err)
	}

	opts := install.InstallOptions{
		Force:  true, // Always overwrite when updating
		Update: true,
	}

	ui.Info("Reinstalling from source...")
	result, err := install.Install(source, skillPath, opts)
	if err != nil {
		return fmt.Errorf("update failed: %w", err)
	}

	ui.Success("Updated: %s", result.SkillPath)
	for _, warning := range result.Warnings {
		ui.Warning("%s", warning)
	}

	fmt.Println()
	ui.Info("Run 'skillshare sync' to distribute updated skill to targets")

	return nil
}

func gitPullRepo(repoPath string) error {
	cmd := exec.Command("git", "pull")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s: %w", strings.TrimSpace(string(output)), err)
	}
	return nil
}

func showDirtyFiles(repoPath string) {
	cmd := exec.Command("git", "status", "--short")
	cmd.Dir = repoPath
	output, _ := cmd.Output()
	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		if line != "" {
			fmt.Printf("  %s\n", line)
		}
	}
}

func printUpdateHelp() {
	fmt.Println(`Usage: skillshare update <name> [options]
       skillshare update --all [options]

Update a skill or tracked repository.

For tracked repos (_repo-name): runs git pull
For regular skills: reinstalls from stored source metadata

Safety: Tracked repos with uncommitted changes are skipped by default.
Use --force to override (may cause merge conflicts).

Arguments:
  name                Skill name or tracked repo name

Options:
  --all, -a           Update all tracked repositories
  --force, -f         Force update even with uncommitted changes
  --dry-run, -n       Preview without making changes
  --help, -h          Show this help

Examples:
  skillshare update my-skill              # Update regular skill from source
  skillshare update _team-skills          # Update tracked repo (git pull)
  skillshare update team-skills           # _ prefix is optional for repos
  skillshare update --all                 # Update all tracked repos
  skillshare update --all --dry-run       # Preview updates
  skillshare update _team --force         # Update even with local changes`)
}
