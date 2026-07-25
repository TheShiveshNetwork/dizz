package integrations

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

func IsRepo() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	err := cmd.Run()
	return err == nil
}

// Commit represents a git commit with metadata
type Commit struct {
	Hash       string
	Message    string
	Time       time.Time
	ChangeSize int
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

func HasUntrackedOrModifiedChanges() bool {
	cmd := exec.Command("git", "status", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) != ""
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

// @ignore-unused
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

// GitCache stores git analysis results to avoid repeated operations
type GitCache struct {
	mu               sync.RWMutex
	lastHeadHash     string
	fileLastModified map[string]time.Time
	functionChurn    map[string]int // key: "file:startLine:endLine"
}

var cache = &GitCache{
	fileLastModified: make(map[string]time.Time),
	functionChurn:    make(map[string]int),
}

// GitBatchResult contains batched git analysis results
type GitBatchResult struct {
	FileLastModified map[string]time.Time
	FunctionChurn    map[string]int
	HeadHash         string
}

// SymbolRange specifies a symbol's location for batch git analysis.
type SymbolRange struct {
	File    string
	Name    string
	Line    int
	EndLine int
}

// BatchGitAnalysis performs git operations in bulk for much better performance
func BatchGitAnalysis(symbols []SymbolRange) (*GitBatchResult, error) {
	result := &GitBatchResult{
		FileLastModified: make(map[string]time.Time),
		FunctionChurn:    make(map[string]int),
	}

	// Get current HEAD hash
	headHash, err := GetCurrentCommit()
	if err != nil {
		return nil, err
	}
	result.HeadHash = headHash

	// Check cache first
	cache.mu.RLock()
	if cache.lastHeadHash == headHash && len(cache.fileLastModified) > 0 {
		result.FileLastModified = make(map[string]time.Time)
		result.FunctionChurn = make(map[string]int)
		for k, v := range cache.fileLastModified {
			result.FileLastModified[k] = v
		}
		for k, v := range cache.functionChurn {
			result.FunctionChurn[k] = v
		}
		cache.mu.RUnlock()
		return result, nil
	}
	cache.mu.RUnlock()

	// Collect unique files and function ranges
	files := make(map[string]bool, len(symbols))
	functionRanges := make(map[string]bool, len(symbols))

	for _, s := range symbols {
		files[s.File] = true
		rangeKey := fmt.Sprintf("%s:%d:%d", s.File, s.Line, s.EndLine)
		functionRanges[rangeKey] = true
	}

	// Batch file last modified analysis
	if len(files) > 0 {
		if err := batchFileLastModified(files, result); err != nil {
			return nil, err
		}
	}

	// Batch function churn analysis
	if len(functionRanges) > 0 {
		if err := batchFunctionChurn(functionRanges, result); err != nil {
			return nil, err
		}
	}

	// Update cache
	cache.mu.Lock()
	cache.lastHeadHash = headHash
	cache.fileLastModified = result.FileLastModified
	cache.functionChurn = result.FunctionChurn
	cache.mu.Unlock()

	return result, nil
}

// batchFileLastModified gets timestamps for all files in one git command
func batchFileLastModified(files map[string]bool, result *GitBatchResult) error {
	fileList := make([]string, 0, len(files))
	for file := range files {
		fileList = append(fileList, file)
	}

	// Single git command to get all file timestamps
	args := []string{"log", "--name-only", "--format=%ct", "--"}
	args = append(args, fileList...)

	cmd := exec.Command("git", args...)
	output, err := cmd.Output()
	if err != nil {
		return err
	}

	// Resolve the git working tree root so we can convert relative paths from
	// --name-only output back to absolute paths (the caller uses absolute keys).
	gitRootCmd := exec.Command("git", "rev-parse", "--show-toplevel")
	gitRootBytes, gitRootErr := gitRootCmd.Output()
	gitRoot := ""
	if gitRootErr == nil {
		gitRoot = strings.TrimSpace(string(gitRootBytes))
	}

	lines := strings.Split(string(output), "\n")
	var currentFile string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check if line is a timestamp (digits only)
		if timestamp, err := strconv.ParseInt(line, 10, 64); err == nil {
			if currentFile != "" {
				result.FileLastModified[currentFile] = time.Unix(timestamp, 0)
				currentFile = ""
			}
		} else {
			// This is a filename (relative to git root).
			// Convert to absolute for consistent lookup.
			if gitRoot != "" && !filepath.IsAbs(line) {
				currentFile = filepath.Join(gitRoot, line)
			} else {
				currentFile = line
			}
		}
	}

	// For any files not found in git log, fall back to individual calls
	for file := range files {
		if _, exists := result.FileLastModified[file]; !exists {
			if lastMod, err := GetFileLastModified(file); err == nil {
				result.FileLastModified[file] = lastMod
			}
		}
	}

	return nil
}

// batchFunctionChurn gets churn for all function ranges using parallel git calls.
func batchFunctionChurn(functionRanges map[string]bool, result *GitBatchResult) error {
	ranges := make([]string, 0, len(functionRanges))
	for rangeKey := range functionRanges {
		ranges = append(ranges, rangeKey)
	}
	if len(ranges) == 0 {
		return nil
	}

	// Process in parallel using a worker pool.
	// git log -L calls are I/O-bound, so parallelism helps significantly.
	workers := 8
	if len(ranges) < workers {
		workers = len(ranges)
	}
	jobs := make(chan string, len(ranges))
	var wg sync.WaitGroup
	var mu sync.Mutex

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for rangeKey := range jobs {
				parts := strings.Split(rangeKey, ":")
				if len(parts) != 3 {
					continue
				}
				file := parts[0]
				startLine, _ := strconv.Atoi(parts[1])
				endLine, _ := strconv.Atoi(parts[2])

				if churn, err := GetFunctionChurn(file, "", startLine, endLine, 50); err == nil {
					mu.Lock()
					result.FunctionChurn[rangeKey] = churn
					mu.Unlock()
				}
			}
		}()
	}

	for _, rangeKey := range ranges {
		jobs <- rangeKey
	}
	close(jobs)
	wg.Wait()
	return nil
}

