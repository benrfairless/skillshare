package main

import (
	"fmt"
	"path/filepath"
	"strings"

	"skillshare/internal/config"
	"skillshare/internal/install"
	"skillshare/internal/sync"
	"skillshare/internal/ui"
	"skillshare/internal/utils"
)

func cmdList(args []string) error {
	var verbose bool

	// Parse arguments
	for _, arg := range args {
		switch arg {
		case "--verbose", "-v":
			verbose = true
		case "--help", "-h":
			printListHelp()
			return nil
		default:
			if strings.HasPrefix(arg, "-") {
				return fmt.Errorf("unknown option: %s", arg)
			}
		}
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	// Discover all skills recursively
	discovered, err := sync.DiscoverSourceSkills(cfg.Source)
	if err != nil {
		return fmt.Errorf("cannot discover skills: %w", err)
	}

	// Get tracked repos
	trackedRepos, _ := install.GetTrackedRepos(cfg.Source)

	// Build skill entries with metadata
	var skills []skillEntry
	for _, d := range discovered {
		entry := skillEntry{
			Name:     d.FlatName,
			IsNested: d.IsInRepo || utils.HasNestedSeparator(d.FlatName),
		}

		// Determine repo name if in tracked repo
		if d.IsInRepo {
			parts := strings.SplitN(d.RelPath, "/", 2)
			if len(parts) > 0 {
				entry.RepoName = parts[0]
			}
		}

		// Read metadata if available
		if meta, err := install.ReadMeta(d.SourcePath); err == nil && meta != nil {
			entry.Source = meta.Source
			entry.Type = meta.Type
			entry.InstalledAt = meta.InstalledAt.Format("2006-01-02")
		}

		skills = append(skills, entry)
	}

	if len(skills) == 0 && len(trackedRepos) == 0 {
		ui.Info("No skills installed")
		ui.Info("Use 'skillshare install <source>' to install a skill")
		return nil
	}

	// Display skills
	if len(skills) > 0 {
		ui.Header("Installed skills")
		fmt.Println(strings.Repeat("-", 55))

		for _, s := range skills {
			if verbose {
				fmt.Printf("  %s\n", s.Name)
				if s.RepoName != "" {
					fmt.Printf("    Tracked repo: %s\n", s.RepoName)
				}
				if s.Source != "" {
					fmt.Printf("    Source: %s\n", s.Source)
					fmt.Printf("    Type: %s\n", s.Type)
					fmt.Printf("    Installed: %s\n", s.InstalledAt)
				} else {
					fmt.Printf("    Source: (local - no metadata)\n")
				}
				fmt.Println()
			} else {
				// Determine display suffix
				var suffix string
				if s.RepoName != "" {
					suffix = fmt.Sprintf("(tracked: %s)", s.RepoName)
				} else if s.Source != "" {
					suffix = abbreviateSource(s.Source)
				} else {
					suffix = "(local)"
				}
				fmt.Printf("  %-30s  %s\n", s.Name, suffix)
			}
		}
	}

	// Display tracked repos section
	if len(trackedRepos) > 0 {
		fmt.Println()
		ui.Header("Tracked repositories")
		fmt.Println(strings.Repeat("-", 55))

		for _, repoName := range trackedRepos {
			repoPath := filepath.Join(cfg.Source, repoName)
			// Count skills in this repo
			skillCount := 0
			for _, d := range discovered {
				if d.IsInRepo && strings.HasPrefix(d.RelPath, repoName+"/") {
					skillCount++
				}
			}
			// Check git status
			statusStr := "up-to-date"
			if isDirty, _ := isRepoDirty(repoPath); isDirty {
				statusStr = "has changes"
			}
			fmt.Printf("  %-20s  %d skills, %s\n", repoName, skillCount, statusStr)
		}
	}

	if !verbose && len(skills) > 0 {
		fmt.Println()
		ui.Info("Use --verbose for more details")
	}

	return nil
}

type skillEntry struct {
	Name        string
	Source      string
	Type        string
	InstalledAt string
	IsNested    bool
	RepoName    string
}

// abbreviateSource shortens long sources for display
func abbreviateSource(source string) string {
	// Remove https:// prefix
	source = strings.TrimPrefix(source, "https://")
	source = strings.TrimPrefix(source, "http://")

	// Truncate if too long
	if len(source) > 50 {
		return source[:47] + "..."
	}
	return source
}

func printListHelp() {
	fmt.Println(`Usage: skillshare list [options]

List all installed skills in the source directory.

Options:
  --verbose, -v   Show detailed information (source, type, install date)
  --help, -h      Show this help

Examples:
  skillshare list
  skillshare list --verbose`)
}
