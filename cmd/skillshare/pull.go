package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"skillshare/internal/config"
	"skillshare/internal/sync"
	"skillshare/internal/ui"
)

func cmdPull(args []string) error {
	dryRun := false
	force := false
	pullAll := false
	pullRemote := false
	var targetName string

	for _, arg := range args {
		switch arg {
		case "--dry-run", "-n":
			dryRun = true
		case "--force", "-f":
			force = true
		case "--all", "-a":
			pullAll = true
		case "--remote", "-r":
			pullRemote = true
		default:
			if targetName == "" && !strings.HasPrefix(arg, "-") {
				targetName = arg
			}
		}
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Handle --remote: pull from git remote and sync
	if pullRemote {
		return pullFromRemote(cfg, dryRun)
	}

	// Select targets to pull from
	targets, err := selectPullTargets(cfg, targetName, pullAll)
	if err != nil {
		return err
	}
	if targets == nil {
		return nil // User needs to specify target
	}

	// Collect all local skills
	allLocalSkills := collectLocalSkills(targets, cfg.Source)

	if len(allLocalSkills) == 0 {
		ui.Info("No local skills to pull")
		return nil
	}

	// Display found skills
	displayLocalSkills(allLocalSkills)

	if dryRun {
		ui.Info("Dry run - no changes made")
		return nil
	}

	// Confirm unless --force
	if !force {
		if !confirmPull() {
			ui.Info("Cancelled")
			return nil
		}
	}

	// Execute pull
	return executePull(allLocalSkills, cfg.Source, dryRun, force)
}

func selectPullTargets(cfg *config.Config, targetName string, pullAll bool) (map[string]config.TargetConfig, error) {
	if targetName != "" {
		if t, exists := cfg.Targets[targetName]; exists {
			return map[string]config.TargetConfig{targetName: t}, nil
		}
		return nil, fmt.Errorf("target '%s' not found", targetName)
	}

	if pullAll || len(cfg.Targets) == 1 {
		return cfg.Targets, nil
	}

	// If no target specified and multiple targets exist, ask or require --all
	ui.Warning("Multiple targets found. Specify a target name or use --all")
	fmt.Println("  Available targets:")
	for name := range cfg.Targets {
		fmt.Printf("    - %s\n", name)
	}
	return nil, nil
}

func collectLocalSkills(targets map[string]config.TargetConfig, source string) []sync.LocalSkillInfo {
	var allLocalSkills []sync.LocalSkillInfo
	for name, target := range targets {
		skills, err := sync.FindLocalSkills(target.Path, source)
		if err != nil {
			ui.Warning("%s: %v", name, err)
			continue
		}
		for i := range skills {
			skills[i].TargetName = name
		}
		allLocalSkills = append(allLocalSkills, skills...)
	}
	return allLocalSkills
}

func displayLocalSkills(skills []sync.LocalSkillInfo) {
	ui.Header("Local skills found")
	for _, skill := range skills {
		ui.ListItem("info", skill.Name, fmt.Sprintf("[%s] %s", skill.TargetName, skill.Path))
	}
}

func confirmPull() bool {
	fmt.Println()
	fmt.Print("Pull these skills to source? [y/N]: ")
	var input string
	fmt.Scanln(&input)
	input = strings.ToLower(strings.TrimSpace(input))
	return input == "y" || input == "yes"
}

func executePull(skills []sync.LocalSkillInfo, source string, dryRun, force bool) error {
	ui.Header("Pulling skills")
	result, err := sync.PullSkills(skills, source, sync.PullOptions{
		DryRun: dryRun,
		Force:  force,
	})
	if err != nil {
		return err
	}

	// Display results
	for _, name := range result.Pulled {
		ui.Success("%s: copied to source", name)
	}
	for _, name := range result.Skipped {
		ui.Warning("%s: skipped (already exists in source, use --force to overwrite)", name)
	}
	for name, err := range result.Failed {
		ui.Error("%s: %v", name, err)
	}

	if len(result.Pulled) > 0 {
		showPullNextSteps(source)
	}

	return nil
}

func showPullNextSteps(source string) {
	fmt.Println()
	ui.Info("Run 'skillshare sync' to distribute to all targets")

	// Check if source has git
	gitDir := filepath.Join(source, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		ui.Info("Commit changes: cd %s && git add . && git commit", source)
	}
}

// pullFromRemote pulls from git remote and syncs to all targets
func pullFromRemote(cfg *config.Config, dryRun bool) error {
	ui.Header("Pulling from remote")

	// Check if source is a git repo
	gitDir := filepath.Join(cfg.Source, ".git")
	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		ui.Error("Source is not a git repository")
		ui.Info("  Run: cd %s && git init", cfg.Source)
		return nil
	}

	// Check if remote exists
	cmd := exec.Command("git", "remote")
	cmd.Dir = cfg.Source
	output, err := cmd.Output()
	if err != nil || strings.TrimSpace(string(output)) == "" {
		ui.Error("No git remote configured")
		ui.Info("  Run: cd %s && git remote add origin <url>", cfg.Source)
		return nil
	}

	// Check for uncommitted changes
	cmd = exec.Command("git", "status", "--porcelain")
	cmd.Dir = cfg.Source
	output, err = cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to check git status: %w", err)
	}

	if len(strings.TrimSpace(string(output))) > 0 {
		ui.Error("Local changes detected - commit or stash first")
		ui.Info("  Run: skillshare push")
		ui.Info("  Or:  cd %s && git stash", cfg.Source)
		return nil
	}

	if dryRun {
		ui.Warning("[dry-run] No changes will be made")
		fmt.Println()
		ui.Info("Would run: git pull")
		ui.Info("Would run: skillshare sync")
		return nil
	}

	// Git pull
	ui.Info("Running git pull...")
	cmd = exec.Command("git", "pull")
	cmd.Dir = cfg.Source
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git pull failed: %w", err)
	}

	// Sync to all targets
	fmt.Println()
	ui.Info("Syncing to all targets...")
	return cmdSync([]string{})
}
