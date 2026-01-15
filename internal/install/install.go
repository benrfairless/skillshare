package install

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// InstallOptions configures the install behavior
type InstallOptions struct {
	Name   string // Override skill name
	Force  bool   // Overwrite existing
	DryRun bool   // Preview only
	Update bool   // Update existing installation
}

// InstallResult reports the outcome of an installation
type InstallResult struct {
	SkillName string
	SkillPath string
	Source    string
	Action    string // "cloned", "copied", "updated", "skipped"
	Warnings  []string
}

// SkillInfo represents a discovered skill in a repository
type SkillInfo struct {
	Name string // Skill name (directory name)
	Path string // Relative path from repo root
}

// DiscoveryResult contains discovered skills from a repository
type DiscoveryResult struct {
	RepoPath string      // Temp directory where repo was cloned
	Skills   []SkillInfo // Discovered skills
	Source   *Source     // Original source
}

// Install executes the installation from source to destination
func Install(source *Source, destPath string, opts InstallOptions) (*InstallResult, error) {
	result := &InstallResult{
		SkillName: source.Name,
		Source:    source.Raw,
	}

	// Check if destination exists
	destInfo, destErr := os.Stat(destPath)
	destExists := destErr == nil

	if destExists {
		if opts.Update {
			return handleUpdate(source, destPath, result, opts)
		}
		if !opts.Force {
			return nil, fmt.Errorf("skill '%s' already exists. Use --force to overwrite or --update to update", source.Name)
		}
		// Force mode: remove existing
		if !opts.DryRun {
			if err := os.RemoveAll(destPath); err != nil {
				return nil, fmt.Errorf("failed to remove existing skill: %w", err)
			}
		}
	} else if destInfo != nil && !destInfo.IsDir() {
		return nil, fmt.Errorf("destination exists but is not a directory")
	}

	result.SkillPath = destPath

	// Execute installation based on source type
	switch source.Type {
	case SourceTypeLocalPath:
		return installFromLocal(source, destPath, result, opts)
	case SourceTypeGitHub, SourceTypeGitHTTPS, SourceTypeGitSSH:
		return installFromGit(source, destPath, result, opts)
	default:
		return nil, fmt.Errorf("unsupported source type: %s", source.Type)
	}
}

func installFromLocal(source *Source, destPath string, result *InstallResult, opts InstallOptions) (*InstallResult, error) {
	// Verify source exists
	srcInfo, err := os.Stat(source.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("source path does not exist: %s", source.Path)
		}
		return nil, fmt.Errorf("cannot access source path: %w", err)
	}
	if !srcInfo.IsDir() {
		return nil, fmt.Errorf("source path is not a directory: %s", source.Path)
	}

	if opts.DryRun {
		result.Action = "would copy"
		return result, nil
	}

	// Copy directory
	if err := copyDir(source.Path, destPath); err != nil {
		return nil, fmt.Errorf("failed to copy skill: %w", err)
	}

	// Write metadata
	meta := NewMetaFromSource(source)
	if err := WriteMeta(destPath, meta); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("failed to write metadata: %v", err))
	}

	// Check for SKILL.md
	checkSkillFile(destPath, result)

	result.Action = "copied"
	return result, nil
}

func installFromGit(source *Source, destPath string, result *InstallResult, opts InstallOptions) (*InstallResult, error) {
	// Check if git is available
	if !isGitInstalled() {
		return nil, fmt.Errorf("git is not installed or not in PATH")
	}

	// If subdir is specified, install directly
	if source.HasSubdir() {
		return installFromGitSubdir(source, destPath, result, opts)
	}

	// No subdir specified - this should be handled by DiscoverFromGit first
	// If we get here, treat it as "install entire repo as one skill"
	if opts.DryRun {
		result.Action = "would clone"
		return result, nil
	}

	// Clone the repository
	if err := cloneRepo(source.CloneURL, destPath, true); err != nil {
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	// Write metadata
	meta := NewMetaFromSource(source)
	// Try to get the commit hash
	if hash, err := getGitCommit(destPath); err == nil {
		meta.Version = hash
	}
	if err := WriteMeta(destPath, meta); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("failed to write metadata: %v", err))
	}

	// Check for SKILL.md
	checkSkillFile(destPath, result)

	result.Action = "cloned"
	return result, nil
}

