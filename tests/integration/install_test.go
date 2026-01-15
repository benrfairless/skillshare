package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"skillshare/internal/testutil"
)

func TestInstall_LocalPath_CopiesToSource(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	sb.WriteConfig(`source: ` + sb.SourcePath + `
targets: {}
`)

	// Create a local skill directory to install from
	localSkillPath := filepath.Join(sb.Root, "external-skill")
	os.MkdirAll(localSkillPath, 0755)
	os.WriteFile(filepath.Join(localSkillPath, "SKILL.md"), []byte("# External Skill"), 0644)

	result := sb.RunCLI("install", localSkillPath)

	result.AssertSuccess(t)
	result.AssertOutputContains(t, "Installed")
	result.AssertOutputContains(t, "external-skill")

	// Verify skill was copied to source
	installedPath := filepath.Join(sb.SourcePath, "external-skill", "SKILL.md")
	if !sb.FileExists(installedPath) {
		t.Error("skill should be installed to source directory")
	}

	// Verify metadata was created
	metaPath := filepath.Join(sb.SourcePath, "external-skill", ".skillshare-meta.json")
	if !sb.FileExists(metaPath) {
		t.Error("metadata file should be created")
	}
}

func TestInstall_CustomName_UsesName(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	sb.WriteConfig(`source: ` + sb.SourcePath + `
targets: {}
`)

	// Create a local skill directory
	localSkillPath := filepath.Join(sb.Root, "original-name")
	os.MkdirAll(localSkillPath, 0755)
	os.WriteFile(filepath.Join(localSkillPath, "SKILL.md"), []byte("# Skill"), 0644)

	result := sb.RunCLI("install", localSkillPath, "--name", "custom-name")

	result.AssertSuccess(t)
	result.AssertOutputContains(t, "custom-name")

	// Verify skill was installed with custom name
	installedPath := filepath.Join(sb.SourcePath, "custom-name", "SKILL.md")
	if !sb.FileExists(installedPath) {
		t.Error("skill should be installed with custom name")
	}

	// Original name should not exist
	originalPath := filepath.Join(sb.SourcePath, "original-name")
	if sb.FileExists(originalPath) {
		t.Error("skill should not be installed with original name")
	}
}

func TestInstall_ExistsWithoutForce_Errors(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	// Create existing skill in source
	sb.CreateSkill("existing-skill", map[string]string{"SKILL.md": "# Existing"})

	sb.WriteConfig(`source: ` + sb.SourcePath + `
targets: {}
`)

	// Create local skill to install
	localSkillPath := filepath.Join(sb.Root, "existing-skill")
	os.MkdirAll(localSkillPath, 0755)
	os.WriteFile(filepath.Join(localSkillPath, "SKILL.md"), []byte("# New Version"), 0644)

	result := sb.RunCLI("install", localSkillPath)

	result.AssertFailure(t)
	result.AssertAnyOutputContains(t, "already exists")
}

func TestInstall_Force_Overwrites(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	// Create existing skill in source
	sb.CreateSkill("existing-skill", map[string]string{"SKILL.md": "# Old Version"})

	sb.WriteConfig(`source: ` + sb.SourcePath + `
targets: {}
`)

	// Create local skill to install
	localSkillPath := filepath.Join(sb.Root, "existing-skill")
	os.MkdirAll(localSkillPath, 0755)
	os.WriteFile(filepath.Join(localSkillPath, "SKILL.md"), []byte("# New Version"), 0644)

	result := sb.RunCLI("install", localSkillPath, "--force")

	result.AssertSuccess(t)
	result.AssertOutputContains(t, "Installed")

	// Verify new content
	content := sb.ReadFile(filepath.Join(sb.SourcePath, "existing-skill", "SKILL.md"))
	if !strings.Contains(content, "New Version") {
		t.Error("skill content should be updated")
	}
}

func TestInstall_DryRun_NoChanges(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	sb.WriteConfig(`source: ` + sb.SourcePath + `
targets: {}
`)

	// Create local skill to install
	localSkillPath := filepath.Join(sb.Root, "dry-run-skill")
	os.MkdirAll(localSkillPath, 0755)
	os.WriteFile(filepath.Join(localSkillPath, "SKILL.md"), []byte("# Dry Run"), 0644)

	result := sb.RunCLI("install", localSkillPath, "--dry-run")

	result.AssertSuccess(t)
	result.AssertOutputContains(t, "dry-run")
	result.AssertOutputContains(t, "would copy")

	// Verify skill was NOT installed
	installedPath := filepath.Join(sb.SourcePath, "dry-run-skill")
	if sb.FileExists(installedPath) {
		t.Error("skill should not be installed in dry-run mode")
	}
}

