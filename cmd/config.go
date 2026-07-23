package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/TheShiveshNetwork/dizz/config"
	commonPkg "github.com/TheShiveshNetwork/dizz/internal/common"
	"github.com/TheShiveshNetwork/dizz/internal/store"
	"github.com/spf13/cobra"
)

var (
	configAddInstructionRule  string
	configAddInstructionScope string
	configAddGuardrailID      string
	configAddGuardrailPaths   []string
	configAddGuardrailRequire bool
	configAddGuardrailAction  string
	configAddGuardrailReason  string

	configShowJSON              bool
	configShowNameOnly          bool
	configShowDescriptionOnly   bool
	configShowInstructionsOnly  bool
	configShowGuardrailsOnly    bool
	configShowCommandsOnly      bool
	configShowSeverityScaleOnly bool
	configShowAgentDefaultsOnly bool
	configShowLinksOnly         bool
	configShowVersionOnly       bool
	configShowIncludeOnly       bool
	configShowExcludeOnly       bool
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage dizz project configuration",
}

var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show current dizz config",
	RunE:  runConfigShow,
}

var configAddInstructionCmd = &cobra.Command{
	Use:   "add-instruction --rule \"<rule>\" [--scope \"<scope>\"]",
	Short: "Add an instruction to the config",
	Long: `Add a coding instruction to the project config.

Instructions can be bare strings (applies globally) or scoped to a glob:
  dizz config add-instruction --rule "Run tests before merge"
  dizz config add-instruction --rule "No class components" --scope "*.tsx"`,
	RunE: runConfigAddInstruction,
}

var configAddGuardrailCmd = &cobra.Command{
	Use:   "add-guardrail",
	Short: "Add a guardrail to the config",
	Long: `Add an enforceable guardrail rule to the project config.

Required flags:
  --action    read_only | require_review | warn | skip | forbid
  --reason    Human-readable reason

Optional flags:
  --id        Stable identifier (e.g. "gr-generated-code")
  --paths     Glob patterns (repeatable; omit for global guardrails)
  --require-all  Fire only when ALL paths are touched together (default: any)

Examples:
  dizz config add-guardrail --action forbid --reason "no force-push"
  dizz config add-guardrail --id gr-gen --paths "generated/**" --action read_only --reason "auto-generated"
  dizz config add-guardrail --id gr-api --paths "**/*.proto" --paths "gen/client/**" --require-all --action require_review --reason "schema and client must move together"`,
	RunE: runConfigAddGuardrail,
}

var configSetDescriptionCmd = &cobra.Command{
	Use:   "set-description \"text\"",
	Short: "Set agentic description in config",
	Args:  cobra.ExactArgs(1),
	RunE:  runConfigSetDescription,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configAddInstructionCmd)
	configCmd.AddCommand(configAddGuardrailCmd)
	configCmd.AddCommand(configSetDescriptionCmd)

	configShowCmd.Flags().BoolVar(&configShowJSON, "json", false, "Output compact JSON (optimized for agents)")
	configShowCmd.Flags().BoolVarP(&configShowNameOnly, "name", "n", false, "Show only the project name")
	configShowCmd.Flags().BoolVarP(&configShowDescriptionOnly, "description", "d", false, "Show only the description")
	configShowCmd.Flags().BoolVarP(&configShowInstructionsOnly, "instructions", "i", false, "Show only instructions")
	configShowCmd.Flags().BoolVarP(&configShowGuardrailsOnly, "guardrails", "g", false, "Show only guardrails")
	configShowCmd.Flags().BoolVarP(&configShowCommandsOnly, "commands", "c", false, "Show only commands")
	configShowCmd.Flags().BoolVar(&configShowSeverityScaleOnly, "severity-scale", false, "Show only severity scale")
	configShowCmd.Flags().BoolVar(&configShowAgentDefaultsOnly, "agent-defaults", false, "Show only agent defaults")
	configShowCmd.Flags().BoolVar(&configShowLinksOnly, "links", false, "Show only links")
	configShowCmd.Flags().BoolVar(&configShowVersionOnly, "version", false, "Show only the config version")
	configShowCmd.Flags().BoolVar(&configShowIncludeOnly, "include", false, "Show only include patterns")
	configShowCmd.Flags().BoolVar(&configShowExcludeOnly, "exclude", false, "Show only exclude patterns")

	configAddInstructionCmd.Flags().StringVar(&configAddInstructionRule, "rule", "", "Rule to add")
	configAddInstructionCmd.Flags().StringVar(&configAddInstructionScope, "scope", "", "Scope for the rule (e.g., internal/**)")

	configAddGuardrailCmd.Flags().StringVar(&configAddGuardrailID, "id", "", "Stable identifier (e.g. gr-generated-code)")
	configAddGuardrailCmd.Flags().StringArrayVar(&configAddGuardrailPaths, "paths", nil, "Glob pattern (repeatable; omit for global)")
	configAddGuardrailCmd.Flags().BoolVar(&configAddGuardrailRequire, "require-all", false, "Fire only when ALL paths are touched together")
	configAddGuardrailCmd.Flags().StringVar(&configAddGuardrailAction, "action", "", "Action: read_only | require_review | warn | skip | forbid")
	configAddGuardrailCmd.Flags().StringVar(&configAddGuardrailReason, "reason", "", "Reason for the guardrail")
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	_, cfg, err := loadProjectConfig()
	if err != nil {
		return err
	}

	if configShowDescriptionOnly && !hasOtherFilters() {
		if configShowJSON {
			data, _ := json.Marshal(map[string]string{"description": cfg.Description})
			fmt.Println(string(data))
		} else {
			fmt.Println(cfg.Description)
		}
		return nil
	}

	activeFilters := getActiveFilters()
	fmt.Println(orderedJSON(cfg, activeFilters, configShowJSON))
	return nil
}

