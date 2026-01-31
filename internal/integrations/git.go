package integrations

import (
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

func IsRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	err := cmd.Run()
	return err == nil
}

// @returns the current commit hash
func GetCurrentCommit() (string, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// @returns the current branch name
func GetCurrentBranch() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(output)), nil
}

// @returns list of modified files since last commit
func GetChangedFiles() ([]string, error) {
	cmd := exec.Command("git", "diff", "--name-only", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	var files []string
	for _, line := range lines {
		if line != "" {
			files = append(files, line)
		}
	}

	return files, nil
}

// @returns the timestamp of the last commit
func GetLastCommitTime() (time.Time, error) {
	cmd := exec.Command("git", "log", "-1", "--format=%ct")
	output, err := cmd.Output()
	if err != nil {
		return time.Time{}, err
	}

	timestamp, err := strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(timestamp, 0), nil
}

// @returns how many times a file has been modified
func GetFileChurn(filePath string, depth int) (int, error) {
	args := []string{"log", "--oneline", "--follow", "--"}
	if depth > 0 {
		args = append([]string{"log", "-" + strconv.Itoa(depth), "--oneline", "--follow", "--"}, filePath)
	} else {
		args = append(args, filePath)
	}
	
	cmd := exec.Command("git", args...)
	output, err := cmd.Output()
	if err != nil {
		return 0, nil // File might not be tracked, return 0
	}

	lines := strings.Split(string(output), "\n")
	count := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			count++
		}
	}

	return count, nil
}

// @returns when a file was last modified in git
func GetFileLastModified(filePath string) (time.Time, error) {
	cmd := exec.Command("git", "log", "-1", "--format=%ct", "--", filePath)
	output, err := cmd.Output()
	if err != nil {
		return time.Time{}, err
	}

	timestamp, err := strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64)
	if err != nil {
		return time.Time{}, err
	}

	return time.Unix(timestamp, 0), nil
}

func getPostCommitHookPath() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--git-path", "hooks/post-commit")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func InstallPostCommitHook(hookContent string) error {
	hookPath, err := getPostCommitHookPath()
	if err != nil {
		return err
	}
	content := "#!/bin/sh\n" + hookContent

	return os.WriteFile(hookPath, []byte(content), 0755)
}

func GetHookContent(appName string) string {
	return `
DIZZ_BIN="$(command -v ` + appName + ` || true)"
if [ -x "$DIZZ_BIN" ]; then
    "$DIZZ_BIN" snapshot --auto >/dev/null 2>&1 || true
fi
`
}

