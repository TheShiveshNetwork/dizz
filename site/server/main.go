package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

type ReleaseInfo struct {
	Version string `json:"version"`
	Tag     string `json:"tag"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Serve static files from public directory
	fs := http.FileServer(http.Dir("../public"))
	http.Handle("/", fs)

	// API endpoint for version info
	http.HandleFunc("/api/version", func(w http.ResponseWriter, r *http.Request) {
		release, err := getLatestRelease()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(release)
	})

	// API endpoint for CLI commands
	http.HandleFunc("/api/commands", func(w http.ResponseWriter, r *http.Request) {
		commands, err := getCLICommands()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(commands)
	})

	fmt.Printf("Server starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func getLatestRelease() (*ReleaseInfo, error) {
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get latest tag: %v", err)
	}

	tag := strings.TrimSpace(string(output))
	version := strings.TrimPrefix(tag, "v")

	return &ReleaseInfo{
		Version: version,
		Tag:     tag,
	}, nil
}

type CLICommand struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Usage       string `json:"usage"`
}

func getCLICommands() ([]CLICommand, error) {
	// For now, return hardcoded commands. In a real implementation,
	// you might parse the CLI help output or maintain a static list.
	commands := []CLICommand{
		{
			Name:        "dizz",
			Description: "Main command - shows help and available commands",
			Usage:       "dizz [command] [flags]",
		},
		{
			Name:        "dizz version",
			Description: "Display the current version",
			Usage:       "dizz version",
		},
		{
			Name:        "dizz install",
			Description: "Install dependencies and setup the project",
			Usage:       "dizz install [flags]",
		},
		{
			Name:        "dizz build",
			Description: "Build the project",
			Usage:       "dizz build [flags]",
		},
		{
			Name:        "dizz run",
			Description: "Run the application",
			Usage:       "dizz run [flags]",
		},
	}

	return commands, nil
}
