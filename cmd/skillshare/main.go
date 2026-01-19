package main

import (
	"fmt"
	"os"

	"skillshare/internal/ui"
)

var version = "dev"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	var err error
	switch cmd {
	case "init":
		err = cmdInit(args)
	case "install":
		err = cmdInstall(args)
	case "uninstall":
		err = cmdUninstall(args)
	case "list":
		err = cmdList(args)
	case "sync":
		err = cmdSync(args)
	case "status":
		err = cmdStatus(args)
	case "diff":
		err = cmdDiff(args)
	case "backup":
		err = cmdBackup(args)
	case "restore":
		err = cmdRestore(args)
	case "pull":
		err = cmdPull(args)
	case "push":
		err = cmdPush(args)
	case "doctor":
		err = cmdDoctor(args)
	case "target":
		err = cmdTarget(args)
	case "upgrade":
		err = cmdUpgrade(args)
	case "update":
		err = cmdUpdate(args)
	case "version", "-v", "--version":
		fmt.Printf("skillshare %s\n", version)
	case "help", "-h", "--help":
		printUsage()
	default:
		ui.Error("Unknown command: %s", cmd)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		ui.Error("%v", err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`skillshare - Share skills across AI CLI tools

Usage:
  skillshare <command> [options]

Commands:
  init [--source PATH] [--remote URL] [--dry-run] Initialize skillshare
  install <source> [--dry-run]      Install a skill from local path or git repo
  uninstall <name> [--force]        Remove a skill from source directory
  list [--verbose]                  List all installed skills
  sync [--dry-run] [--force]        Sync skills to all targets
  status                            Show status of all targets
  diff [--target name]              Show differences between source and targets
  backup [--target name] [--dry-run] Create backup of target(s)
  backup --list                     List all backups
  backup --cleanup [--dry-run]      Clean up old backups
  restore <target> [--force] [--dry-run] Restore target from latest backup
  restore <target> --from TS [--force] [--dry-run] Restore target from specific backup
  pull [target] [--dry-run]         Pull local skills from target(s) to source
  pull --all [--dry-run]            Pull from all targets
  pull --remote [--dry-run]         Pull from git remote and sync to all targets
  push [-m MSG] [--dry-run]         Commit and push skills to git remote
  doctor                            Check environment and diagnose issues
  target <name>                     Show target info
  target <name> --mode MODE         Set target sync mode (merge|symlink)
  target add <name> <path>          Add a target
  target remove <name> [--dry-run]  Unlink target and restore skills
  target remove --all [--dry-run]   Unlink all targets
  target list                       List all targets
  upgrade [--skill|--cli] [--force] Upgrade CLI and/or skillshare skill
  update <name> [--dry-run]         Update a skill or tracked repository
  update --all [--dry-run]          Update all tracked repositories
  version                           Show version
  help                              Show this help

Examples:
  skillshare init --source ~/.skills
  skillshare install github.com/user/skill-repo
  skillshare install github.com/org/repo/path/to/skill --dry-run
  skillshare target add claude ~/.claude/skills
  skillshare sync
  skillshare status
  skillshare backup --list
  skillshare restore claude`)
}
