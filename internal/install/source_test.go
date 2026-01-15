package install

import (
	"testing"
)

func TestParseSource_LocalPath(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantType SourceType
		wantName string
	}{
		{
			name:     "absolute path",
			input:    "/path/to/my-skill",
			wantType: SourceTypeLocalPath,
			wantName: "my-skill",
		},
		{
			name:     "tilde path",
			input:    "~/skills/my-skill",
			wantType: SourceTypeLocalPath,
			wantName: "my-skill",
		},
		{
			name:     "relative path with dot",
			input:    "./local-skill",
			wantType: SourceTypeLocalPath,
			wantName: "local-skill",
		},
		{
			name:     "parent directory path",
			input:    "../other-skill",
			wantType: SourceTypeLocalPath,
			wantName: "other-skill",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source, err := ParseSource(tt.input)
			if err != nil {
				t.Fatalf("ParseSource() error = %v", err)
			}
			if source.Type != tt.wantType {
				t.Errorf("Type = %v, want %v", source.Type, tt.wantType)
			}
			if source.Name != tt.wantName {
				t.Errorf("Name = %v, want %v", source.Name, tt.wantName)
			}
		})
	}
}

func TestParseSource_GitHubShorthand(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantCloneURL string
		wantSubdir   string
		wantName     string
	}{
		{
			name:         "basic github shorthand",
			input:        "github.com/user/repo",
			wantCloneURL: "https://github.com/user/repo.git",
			wantSubdir:   "",
			wantName:     "repo",
		},
		{
			name:         "github shorthand with .git",
			input:        "github.com/user/repo.git",
			wantCloneURL: "https://github.com/user/repo.git",
			wantSubdir:   "",
			wantName:     "repo",
		},
		{
			name:         "github with subdirectory",
			input:        "github.com/user/repo/path/to/skill",
			wantCloneURL: "https://github.com/user/repo.git",
			wantSubdir:   "path/to/skill",
			wantName:     "skill",
		},
		{
			name:         "github with https prefix",
			input:        "https://github.com/user/repo",
			wantCloneURL: "https://github.com/user/repo.git",
			wantSubdir:   "",
			wantName:     "repo",
		},
		{
			name:         "github https with .git",
			input:        "https://github.com/user/repo.git",
			wantCloneURL: "https://github.com/user/repo.git",
			wantSubdir:   "",
			wantName:     "repo",
		},
		{
			name:         "github web URL with tree/main",
			input:        "https://github.com/user/repo/tree/main/path/to/skill",
			wantCloneURL: "https://github.com/user/repo.git",
			wantSubdir:   "path/to/skill",
			wantName:     "skill",
		},
		{
			name:         "github web URL with tree/master",
			input:        "github.com/user/repo/tree/master/skills/my-skill",
			wantCloneURL: "https://github.com/user/repo.git",
			wantSubdir:   "skills/my-skill",
			wantName:     "my-skill",
		},
		{
			name:         "github web URL with blob (file view)",
			input:        "https://github.com/user/repo/blob/main/path/to/skill",
			wantCloneURL: "https://github.com/user/repo.git",
			wantSubdir:   "path/to/skill",
			wantName:     "skill",
		},
		{
			name:         "github web URL tree/branch only (no subdir)",
			input:        "https://github.com/user/repo/tree/main",
			wantCloneURL: "https://github.com/user/repo.git",
			wantSubdir:   "",
			wantName:     "repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source, err := ParseSource(tt.input)
			if err != nil {
				t.Fatalf("ParseSource() error = %v", err)
			}
			if source.Type != SourceTypeGitHub {
				t.Errorf("Type = %v, want %v", source.Type, SourceTypeGitHub)
			}
			if source.CloneURL != tt.wantCloneURL {
				t.Errorf("CloneURL = %v, want %v", source.CloneURL, tt.wantCloneURL)
			}
			if source.Subdir != tt.wantSubdir {
				t.Errorf("Subdir = %v, want %v", source.Subdir, tt.wantSubdir)
			}
			if source.Name != tt.wantName {
				t.Errorf("Name = %v, want %v", source.Name, tt.wantName)
			}
		})
	}
}

