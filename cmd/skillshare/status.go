package main

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"skillshare/internal/config"
	"skillshare/internal/install"
	"skillshare/internal/sync"
	"skillshare/internal/ui"
	"skillshare/internal/utils"
)

func cmdStatus(args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	printSourceStatus(cfg)
	printTrackedReposStatus(cfg)
	printTargetsStatus(cfg)
	checkSkillVersion(cfg)

	return nil
}

func printSourceStatus(cfg *config.Config) {
	ui.Header("Source")
	info, err := os.Stat(cfg.Source)
	if err != nil {
		ui.Error("%s (not found)", cfg.Source)
		return
	}

	entries, _ := os.ReadDir(cfg.Source)
	skillCount := 0
	for _, e := range entries {
		if e.IsDir() && !utils.IsHidden(e.Name()) {
			skillCount++
		}
	}
	ui.Success("%s (%d skills, %s)", cfg.Source, skillCount, info.ModTime().Format("2006-01-02 15:04"))
}

func printTrackedReposStatus(cfg *config.Config) {
	trackedRepos, err := install.GetTrackedRepos(cfg.Source)
	if err != nil || len(trackedRepos) == 0 {
		return // No tracked repos, skip this section
	}

	ui.Header("Tracked Repositories")
	for _, repoName := range trackedRepos {
		repoPath := filepath.Join(cfg.Source, repoName)

		// Count skills in this repo
		discovered, _ := sync.DiscoverSourceSkills(cfg.Source)
		skillCount := 0
		for _, d := range discovered {
			if d.IsInRepo && strings.HasPrefix(d.RelPath, repoName+"/") {
				skillCount++
			}
		}

		// Check git status
		statusStr := "up-to-date"
		statusIcon := "✓"
		if isDirty, _ := checkRepoDirty(repoPath); isDirty {
			statusStr = "has uncommitted changes"
			statusIcon = "!"
		}

		ui.Status(repoName, statusIcon, fmt.Sprintf("%d skills, %s", skillCount, statusStr))
	}
}

// checkRepoDirty checks if a git repository has uncommitted changes
func checkRepoDirty(repoPath string) (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return len(strings.TrimSpace(string(output))) > 0, nil
}

func printTargetsStatus(cfg *config.Config) {
	ui.Header("Targets")
	for name, target := range cfg.Targets {
		mode := getTargetMode(target.Mode, cfg.Mode)
		statusStr, detail := getTargetStatusDetail(target, cfg.Source, mode)
		ui.Status(name, statusStr, detail)
	}
}

func getTargetMode(targetMode, globalMode string) string {
	if targetMode != "" {
		return targetMode
	}
	if globalMode != "" {
		return globalMode
	}
	return "merge"
}

func getTargetStatusDetail(target config.TargetConfig, source, mode string) (string, string) {
	if mode == "merge" {
		return getMergeStatusDetail(target, source, mode)
	}
	return getSymlinkStatusDetail(target, source, mode)
}

func getMergeStatusDetail(target config.TargetConfig, source, mode string) (string, string) {
	status, linkedCount, localCount := sync.CheckStatusMerge(target.Path, source)

	switch status {
	case sync.StatusMerged:
		return "merged", fmt.Sprintf("[%s] %s (%d shared, %d local)", mode, target.Path, linkedCount, localCount)
	case sync.StatusLinked:
		// Configured as merge but actually using symlink - needs resync
		return "linked", fmt.Sprintf("[%s->needs sync] %s", mode, target.Path)
	default:
		return status.String(), fmt.Sprintf("[%s] %s (%d local)", mode, target.Path, localCount)
	}
}

func getSymlinkStatusDetail(target config.TargetConfig, source, mode string) (string, string) {
	status := sync.CheckStatus(target.Path, source)
	detail := fmt.Sprintf("[%s] %s", mode, target.Path)

	switch status {
	case sync.StatusConflict:
		link, _ := os.Readlink(target.Path)
		detail = fmt.Sprintf("[%s] %s -> %s", mode, target.Path, link)
	case sync.StatusMerged:
		// Configured as symlink but actually using merge - needs resync
		detail = fmt.Sprintf("[%s->needs sync] %s", mode, target.Path)
	}

	return status.String(), detail
}

func checkSkillVersion(cfg *config.Config) {
	skillFile := filepath.Join(cfg.Source, "skillshare", "SKILL.md")
	localVersion := readSkillVersion(skillFile)

	if localVersion == "" {
		// No skill or no version - suggest update
		ui.Header("Skill Update")
		ui.Warning("skillshare skill not found or missing version")
		ui.Info("Run: skillshare update")
		return
	}

	// Fetch remote version (with short timeout)
	remoteVersion := fetchRemoteSkillVersion()
	if remoteVersion == "" {
		return // Network error, skip check silently
	}

	// Compare local vs remote
	if localVersion != remoteVersion {
		ui.Header("Skill Update")
		ui.Warning("skillshare skill update available (%s → %s)", localVersion, remoteVersion)
		ui.Info("Run: skillshare update && skillshare sync")
	}
}

func fetchRemoteSkillVersion() string {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(skillshareSkillURL)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	scanner := bufio.NewScanner(resp.Body)
	inFrontmatter := false

	for scanner.Scan() {
		line := scanner.Text()

		if line == "---" {
			if !inFrontmatter {
				inFrontmatter = true
				continue
			}
			break
		}

		if inFrontmatter && strings.HasPrefix(line, "version:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}

	return ""
}

func readSkillVersion(skillFile string) string {
	file, err := os.Open(skillFile)
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	inFrontmatter := false

	for scanner.Scan() {
		line := scanner.Text()

		if line == "---" {
			if !inFrontmatter {
				inFrontmatter = true
				continue
			}
			// End of frontmatter
			break
		}

		if inFrontmatter && strings.HasPrefix(line, "version:") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}

	return ""
}