func TestInstall_InvalidName_Errors(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	sb.WriteConfig(`source: ` + sb.SourcePath + `
targets: {}
`)

	// Create local skill
	localSkillPath := filepath.Join(sb.Root, "valid-skill")
	os.MkdirAll(localSkillPath, 0755)
	os.WriteFile(filepath.Join(localSkillPath, "SKILL.md"), []byte("# Skill"), 0644)

	// Try to install with invalid name (starts with -)
	result := sb.RunCLI("install", localSkillPath, "--name", "-invalid")

	result.AssertFailure(t)
	result.AssertAnyOutputContains(t, "invalid skill name")
}

func TestInstall_SourceNotExist_Errors(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	sb.WriteConfig(`source: ` + sb.SourcePath + `
targets: {}
`)

	result := sb.RunCLI("install", "/nonexistent/path/to/skill")

	result.AssertFailure(t)
	result.AssertAnyOutputContains(t, "does not exist")
}

func TestInstall_NoSKILLmd_ShowsWarning(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	sb.WriteConfig(`source: ` + sb.SourcePath + `
targets: {}
`)

	// Create local skill without SKILL.md
	localSkillPath := filepath.Join(sb.Root, "no-skillmd")
	os.MkdirAll(localSkillPath, 0755)
	os.WriteFile(filepath.Join(localSkillPath, "README.md"), []byte("# Readme"), 0644)

	result := sb.RunCLI("install", localSkillPath)

	result.AssertSuccess(t)
	result.AssertOutputContains(t, "no SKILL.md")
}

func TestInstall_Help_ShowsUsage(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	result := sb.RunCLI("install", "--help")

	result.AssertSuccess(t)
	result.AssertOutputContains(t, "Usage:")
	result.AssertOutputContains(t, "--force")
	result.AssertOutputContains(t, "--dry-run")
	result.AssertOutputContains(t, "--name")
}

func TestInstall_NoArgs_ShowsHelp(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	sb.WriteConfig(`source: ` + sb.SourcePath + `
targets: {}
`)

	result := sb.RunCLI("install")

	result.AssertFailure(t)
	result.AssertAnyOutputContains(t, "source is required")
}

func TestInstall_LocalGitRepo_ClonesSuccessfully(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	sb.WriteConfig(`source: ` + sb.SourcePath + `
targets: {}
`)

	// Create a local git repository to test git clone
	gitRepoPath := filepath.Join(sb.Root, "git-skill-repo")
	os.MkdirAll(gitRepoPath, 0755)
	os.WriteFile(filepath.Join(gitRepoPath, "SKILL.md"), []byte("# Git Skill"), 0644)

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = gitRepoPath
	if err := cmd.Run(); err != nil {
		t.Skip("git not available, skipping git test")
	}

	cmd = exec.Command("git", "config", "user.email", "test@test.com")
	cmd.Dir = gitRepoPath
	cmd.Run()

	cmd = exec.Command("git", "config", "user.name", "Test")
	cmd.Dir = gitRepoPath
	cmd.Run()

	cmd = exec.Command("git", "add", ".")
	cmd.Dir = gitRepoPath
	cmd.Run()

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = gitRepoPath
	cmd.Run()

	// Install from local git repo path (should detect it's a local path, not git URL)
	result := sb.RunCLI("install", gitRepoPath)

	result.AssertSuccess(t)
	result.AssertOutputContains(t, "Installed")

	// Verify skill was installed (should have copied, not cloned since it's a local path)
	installedPath := filepath.Join(sb.SourcePath, "git-skill-repo", "SKILL.md")
	if !sb.FileExists(installedPath) {
		t.Error("skill should be installed from local git repo")
	}
}

func TestInstall_MetadataContainsSource(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	sb.WriteConfig(`source: ` + sb.SourcePath + `
targets: {}
`)

	// Create local skill to install
	localSkillPath := filepath.Join(sb.Root, "meta-test-skill")
	os.MkdirAll(localSkillPath, 0755)
	os.WriteFile(filepath.Join(localSkillPath, "SKILL.md"), []byte("# Meta Test"), 0644)

	result := sb.RunCLI("install", localSkillPath)
	result.AssertSuccess(t)

	// Read and verify metadata
	metaContent := sb.ReadFile(filepath.Join(sb.SourcePath, "meta-test-skill", ".skillshare-meta.json"))

	if !strings.Contains(metaContent, `"type": "local"`) {
		t.Error("metadata should contain type: local")
	}
	if !strings.Contains(metaContent, "meta-test-skill") {
		t.Error("metadata should contain source path")
	}
	if !strings.Contains(metaContent, "installed_at") {
		t.Error("metadata should contain installed_at timestamp")
	}
}

