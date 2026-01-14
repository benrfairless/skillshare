package main

import (
	"fmt"
	"os"
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
	var targetName string

	for _, arg := range args {
		switch arg {
		case "--dry-run", "-n":
			dryRun = true
		case "--force", "-f":
			force = true
		case "--all", "-a":
			pullAll = true
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
		fmt.Printf("  %-20s [%s] %s\n", skill.Name, skill.TargetName, skill.Path)
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
