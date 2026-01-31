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

// parseListArgs parses list command arguments
func parseListArgs(args []string) (verbose bool, showHelp bool, err error) {
	for _, arg := range args {
		switch arg {
		case "--verbose", "-v":
			verbose = true
		case "--help", "-h":
			return false, true, nil
		default:
			if strings.HasPrefix(arg, "-") {
				return false, false, fmt.Errorf("unknown option: %s", arg)
			}
		}
	}
	return verbose, false, nil
}

// buildSkillEntries builds skill entries from discovered skills
func buildSkillEntries(discovered []sync.DiscoveredSkill) []skillEntry {
	var skills []skillEntry
	for _, d := range discovered {
		entry := skillEntry{
			Name:     d.FlatName,
			IsNested: d.IsInRepo || utils.HasNestedSeparator(d.FlatName),
		}

		if d.IsInRepo {
			parts := strings.SplitN(d.RelPath, "/", 2)
			if len(parts) > 0 {
				entry.RepoName = parts[0]
			}
		}

		if meta, err := install.ReadMeta(d.SourcePath); err == nil && meta != nil {
			entry.Source = meta.Source
			entry.Type = meta.Type
			entry.InstalledAt = meta.InstalledAt.Format("2006-01-02")
		}

		skills = append(skills, entry)
	}
	return skills
}

// displaySkillsVerbose displays skills in verbose mode
func displaySkillsVerbose(skills []skillEntry) {
	for _, s := range skills {
		fmt.Printf("  %s%s%s\n", ui.Cyan, s.Name, ui.Reset)
		if s.RepoName != "" {
			fmt.Printf("    %sTracked repo:%s %s\n", ui.Gray, ui.Reset, s.RepoName)
		}
		if s.Source != "" {
			fmt.Printf("    %sSource:%s      %s\n", ui.Gray, ui.Reset, s.Source)
			fmt.Printf("    %sType:%s        %s\n", ui.Gray, ui.Reset, s.Type)
			fmt.Printf("    %sInstalled:%s   %s\n", ui.Gray, ui.Reset, s.InstalledAt)
		} else {
			fmt.Printf("    %sSource:%s      (local - no metadata)\n", ui.Gray, ui.Reset)
		}
		fmt.Println()
	}
}

// displaySkillsCompact displays skills in compact mode
func displaySkillsCompact(skills []skillEntry) {
	maxNameLen := 0
	for _, s := range skills {
		if len(s.Name) > maxNameLen {
			maxNameLen = len(s.Name)
		}
	}

	for _, s := range skills {
		suffix := getSkillSuffix(s)
		format := fmt.Sprintf("  %s→%s %%-%ds  %s%%s%s\n", ui.Cyan, ui.Reset, maxNameLen, ui.Gray, ui.Reset)
		fmt.Printf(format, s.Name, suffix)
	}
}

// getSkillSuffix returns the display suffix for a skill
func getSkillSuffix(s skillEntry) string {
	if s.RepoName != "" {
		return fmt.Sprintf("tracked: %s", s.RepoName)
	}
	if s.Source != "" {
		return abbreviateSource(s.Source)
	}
	return "local"
}

// displayTrackedRepos displays the tracked repositories section
func displayTrackedRepos(trackedRepos []string, discovered []sync.DiscoveredSkill, sourcePath string) {
	fmt.Println()
	ui.Header("Tracked repositories")

	for _, repoName := range trackedRepos {
		repoPath := filepath.Join(sourcePath, repoName)
		skillCount := countRepoSkills(repoName, discovered)

		if isDirty, _ := isRepoDirty(repoPath); isDirty {
			ui.ListItem("warning", repoName, fmt.Sprintf("%d skills, has changes", skillCount))
		} else {
			ui.ListItem("success", repoName, fmt.Sprintf("%d skills, up-to-date", skillCount))
		}
	}
}

// countRepoSkills counts skills in a tracked repo
func countRepoSkills(repoName string, discovered []sync.DiscoveredSkill) int {
	count := 0
	for _, d := range discovered {
		if d.IsInRepo && strings.HasPrefix(d.RelPath, repoName+"/") {
			count++
		}
	}
	return count
}

func cmdList(args []string) error {
	verbose, showHelp, err := parseListArgs(args)
	if showHelp {
		printListHelp()
		return nil
	}
	if err != nil {
		return err
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	discovered, err := sync.DiscoverSourceSkills(cfg.Source)
	if err != nil {
		return fmt.Errorf("cannot discover skills: %w", err)
	}

	trackedRepos, _ := install.GetTrackedRepos(cfg.Source)
	skills := buildSkillEntries(discovered)

	if len(skills) == 0 && len(trackedRepos) == 0 {
		ui.Info("No skills installed")
		ui.Info("Use 'skillshare install <source>' to install a skill")
		return nil
	}

	if len(skills) > 0 {
		ui.Header("Installed skills")
		if verbose {
			displaySkillsVerbose(skills)
		} else {
			displaySkillsCompact(skills)
		}
	}

	if len(trackedRepos) > 0 {
		displayTrackedRepos(trackedRepos, discovered, cfg.Source)
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
