package notify

import "fmt"

// Report prints a status line. Verbose enables a debug line.
func Report(verbose bool) {
	if verbose {
		fmt.Println("[debug] notifications enabled")
	}
	fmt.Println("notifications: 0 pending")
}
