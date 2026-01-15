package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"skillshare/internal/config"
	"skillshare/internal/ui"
	"skillshare/internal/utils"
)

func cmdInit(args []string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("cannot determine home directory: %w", err)
	}
	sourcePath := "" // Will be determined
	dryRun := false

	// Parse args
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--source", "-s":
			if i+1 >= len(args) {
				return fmt.Errorf("--source requires a path argument")
			}
			sourcePath = args[i+1]
			i++
		case "--dry-run", "-n":
			dryRun = true
		}
	}

	// Expand ~ in path
	if utils.HasTildePrefix(sourcePath) {
		sourcePath = filepath.Join(home, sourcePath[1:])
	}

	// Check if already initialized
	if _, err := os.Stat(config.ConfigPath()); err == nil {
		return fmt.Errorf("already initialized. Config at: %s", config.ConfigPath())
	}

	// Detect existing CLI skills directories
	detected := detectCLIDirectories(home)

	// Default source path (same location as config)
	if sourcePath == "" {
		sourcePath = filepath.Join(home, ".config", "skillshare", "skills")
	}

	// Find directories with skills to potentially copy from
	var withSkills []detectedDir
	for _, d := range detected {
		if d.hasSkills {
			withSkills = append(withSkills, d)
		}
	}

	// Ask user if they want to initialize from existing skills
	copyFrom, copyFromName := promptCopyFrom(withSkills)

	if dryRun {
		ui.Warning("Dry run mode - no changes will be made")
	}

	// Create source directory if needed
	if dryRun {
		if _, err := os.Stat(sourcePath); err == nil {
			ui.Info("Source directory exists: %s", sourcePath)
		} else {
			ui.Info("Would create source directory: %s", sourcePath)
		}
	} else if err := os.MkdirAll(sourcePath, 0755); err != nil {
		return fmt.Errorf("failed to create source directory: %w", err)
	}

	// Copy skills from selected directory
	if copyFrom != "" {
		copySkillsToSource(copyFrom, sourcePath, dryRun)
	}

	// Build targets list
	targets := buildTargetsList(detected, copyFrom, copyFromName)

	// Create config
	cfg := &config.Config{
		Source:  sourcePath,
		Mode:    "merge",
		Targets: targets,
		Ignore: []string{
			"**/.DS_Store",
			"**/.git/**",
		},
	}

	if dryRun {
		summarizeInitConfig(cfg)
	} else if err := cfg.Save(); err != nil {
		return err
	}

	// Initialize git in source directory for safety
	initGitIfNeeded(sourcePath, dryRun)

	// Create default skillshare skill
	createDefaultSkill(sourcePath, dryRun)

	if dryRun {
		ui.Header("Dry run complete")
		ui.Info("Would write config: %s", config.ConfigPath())
		ui.Info("Run 'skillshare init' to apply these changes")
		return nil
	}

	ui.Header("Initialized successfully")
	ui.Success("Source: %s", sourcePath)
	ui.Success("Config: %s", config.ConfigPath())
	ui.Info("Run 'skillshare sync' to sync your skills")

	return nil
}

type detectedDir struct {
	name       string
	path       string
	skillCount int
	hasSkills  bool
	exists     bool // true if skills dir exists, false if only parent exists
}

func detectCLIDirectories(home string) []detectedDir {
	ui.Header("Detecting CLI skills directories")
	defaultTargets := config.DefaultTargets()
	var detected []detectedDir

	for name, target := range defaultTargets {
		if info, err := os.Stat(target.Path); err == nil && info.IsDir() {
			// Skills directory exists - count skills
			entries, _ := os.ReadDir(target.Path)
			skillCount := 0
			for _, e := range entries {
				if e.IsDir() && !utils.IsHidden(e.Name()) {
					skillCount++
				}
			}
			detected = append(detected, detectedDir{
				name:       name,
				path:       target.Path,
				skillCount: skillCount,
				hasSkills:  skillCount > 0,
				exists:     true,
			})
			if skillCount > 0 {
				ui.Success("Found: %s (%d skills)", name, skillCount)
			} else {
				ui.Info("Found: %s (empty)", name)
			}
		} else {
			// Skills directory doesn't exist - check if parent exists (CLI installed)
			parent := filepath.Dir(target.Path)
			if _, err := os.Stat(parent); err == nil {
				detected = append(detected, detectedDir{
					name:       name,
					path:       target.Path,
					skillCount: 0,
					hasSkills:  false,
					exists:     false,
				})
				ui.Info("Found: %s (not initialized)", name)
			}
		}
	}

	return detected
}