func hasOtherFilters() bool {
	return configShowNameOnly || configShowInstructionsOnly || configShowGuardrailsOnly ||
		configShowCommandsOnly || configShowSeverityScaleOnly || configShowAgentDefaultsOnly ||
		configShowLinksOnly || configShowVersionOnly || configShowIncludeOnly || configShowExcludeOnly
}

func getActiveFilters() []string {
	var filters []string
	if configShowNameOnly {
		filters = append(filters, "project_name")
	}
	if configShowDescriptionOnly {
		filters = append(filters, "description")
	}
	if configShowInstructionsOnly {
		filters = append(filters, "instructions")
	}
	if configShowGuardrailsOnly {
		filters = append(filters, "guardrails")
	}
	if configShowCommandsOnly {
		filters = append(filters, "commands")
	}
	if configShowSeverityScaleOnly {
		filters = append(filters, "severity_scale")
	}
	if configShowAgentDefaultsOnly {
		filters = append(filters, "agent_defaults")
	}
	if configShowLinksOnly {
		filters = append(filters, "links")
	}
	if configShowVersionOnly {
		filters = append(filters, "version")
	}
	if configShowIncludeOnly {
		filters = append(filters, "include")
	}
	if configShowExcludeOnly {
		filters = append(filters, "exclude")
	}
	return filters
}

func configFieldOrder() []string {
	return []string{
		"project_name", "description", "instructions", "guardrails", "commands",
		"severity_scale", "agent_defaults", "links", "version", "include", "exclude",
	}
}

func orderedJSON(cfg *config.Config, activeFilters []string, compact bool) string {
	entries := map[string]interface{}{
		"project_name":   cfg.ProjectName,
		"description":    cfg.Description,
		"instructions":   cfg.Instructions,
		"guardrails":     cfg.Guardrails,
		"commands":       cfg.Commands,
		"severity_scale": cfg.SeverityScale,
		"agent_defaults": cfg.AgentDefaults,
		"links":          cfg.Links,
		"version":        cfg.Version,
		"include":        cfg.Include,
		"exclude":        cfg.Exclude,
	}

	if len(activeFilters) > 0 {
		result := map[string]interface{}{}
		for _, f := range activeFilters {
			result[f] = entries[f]
		}
		if compact {
			data, _ := json.Marshal(result)
			return string(data)
		}
		return prettyOrderedJSON(result, activeFilters)
	}

	if compact {
		result := map[string]interface{}{}
		for _, key := range configFieldOrder() {
			result[key] = entries[key]
		}
		data, _ := json.Marshal(result)
		return string(data)
	}

	return prettyOrderedJSON(entries, configFieldOrder())
}

