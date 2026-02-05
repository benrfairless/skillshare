package main

import (
	"fmt"
	"strings"

	"skillshare/internal/config"
	"skillshare/internal/ui"
)

func cmdCollectProject(args []string, root string) error {
	dryRun := false
	force := false
	collectAll := false
	var targetName string

	for _, arg := range args {
		switch arg {
		case "--dry-run", "-n":
			dryRun = true
		case "--force", "-f":
			force = true
		case "--all", "-a":
			collectAll = true
		default:
			if targetName == "" && !strings.HasPrefix(arg, "-") {
				targetName = arg
			}
		}
	}

	runtime, err := loadProjectRuntime(root)
	if err != nil {
		return err
	}

	targets, err := selectCollectProjectTargets(runtime, targetName, collectAll)
	if err != nil {
		return err
	}
	if targets == nil {
		return nil
	}

	allLocalSkills := collectLocalSkills(targets, runtime.sourcePath)

	if len(allLocalSkills) == 0 {
		ui.Info("No local skills to collect")
		return nil
	}

	displayLocalSkills(allLocalSkills)

	if dryRun {
		ui.Info("Dry run - no changes made")
		return nil
	}

	if !force {
		if !confirmCollect() {
			ui.Info("Cancelled")
			return nil
		}
	}

	return executeCollect(allLocalSkills, runtime.sourcePath, dryRun, force)
}

func selectCollectProjectTargets(runtime *projectRuntime, targetName string, collectAll bool) (map[string]config.TargetConfig, error) {
	if targetName != "" {
		if t, ok := runtime.targets[targetName]; ok {
			return map[string]config.TargetConfig{targetName: t}, nil
		}
		return nil, fmt.Errorf("target '%s' not found in project config", targetName)
	}

	if collectAll || len(runtime.targets) == 1 {
		return runtime.targets, nil
	}

	ui.Warning("Multiple targets found. Specify a target name or use --all")
	fmt.Println("  Available targets:")
	for _, entry := range runtime.config.Targets {
		fmt.Printf("    - %s\n", entry.Name)
	}
	return nil, nil
}
