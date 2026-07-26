package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

const tuiBinaryName = "dizzie"

var uiCmd = &cobra.Command{
	Use:   "ui",
	Short: "Launch or install the dizz Terminal UI",
	Long: `Opens the dizz interactive terminal interface (dizzie).

If dizzie is not installed, it will be downloaded and installed
automatically before launching.`,
	RunE: runUI,
}

// @dizz-ignore-unused
func init() {
	rootCmd.AddCommand(uiCmd)
}

// @dizz-ignore-unused
func runUI(cmd *cobra.Command, args []string) error {
	path, err := exec.LookPath(tuiBinaryName)
	if err == nil {
		return execTUI(path)
	}

	installPaths := []string{
		filepath.Join(homeDir(), ".dizz", "bin", tuiBinaryName),
		filepath.Join(homeDir(), ".local", "bin", tuiBinaryName),
	}
	for _, p := range installPaths {
		if _, err := os.Stat(p); err == nil {
			return execTUI(p)
		}
	}

	fmt.Println("dizzie not found. Installing...")
	if err := installTUI(); err != nil {
		return fmt.Errorf("failed to install dizzie: %w", err)
	}

	path, err = exec.LookPath(tuiBinaryName)
	if err != nil {
		for _, p := range installPaths {
			if _, err := os.Stat(p); err == nil {
				return execTUI(p)
			}
		}
		return fmt.Errorf("dizzie installed but not found on PATH; look in ~/.dizz/bin/ or ~/.local/bin/")
	}

	return execTUI(path)
}

func execTUI(path string) error {
	c := exec.Command(path)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	if err := c.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return err
	}
	return nil
}

func installTUI() error {
	targetDir := filepath.Join(homeDir(), ".dizz", "bin")
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("cannot create install directory: %w", err)
	}

	targetPath := filepath.Join(targetDir, tuiBinaryName)
	if runtime.GOOS == "windows" {
		targetPath += ".exe"
	}

	url := fmt.Sprintf(
		"https://github.com/TheShiveshNetwork/dizz/releases/latest/download/%s-%s-%s",
		tuiBinaryName, runtime.GOOS, runtime.GOARCH,
	)

	fmt.Printf("  Downloading from %s\n", url)

	resp, err := http.Get(url)
	if err == nil && resp.StatusCode == 200 {
		defer resp.Body.Close()
		f, err := os.Create(targetPath)
		if err != nil {
			return fmt.Errorf("cannot create binary: %w", err)
		}
		defer f.Close()

		if _, err := io.Copy(f, resp.Body); err != nil {
			return fmt.Errorf("download failed: %w", err)
		}
		f.Close()

		if err := os.Chmod(targetPath, 0755); err != nil {
			return fmt.Errorf("cannot make binary executable: %w", err)
		}

		fmt.Printf("  Installed to %s\n", targetPath)
		return nil
	}
	if resp != nil {
		resp.Body.Close()
	}

	fmt.Println("  Binary download unavailable, trying go install...")
	installCmd := exec.Command("go", "install", "github.com/TheShiveshNetwork/dizz/tui@latest")
	installCmd.Stdout = os.Stdout
	installCmd.Stderr = os.Stderr
	if err := installCmd.Run(); err != nil {
		return fmt.Errorf("go install failed (try building from source): %w", err)
	}

	fmt.Println("  Installed via go install")
	return nil
}

func homeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return home
}