func promptCopyFrom(withSkills []detectedDir) (copyFrom, copyFromName string) {
	if len(withSkills) == 0 {
		return "", ""
	}

	ui.Header("Initialize from existing skills?")
	fmt.Println("  Copy skills from an existing directory to the shared source?")
	fmt.Println()

	for i, d := range withSkills {
		fmt.Printf("  [%d] Copy from %s (%d skills)\n", i+1, d.name, d.skillCount)
	}
	fmt.Printf("  [%d] Start fresh (empty source)\n", len(withSkills)+1)
	fmt.Println()

	fmt.Print("  Enter choice [1]: ")
	var input string
	fmt.Scanln(&input)

	choice := 1
	if input != "" {
		fmt.Sscanf(input, "%d", &choice)
	}

	if choice >= 1 && choice <= len(withSkills) {
		copyFrom = withSkills[choice-1].path
		copyFromName = withSkills[choice-1].name
		ui.Success("Will copy skills from %s", copyFromName)
	} else {
		ui.Info("Starting with empty source")
	}

	return copyFrom, copyFromName
}

func copySkillsToSource(copyFrom, sourcePath string, dryRun bool) {
	entries, err := os.ReadDir(copyFrom)
	if err != nil {
		ui.Warning("Failed to read %s: %v", copyFrom, err)
		return
	}

	if dryRun {
		copyCount := 0
		for _, entry := range entries {
			if entry.IsDir() && !utils.IsHidden(entry.Name()) {
				copyCount++
			}
		}
		ui.Info("Would copy %d skills to %s", copyCount, sourcePath)
		return
	}

	ui.Info("Copying skills to %s...", sourcePath)
	copied := 0
	for _, entry := range entries {
		if !entry.IsDir() || utils.IsHidden(entry.Name()) {
			continue
		}
		srcPath := filepath.Join(copyFrom, entry.Name())
		dstPath := filepath.Join(sourcePath, entry.Name())

		// Skip if already exists
		if _, err := os.Stat(dstPath); err == nil {
			continue
		}

		// Copy directory
		if err := copyDir(srcPath, dstPath); err != nil {
			ui.Warning("Failed to copy %s: %v", entry.Name(), err)
			continue
		}
		copied++
	}
	ui.Success("Copied %d skills to source", copied)
}

func buildTargetsList(detected []detectedDir, copyFrom, copyFromName string) map[string]config.TargetConfig {
	defaultTargets := config.DefaultTargets()
	targets := make(map[string]config.TargetConfig)

	// Add the directory user chose to copy from
	if copyFromName != "" {
		targets[copyFromName] = config.TargetConfig{Path: copyFrom}
	}

	// Find other available targets (detected directories)
	var otherTargets []string
	for _, d := range detected {
		if d.name == copyFromName {
			continue // Already added
		}
		otherTargets = append(otherTargets, d.name)
	}

	// Ask if user wants to add other targets
	if len(otherTargets) > 0 {
		ui.Header("Add other CLI targets?")
		fmt.Println("  Other CLI tools detected on your system:")
		for _, name := range otherTargets {
			fmt.Printf("    - %s\n", name)
		}
		fmt.Println()
		fmt.Print("  Add these targets? [Y/n]: ")
		var input string
		fmt.Scanln(&input)
		input = strings.ToLower(strings.TrimSpace(input))

		if input == "" || input == "y" || input == "yes" {
			for _, name := range otherTargets {
				targets[name] = defaultTargets[name]
			}
			ui.Success("Added %d additional targets", len(otherTargets))
		} else {
			ui.Info("Skipped additional targets")
		}
	}

	if len(targets) == 0 {
		ui.Warning("No CLI skills directories detected.")
	}

	return targets
}

func summarizeInitConfig(cfg *config.Config) {
	ui.Header("Planned configuration")
	ui.Info("Source: %s", cfg.Source)

	if len(cfg.Targets) == 0 {
		ui.Info("Targets: none")
		return
	}

	ui.Info("Targets: %d", len(cfg.Targets))
	for name, target := range cfg.Targets {
		mode := target.Mode
		if mode == "" {
			mode = cfg.Mode
		}
		if mode == "" {
			mode = "merge"
		}
		fmt.Printf("  %-12s %s (%s)\n", name, target.Path, mode)
	}
}

