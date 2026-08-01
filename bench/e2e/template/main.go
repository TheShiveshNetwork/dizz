package main

import (
	"flag"
	"fmt"
	"os"

	"taskforge/internal/notify"
	"taskforge/internal/plan"
	"taskforge/internal/render"
	"taskforge/internal/store"
)

func main() {
	verbose := flag.Bool("verbose", false, "enable verbose output")
	flag.Parse()
	if len(flag.Args()) == 0 {
		usage()
		os.Exit(1)
	}

	// BUG: the --verbose flag is parsed above but ignored here.
	// notify.Report should receive *verbose.
	notify.Report(false)

	tasks, err := store.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "load: %v\n", err)
		os.Exit(1)
	}

	switch flag.Arg(0) {
	case "list":
		fs := flag.NewFlagSet("list", flag.ExitOnError)
		all := fs.Bool("all", false, "include completed tasks")
		fs.Parse(flag.Args()[1:])
		render.List(tasks, *all, *verbose)
	case "plan":
		fs := flag.NewFlagSet("plan", flag.ExitOnError)
		sortBy := fs.String("sort", "priority", "sort order: priority or due")
		fs.Parse(flag.Args()[1:])
		ordered, err := plan.Sort(tasks, *sortBy)
		if err != nil {
			fmt.Fprintf(os.Stderr, "plan: %v\n", err)
			os.Exit(1)
		}
		render.Plan(ordered)
	case "add":
		fs := flag.NewFlagSet("add", flag.ExitOnError)
		due := fs.String("due", "", "due date YYYY-MM-DD")
		fs.Parse(flag.Args()[1:])
		if fs.NArg() == 0 {
			usage()
			os.Exit(1)
		}
		if err := store.Add(fs.Arg(0), *due); err != nil {
			fmt.Fprintf(os.Stderr, "add: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("added: %s\n", fs.Arg(0))
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", flag.Arg(0))
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: taskforge [--verbose] <list|plan|add> [flags]")
}
