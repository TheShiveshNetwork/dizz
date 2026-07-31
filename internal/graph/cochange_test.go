package graph

import "testing"

func TestComputeCoChanges(t *testing.T) {
	commits := []CommitInfo{
		{Hash: "aaa00000000000000000000000000000000000001", Files: []string{"/p/a.go", "/p/b.go"}},
		{Hash: "aaa00000000000000000000000000000000000002", Files: []string{"/p/a.go", "/p/b.go"}},
		{Hash: "aaa00000000000000000000000000000000000003", Files: []string{"/p/a.go", "/p/b.go", "/p/c.go"}},
		{Hash: "aaa00000000000000000000000000000000000004", Files: []string{"/p/b.go"}},
	}
	relevant := map[string]bool{"/p/a.go": true, "/p/b.go": true, "/p/c.go": true}

	out := ComputeCoChanges(commits, relevant, 0.3, 3)
	if len(out) != 1 {
		t.Fatalf("expected exactly 1 co-change pair, got %d: %+v", len(out), out)
	}
	cc := out[0]
	if cc.FileA != "/p/a.go" || cc.FileB != "/p/b.go" {
		t.Errorf("unexpected pair: %s <-> %s", cc.FileA, cc.FileB)
	}
	if cc.CoOccurrences != 3 {
		t.Errorf("co-occurrences = %d, want 3", cc.CoOccurrences)
	}
	if cc.CommitsA != 3 || cc.CommitsB != 4 {
		t.Errorf("commits = %d/%d, want 3/4", cc.CommitsA, cc.CommitsB)
	}
	// jaccard = 3 / (3+4-3) = 0.75
	if want := 0.75; cc.Jaccard != want {
		t.Errorf("jaccard = %v, want %v", cc.Jaccard, want)
	}
}

func TestComputeCoChangesFiltersByMinCommits(t *testing.T) {
	commits := []CommitInfo{
		{Hash: "aaa00000000000000000000000000000000000001", Files: []string{"/p/a.go", "/p/b.go"}},
		{Hash: "aaa00000000000000000000000000000000000002", Files: []string{"/p/a.go"}},
	}
	relevant := map[string]bool{"/p/a.go": true, "/p/b.go": true}
	out := ComputeCoChanges(commits, relevant, 0.3, 3)
	if len(out) != 0 {
		t.Errorf("expected no pairs (files below minCommits), got %d", len(out))
	}
}

func TestComputeCoChangesIgnoresIrrelevantFiles(t *testing.T) {
	commits := []CommitInfo{
		{Hash: "aaa00000000000000000000000000000000000001", Files: []string{"/p/a.go", "/other/x.go"}},
		{Hash: "aaa00000000000000000000000000000000000002", Files: []string{"/p/a.go", "/other/x.go"}},
		{Hash: "aaa00000000000000000000000000000000000003", Files: []string{"/p/a.go", "/other/x.go"}},
	}
	relevant := map[string]bool{"/p/a.go": true}
	out := ComputeCoChanges(commits, relevant, 0.3, 3)
	if len(out) != 0 {
		t.Errorf("expected no pairs when only one relevant file, got %d", len(out))
	}
}

func TestGitLogWithFilesSkipsGracefully(t *testing.T) {
	commits, err := GitLogWithFiles(".", 10)
	if err != nil {
		t.Skipf("git log unavailable: %v", err)
	}
	if len(commits) == 0 {
		t.Skip("no commits in repository")
	}
	if commits[0].Hash == "" {
		t.Errorf("first commit has empty hash")
	}
}
