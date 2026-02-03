package cmd

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"

	"github.com/spf13/cobra"

	"github.com/TheShiveshNetwork/dizz/internal/config"
)

var upgradeCmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade dizz to the latest version",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runUpgrade()
	},
}

// @ignore-unused
func init() {
	rootCmd.AddCommand(upgradeCmd)
}

func runUpgrade() error {
	url := downloadURL("latest")
	fmt.Println("Downloading:", url)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("failed to download binary")
	}

	exe, err := os.Executable()
	if err != nil {
		return err
	}

	tmp := exe + ".tmp"
	out, err := os.Create(tmp)
	if err != nil {
		return err
	}

	if _, err := io.Copy(out, resp.Body); err != nil {
		return err
	}
	out.Close()

	if err := os.Chmod(tmp, 0755); err != nil {
		return err
	}

	if err := os.Rename(tmp, exe); err != nil {
		return err
	}

	fmt.Println("✅ dizz upgraded successfully")
	return nil
}

func downloadURL(version string) string {
	name := config.AppName
	os := runtime.GOOS
	arch := runtime.GOARCH
	return fmt.Sprintf(
		"https://github.com/TheShiveshNetwork/dizz/releases/latest/download/%s-%s-%s",
		name, os, arch,
	)
}
