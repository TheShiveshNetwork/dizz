package main

import (
	"fmt"
	"syscall/js"
)

// @ignore-unused
func main() {
	// Make functions available to JavaScript
	js.Global().Set("getWasmVersion", js.FuncOf(getWasmVersion))
	js.Global().Set("getWasmCommands", js.FuncOf(getWasmCommands))
	js.Global().Set("runWasmCommand", js.FuncOf(runWasmCommand))

	// Keep the program running
	select {}
}

// @ignore-unused
func getWasmVersion(this js.Value, args []js.Value) interface{} {
	// This would typically read from the embedded version info
	// For now, return a hardcoded version
	return map[string]interface{}{
		"version": "1.0.0",
		"source":  "wasm",
	}
}

// @ignore-unused
func getWasmCommands(this js.Value, args []js.Value) interface{} {
	// Return available CLI commands that can be executed
	commands := []map[string]interface{}{
		{
			"name":        "version",
			"description": "Display version information",
			"category":    "info",
		},
		{
			"name":        "help",
			"description": "Show help information",
			"category":    "info",
		},
		{
			"name":        "build",
			"description": "Build the project",
			"category":    "build",
		},
		{
			"name":        "install",
			"description": "Install dependencies",
			"category":    "setup",
		},
		{
			"name":        "run",
			"description": "Run the application",
			"category":    "execution",
		},
	}

	// Convert to JavaScript array
	result := js.Global().Get("Array").New(len(commands))
	for i, cmd := range commands {
		cmdObj := js.Global().Get("Object").New()
		for key, value := range cmd {
			cmdObj.Set(key, value)
		}
		result.SetIndex(i, cmdObj)
	}

	return result
}

// @ignore-unused
func runWasmCommand(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return map[string]interface{}{
			"success": false,
			"error":   "No command provided",
		}
	}

	command := args[0].String()

	// Simulate command execution (in a real scenario, this would interface with the actual CLI)
	output := fmt.Sprintf("Simulated execution of '%s' command", command)

	return map[string]interface{}{
		"success": true,
		"command": command,
		"output":  output,
	}
}