// DiscoverFromGit clones a repo and discovers available skills
func DiscoverFromGit(source *Source) (*DiscoveryResult, error) {
	if !isGitInstalled() {
		return nil, fmt.Errorf("git is not installed or not in PATH")
	}

	// Clone to temp directory
	tempDir, err := os.MkdirTemp("", "skillshare-discover-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	repoPath := filepath.Join(tempDir, "repo")
	if err := cloneRepo(source.CloneURL, repoPath, true); err != nil {
		os.RemoveAll(tempDir)
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	// Discover skills
	skills := discoverSkills(repoPath)

	return &DiscoveryResult{
		RepoPath: tempDir,
		Skills:   skills,
		Source:   source,
	}, nil
}

// discoverSkills finds directories containing SKILL.md
func discoverSkills(repoPath string) []SkillInfo {
	var skills []SkillInfo

	filepath.Walk(repoPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip hidden directories and .git
		if info.IsDir() && (info.Name() == ".git" || (len(info.Name()) > 0 && info.Name()[0] == '.')) {
			return filepath.SkipDir
		}

		// Check if this is a SKILL.md file
		if !info.IsDir() && info.Name() == "SKILL.md" {
			skillDir := filepath.Dir(path)
			relPath, _ := filepath.Rel(repoPath, skillDir)

			// Skip root level SKILL.md (repo itself is not a skill container)
			if relPath != "." {
				skills = append(skills, SkillInfo{
					Name: filepath.Base(skillDir),
					Path: relPath,
				})
			}
		}

		return nil
	})

	return skills
}

// CleanupDiscovery removes the temporary directory from discovery
func CleanupDiscovery(result *DiscoveryResult) {
	if result != nil && result.RepoPath != "" {
		os.RemoveAll(result.RepoPath)
	}
}

// InstallFromDiscovery installs a skill from a discovered repository
func InstallFromDiscovery(discovery *DiscoveryResult, skill SkillInfo, destPath string, opts InstallOptions) (*InstallResult, error) {
	result := &InstallResult{
		SkillName: skill.Name,
		Source:    discovery.Source.Raw + "/" + skill.Path,
	}

	// Check if destination exists
	if _, err := os.Stat(destPath); err == nil {
		if !opts.Force {
			return nil, fmt.Errorf("skill '%s' already exists. Use --force to overwrite", skill.Name)
		}
		if !opts.DryRun {
			if err := os.RemoveAll(destPath); err != nil {
				return nil, fmt.Errorf("failed to remove existing skill: %w", err)
			}
		}
	}

	result.SkillPath = destPath

	if opts.DryRun {
		result.Action = "would install"
		return result, nil
	}

	// Copy from temp repo
	srcPath := filepath.Join(discovery.RepoPath, "repo", skill.Path)
	if err := copyDir(srcPath, destPath); err != nil {
		return nil, fmt.Errorf("failed to copy skill: %w", err)
	}

	// Write metadata
	source := &Source{
		Type:     discovery.Source.Type,
		Raw:      discovery.Source.Raw + "/" + skill.Path,
		CloneURL: discovery.Source.CloneURL,
		Subdir:   skill.Path,
		Name:     skill.Name,
	}
	meta := NewMetaFromSource(source)
	if hash, err := getGitCommit(filepath.Join(discovery.RepoPath, "repo")); err == nil {
		meta.Version = hash
	}
	if err := WriteMeta(destPath, meta); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("failed to write metadata: %v", err))
	}

	result.Action = "installed"
	return result, nil
}

