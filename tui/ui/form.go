package ui

import (
	"github.com/TheShiveshNetwork/dizz/tui/render"
	"github.com/gdamore/tcell/v2"
)

type FormField struct {
	Label     string
	Value     string
	Focused   bool
	Validator func(string) string
}

type FormModel struct {
	Fields   []FormField
	Selected int
	Title    string
	SubmitFn func([]string)
	CancelFn func()
}