func TestParseSource_GitSSH(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantCloneURL string
		wantName     string
	}{
		{
			name:         "github ssh",
			input:        "git@github.com:user/repo.git",
			wantCloneURL: "git@github.com:user/repo.git",
			wantName:     "repo",
		},
		{
			name:         "gitlab ssh",
			input:        "git@gitlab.com:user/repo.git",
			wantCloneURL: "git@gitlab.com:user/repo.git",
			wantName:     "repo",
		},
		{
			name:         "ssh without .git",
			input:        "git@github.com:user/my-skill",
			wantCloneURL: "git@github.com:user/my-skill.git",
			wantName:     "my-skill",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source, err := ParseSource(tt.input)
			if err != nil {
				t.Fatalf("ParseSource() error = %v", err)
			}
			if source.Type != SourceTypeGitSSH {
				t.Errorf("Type = %v, want %v", source.Type, SourceTypeGitSSH)
			}
			if source.CloneURL != tt.wantCloneURL {
				t.Errorf("CloneURL = %v, want %v", source.CloneURL, tt.wantCloneURL)
			}
			if source.Name != tt.wantName {
				t.Errorf("Name = %v, want %v", source.Name, tt.wantName)
			}
		})
	}
}

func TestParseSource_GitHTTPS(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantCloneURL string
		wantName     string
	}{
		{
			name:         "gitlab https",
			input:        "https://gitlab.com/user/repo",
			wantCloneURL: "https://gitlab.com/user/repo.git",
			wantName:     "repo",
		},
		{
			name:         "bitbucket https",
			input:        "https://bitbucket.org/user/repo.git",
			wantCloneURL: "https://bitbucket.org/user/repo.git",
			wantName:     "repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			source, err := ParseSource(tt.input)
			if err != nil {
				t.Fatalf("ParseSource() error = %v", err)
			}
			if source.Type != SourceTypeGitHTTPS {
				t.Errorf("Type = %v, want %v", source.Type, SourceTypeGitHTTPS)
			}
			if source.CloneURL != tt.wantCloneURL {
				t.Errorf("CloneURL = %v, want %v", source.CloneURL, tt.wantCloneURL)
			}
			if source.Name != tt.wantName {
				t.Errorf("Name = %v, want %v", source.Name, tt.wantName)
			}
		})
	}
}

func TestParseSource_Errors(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "empty string",
			input: "",
		},
		{
			name:  "whitespace only",
			input: "   ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseSource(tt.input)
			if err == nil {
				t.Error("ParseSource() should return error")
			}
		})
	}
}

func TestSource_HasSubdir(t *testing.T) {
	source := &Source{Subdir: "path/to/skill"}
	if !source.HasSubdir() {
		t.Error("HasSubdir() should return true")
	}

	source = &Source{Subdir: ""}
	if source.HasSubdir() {
		t.Error("HasSubdir() should return false")
	}
}

func TestSource_IsGit(t *testing.T) {
	tests := []struct {
		sourceType SourceType
		wantIsGit  bool
	}{
		{SourceTypeGitHub, true},
		{SourceTypeGitHTTPS, true},
		{SourceTypeGitSSH, true},
		{SourceTypeLocalPath, false},
		{SourceTypeUnknown, false},
	}

	for _, tt := range tests {
		source := &Source{Type: tt.sourceType}
		if source.IsGit() != tt.wantIsGit {
			t.Errorf("IsGit() for %v = %v, want %v", tt.sourceType, source.IsGit(), tt.wantIsGit)
		}
	}
}

func TestSource_MetaType(t *testing.T) {
	tests := []struct {
		source   *Source
		wantType string
	}{
		{
			source:   &Source{Type: SourceTypeGitHub},
			wantType: "github",
		},
		{
			source:   &Source{Type: SourceTypeGitHub, Subdir: "path"},
			wantType: "github-subdir",
		},
		{
			source:   &Source{Type: SourceTypeLocalPath},
			wantType: "local",
		},
	}

	for _, tt := range tests {
		if tt.source.MetaType() != tt.wantType {
			t.Errorf("MetaType() = %v, want %v", tt.source.MetaType(), tt.wantType)
		}
	}
}

func TestParseSource_GeminiCLIMonorepo(t *testing.T) {
	// Real-world test case from the plan
	input := "github.com/google-gemini/gemini-cli/packages/core/src/skills/builtin/skill-creator"

	source, err := ParseSource(input)
	if err != nil {
		t.Fatalf("ParseSource() error = %v", err)
	}

	if source.Type != SourceTypeGitHub {
		t.Errorf("Type = %v, want %v", source.Type, SourceTypeGitHub)
	}
	if source.CloneURL != "https://github.com/google-gemini/gemini-cli.git" {
		t.Errorf("CloneURL = %v, want https://github.com/google-gemini/gemini-cli.git", source.CloneURL)
	}
	if source.Subdir != "packages/core/src/skills/builtin/skill-creator" {
		t.Errorf("Subdir = %v, want packages/core/src/skills/builtin/skill-creator", source.Subdir)
	}
	if source.Name != "skill-creator" {
		t.Errorf("Name = %v, want skill-creator", source.Name)
	}
}
