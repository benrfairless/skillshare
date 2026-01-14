package main

import (
	"fmt"
	"strings"

	"skillshare/internal/backup"
	"skillshare/internal/config"
	"skillshare/internal/ui"
)

func cmdBackup(args []string) error {
	var targetName string
	doList := false
	doCleanup := false

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--list", "-l":
			doList = true
		case "--cleanup", "-c":
			doCleanup = true
		case "--target", "-t":
			if i+1 < len(args) {
				targetName = args[i+1]
				i++
			}
		default:
			targetName = args[i]
		}
	}

	if doList {
		return backupList()
	}

	if doCleanup {
		return backupCleanup()
	}

	return createBackup(targetName)
}

func createBackup(targetName string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	targets := cfg.Targets
	if targetName != "" {
		if t, exists := cfg.Targets[targetName]; exists {
			targets = map[string]config.TargetConfig{targetName: t}
		} else {
			return fmt.Errorf("target '%s' not found", targetName)
		}
	}

	ui.Header("Creating backup")
	created := 0
	for name, target := range targets {
		backupPath, err := backup.Create(name, target.Path)
		if err != nil {
			ui.Warning("Failed to backup %s: %v", name, err)
			continue
		}
		if backupPath != "" {
			ui.Success("%s -> %s", name, backupPath)
			created++
		} else {
			ui.Info("%s: nothing to backup (empty or symlink)", name)
		}
	}

	if created == 0 {
		ui.Info("No backups created")
	}

	// List recent backups
	backups, _ := backup.List()
	if len(backups) > 0 {
		ui.Header("Recent backups")
		limit := 5
		if len(backups) < limit {
			limit = len(backups)
		}
		for i := 0; i < limit; i++ {
			b := backups[i]
			fmt.Printf("  %s %s (%s)\n", b.Timestamp, ui.Gray+strings.Join(b.Targets, ", ")+ui.Reset, b.Path)
		}
	}

	return nil
}

func backupList() error {
	backups, err := backup.List()
	if err != nil {
		return err
	}

	if len(backups) == 0 {
		ui.Info("No backups found")
		return nil
	}

	totalSize, _ := backup.TotalSize()
	ui.Header(fmt.Sprintf("All backups (%.1f MB total)", float64(totalSize)/(1024*1024)))

	for _, b := range backups {
		size := backup.Size(b.Path)
		fmt.Printf("  %s  %-20s  %6.1f MB  %s\n",
			b.Timestamp,
			strings.Join(b.Targets, ", "),
			float64(size)/(1024*1024),
			b.Path)
	}

	return nil
}

func backupCleanup() error {
	ui.Header("Cleaning up old backups")

	// Show current state
	backups, err := backup.List()
	if err != nil {
		return err
	}

	if len(backups) == 0 {
		ui.Info("No backups to clean up")
		return nil
	}

	totalSize, _ := backup.TotalSize()
	ui.Info("Current: %d backups, %.1f MB total", len(backups), float64(totalSize)/(1024*1024))

	// Use default cleanup config
	cfg := backup.DefaultCleanupConfig()
	removed, err := backup.Cleanup(cfg)
	if err != nil {
		return err
	}

	if removed > 0 {
		newSize, _ := backup.TotalSize()
		ui.Success("Removed %d old backups (freed %.1f MB)",
			removed,
			float64(totalSize-newSize)/(1024*1024))
	} else {
		ui.Info("No backups needed to be removed")
	}

	return nil
}

func cmdRestore(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: skillshare restore <target> [--from <timestamp>] [--force]")
	}

	var targetName string
	var fromTimestamp string
	force := false

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--from", "-f":
			if i+1 < len(args) {
				fromTimestamp = args[i+1]
				i++
			}
		case "--force":
			force = true
		default:
			if targetName == "" {
				targetName = args[i]
			}
		}
	}

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	target, exists := cfg.Targets[targetName]
	if !exists {
		return fmt.Errorf("target '%s' not found in config", targetName)
	}

	ui.Header(fmt.Sprintf("Restoring %s", targetName))

	opts := backup.RestoreOptions{Force: force}

	if fromTimestamp != "" {
		return restoreFromTimestamp(targetName, target.Path, fromTimestamp, opts)
	}

	return restoreFromLatest(targetName, target.Path, opts)
}

func restoreFromTimestamp(targetName, targetPath, timestamp string, opts backup.RestoreOptions) error {
	backupInfo, err := backup.GetBackupByTimestamp(timestamp)
	if err != nil {
		return err
	}

	if err := backup.RestoreToPath(backupInfo.Path, targetName, targetPath, opts); err != nil {
		return err
	}
	ui.Success("Restored %s from backup %s", targetName, timestamp)
	return nil
}

func restoreFromLatest(targetName, targetPath string, opts backup.RestoreOptions) error {
	timestamp, err := backup.RestoreLatest(targetName, targetPath, opts)
	if err != nil {
		return err
	}
	ui.Success("Restored %s from latest backup (%s)", targetName, timestamp)
	return nil
}