func installFromGitSubdir(source *Source, destPath string, result *InstallResult, opts InstallOptions) (*InstallResult, error) {
	if opts.DryRun {
		result.Action = "would clone and extract"
		return result, nil
	}

	// Clone to temp directory
	tempDir, err := os.MkdirTemp("", "skillshare-install-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	tempRepoPath := filepath.Join(tempDir, "repo")
	if err := cloneRepo(source.CloneURL, tempRepoPath, true); err != nil {
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	// Verify subdirectory exists
	subdirPath := filepath.Join(tempRepoPath, source.Subdir)
	info, err := os.Stat(subdirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("subdirectory '%s' does not exist in repository", source.Subdir)
		}
		return nil, fmt.Errorf("cannot access subdirectory: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("'%s' is not a directory", source.Subdir)
	}

	// Copy subdirectory to destination
	if err := copyDir(subdirPath, destPath); err != nil {
		return nil, fmt.Errorf("failed to copy skill: %w", err)
	}

	// Write metadata
	meta := NewMetaFromSource(source)
	// Try to get the commit hash from temp repo
	if hash, err := getGitCommit(tempRepoPath); err == nil {
		meta.Version = hash
	}
	if err := WriteMeta(destPath, meta); err != nil {
		result.Warnings = append(result.Warnings, fmt.Sprintf("failed to write metadata: %v", err))
	}

	// Check for SKILL.md
	checkSkillFile(destPath, result)

	result.Action = "cloned and extracted"
	return result, nil
}

func handleUpdate(source *Source, destPath string, result *InstallResult, opts InstallOptions) (*InstallResult, error) {
	result.SkillPath = destPath

	// For git repos without subdir, try git pull
	if source.IsGit() && !source.HasSubdir() && isGitRepo(destPath) {
		if opts.DryRun {
			result.Action = "would update (git pull)"
			return result, nil
		}

		if err := gitPull(destPath); err != nil {
			return nil, fmt.Errorf("failed to update: %w", err)
		}

		// Update metadata timestamp
		meta, _ := ReadMeta(destPath)
		if meta != nil {
			if hash, err := getGitCommit(destPath); err == nil {
				meta.Version = hash
			}
			WriteMeta(destPath, meta)
		}

		result.Action = "updated"
		return result, nil
	}

	// For other cases, need to reinstall
	if !opts.Force {
		return nil, fmt.Errorf("cannot update '%s': use --force to reinstall", source.Name)
	}

	// Remove and reinstall
	if !opts.DryRun {
		if err := os.RemoveAll(destPath); err != nil {
			return nil, fmt.Errorf("failed to remove existing skill: %w", err)
		}
	}

	return Install(source, destPath, InstallOptions{
		Name:   opts.Name,
		Force:  false, // Already handled
		DryRun: opts.DryRun,
		Update: false,
	})
}

// checkSkillFile adds a warning if SKILL.md is not found
func checkSkillFile(skillPath string, result *InstallResult) {
	skillFile := filepath.Join(skillPath, "SKILL.md")
	if _, err := os.Stat(skillFile); os.IsNotExist(err) {
		result.Warnings = append(result.Warnings, "no SKILL.md found in skill directory")
	}
}

// isGitInstalled checks if git command is available
func isGitInstalled() bool {
	_, err := exec.LookPath("git")
	return err == nil
}

// isGitRepo checks if path is a git repository
func isGitRepo(path string) bool {
	gitDir := filepath.Join(path, ".git")
	info, err := os.Stat(gitDir)
	return err == nil && info.IsDir()
}

// cloneRepo performs a git clone
func cloneRepo(url, destPath string, shallow bool) error {
	args := []string{"clone"}
	if shallow {
		args = append(args, "--depth", "1")
	}
	args = append(args, url, destPath)

	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// gitPull performs a git pull
func gitPull(repoPath string) error {
	cmd := exec.Command("git", "pull")
	cmd.Dir = repoPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// getGitCommit returns the current HEAD commit hash
func getGitCommit(repoPath string) (string, error) {
	cmd := exec.Command("git", "rev-parse", "--short", "HEAD")
	cmd.Dir = repoPath
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output[:len(output)-1]), nil // Remove trailing newline
}

// copyDir copies a directory recursively
func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, _ := filepath.Rel(src, path)
		dstPath := filepath.Join(dst, relPath)

		// Skip .git directory
		if info.IsDir() && info.Name() == ".git" {
			return filepath.SkipDir
		}

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		return copyFile(path, dstPath)
	})
}

// copyFile copies a single file
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}
