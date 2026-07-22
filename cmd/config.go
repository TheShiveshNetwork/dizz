package cmd

import (
	"encoding/json"
	"fmt"

	commonPkg "github.com/TheShiveshNetwork/dizz/internal/common"
	"github.com/TheShiveshNetwork/dizz/internal/config"
	"github.com/TheShiveshNetwork/dizz/internal/store"
	"github.com/spf13/cobra"
)

var (
	configAddRule        string
	configAddStandard    string
	configAddInstruction string
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

var configAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add agentic config entries (rule/standard/instruction)",
	RunE:  runConfigAdd,
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
	configCmd.AddCommand(configAddCmd)
	configCmd.AddCommand(configSetDescriptionCmd)

	configAddCmd.Flags().StringVar(&configAddRule, "rule", "", "Rule to add")
	configAddCmd.Flags().StringVar(&configAddStandard, "standard", "", "Standard to add")
	configAddCmd.Flags().StringVar(&configAddInstruction, "instruction", "", "Instruction to add")
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

func runConfigAdd(cmd *cobra.Command, args []string) error {
	entries := 0
	if configAddRule != "" {
		entries++
	}
	if configAddStandard != "" {
		entries++
	}
	if configAddInstruction != "" {
		entries++
	}
	if entries != 1 {
		return fmt.Errorf("exactly one of --rule, --standard, or --instruction must be provided")
	}

	configStore, cfg, err := loadProjectConfig()
	if err != nil {
		return err
	}

	switch {
	case configAddRule != "":
		appendUnique(&cfg.Agentic.Rules, configAddRule)
		fmt.Println("✓ Added rule")
	case configAddStandard != "":
		appendUnique(&cfg.Agentic.Standards, configAddStandard)
		fmt.Println("✓ Added standard")
	case configAddInstruction != "":
		appendUnique(&cfg.Agentic.Instructions, configAddInstruction)
		fmt.Println("✓ Added instruction")
	}

	return configStore.SaveConfig(cfg)
}

func runConfigSetDescription(cmd *cobra.Command, args []string) error {
	configStore, cfg, err := loadProjectConfig()
	if err != nil {
		return err
	}

	cfg.Agentic.Description = args[0]
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

func appendUnique(values *[]string, value string) {
	for _, existing := range *values {
		if existing == value {
			return
		}
	}
	*values = append(*values, value)
}
