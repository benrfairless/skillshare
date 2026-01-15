package integration

import (
	"os"
	"path/filepath"
	"testing"

	"skillshare/internal/testutil"
)

func TestInit_Fresh_CreatesConfigAndSource(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	// Remove config file to simulate fresh state
	os.Remove(sb.ConfigPath)

	// Run init with input to skip interactive prompts
	// Input: "2" to start fresh, "n" to skip adding other targets, "n" to skip git
	result := sb.RunCLIWithInput("2\nn\nn\n", "init")

	result.AssertSuccess(t)
	result.AssertOutputContains(t, "Initialized successfully")

	// Verify config was created
	if !sb.FileExists(sb.ConfigPath) {
		t.Error("config file should be created")
	}

	// Verify source directory was created
	if !sb.FileExists(sb.SourcePath) {
		t.Error("source directory should be created")
	}
}

func TestInit_WithSourceFlag_UsesCustomPath(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	// Remove config file
	os.Remove(sb.ConfigPath)

	customSource := filepath.Join(sb.Home, "my-skills")

	// Input: "1" to start fresh (no skills detected), "n" to skip git
	result := sb.RunCLIWithInput("1\nn\n", "init", "--source", customSource)

	result.AssertSuccess(t)
	result.AssertOutputContains(t, customSource)

	// Verify custom source was created
	if !sb.FileExists(customSource) {
		t.Error("custom source directory should be created")
	}
}

func TestInit_DryRun_DoesNotWrite(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	os.Remove(sb.ConfigPath)
	os.RemoveAll(sb.SourcePath)

	result := sb.RunCLIWithInput("n\nn\n", "init", "--dry-run")

	result.AssertSuccess(t)
	result.AssertOutputContains(t, "Dry run")

	if sb.FileExists(sb.ConfigPath) {
		t.Error("dry-run should not create config")
	}

	if sb.FileExists(sb.SourcePath) {
		t.Error("dry-run should not create source directory")
	}

	defaultSkillPath := filepath.Join(sb.SourcePath, "skillshare", "SKILL.md")
	if sb.FileExists(defaultSkillPath) {
		t.Error("dry-run should not create default skill")
	}
}

func TestInit_AlreadyInitialized_ReturnsError(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	// Create config to simulate already initialized
	sb.WriteConfig(`source: ` + sb.SourcePath + `
targets: {}
`)

	result := sb.RunCLI("init")

	result.AssertFailure(t)
	result.AssertAnyOutputContains(t, "already initialized")
}

func TestInit_CreatesDefaultSkill(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	// Remove config file
	os.Remove(sb.ConfigPath)

	// Input: "1" to start fresh, "n" to skip git
	result := sb.RunCLIWithInput("1\nn\n", "init")

	result.AssertSuccess(t)

	// Verify default skillshare skill was created
	defaultSkillPath := filepath.Join(sb.SourcePath, "skillshare", "SKILL.md")
	if !sb.FileExists(defaultSkillPath) {
		t.Error("default skillshare skill should be created")
	}
}

func TestInit_DetectsCLI_OffersImport(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	// Remove config file
	os.Remove(sb.ConfigPath)

	// Create existing claude skills directory with a skill
	claudeSkillsPath := filepath.Join(sb.Home, ".claude", "skills")
	os.MkdirAll(claudeSkillsPath, 0755)
	testSkillPath := filepath.Join(claudeSkillsPath, "test-skill")
	os.MkdirAll(testSkillPath, 0755)
	os.WriteFile(filepath.Join(testSkillPath, "SKILL.md"), []byte("# Test"), 0644)

	// Input: "2" to start fresh (not copy), "y" to add claude as target, "n" to skip git
	result := sb.RunCLIWithInput("2\ny\nn\n", "init")

	result.AssertSuccess(t)
	result.AssertOutputContains(t, "claude")
}

func TestInit_WithSkills_CopiesOnConfirm(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	// Remove config file
	os.Remove(sb.ConfigPath)

	// Create existing claude skills directory with a skill
	claudeSkillsPath := filepath.Join(sb.Home, ".claude", "skills")
	os.MkdirAll(claudeSkillsPath, 0755)
	testSkillPath := filepath.Join(claudeSkillsPath, "my-test-skill")
	os.MkdirAll(testSkillPath, 0755)
	os.WriteFile(filepath.Join(testSkillPath, "SKILL.md"), []byte("# My Test Skill"), 0644)

	// Input: "1" to copy from claude, "n" to skip git
	result := sb.RunCLIWithInput("1\nn\n", "init")

	result.AssertSuccess(t)
	result.AssertOutputContains(t, "Copy")

	// Check if skill was copied to source
	copiedSkillPath := filepath.Join(sb.SourcePath, "my-test-skill", "SKILL.md")
	if !sb.FileExists(copiedSkillPath) {
		t.Error("skill should be copied to source")
	}
}