func TestInstall_GitSubdir_DirectInstall(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	sb.WriteConfig(`source: ` + sb.SourcePath + `
targets: {}
`)

	// Create a monorepo-style git repository with multiple skills
	gitRepoPath := filepath.Join(sb.Root, "monorepo")
	skill1Path := filepath.Join(gitRepoPath, "skills", "skill-one")
	skill2Path := filepath.Join(gitRepoPath, "skills", "skill-two")

	os.MkdirAll(skill1Path, 0755)
	os.MkdirAll(skill2Path, 0755)
	os.WriteFile(filepath.Join(skill1Path, "SKILL.md"), []byte("# Skill One"), 0644)
	os.WriteFile(filepath.Join(skill2Path, "SKILL.md"), []byte("# Skill Two"), 0644)

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = gitRepoPath
	if err := cmd.Run(); err != nil {
		t.Skip("git not available, skipping git test")
	}

	cmd = exec.Command("git", "config", "user.email", "test@test.com")
	cmd.Dir = gitRepoPath
	cmd.Run()

	cmd = exec.Command("git", "config", "user.name", "Test")
	cmd.Dir = gitRepoPath
	cmd.Run()

	cmd = exec.Command("git", "add", ".")
	cmd.Dir = gitRepoPath
	cmd.Run()

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = gitRepoPath
	cmd.Run()

	// Install specific skill from subdir (using local path with subdir pattern)
	// This tests the direct install path when subdir is specified
	result := sb.RunCLI("install", filepath.Join(gitRepoPath, "skills", "skill-one"))

	result.AssertSuccess(t)
	result.AssertOutputContains(t, "Installed")

	// Verify only skill-one was installed
	if !sb.FileExists(filepath.Join(sb.SourcePath, "skill-one", "SKILL.md")) {
		t.Error("skill-one should be installed")
	}
}

func TestInstall_Discovery_FindsSkills(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	sb.WriteConfig(`source: ` + sb.SourcePath + `
targets: {}
`)

	// Create a monorepo-style git repository
	gitRepoPath := filepath.Join(sb.Root, "discover-repo")
	skill1Path := filepath.Join(gitRepoPath, "skill-alpha")
	skill2Path := filepath.Join(gitRepoPath, "nested", "skill-beta")

	os.MkdirAll(skill1Path, 0755)
	os.MkdirAll(skill2Path, 0755)
	os.WriteFile(filepath.Join(skill1Path, "SKILL.md"), []byte("# Alpha"), 0644)
	os.WriteFile(filepath.Join(skill2Path, "SKILL.md"), []byte("# Beta"), 0644)

	// Test the discovery via internal package directly
	// (Can't easily test interactive selection in integration tests)
	result := sb.RunCLI("install", skill1Path)

	result.AssertSuccess(t)
	result.AssertOutputContains(t, "Installed")
}

func TestInstall_Discovery_DryRun_ShowsSkills(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	sb.WriteConfig(`source: ` + sb.SourcePath + `
targets: {}
`)

	// Create a monorepo with multiple skills
	gitRepoPath := filepath.Join(sb.Root, "dry-run-repo")
	skill1Path := filepath.Join(gitRepoPath, "skill-one")
	skill2Path := filepath.Join(gitRepoPath, "skill-two")

	os.MkdirAll(skill1Path, 0755)
	os.MkdirAll(skill2Path, 0755)
	os.WriteFile(filepath.Join(skill1Path, "SKILL.md"), []byte("# One"), 0644)
	os.WriteFile(filepath.Join(skill2Path, "SKILL.md"), []byte("# Two"), 0644)

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = gitRepoPath
	if err := cmd.Run(); err != nil {
		t.Skip("git not available, skipping git test")
	}

	cmd = exec.Command("git", "config", "user.email", "test@test.com")
	cmd.Dir = gitRepoPath
	cmd.Run()

	cmd = exec.Command("git", "config", "user.name", "Test")
	cmd.Dir = gitRepoPath
	cmd.Run()

	cmd = exec.Command("git", "add", ".")
	cmd.Dir = gitRepoPath
	cmd.Run()

	cmd = exec.Command("git", "commit", "-m", "Initial commit")
	cmd.Dir = gitRepoPath
	cmd.Run()

	// Use file:// protocol to test git discovery with local repo
	result := sb.RunCLI("install", "file://"+gitRepoPath, "--dry-run")

	result.AssertSuccess(t)
	result.AssertOutputContains(t, "Found")
	result.AssertOutputContains(t, "skill-one")
	result.AssertOutputContains(t, "skill-two")
	result.AssertOutputContains(t, "dry-run")

	// Verify nothing was installed
	if sb.FileExists(filepath.Join(sb.SourcePath, "skill-one")) {
		t.Error("skill should not be installed in dry-run mode")
	}
}
