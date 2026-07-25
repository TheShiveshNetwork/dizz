package main

import (
	"fmt"
	"os"

	"github.com/TheShiveshNetwork/dizz/tui/app"
	tea "github.com/charmbracelet/bubbletea"
)

var version = "dev"

// @dizz-ignore-unused
func main() {
	p := tea.NewProgram(
		app.NewModel(version),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "dizzie: %v\n", err)
		os.Exit(1)
	}
}
