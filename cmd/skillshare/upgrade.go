package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"skillshare/internal/config"
	"skillshare/internal/install"
	"skillshare/internal/ui"
	versionpkg "skillshare/internal/version"
)

func cmdUpgrade(args []string) error {
	dryRun := false
	force := false
	skillOnly := false
	cliOnly := false

	// Parse args
	for _, arg := range args {
		switch arg {
		case "--dry-run", "-n":
			dryRun = true
		case "--force", "-f":
			force = true
		case "--skill":
			skillOnly = true
		case "--cli":
			cliOnly = true
		case "--help", "-h":
			printUpgradeHelp()
			return nil
		}
	}

	// Show logo
	ui.Logo(version)

	// Default: upgrade both
	upgradeCLI := !skillOnly
	upgradeSkill := !cliOnly

	if dryRun {
		ui.Warning("Dry run mode - no changes will be made")
		fmt.Println()
	}

	var cliErr, skillErr error

	// Upgrade CLI
	if upgradeCLI {
		cliErr = upgradeCLIBinary(dryRun, force)
	}

	// Upgrade skill
	if upgradeSkill {
		if upgradeCLI {
			fmt.Println()
		}
		skillErr = upgradeSkillshareSkill(dryRun, force)
	}

	// Return first error
	if cliErr != nil {
		return cliErr
	}
	return skillErr
}

func upgradeCLIBinary(dryRun, force bool) error {
	// Step 1: Show current version
	ui.StepStart("CLI", fmt.Sprintf("v%s", version))

	// Get current executable path
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}
	execPath, err = filepath.EvalSymlinks(execPath)
	if err != nil {
		return fmt.Errorf("failed to resolve symlink: %w", err)
	}

	// Check if installed via Homebrew
	if isHomebrewInstall(execPath) {
		ui.StepContinue("Install", "Homebrew")
		if dryRun {
			ui.StepEnd("Action", "Would run: brew upgrade runkids/tap/skillshare")
			return nil
		}
		return runBrewUpgrade()
	}

	// Get latest version from GitHub
	treeSpinner := ui.StartTreeSpinner("Checking latest version...", false)
	release, err := versionpkg.FetchLatestRelease()
	if err != nil {
		treeSpinner.Fail("Failed to check")
		return fmt.Errorf("failed to check latest version: %w", err)
	}
	latestVersion := release.Version
	treeSpinner.Success(fmt.Sprintf("Latest: v%s", latestVersion))

	if version == latestVersion && !force {
		ui.StepEnd("Status", "Already up to date ✓")
		return nil
	}

	if dryRun {
		ui.StepEnd("Action", fmt.Sprintf("Would download v%s", latestVersion))
		return nil
	}

	// Confirm if not forced
	if !force {
		fmt.Println()
		fmt.Printf("  Upgrade to v%s? [y/N]: ", latestVersion)
		var input string
		fmt.Scanln(&input)
		if input != "y" && input != "Y" && input != "yes" {
			ui.Info("Cancelled")
			return nil
		}
	}

	// Get download URL for current platform
	downloadURL, err := release.GetDownloadURL()
	if err != nil {
		return fmt.Errorf("failed to get download URL: %w", err)
	}

	// Download
	fmt.Println()
	downloadSpinner := ui.StartSpinner(fmt.Sprintf("Downloading v%s...", latestVersion))
	if err := downloadAndReplace(downloadURL, execPath); err != nil {
		downloadSpinner.Fail("Failed to download")
		return fmt.Errorf("failed to upgrade: %w", err)
	}
	downloadSpinner.Success(fmt.Sprintf("Upgraded to v%s", latestVersion))

	// Clear version cache so next check fetches fresh data
	versionpkg.ClearCache()

	return nil
}

