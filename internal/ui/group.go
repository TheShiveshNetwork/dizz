package ui

import (
	"sort"

	"github.com/TheShiveshNetwork/dizz/internal/state"
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
