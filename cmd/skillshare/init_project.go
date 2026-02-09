package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/AlecAivazis/survey/v2"
	"skillshare/internal/config"
	"skillshare/internal/install"
	"skillshare/internal/ui"
)

type projectInitOptions struct {
	dryRun    bool
	targets   []string // Non-interactive target list
	discover  bool
	selectArg string // Non-interactive selection for --discover
}

type detectedProjectTarget struct {
	name         string
	path         string
	exists       bool
	parentExists bool
}

func parseProjectInitArgs(args []string) (projectInitOptions, bool, error) {
	opts := projectInitOptions{}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--dry-run" || arg == "-n":
			opts.dryRun = true
		case arg == "--help" || arg == "-h":
			return opts, true, nil
		case arg == "--discover" || arg == "-d":
			opts.discover = true
		case arg == "--select":
			if i+1 >= len(args) {
				return opts, false, fmt.Errorf("--select requires a value")
			}
			i++
			opts.selectArg = args[i]
		case arg == "--targets":
			if i+1 >= len(args) {
				return opts, false, fmt.Errorf("--targets requires a value")
			}
			i++
			opts.targets = strings.Split(args[i], ",")
		case strings.HasPrefix(arg, "-"):
			return opts, false, fmt.Errorf("unknown option: %s", arg)
		}
	}

	if opts.selectArg != "" && !opts.discover {
		return opts, false, fmt.Errorf("--select requires --discover flag")
	}

	return opts, false, nil
}

func cmdInitProject(args []string) error {
	opts, showHelp, err := parseProjectInitArgs(args)
	if showHelp {
		fmt.Println("Usage: skillshare init -p [--dry-run]")
		return nil
	}
	if err != nil {
		return err
	}

	root, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("cannot determine project directory: %w", err)
	}

	return performProjectInit(root, opts)
}

func performProjectInit(root string, opts projectInitOptions) error {
	projectDir := filepath.Join(root, ".skillshare")
	configPath := config.ProjectConfigPath(root)
	if _, err := os.Stat(projectDir); err == nil {
		if opts.discover {
			return reinitProjectWithDiscover(root, opts)
		}
		return fmt.Errorf("project already initialized: %s\nRun 'skillshare init -p --discover' to add new targets", projectDir)
	}

	ui.Logo(version)
	ui.Header("Initializing project-level skills")

	var selected []config.ProjectTargetEntry

	// If --targets provided, skip interactive prompt
	if len(opts.targets) > 0 {
		selected = make([]config.ProjectTargetEntry, 0, len(opts.targets))
		for _, name := range opts.targets {
			name = strings.TrimSpace(name)
			if name != "" {
				selected = append(selected, config.ProjectTargetEntry{Name: name})
			}
		}
	} else {
		detected := detectProjectCLIDirectories(root)
		available := detected
		if len(available) == 0 {
			ui.Warning("No AI CLI directories detected.")
			available = listAllProjectTargets()
		}

		var err error
		selected, err = promptProjectTargets(available)
		if err != nil {
			return err
		}
	}

	if opts.dryRun {
		ui.Header("Dry run complete (project)")
		ui.Info("Would create .skillshare/skills/")
		ui.Info("Would write config: %s", configPath)
		return nil
	}

	if err := os.MkdirAll(filepath.Join(root, ".skillshare", "skills"), 0755); err != nil {
		return fmt.Errorf("failed to create .skillshare/skills: %w", err)
	}

	if err := ensureProjectGitignore(root); err != nil {
		return err
	}

	cfg := &config.ProjectConfig{
		Targets: selected,
	}
	if err := cfg.Save(root); err != nil {
		return err
	}

	if err := createProjectTargetDirs(root, selected); err != nil {
		return err
	}

	ui.Success("Created .skillshare/config.yaml")
	ui.Success("Created .skillshare/skills/")
	ui.Success("Added %d target(s)", len(selected))
	fmt.Println()

	ui.Header("Initialized successfully (project)")
	ui.Success("Source: .skillshare/skills/")
	ui.Success("Config: %s", config.ProjectConfigPath(root))
	fmt.Println()
	ui.Info("Next steps:")
	fmt.Println("  skillshare install <skill> -p    # Install a skill")
	fmt.Println("  skillshare sync                  # Sync to all targets")

	return nil
}

func detectProjectCLIDirectories(root string) []detectedProjectTarget {
	ui.Header("Detecting AI CLI directories")

	projectTargets := config.ProjectTargets()
	names := make([]string, 0, len(projectTargets))
	for name := range projectTargets {
		names = append(names, name)
	}
	sort.Strings(names)

	var detected []detectedProjectTarget
	for _, name := range names {
		target := projectTargets[name]
		relPath := filepath.FromSlash(target.Path)
		fullPath := filepath.Join(root, relPath)

		if info, err := os.Stat(fullPath); err == nil && info.IsDir() {
			ui.Success("Found: %s (%s)", name, relPath)
			detected = append(detected, detectedProjectTarget{name: name, path: relPath, exists: true})
			continue
		}

		parentRel := filepath.Dir(relPath)
		if parentRel != "." {
			parentPath := filepath.Join(root, parentRel)
			if info, err := os.Stat(parentPath); err == nil && info.IsDir() {
				ui.Info("Found: %s (not initialized)", name)
				detected = append(detected, detectedProjectTarget{name: name, path: relPath, parentExists: true})
			}
		}
	}

	return detected
}