func upgradeSkillshareSkill(dryRun, force bool) error {
	// Step 1: Show skill info
	ui.StepStart("Skill", "skillshare")

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config not found: run 'skillshare init' first")
	}

	skillshareSkillDir := filepath.Join(cfg.Source, "skillshare")
	skillshareSkillFile := filepath.Join(skillshareSkillDir, "SKILL.md")

	// Check if skill exists
	exists := false
	if _, err := os.Stat(skillshareSkillFile); err == nil {
		exists = true
	}

	// If exists and not forced, just show status and return
	if exists && !force {
		ui.StepEnd("Status", "Already up to date ✓")
		return nil
	}

	status := "Not installed"
	if exists {
		status = "Will upgrade"
	}
	ui.StepContinue("Status", status)

	if dryRun {
		action := "Would download"
		if exists {
			action = "Would upgrade"
		}
		ui.StepEnd("Action", action)
		return nil
	}

	// Install using install package
	treeSpinner := ui.StartTreeSpinner("Downloading from GitHub...", true)

	source, err := install.ParseSource(skillshareSkillSource)
	if err != nil {
		treeSpinner.Fail("Failed to parse source")
		return err
	}
	source.Name = "skillshare"

	_, err = install.Install(source, skillshareSkillDir, install.InstallOptions{
		Force:  true,
		DryRun: false,
	})
	if err != nil {
		treeSpinner.Fail("Failed to download")
		return err
	}

	treeSpinner.Success("Upgraded")

	fmt.Println()
	ui.Info("Run 'skillshare sync' to distribute to all targets")

	return nil
}

func downloadAndReplace(url, destPath string) error {
	// Download tarball
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Extract from tar.gz
	gzr, err := gzip.NewReader(resp.Body)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	// Find skillshare binary in archive
	for {
		header, err := tr.Next()
		if err == io.EOF {
			return fmt.Errorf("skillshare binary not found in archive")
		}
		if err != nil {
			return err
		}

		if header.Name == "skillshare" || header.Name == "./skillshare" {
			// Write to temp file first
			tmpFile, err := os.CreateTemp(filepath.Dir(destPath), "skillshare-upgrade-*")
			if err != nil {
				return err
			}
			tmpPath := tmpFile.Name()

			if _, err := io.Copy(tmpFile, tr); err != nil {
				tmpFile.Close()
				os.Remove(tmpPath)
				return err
			}
			tmpFile.Close()

			// Make executable
			if err := os.Chmod(tmpPath, 0755); err != nil {
				os.Remove(tmpPath)
				return err
			}

			// Replace original
			if err := os.Rename(tmpPath, destPath); err != nil {
				os.Remove(tmpPath)
				return err
			}

			return nil
		}
	}
}

func isHomebrewInstall(execPath string) bool {
	// Check common Homebrew paths
	homebrewPaths := []string{
		"/usr/local/Cellar/skillshare",
		"/opt/homebrew/Cellar/skillshare",
		"/home/linuxbrew/.linuxbrew/Cellar/skillshare",
	}
	for _, prefix := range homebrewPaths {
		if strings.HasPrefix(execPath, prefix) {
			return true
		}
	}
	return false
}

func runBrewUpgrade() error {
	// First update the tap to get latest formula
	ui.Info("Updating tap...")
	updateCmd := exec.Command("brew", "update", "--quiet")
	updateCmd.Stdout = os.Stdout
	updateCmd.Stderr = os.Stderr
	if err := updateCmd.Run(); err != nil {
		ui.Warning("brew update failed, trying upgrade anyway...")
	}

	// Then upgrade
	ui.Info("Upgrading...")
	cmd := exec.Command("brew", "upgrade", "runkids/tap/skillshare")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err == nil {
		// Clear version cache so next check fetches fresh data
		versionpkg.ClearCache()
	}
	return err
}

func printUpgradeHelp() {
	fmt.Println(`Usage: skillshare upgrade [options]

Upgrade the CLI binary and/or built-in skillshare skill.

Options:
  --skill       Upgrade skill only
  --cli         Upgrade CLI only
  --force, -f   Skip confirmation prompts
  --dry-run, -n Preview without making changes
  --help, -h    Show this help

Examples:
  skillshare upgrade              # Upgrade both CLI and skill
  skillshare upgrade --cli        # Upgrade CLI only
  skillshare upgrade --skill      # Upgrade skill only
  skillshare upgrade --dry-run    # Preview upgrades`)
}
