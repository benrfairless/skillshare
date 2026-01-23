package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"skillshare/internal/config"
	"skillshare/internal/git"
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

	// Header
	ui.HeaderBox("skillshare update --all",
		fmt.Sprintf("Updating %d tracked repositories", len(repos)))
	fmt.Println()

	var updated, skipped, failed int
	var failedRepos []string

	for _, repo := range repos {
		repoPath := filepath.Join(cfg.Source, repo)

		// Check for uncommitted changes
		if isDirty, _ := git.IsDirty(repoPath); isDirty {
			if !force {
				ui.ListItem("warning", repo, "has uncommitted changes (use --force)")
				skipped++
				continue
			}
			if !dryRun {
				if err := git.Restore(repoPath); err != nil {
					ui.ListItem("error", repo, fmt.Sprintf("failed to discard changes: %v", err))
					failed++
					failedRepos = append(failedRepos, repo)
					continue
				}
			}
		}

		if dryRun {
			ui.ListItem("info", repo, "[dry-run] would update")
			continue
		}

		// Pull (use ForcePull if --force to handle force push)
		var info *git.UpdateInfo
		if force {
			info, err = git.ForcePull(repoPath)
		} else {
			info, err = git.Pull(repoPath)
		}
		if err != nil {
			ui.ListItem("error", repo, fmt.Sprintf("failed: %v", err))
			failed++
			failedRepos = append(failedRepos, repo)
			continue
		}

		if info.UpToDate {
			ui.ListItem("success", repo, "Already up to date")
		} else {
			detail := fmt.Sprintf("%d commits, %d files", len(info.Commits), info.Stats.FilesChanged)
			ui.ListItem("success", repo, detail)
		}
		updated++
	}

	// Summary
	if !dryRun {
		fmt.Println()
		ui.Box("Summary",
			"",
			fmt.Sprintf("  Updated:  %d", updated),
			fmt.Sprintf("  Skipped:  %d", skipped),
			fmt.Sprintf("  Failed:   %d", failed),
			"",
		)
	}

	if updated > 0 {
		fmt.Println()
		ui.Info("Run 'skillshare sync' to distribute changes")
	}

	for _, repo := range failedRepos {
		ui.Warning("%s failed to update", repo)
	}

	if failed > 0 {
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

	// Header box
	ui.HeaderBox("skillshare update", fmt.Sprintf("Updating: %s", repoName))
	fmt.Println()

	// Check for uncommitted changes
	spinner := ui.StartSpinner("Checking repository status...")

	isDirty, _ := git.IsDirty(repoPath)
	if isDirty {
		spinner.Stop()
		files, _ := git.GetDirtyFiles(repoPath)

		if !force {
			lines := []string{
				"",
				"Repository has uncommitted changes:",
				"",
			}
			lines = append(lines, files...)
			lines = append(lines, "", "Use --force to discard changes and update", "")

			ui.WarningBox("Warning", lines...)
			fmt.Println()
			ui.ErrorMsg("Update aborted")
			return fmt.Errorf("uncommitted changes in repository")
		}

		ui.Warning("Discarding local changes (--force)")
		if !dryRun {
			if err := git.Restore(repoPath); err != nil {
				return fmt.Errorf("failed to discard changes: %w", err)
			}
		}
		spinner = ui.StartSpinner("Fetching from origin...")
	}

	if dryRun {
		spinner.Stop()
		ui.Warning("[dry-run] Would run: git pull")
		return nil
	}

	spinner.Update("Fetching from origin...")

	// Use ForcePull if --force to handle force push
	var info *git.UpdateInfo
	var err error
	if force {
		info, err = git.ForcePull(repoPath)
	} else {
		info, err = git.Pull(repoPath)
	}
	if err != nil {
		spinner.Fail("Failed to update")
		return fmt.Errorf("git pull failed: %w", err)
	}

	if info.UpToDate {
		spinner.Success("Already up to date")
		return nil
	}

	spinner.Stop()
	fmt.Println()

	// Show changes box
	lines := []string{
		"",
		fmt.Sprintf("  Commits:  %d new", len(info.Commits)),
		fmt.Sprintf("  Files:    %d changed (+%d / -%d)",
			info.Stats.FilesChanged, info.Stats.Insertions, info.Stats.Deletions),
		"",
	}

	// Show up to 5 commits
	maxCommits := 5
	for i, c := range info.Commits {
		if i >= maxCommits {
			lines = append(lines, fmt.Sprintf("  ... and %d more", len(info.Commits)-maxCommits))
			break
		}
		lines = append(lines, fmt.Sprintf("  %s  %s", c.Hash, truncateString(c.Message, 40)))
	}
	lines = append(lines, "")

	ui.Box("Changes", lines...)
	fmt.Println()

	ui.SuccessMsg("Updated %s", repoName)
	fmt.Println()
	ui.Info("Run 'skillshare sync' to distribute changes")

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

	// Header box
	ui.HeaderBox("skillshare update",
		fmt.Sprintf("Updating: %s\nSource: %s", skillName, meta.Source))
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

	spinner := ui.StartSpinner("Cloning source repository...")

	opts := install.InstallOptions{
		Force:  true,
		Update: true,
	}

	result, err := install.Install(source, skillPath, opts)
	if err != nil {
		spinner.Fail("Failed to update")
		return fmt.Errorf("update failed: %w", err)
	}

	spinner.Success(fmt.Sprintf("Updated %s", skillName))

	for _, warning := range result.Warnings {
		ui.Warning("%s", warning)
	}

	fmt.Println()
	ui.Info("Run 'skillshare sync' to distribute changes")

	return nil
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

func printUpdateHelp() {
	fmt.Println(`Usage: skillshare update <name> [options]
       skillshare update --all [options]

Update a skill or tracked repository.

For tracked repos (_repo-name): runs git pull
For regular skills: reinstalls from stored source metadata

Safety: Tracked repos with uncommitted changes are skipped by default.
Use --force to discard local changes and update.

Arguments:
  name                Skill name or tracked repo name

Options:
  --all, -a           Update all tracked repositories
  --force, -f         Discard local changes and force update
  --dry-run, -n       Preview without making changes
  --help, -h          Show this help

Examples:
  skillshare update my-skill              # Update regular skill from source
  skillshare update _team-skills          # Update tracked repo (git pull)
  skillshare update team-skills           # _ prefix is optional for repos
  skillshare update --all                 # Update all tracked repos
  skillshare update --all --dry-run       # Preview updates
  skillshare update _team --force         # Discard changes and update`)
}
