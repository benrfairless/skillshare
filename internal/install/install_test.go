package install

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverSkills_RootOnly(t *testing.T) {
	// Setup: repo with SKILL.md at root only
	repoPath := t.TempDir()
	if err := os.WriteFile(filepath.Join(repoPath, "SKILL.md"), []byte("---\nname: test\n---\n# Test"), 0644); err != nil {
		t.Fatal(err)
	}

	skills := discoverSkills(repoPath, true)
	if len(skills) != 1 {
		t.Fatalf("expected 1 skill, got %d", len(skills))
	}
	if skills[0].Path != "." {
		t.Errorf("Path = %q, want %q", skills[0].Path, ".")
	}
}

func TestDiscoverSkills_RootOnly_ExcludeRoot(t *testing.T) {
	// Setup: repo with SKILL.md at root only, includeRoot=false
	repoPath := t.TempDir()
	if err := os.WriteFile(filepath.Join(repoPath, "SKILL.md"), []byte("---\nname: test\n---\n# Test"), 0644); err != nil {
		t.Fatal(err)
	}

	skills := discoverSkills(repoPath, false)
	if len(skills) != 0 {
		t.Fatalf("expected 0 skills with includeRoot=false, got %d", len(skills))
	}
}

func TestDiscoverSkills_RootAndChildren(t *testing.T) {
	// Setup: repo with SKILL.md at root AND child directories
	repoPath := t.TempDir()
	if err := os.WriteFile(filepath.Join(repoPath, "SKILL.md"), []byte("---\nname: root\n---\n# Root"), 0644); err != nil {
		t.Fatal(err)
	}

	childDir := filepath.Join(repoPath, "child-skill")
	if err := os.MkdirAll(childDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(childDir, "SKILL.md"), []byte("---\nname: child\n---\n# Child"), 0644); err != nil {
		t.Fatal(err)
	}

	skills := discoverSkills(repoPath, true)
	if len(skills) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(skills))
	}

	// Verify we have both root and child
	var hasRoot, hasChild bool
	for _, s := range skills {
		if s.Path == "." {
			hasRoot = true
		}
		if s.Path == "child-skill" && s.Name == "child-skill" {
			hasChild = true
		}
	}
	if !hasRoot {
		t.Error("missing root skill (Path='.')")
	}
	if !hasChild {
		t.Error("missing child skill (Path='child-skill')")
	}
}

func TestDiscoverSkills_ChildrenOnly(t *testing.T) {
	// Setup: orchestrator repo with no root SKILL.md, only children
	repoPath := t.TempDir()

	for _, name := range []string{"skill-a", "skill-b"} {
		dir := filepath.Join(repoPath, name)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "SKILL.md"), []byte("---\nname: "+name+"\n---\n# "+name), 0644); err != nil {
			t.Fatal(err)
		}
	}

	skills := discoverSkills(repoPath, true)
	if len(skills) != 2 {
		t.Fatalf("expected 2 skills, got %d", len(skills))
	}
	for _, s := range skills {
		if s.Path == "." {
			t.Error("should not have root skill when no root SKILL.md exists")
		}
	}
}

func TestDiscoverSkills_SkipsHiddenDirs(t *testing.T) {
	repoPath := t.TempDir()

	// Create .git directory with SKILL.md (should be skipped)
	gitDir := filepath.Join(repoPath, ".git")
	if err := os.MkdirAll(gitDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(gitDir, "SKILL.md"), []byte("---\nname: git\n---"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create .hidden directory with SKILL.md (should be skipped)
	hiddenDir := filepath.Join(repoPath, ".hidden")
	if err := os.MkdirAll(hiddenDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(hiddenDir, "SKILL.md"), []byte("---\nname: hidden\n---"), 0644); err != nil {
		t.Fatal(err)
	}

	skills := discoverSkills(repoPath, true)
	if len(skills) != 0 {
		t.Errorf("expected 0 skills (hidden dirs skipped), got %d", len(skills))
	}
}
