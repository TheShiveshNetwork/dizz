package integrations

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
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

// @returns the current commit with message
func GetCurrentCommitWithMessage() (Commit, error) {
	cmd := exec.Command("git", "log", "-1", "--format=%H %ct %s")
	output, err := cmd.Output()
	if err != nil {
		return Commit{}, err
	}

	parts := strings.Split(strings.TrimSpace(string(output)), " ")
	if len(parts) < 3 {
		return Commit{}, fmt.Errorf("unexpected git output format")
	}

	hash := parts[0]
	timestamp, _ := strconv.ParseInt(parts[1], 10, 64)
	message := strings.Join(parts[2:], " ")

	return Commit{
		Hash:    hash,
		Time:    time.Unix(timestamp, 0),
		Message: message,
	}, nil
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

// @returns how many times a function/symbol has been modified
func GetFunctionChurn(filePath string, functionName string, startLine, endLine int, depth int) (int, error) {
	// Use git log with line-level tracking for the specific function range
	args := []string{"log", "--oneline", "-L"}
	if depth > 0 {
		args = []string{"log", "-" + strconv.Itoa(depth), "--oneline", "-L"}
	}

	// Format: -L startLine,endLine:filePath
	lineRange := fmt.Sprintf("%d,%d:%s", startLine, endLine, filePath)
	args = append(args, lineRange)

	cmd := exec.Command("git", args...)
	output, err := cmd.Output()
	if err != nil {
		return 0, nil // Function might not be trackable, return 0
	}

	lines := strings.Split(string(output), "\n")
	count := 0
	commitHashRegex := regexp.MustCompile(`^[a-f0-9]{7,}`)
	for _, line := range lines {
		if commitHashRegex.MatchString(strings.TrimSpace(line)) {
			count++
		}
	}

	return count, nil
}

// Commit represents a git commit with metadata
type Commit struct {
	Hash       string
	Message    string
	Time       time.Time
	ChangeSize int
}

// @returns detailed commit history for a function with change sizes
func GetFunctionCommits(filePath string, startLine, endLine int, depth int) ([]Commit, error) {
	args := []string{"log", "--format=%H %ct %s", "-L"}
	if depth > 0 {
		args = []string{"log", "-" + strconv.Itoa(depth), "--format=%H %ct %s", "-L"}
	}

	lineRange := fmt.Sprintf("%d,%d:%s", startLine, endLine, filePath)
	args = append(args, lineRange)

	cmd := exec.Command("git", args...)
	output, err := cmd.Output()
	if err != nil {
		return nil, nil // Function might not be trackable, return empty
	}

	lines := strings.Split(string(output), "\n")
	var commits []Commit
	commitHashRegex := regexp.MustCompile(`^([a-f0-9]{7,}) (\d+) (.*)$`)

	for _, line := range lines {
		matches := commitHashRegex.FindStringSubmatch(strings.TrimSpace(line))
		if len(matches) == 4 {
			hash := matches[1]
			timestamp, _ := strconv.ParseInt(matches[2], 10, 64)
			message := strings.TrimSpace(matches[3])
			time := time.Unix(timestamp, 0)

			// Get change size for this commit
			changeSize := getCommitChangeSize(hash, filePath, startLine, endLine)

			commits = append(commits, Commit{
				Hash:       hash,
				Message:    message,
				Time:       time,
				ChangeSize: changeSize,
			})
		}
	}

	return commits, nil
}

// @returns the number of lines changed in a specific commit for the function range
func getCommitChangeSize(commitHash, filePath string, startLine, endLine int) int {
	args := []string{"show", commitHash, "--format=", "--unified=0", "--", filePath}
	cmd := exec.Command("git", args...)
	output, err := cmd.Output()
	if err != nil {
		return 1 // Default to 1 if we can't determine size
	}

	// Check if this is a file creation commit (new file mode)
	if strings.Contains(string(output), "new file mode") {
		// For file creation, use a more reasonable change size
		return min(10, endLine-startLine+1)
	}

	lines := strings.Split(string(output), "\n")
	changes := 0
	inFunctionRange := false

	for _, line := range lines {
		if strings.HasPrefix(line, "@@") {
			// Parse line numbers from hunk header
			if start := parseHunkHeader(line, startLine, endLine); start != -1 {
				inFunctionRange = true
			} else {
				inFunctionRange = false
			}
		} else if inFunctionRange && (strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-")) {
			changes++
		}
	}

	// Cap the change size to avoid extremely large values from file creation
	if changes == 0 {
		return 1 // At least 1 change if this commit touched the function
	}
	return min(changes, 20) // Cap at 20 lines per commit
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// @returns whether a hunk overlaps with the function range
func parseHunkHeader(hunkLine string, funcStart, funcEnd int) int {
	// Format: @@ -start,count +start,count @@
	re := regexp.MustCompile(`@@ -\d+,?\d* \+(\d+),?(\d*) @@`)
	matches := re.FindStringSubmatch(hunkLine)
	if len(matches) < 2 {
		return -1
	}

	start, _ := strconv.Atoi(matches[1])
	count := 1
	if len(matches) > 2 && matches[2] != "" {
		count, _ = strconv.Atoi(matches[2])
	}
	end := start + count - 1

	// Check if this hunk overlaps with our function range
	if start > funcEnd || end < funcStart {
		return -1 // No overlap
	}

	return start // Only return start since end not used
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