func listAllProjectTargets() []detectedProjectTarget {
	projectTargets := config.ProjectTargets()
	names := make([]string, 0, len(projectTargets))
	for name := range projectTargets {
		names = append(names, name)
	}
	sort.Strings(names)

	var available []detectedProjectTarget
	for _, name := range names {
		target := projectTargets[name]
		available = append(available, detectedProjectTarget{name: name, path: filepath.FromSlash(target.Path)})
	}
	return available
}

func promptProjectTargets(available []detectedProjectTarget) ([]config.ProjectTargetEntry, error) {
	ui.Header("Select targets to sync")

	options := make([]string, len(available))
	defaultIndices := []int{}
	for i, target := range available {
		status := ""
		if target.exists {
			status = ""
		} else if target.parentExists {
			status = "(not initialized)"
		}
		if status != "" {
			options[i] = fmt.Sprintf("%-12s %s %s", target.name, target.path, status)
		} else {
			options[i] = fmt.Sprintf("%-12s %s", target.name, target.path)
		}

		if target.exists || target.parentExists {
			defaultIndices = append(defaultIndices, i)
		}
	}

	var selectedIndices []int
	prompt := &survey.MultiSelect{
		Message:  "Select targets to sync:",
		Options:  options,
		Default:  defaultIndices,
		PageSize: 15,
		Help:     "Use arrow keys to navigate, space to select, enter to confirm",
	}

	if err := survey.AskOne(prompt, &selectedIndices); err != nil {
		return nil, nil
	}

	selected := make([]config.ProjectTargetEntry, 0, len(selectedIndices))
	for _, idx := range selectedIndices {
		name := available[idx].name
		selected = append(selected, config.ProjectTargetEntry{Name: name})
	}

	return selected, nil
}

func createProjectTargetDirs(root string, targets []config.ProjectTargetEntry) error {
	if len(targets) == 0 {
		return nil
	}

	knownTargets := config.ProjectTargets()
	created := map[string]bool{}

	for _, target := range targets {
		name := target.Name
		path := target.Path
		if path == "" {
			if known, ok := knownTargets[name]; ok {
				path = known.Path
			} else {
				continue
			}
		}

		absPath := path
		if !filepath.IsAbs(path) {
			absPath = filepath.Join(root, filepath.FromSlash(path))
		}

		if created[absPath] {
			continue
		}
		created[absPath] = true

		if err := os.MkdirAll(absPath, 0755); err != nil {
			return fmt.Errorf("failed to create target directory: %w", err)
		}
	}

	return nil
}

// reinitProjectWithDiscover detects new targets and adds them to existing project config
func reinitProjectWithDiscover(root string, opts projectInitOptions) error {
	ui.Logo(version)
	ui.Header("Discovering new targets")

	cfg, err := config.LoadProject(root)
	if err != nil {
		return err
	}

	// Build set of existing target names
	existing := make(map[string]bool)
	for _, t := range cfg.Targets {
		existing[t.Name] = true
	}

	// Detect AI CLI directories in project
	detected := detectProjectCLIDirectories(root)
	if len(detected) == 0 {
		detected = listAllProjectTargets()
	}

	// Filter out already-configured targets
	var newTargets []detectedProjectTarget
	for _, d := range detected {
		if !existing[d.name] {
			newTargets = append(newTargets, d)
		}
	}

	if len(newTargets) == 0 {
		ui.Info("No new targets detected")
		return nil
	}

	ui.Success("Found %d new target(s)", len(newTargets))

	var selected []config.ProjectTargetEntry

	// Non-interactive: --select
	if opts.selectArg != "" {
		names := strings.Split(opts.selectArg, ",")
		for _, name := range names {
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}
			found := false
			for _, d := range newTargets {
				if d.name == name {
					found = true
					break
				}
			}
			if !found {
				if existing[name] {
					ui.Info("Target already in config: %s (skipped)", name)
				} else {
					ui.Warning("Target not detected: %s (skipped)", name)
				}
				continue
			}
			selected = append(selected, config.ProjectTargetEntry{Name: name})
		}
	} else {
		// Interactive selection
		var err error
		selected, err = promptProjectTargets(newTargets)
		if err != nil {
			return err
		}
	}

	if len(selected) == 0 {
		ui.Info("No new targets added")
		return nil
	}

	if opts.dryRun {
		ui.Warning("Dry run - would add %d target(s) to config", len(selected))
		for _, t := range selected {
			fmt.Printf("  + %s\n", t.Name)
		}
		return nil
	}

	// Add to config and save
	cfg.Targets = append(cfg.Targets, selected...)
	if err := cfg.Save(root); err != nil {
		return err
	}

	// Create target directories
	if err := createProjectTargetDirs(root, selected); err != nil {
		return err
	}

	ui.Success("Added %d target(s) to config", len(selected))
	for _, t := range selected {
		fmt.Printf("  + %s\n", t.Name)
	}
	ui.Info("Run 'skillshare sync' to sync skills to new targets")

	return nil
}

func ensureProjectGitignore(root string) error {
	gitignoreDir := filepath.Join(root, ".skillshare")
	gitignorePath := filepath.Join(gitignoreDir, ".gitignore")
	if _, err := os.Stat(gitignorePath); os.IsNotExist(err) {
		if err := os.WriteFile(gitignorePath, []byte(""), 0644); err != nil {
			return fmt.Errorf("failed to create .skillshare/.gitignore: %w", err)
		}
	}

	// Always ignore project logs to avoid committing operational noise.
	if err := install.UpdateGitIgnore(gitignoreDir, "logs"); err != nil {
		return fmt.Errorf("failed to update .skillshare/.gitignore: %w", err)
	}

	return nil
}
