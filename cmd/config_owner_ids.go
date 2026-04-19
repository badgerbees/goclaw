package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/titanous/json5"

	"github.com/nextlevelbuilder/goclaw/internal/config"
)

func configOwnerIDsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "owner-ids",
		Short: "Manage configured owner user IDs",
	}
	cmd.AddCommand(configOwnerIDsSetCmd())
	return cmd
}

func configOwnerIDsSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set [owner-id...]",
		Short: "Set owner user IDs in the config file",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cfgPath := resolveConfigPath()
			if err := updateConfigOwnerIDs(cfgPath, args); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Updated owner IDs in %s\n", cfgPath)
		},
	}
}

func updateConfigOwnerIDs(cfgPath string, ownerIDs []string) error {
	cleaned := config.NormalizeOwnerIDs(ownerIDs)
	if len(cleaned) == 0 {
		return fmt.Errorf("at least one owner ID is required")
	}

	cfg, err := loadConfigFileForEdit(cfgPath)
	if err != nil {
		return err
	}
	cfg.Gateway.OwnerIDs = cleaned
	if err := config.Save(cfgPath, cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}
	return nil
}

func loadConfigFileForEdit(cfgPath string) (*config.Config, error) {
	cfg := config.Default()
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, fmt.Errorf("read config: %w", err)
	}
	if err := json5.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}
