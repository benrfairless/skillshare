package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"skillshare/internal/config"
	"skillshare/internal/install"
	"skillshare/internal/ui"
	"skillshare/internal/validate"
)

func cmdInstall(args []string) error {
	opts := install.InstallOptions{}
	var sourceArg string

	// Parse arguments
	i := 0
	for i < len(args) {
		arg := args[i]
		switch {
		case arg == "--name":
			if i+1 >= len(args) {
				return fmt.Errorf("--name requires a value")
			}
			i++
			opts.Name = args[i]
		case arg == "--force" || arg == "-f":
			opts.Force = true
		case arg == "--update" || arg == "-u":
			opts.Update = true
		case arg == "--dry-run" || arg == "-n":
			opts.DryRun = true
		case arg == "--help" || arg == "-h":
			printInstallHelp()
			return nil
		case strings.HasPrefix(arg, "-"):
			return fmt.Errorf("unknown option: %s", arg)
		default:
			if sourceArg != "" {
				return fmt.Errorf("unexpected argument: %s", arg)
			}
			sourceArg = arg
		}
		i++
	}

	if sourceArg == "" {
		printInstallHelp()
		return fmt.Errorf("source is required")
	}

	// Load config to get source directory
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Parse the source
	source, err := install.ParseSource(sourceArg)
	if err != nil {
		return fmt.Errorf("invalid source: %w", err)
	}

	// If git source without subdir, discover skills and let user choose
	if source.IsGit() && !source.HasSubdir() {
		return handleGitDiscovery(source, cfg, opts)
	}

	// Direct installation (local path or git with subdir)
	return handleDirectInstall(source, cfg, opts)
}

func handleGitDiscovery(source *install.Source, cfg *config.Config, opts install.InstallOptions) error {
	ui.Header("Discovering skills")
	fmt.Println(strings.Repeat("-", 45))
	ui.Info("Source: %s", source.Raw)
	ui.Info("Cloning repository...")
	fmt.Println()

	// Discover skills
	discovery, err := install.DiscoverFromGit(source)
	if err != nil {
		return err
	}
	defer install.CleanupDiscovery(discovery)

	if len(discovery.Skills) == 0 {
		ui.Warning("No skills found in repository (no SKILL.md files)")
		return nil
	}

	// Display discovered skills
	ui.Success("Found %d skill(s):", len(discovery.Skills))
	fmt.Println()
	for i, skill := range discovery.Skills {
		fmt.Printf("  [%d] %s\n", i+1, skill.Name)
		fmt.Printf("      %s\n", skill.Path)
	}
	fmt.Println()

	if opts.DryRun {
		ui.Warning("[dry-run] Would prompt for selection")
		return nil
	}

	// Prompt for selection
	selected, err := promptSkillSelection(discovery.Skills)
	if err != nil {
		return err
	}

	if len(selected) == 0 {
		ui.Info("No skills selected")
		return nil
	}

	// Install selected skills
	fmt.Println()
	ui.Header("Installing selected skills")
	fmt.Println(strings.Repeat("-", 45))

	var installed int
	for _, skill := range selected {
		destPath := filepath.Join(cfg.Source, skill.Name)

		result, err := install.InstallFromDiscovery(discovery, skill, destPath, opts)
		if err != nil {
			ui.Error("%s: %v", skill.Name, err)
			continue
		}

		ui.Success("Installed: %s", result.SkillPath)
		for _, warning := range result.Warnings {
			ui.Warning("  %s", warning)
		}
		installed++
	}

	fmt.Println()
	if installed > 0 {
		ui.Info("Run 'skillshare sync' to distribute to all targets")
	}

	return nil
}

func promptSkillSelection(skills []install.SkillInfo) ([]install.SkillInfo, error) {
	fmt.Print("Enter numbers to install (e.g., 1,2,3 or 'all' or 'q' to quit): ")

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return nil, err
	}

	input = strings.TrimSpace(strings.ToLower(input))

	if input == "q" || input == "quit" || input == "" {
		return nil, nil
	}

	if input == "all" || input == "a" {
		return skills, nil
	}

	// Parse comma-separated numbers
	var selected []install.SkillInfo
	parts := strings.Split(input, ",")
	seen := make(map[int]bool)

	for _, part := range parts {
		part = strings.TrimSpace(part)
		num, err := strconv.Atoi(part)
		if err != nil {
			ui.Warning("Invalid number: %s", part)
			continue
		}

		if num < 1 || num > len(skills) {
			ui.Warning("Out of range: %d", num)
			continue
		}

		if !seen[num] {
			seen[num] = true
			selected = append(selected, skills[num-1])
		}
	}

	return selected, nil
}

func handleDirectInstall(source *install.Source, cfg *config.Config, opts install.InstallOptions) error {
	// Determine skill name
	skillName := source.Name
	if opts.Name != "" {
		skillName = opts.Name
	}

	// Validate skill name
	if err := validate.SkillName(skillName); err != nil {
		return fmt.Errorf("invalid skill name '%s': %w", skillName, err)
	}

	// Set the name in source for display
	source.Name = skillName

	// Determine destination path
	destPath := filepath.Join(cfg.Source, skillName)

	// Display installation info
	ui.Header("Installing skill")
	fmt.Println(strings.Repeat("-", 45))
	ui.Info("Source: %s", source.Raw)
	ui.Info("Name: %s", skillName)
	if source.HasSubdir() {
		ui.Info("Subdirectory: %s", source.Subdir)
	}
	ui.Info("Destination: %s", destPath)
	fmt.Println()

	// Execute installation
	result, err := install.Install(source, destPath, opts)
	if err != nil {
		return err
	}

	// Display result
	if opts.DryRun {
		ui.Warning("[dry-run] %s", result.Action)
	} else {
		ui.Success("Installed: %s", result.SkillPath)
	}

	// Display warnings
	for _, warning := range result.Warnings {
		ui.Warning("%s", warning)
	}

	// Show next steps
	if !opts.DryRun {
		fmt.Println()
		ui.Info("Run 'skillshare sync' to distribute to all targets")
	}

	return nil
}

func printInstallHelp() {
	fmt.Println(`Usage: skillshare install <source> [options]

Install a skill from a local path or git repository.

Sources:
  ~/path/to/skill            Local directory
  github.com/user/repo       GitHub repository (discovers skills)
  github.com/user/repo/path  Subdirectory in GitHub repo (direct install)
  https://github.com/...     HTTPS git URL
  git@github.com:...         SSH git URL

Options:
  --name <name>       Override the skill name (only for direct install)
  --force, -f         Overwrite if skill already exists
  --update, -u        Update existing installation (git pull for git sources)
  --dry-run, -n       Preview the installation without making changes
  --help, -h          Show this help

Examples:
  skillshare install ~/my-skill
  skillshare install github.com/ComposioHQ/awesome-claude-skills
  skillshare install github.com/google-gemini/gemini-cli/packages/core/src/skills/builtin/skill-creator
  skillshare install github.com/user/repo/skill --name my-custom-name
  skillshare install github.com/user/repo --force`)
}