func initGitIfNeeded(sourcePath string, dryRun bool) {
	gitDir := filepath.Join(sourcePath, ".git")
	if _, err := os.Stat(gitDir); err == nil {
		ui.Info("Git already initialized in source directory")
		return
	}

	ui.Header("Git version control")
	fmt.Println("  Git helps protect your skills from accidental deletion.")
	fmt.Println()
	fmt.Print("  Initialize git in source directory? [Y/n]: ")
	var input string
	fmt.Scanln(&input)
	input = strings.ToLower(strings.TrimSpace(input))

	if input != "" && input != "y" && input != "yes" {
		if dryRun {
			ui.Info("Dry run - skipped git initialization")
			return
		}
		ui.Info("Skipped git initialization")
		ui.Warning("Without git, deleted skills cannot be recovered!")
		return
	}

	if dryRun {
		ui.Info("Dry run - would initialize git in %s", sourcePath)
		return
	}

	// Run git init
	cmd := exec.Command("git", "init")
	cmd.Dir = sourcePath
	if err := cmd.Run(); err != nil {
		ui.Warning("Failed to initialize git: %v", err)
		return
	}

	// Create .gitignore
	gitignore := filepath.Join(sourcePath, ".gitignore")
	if _, err := os.Stat(gitignore); os.IsNotExist(err) {
		os.WriteFile(gitignore, []byte(".DS_Store\n"), 0644)
	}

	// Initial commit if there are files
	entries, _ := os.ReadDir(sourcePath)
	hasFiles := false
	for _, e := range entries {
		if e.Name() != ".git" && e.Name() != ".gitignore" {
			hasFiles = true
			break
		}
	}

	if hasFiles {
		addCmd := exec.Command("git", "add", ".")
		addCmd.Dir = sourcePath
		addCmd.Run()

		commitCmd := exec.Command("git", "commit", "-m", "Initial skills")
		commitCmd.Dir = sourcePath
		commitCmd.Run()
		ui.Success("Git initialized with initial commit")
	} else {
		ui.Success("Git initialized (empty repository)")
	}
	ui.Info("Push to a remote repo for backup: git remote add origin <url>")
}

func createDefaultSkill(sourcePath string, dryRun bool) {
	skillshareSkillDir := filepath.Join(sourcePath, "skillshare")
	skillshareSkillFile := filepath.Join(skillshareSkillDir, "SKILL.md")

	if _, err := os.Stat(skillshareSkillFile); err == nil {
		return
	}

	if dryRun {
		ui.Info("Would create default skill: skillshare")
		return
	}

	if err := os.MkdirAll(skillshareSkillDir, 0755); err != nil {
		return
	}

	skillContent := `---
name: skillshare
description: Manage and sync skills across AI CLI tools
---

# Skillshare CLI

Use skillshare to manage skills shared across multiple AI CLI tools.

## Commands

### Check Status
` + "```bash" + `
skillshare status
` + "```" + `
Shows source directory, skill count, and sync state for all targets.

### Sync Skills
` + "```bash" + `
skillshare sync           # Sync all targets
skillshare sync --dry-run # Preview changes
` + "```" + `
Pushes skills from source to all configured targets.

### Pull Local Skills
` + "```bash" + `
skillshare pull claude    # Pull from specific target
skillshare pull --all     # Pull from all targets
` + "```" + `
Copies skills created in target directories back to source.

### View Differences
` + "```bash" + `
skillshare diff           # All targets
skillshare diff claude    # Specific target
` + "```" + `

### Manage Targets
` + "```bash" + `
skillshare target list              # List all targets
skillshare target add myapp ~/path  # Add custom target
skillshare target remove myapp      # Remove target
` + "```" + `

### Backup & Restore
` + "```bash" + `
skillshare backup --list    # List backups
skillshare backup --cleanup # Clean old backups
skillshare restore claude   # Restore from backup
` + "```" + `

## Typical Workflow

1. Create/edit skills in any target directory (e.g., ~/.claude/skills/)
2. Run ` + "`skillshare pull`" + ` to bring changes to source
3. Run ` + "`skillshare sync`" + ` to distribute to all targets
4. Commit changes: ` + "`cd ~/.config/skillshare/skills && git add . && git commit`" + `

## Tips

- Source directory: ~/.config/skillshare/skills
- Config file: ~/.config/skillshare/config.yaml
- Use ` + "`skillshare doctor`" + ` to diagnose issues
`
	if err := os.WriteFile(skillshareSkillFile, []byte(skillContent), 0644); err == nil {
		ui.Success("Created default skill: skillshare")
	}
}