func prettyOrderedJSON(entries map[string]interface{}, keys []string) string {
	var b strings.Builder
	b.WriteString("{\n")
	first := true
	for _, key := range keys {
		val, ok := entries[key]
		if !ok {
			continue
		}
		if !first {
			b.WriteString(",\n")
		}
		first = false
		keyJSON, _ := json.Marshal(key)
		valJSON, _ := json.MarshalIndent(val, "", "  ")
		valIndented := strings.ReplaceAll(string(valJSON), "\n", "\n  ")
		b.WriteString("  ")
		b.Write(keyJSON)
		b.WriteString(": ")
		b.WriteString(valIndented)
	}
	b.WriteString("\n}")
	return b.String()
}

func runConfigAddInstruction(cmd *cobra.Command, args []string) error {
	if configAddInstructionRule == "" {
		return fmt.Errorf("--rule must be provided")
	}

	configStore, cfg, err := loadProjectConfig()
	if err != nil {
		return err
	}

	instruction := config.Instruction{
		Rule:  configAddInstructionRule,
		Scope: configAddInstructionScope,
	}
	cfg.Instructions = appendUniqueInstruction(cfg.Instructions, instruction)
	fmt.Println("✓ Added instruction")

	return configStore.SaveConfig(cfg)
}

func runConfigAddGuardrail(cmd *cobra.Command, args []string) error {
	if configAddGuardrailAction == "" {
		return fmt.Errorf("--action is required (read_only | require_review | warn | skip | forbid)")
	}
	if configAddGuardrailReason == "" {
		return fmt.Errorf("--reason is required")
	}

	action := config.Action(configAddGuardrailAction)
	if !action.Valid() {
		return fmt.Errorf("--action must be one of: read_only, require_review, warn, skip, forbid")
	}

	configStore, cfg, err := loadProjectConfig()
	if err != nil {
		return err
	}

	guardrail := config.Guardrail{
		ID:         configAddGuardrailID,
		Paths:      configAddGuardrailPaths,
		RequireAll: configAddGuardrailRequire,
		Action:     action,
		Reason:     configAddGuardrailReason,
	}
	cfg.Guardrails = appendUniqueGuardrail(cfg.Guardrails, guardrail)
	fmt.Println("✓ Added guardrail")

	return configStore.SaveConfig(cfg)
}

func runConfigSetDescription(cmd *cobra.Command, args []string) error {
	configStore, cfg, err := loadProjectConfig()
	if err != nil {
		return err
	}

	cfg.Description = args[0]
	if err := configStore.SaveConfig(cfg); err != nil {
		return err
	}
	fmt.Println("✓ Updated description")
	return nil
}

func loadProjectConfig() (*store.ConfigStore, *config.Config, error) {
	projectRoot, err := commonPkg.FindProjectRoot()
	if err != nil {
		return nil, nil, err
	}

	trackDir := config.TrackDirPath(projectRoot)
	configStore := store.NewConfigStore(trackDir)
	cfg, err := configStore.LoadConfig()
	if err != nil {
		return nil, nil, err
	}

	return configStore, cfg, nil
}

func appendUniqueInstruction(instructions []config.Instruction, c config.Instruction) []config.Instruction {
	for _, existing := range instructions {
		if existing.Rule == c.Rule && existing.Scope == c.Scope {
			return instructions
		}
	}
	return append(instructions, c)
}

func appendUniqueGuardrail(guardrails []config.Guardrail, g config.Guardrail) []config.Guardrail {
	for _, existing := range guardrails {
		if existing.ID == g.ID && existing.ID != "" {
			return guardrails
		}
		if existing.Action == g.Action && existing.Reason == g.Reason && strings.Join(existing.Paths, ",") == strings.Join(g.Paths, ",") {
			return guardrails
		}
	}
	return append(guardrails, g)
}

func resetAllShowFlags() {
	configShowJSON = false
	configShowNameOnly = false
	configShowDescriptionOnly = false
	configShowInstructionsOnly = false
	configShowGuardrailsOnly = false
	configShowCommandsOnly = false
	configShowSeverityScaleOnly = false
	configShowAgentDefaultsOnly = false
	configShowLinksOnly = false
	configShowVersionOnly = false
	configShowIncludeOnly = false
	configShowExcludeOnly = false
}