// InstallLocalPostCommitHook writes the dizz post-commit hook to the given path
// inside .dizz/hooks/ and ensures it is executable.
func InstallLocalPostCommitHook(hookPath, hookContent string) error {
	if err := os.MkdirAll(filepath.Dir(hookPath), 0755); err != nil {
		return fmt.Errorf("create hooks dir: %w", err)
	}
	if err := os.WriteFile(hookPath, []byte(hookContent), 0755); err != nil {
		return fmt.Errorf("write hook: %w", err)
	}
	return nil
}

// SetLocalHooksPath runs "git config core.hooksPath .dizz/hooks" for the
// current repository so future commits use the tracked hooks directly.
func SetLocalHooksPath() error {
	hooksPath := "." + string(filepath.Separator) + ".dizz" + string(filepath.Separator) + "hooks"
	cmd := exec.Command("git", "config", "core.hooksPath", hooksPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("set local hooksPath: %s: %w", strings.TrimSpace(string(out)), err)
	}
	return nil
}

// EnsureLocalHooksConfigured checks if .dizz/hooks/post-commit exists and
// configures the local core.hooksPath accordingly. It is a no-op if the config
// is already correct, making it safe to call on every command.
// Returns true if hooks were (re)configured, false if already correct or not applicable.
func EnsureLocalHooksConfigured(trackDir string) bool {
	hooksDir := filepath.Join(trackDir, "hooks")
	hookPath := filepath.Join(hooksDir, "post-commit")

	if _, err := os.Stat(hookPath); os.IsNotExist(err) {
		return false
	}

	// Check current hooksPath
	cmd := exec.Command("git", "config", "core.hooksPath")
	out, err := cmd.Output()
	if err == nil {
		current := strings.TrimSpace(string(out))
		if current == "."+string(filepath.Separator)+".dizz"+string(filepath.Separator)+"hooks" ||
			current == ".dizz/hooks" ||
			current == ".dizz\\hooks" {
			return false
		}
	}

	_ = SetLocalHooksPath()
	return true
}

// @dizz-ignore-unused
// InvalidateCache clears the git cache (useful for testing or force refresh)
func InvalidateCache() {
	cache.mu.Lock()
	defer cache.mu.Unlock()
	cache.lastHeadHash = ""
	cache.fileLastModified = make(map[string]time.Time)
	cache.functionChurn = make(map[string]int)
}
