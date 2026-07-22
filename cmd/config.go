package cmd

import (
	"encoding/json"
	"fmt"

	commonPkg "github.com/TheShiveshNetwork/dizz/internal/common"
	"github.com/TheShiveshNetwork/dizz/config"
	"github.com/TheShiveshNetwork/dizz/internal/store"
	"github.com/spf13/cobra"
)

var (
	configAddConventionRule     string
	configAddConventionScope    string
	configAddGuardrailPath      string
	configAddGuardrailAction    string
	configAddGuardrailReason    string
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

var configAddConventionCmd = &cobra.Command{
	Use:   "add-convention --rule \"<rule>\" --scope \"<scope>\"",
	Short: "Add a convention to the config",
	RunE:  runConfigAddConvention,
}

var configAddGuardrailCmd = &cobra.Command{
	Use:   "add-guardrail --path \"<path>\" --action \"<action>\" --reason \"<reason>\"",
	Short: "Add a guardrail to the config",
	RunE:  runConfigAddGuardrail,
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
	configCmd.AddCommand(configAddConventionCmd)
	configCmd.AddCommand(configAddGuardrailCmd)
	configCmd.AddCommand(configSetDescriptionCmd)

	configAddConventionCmd.Flags().StringVar(&configAddConventionRule, "rule", "", "Rule to add")
	configAddConventionCmd.Flags().StringVar(&configAddConventionScope, "scope", "", "Scope for the rule (e.g., internal/**)")
	configAddGuardrailCmd.Flags().StringVar(&configAddGuardrailPath, "path", "", "Path to apply guardrail (e.g., internal/generated/**)")
	configAddGuardrailCmd.Flags().StringVar(&configAddGuardrailAction, "action", "", "Action to take (read_only, warn, etc.)")
	configAddGuardrailCmd.Flags().StringVar(&configAddGuardrailReason, "reason", "", "Reason for the guardrail")
}

func runConfigShow(cmd *cobra.Command, args []string) error {
	_, cfg, err := loadProjectConfig()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to format config: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func runConfigAddConvention(cmd *cobra.Command, args []string) error {
	if configAddConventionRule == "" || configAddConventionScope == "" {
		return fmt.Errorf("both --rule and --scope must be provided")
	}

	configStore, cfg, err := loadProjectConfig()
	if err != nil {
		return err
	}

	convention := config.Convention{
		Rule:  configAddConventionRule,
		Scope: configAddConventionScope,
	}
	cfg.Conventions = appendUniqueConvention(cfg.Conventions, convention)
	fmt.Println("✓ Added convention")

	return configStore.SaveConfig(cfg)
}

func runConfigAddGuardrail(cmd *cobra.Command, args []string) error {
	if configAddGuardrailPath == "" || configAddGuardrailAction == "" || configAddGuardrailReason == "" {
		return fmt.Errorf("--path, --action, and --reason must be provided")
	}

	configStore, cfg, err := loadProjectConfig()
	if err != nil {
		return err
	}

	guardrail := config.Guardrail{
		Path:   configAddGuardrailPath,
		Action: configAddGuardrailAction,
		Reason: configAddGuardrailReason,
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

func appendUniqueConvention(conventions []config.Convention, c config.Convention) []config.Convention {
	for _, existing := range conventions {
		if existing.Rule == c.Rule && existing.Scope == c.Scope {
			return conventions
		}
	}
	return append(conventions, c)
}

func appendUniqueGuardrail(guardrails []config.Guardrail, g config.Guardrail) []config.Guardrail {
	for _, existing := range guardrails {
		if existing.Path == g.Path && existing.Action == g.Action && existing.Reason == g.Reason {
			return guardrails
		}
	}
	return append(guardrails, g)
}