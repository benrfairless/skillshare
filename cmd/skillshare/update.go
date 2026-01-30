package main

import (
	"fmt"
	"os"
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

	skills, err := getUpdatableSkills(cfg.Source)
	if err != nil {
		return fmt.Errorf("failed to get updatable skills: %w", err)
	}

	if len(repos) == 0 && len(skills) == 0 {
		ui.Info("No tracked repositories or updatable skills found")
		ui.Info("Use 'skillshare install <repo> --track' to add a tracked repository")
		return nil
	}

	// Header
	total := len(repos) + len(skills)
	ui.HeaderBox("skillshare update --all",
		fmt.Sprintf("Updating %d tracked repos + %d skills", len(repos), len(skills)))
	fmt.Println()

	var updated, skipped int

	// Update tracked repos
	for i, repo := range repos {
		repoPath := filepath.Join(cfg.Source, repo)
		progress := fmt.Sprintf("[%d/%d]", i+1, total)

		// Check for uncommitted changes
		if isDirty, _ := git.IsDirty(repoPath); isDirty {
			if !force {
				ui.ListItem("warning", repo, "has uncommitted changes (use --force)")
				skipped++
				continue
			}
			if !dryRun {
				if err := git.Restore(repoPath); err != nil {
					ui.ListItem("warning", repo, fmt.Sprintf("failed to discard changes: %v", err))
					skipped++
					continue
				}
			}
		}

		if dryRun {
			ui.ListItem("info", repo, "[dry-run] would git pull")
			continue
		}

		spinner := ui.StartSpinner(fmt.Sprintf("%s Updating %s...", progress, repo))

		// Pull (use ForcePull if --force to handle force push)
		var info *git.UpdateInfo
		if force {
			info, err = git.ForcePull(repoPath)
		} else {
			info, err = git.Pull(repoPath)
		}
		if err != nil {
			spinner.Warn(fmt.Sprintf("%s %v", repo, err))
			skipped++
			continue
		}

		if info.UpToDate {
			spinner.Success(fmt.Sprintf("%s Already up to date", repo))
		} else {
			detail := fmt.Sprintf("%s %d commits, %d files", repo, len(info.Commits), info.Stats.FilesChanged)
			spinner.Success(detail)
		}
		updated++
	}

	// Update regular skills with metadata
	repoCount := len(repos)
	for i, skill := range skills {
		skillPath := filepath.Join(cfg.Source, skill)
		progress := fmt.Sprintf("[%d/%d]", repoCount+i+1, total)

		if dryRun {
			ui.ListItem("info", skill, "[dry-run] would reinstall from source")
			continue
		}

		spinner := ui.StartSpinner(fmt.Sprintf("%s Updating %s...", progress, skill))

		meta, _ := install.ReadMeta(skillPath)
		source, err := install.ParseSource(meta.Source)
		if err != nil {
			spinner.Warn(fmt.Sprintf("%s invalid source: %v", skill, err))
			skipped++
			continue
		}

		opts := install.InstallOptions{
			Force:  true,
			Update: true,
		}

		_, err = install.Install(source, skillPath, opts)
		if err != nil {
			spinner.Warn(fmt.Sprintf("%s %v", skill, err))
			skipped++
			continue
		}

		spinner.Success(fmt.Sprintf("%s Reinstalled from source", skill))
		updated++
	}

	// Summary
	if !dryRun {
		fmt.Println()
		ui.Box("Summary",
			"",
			fmt.Sprintf("  Total:    %d", total),
			fmt.Sprintf("  Updated:  %d", updated),
			fmt.Sprintf("  Skipped:  %d", skipped),
			"",
		)
	}

	if updated > 0 {
		fmt.Println()
		ui.Info("Run 'skillshare sync' to distribute changes")
	}

	return nil
}

// getUpdatableSkills returns skill names that have metadata with a remote source
func getUpdatableSkills(sourceDir string) ([]string, error) {
	entries, err := os.ReadDir(sourceDir)
	if err != nil {
		return nil, err
	}

	var skills []string
	for _, entry := range entries {
		// Skip tracked repos (start with _) and non-directories
		if !entry.IsDir() || (len(entry.Name()) > 0 && entry.Name()[0] == '_') {
			continue
		}

		skillPath := filepath.Join(sourceDir, entry.Name())
		meta, err := install.ReadMeta(skillPath)
		if err != nil || meta == nil || meta.Source == "" {
			continue // No metadata or no source, skip
		}

		skills = append(skills, entry.Name())
	}
	return skills, nil
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
  --all, -a           Update all tracked repos + skills with metadata
  --force, -f         Discard local changes and force update
  --dry-run, -n       Preview without making changes
  --help, -h          Show this help

Examples:
  skillshare update my-skill              # Update regular skill from source
  skillshare update _team-skills          # Update tracked repo (git pull)
  skillshare update team-skills           # _ prefix is optional for repos
  skillshare update --all                 # Update all tracked repos + skills
  skillshare update --all --dry-run       # Preview updates
  skillshare update _team --force         # Discard changes and update`)
}
