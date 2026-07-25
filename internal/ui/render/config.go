package render

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/TheShiveshNetwork/dizz/config"
)

func ConfigShow(cfg *config.Config, filters []string, compact bool) string {
	if compact {
		return configShowCompact(cfg, filters)
	}
	return configShowPretty(cfg, filters)
}

func configShowCompact(cfg *config.Config, filters []string) string {
	entries := map[string]interface{}{
		"project_name":   cfg.ProjectName,
		"description":    cfg.Description,
		"instructions":   cfg.Instructions,
		"guardrails":     cfg.Guardrails,
		"commands":       cfg.Commands,
		"severity_scale": cfg.SeverityScale,
		"agent_defaults": cfg.AgentDefaults,
		"links":          cfg.Links,
		"include":        cfg.Include,
		"exclude":        cfg.Exclude,
	}

	keys := configFieldOrder()
	if len(filters) > 0 {
		keys = filters
	}

	result := map[string]interface{}{}
	for _, key := range keys {
		result[key] = entries[key]
	}
	data, _ := json.Marshal(result)
	return string(data)
}

func configShowPretty(cfg *config.Config, filters []string) string {
	var b strings.Builder
	keys := configFieldOrder()
	if len(filters) > 0 {
		keys = filters
	}

	for _, key := range keys {
		switch key {
		case "project_name":
			fmt.Fprintf(&b, "project_name: %s\n", cfg.ProjectName)
		case "description":
			if cfg.Description == "" {
				fmt.Fprintln(&b, "description:")
			} else {
				fmt.Fprintf(&b, "description: %s\n", cfg.Description)
			}
		case "instructions":
			writeInstructionSection(&b, cfg.Instructions)
		case "guardrails":
			writeGuardrailSection(&b, cfg.Guardrails)
		case "commands":
			writeCommandSection(&b, cfg.Commands)
		case "severity_scale":
			writeMapSection(&b, "severity_scale", cfg.SeverityScale)
		case "agent_defaults":
			writeAgentDefaultsSection(&b, cfg.AgentDefaults)
		case "links":
			writeLinksSection(&b, cfg.Links)
		case "include":
			writeSliceSection(&b, "include", cfg.Include)
		case "exclude":
			writeSliceSection(&b, "exclude", cfg.Exclude)
		}
	}
	return b.String()
}

func configFieldOrder() []string {
	return []string{
		"project_name", "description", "instructions", "guardrails", "commands",
		"severity_scale", "agent_defaults", "links",
	}
}

func writeInstructionSection(b *strings.Builder, instructions []config.Instruction) {
	if len(instructions) == 0 {
		return
	}
	fmt.Fprintln(b, "instructions:")
	for _, inst := range instructions {
		if inst.Scope != "" {
			fmt.Fprintf(b, "  - %s @ %s\n", inst.Rule, inst.Scope)
		} else {
			fmt.Fprintf(b, "  - %s\n", inst.Rule)
		}
	}
}

func writeGuardrailSection(b *strings.Builder, guardrails []config.Guardrail) {
	if len(guardrails) == 0 {
		return
	}
	fmt.Fprintln(b, "guardrails:")
	for _, g := range guardrails {
		if g.ID != "" {
			fmt.Fprintf(b, "  - [%s] %s\n", g.ID, g.Action)
		} else {
			fmt.Fprintf(b, "  - %s\n", g.Action)
		}
		if len(g.Paths) > 0 {
			fmt.Fprintf(b, "    paths: %s\n", strings.Join(g.Paths, ", "))
		}
		if g.RequireAll {
			fmt.Fprintln(b, "    require_all: true")
		}
		fmt.Fprintf(b, "    reason: %s\n", g.Reason)
	}
}

func writeCommandSection(b *strings.Builder, commands map[string]string) {
	if len(commands) == 0 {
		return
	}
	fmt.Fprintln(b, "commands:")
	keys := make([]string, 0, len(commands))
	for k := range commands {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, name := range keys {
		fmt.Fprintf(b, "  %s: %s\n", name, commands[name])
	}
}

func writeMapSection(b *strings.Builder, name string, m map[string]string) {
	if len(m) == 0 {
		return
	}
	fmt.Fprintf(b, "%s:\n", name)
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(b, "  %s: %s\n", k, m[k])
	}
}

func writeAgentDefaultsSection(b *strings.Builder, defaults config.AgentDefaults) {
	fmt.Fprintln(b, "agent_defaults:")
	if defaults.DefaultLens != "" {
		fmt.Fprintf(b, "  default_lens: %s\n", defaults.DefaultLens)
	}
	if defaults.MinSeverity > 0 {
		fmt.Fprintf(b, "  min_severity: %d\n", defaults.MinSeverity)
	}
}

func writeLinksSection(b *strings.Builder, links config.Links) {
	fmt.Fprintln(b, "links:")
	if links.Contributing != "" {
		fmt.Fprintf(b, "  contributing: %s\n", links.Contributing)
	}
	if links.Docs != "" {
		fmt.Fprintf(b, "  docs: %s\n", links.Docs)
	}
}

func writeSliceSection(b *strings.Builder, name string, items []string) {
	if len(items) == 0 {
		return
	}
	fmt.Fprintf(b, "%s:\n", name)
	for _, item := range items {
		fmt.Fprintf(b, "  - %s\n", item)
	}
}
