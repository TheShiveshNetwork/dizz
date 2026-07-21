package skill

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDetectAgentDirs_None(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	dirs := DetectAgentDirs()
	if len(dirs) != 0 {
		t.Fatalf("expected 0 dirs, got %d", len(dirs))
	}
}

func TestDetectAgentDirs_One(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	skillDir := filepath.Join(home, ".claude", "skills")
	if err := os.MkdirAll(skillDir, 0755); err != nil {
		t.Fatal(err)
	}

	dirs := DetectAgentDirs()
	found := false
	for _, d := range dirs {
		if d.Path == skillDir {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected to find %s in dirs", skillDir)
	}
}

func TestDetectAgentDirs_Multiple(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	paths := []string{
		filepath.Join(home, ".claude", "skills"),
		filepath.Join(home, ".cursor", "skills"),
		filepath.Join(home, ".config", "opencode", "skills"),
	}
	for _, p := range paths {
		if err := os.MkdirAll(p, 0755); err != nil {
			t.Fatal(err)
		}
	}

	dirs := DetectAgentDirs()
	if len(dirs) < 3 {
		t.Fatalf("expected at least 3 dirs, got %d", len(dirs))
	}
}

func TestInstallToDir_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	content := []byte("# Test Skill\n\nInstructions here.\n")

	if err := InstallToDir(content, "TestAgent", dir); err != nil {
		t.Fatal(err)
	}

	skillPath := filepath.Join(dir, "dizz", "SKILL.md")
	if _, err := os.Stat(skillPath); os.IsNotExist(err) {
		t.Fatalf("SKILL.md not created at %s", skillPath)
	}

	data, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(content) {
		t.Fatalf("content mismatch:\ngot:  %q\nwant: %q", string(data), string(content))
	}
}

func TestInstallToDir_Overwrites(t *testing.T) {
	dir := t.TempDir()

	oldContent := []byte("# Old skill")
	if err := InstallToDir(oldContent, "TestAgent", dir); err != nil {
		t.Fatal(err)
	}

	newContent := []byte("# New skill")
	if err := InstallToDir(newContent, "TestAgent", dir); err != nil {
		t.Fatal(err)
	}

	skillPath := filepath.Join(dir, "dizz", "SKILL.md")
	data, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(newContent) {
		t.Fatalf("expected overwritten content, got %q", string(data))
	}
}

func TestInstallToAll_NoAgentDirs(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	results := InstallToAll([]byte("test"))
	if len(results) == 0 {
		t.Fatal("expected at least 1 result")
	}

	allSkipped := true
	for _, r := range results {
		if !r.Skipped {
			allSkipped = false
		}
	}
	if !allSkipped {
		t.Fatal("expected all results to be skipped when no agent dirs exist")
	}
}

func TestInstallToAll_WithDirs(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	claudeDir := filepath.Join(home, ".claude", "skills")
	cursorDir := filepath.Join(home, ".cursor", "skills")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(cursorDir, 0755); err != nil {
		t.Fatal(err)
	}

	results := InstallToAll([]byte("# dizz skill"))
	successCount := 0
	for _, r := range results {
		if r.Err == nil && !r.Skipped {
			successCount++
		}
	}
	if successCount < 2 {
		t.Fatalf("expected at least 2 successful installs, got %d", successCount)
	}

	for _, p := range []string{claudeDir, cursorDir} {
		skillPath := filepath.Join(p, "dizz", "SKILL.md")
		if _, err := os.Stat(skillPath); os.IsNotExist(err) {
			t.Fatalf("SKILL.md not found at %s", skillPath)
		}
	}
}

func TestInstallToAll_Deduplicates(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	singlePath := filepath.Join(home, ".agents", "skills")
	if err := os.MkdirAll(singlePath, 0755); err != nil {
		t.Fatal(err)
	}

	results := InstallToAll([]byte("# dedup test"))
	count := 0
	for _, r := range results {
		if r.Err == nil && !r.Skipped {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("expected 1 install after dedup (shared .agents/skills), got %d", count)
	}
}

func TestFormatInstallResults(t *testing.T) {
	results := []InstallResult{
		{Agent: "Claude Code", Path: "/home/user/.claude/skills/dizz/SKILL.md"},
		{Agent: "Cursor", Path: "/home/user/.cursor/skills/dizz/SKILL.md", Err: os.ErrPermission},
		{Agent: "system", Skipped: true, Err: nil},
	}

	output := FormatInstallResults(results)
	if !strings.Contains(output, "Claude Code") {
		t.Fatal("expected output to contain 'Claude Code'")
	}
	if !strings.Contains(output, "Cursor") {
		t.Fatal("expected output to contain 'Cursor'")
	}
}

func TestGlobalSkillDirs(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	dirs := GlobalSkillDirs()
	// GlobalSkillDirs returns candidate paths regardless of existence
	if len(dirs) == 0 {
		t.Fatal("expected at least some candidate dirs")
	}

	foundClaude := false
	foundAgents := false
	for _, d := range dirs {
		if strings.HasSuffix(d, filepath.Join(".claude", "skills", "dizz")) {
			foundClaude = true
		}
		if strings.HasSuffix(d, filepath.Join(".agents", "skills", "dizz")) {
			foundAgents = true
		}
	}
	if !foundClaude {
		t.Fatal("expected claude dizz skill dir candidate")
	}
	if !foundAgents {
		t.Fatal("expected .agents dizz skill dir candidate")
	}
}

func TestPlatform(t *testing.T) {
	p := Platform()
	if p == "" {
		t.Fatal("expected non-empty platform string")
	}
	if !strings.Contains(p, "/") {
		t.Fatalf("expected platform to contain '/', got %q", p)
	}
}

func TestCleanDirName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"/home/user/.claude/skills/dizz/SKILL.md", "dizz"},
		{"/tmp/.agents/skills/dizz/SKILL.md", "dizz"},
	}
	for _, tt := range tests {
		got := CleanDirName(tt.input)
		if got != tt.want {
			t.Errorf("CleanDirName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
