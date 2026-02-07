//go:build online

package integration

import (
	"testing"

	"skillshare/internal/testutil"
)

// TestInstall_RemoteGitHubSubdir_DryRun validates a network-backed install path.
// This test is excluded from default runs and only enabled with -tags online.
func TestInstall_RemoteGitHubSubdir_DryRun(t *testing.T) {
	sb := testutil.NewSandbox(t)
	defer sb.Cleanup()

	sb.WriteConfig(`source: ` + sb.SourcePath + `
targets: {}
`)

	result := sb.RunCLI("install", "runkids/skillshare/skills/skillshare", "--dry-run")

	result.AssertSuccess(t)
	result.AssertAnyOutputContains(t, "dry-run")
}
