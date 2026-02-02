package render

import (
	"fmt"
	"sort"

	"github.com/TheShiveshNetwork/dizz/internal/state"
	"github.com/TheShiveshNetwork/dizz/internal/ui"
)

type FileGroup struct {
	File    string
	Symbols []state.Symbol
	Count   int
}

func GroupSymbolsByFile(symbols []state.Symbol) []FileGroup {
	fileMap := make(map[string][]state.Symbol)

	for _, sym := range symbols {
		fileMap[sym.File] = append(fileMap[sym.File], sym)
	}

	groups := make([]FileGroup, 0, len(fileMap))
	for file, syms := range fileMap {
		groups = append(groups, FileGroup{
			File:    file,
			Symbols: syms,
			Count:   len(syms),
		})
	}

	// Sort by most items → least
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Count > groups[j].Count
	})

	return groups
}

type RenderArgs struct {
	Title    string
	Subtitle string

	Symbols []state.Symbol
	ShowAll bool

	MaxPerFile int
	ShowChurn  bool
}

func RenderSymbolGroup(args RenderArgs) {
	if len(args.Symbols) == 0 {
		return
	}

	groupState := args.Symbols[0].State

	// ---- Header ----
	title := args.Title
	fmt.Println(
		ColorByState(title, groupState),
		ui.Muted(fmt.Sprintf("(%s)", args.Subtitle)),
	)

	groups := GroupSymbolsByFile(args.Symbols)

	for _, group := range groups {
		// File line
		fmt.Printf(
			"  %s %s\n",
			ColorByState(group.File, groupState),
			ui.Muted(fmt.Sprintf("(%d items)", group.Count)),
		)

		limit := args.MaxPerFile
		if args.ShowAll {
			limit = len(group.Symbols)
		}

		for i, sym := range group.Symbols {
			if i >= limit {
				fmt.Printf(
					ui.Muted("     ... and %d more\n"),
					len(group.Symbols)-limit,
				)
				break
			}
			fmt.Printf(
				"     • %s %d:%d",
				ColorByState(sym.Name, sym.State),
				sym.Line,
				sym.Column,
			)

			if args.ShowChurn {
				fmt.Printf(
					" %s",
					ui.Error(fmt.Sprintf("(churn: %d)", sym.ChurnCount)),
				)
			}

			fmt.Println()
		}

		fmt.Println()
	}
}

func ColorByState(text string, s state.SymbolState) string {
	switch s {
	case state.Planned:
		return ui.Warning(text)
	case state.Unstable:
		return ui.Error(text)
	case state.Unused:
		return ui.Info(text)
	case state.Active:
		return ui.Success(text)
	case state.Abandoned:
		return ui.Muted(text)
	default:
		return text
	}
}
