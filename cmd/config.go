package cmd

import (
	"fmt"
	"strings"

	"github.com/maastrich/gh-next/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage gh-next configuration",
}

var configSetCmd = &cobra.Command{
	Use:   "set <key> <value>",
	Short: "Set a config value",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key, val := args[0], args[1]
		cfg := config.Load()
		switch key {
		case "ignoreFailingJobs":
			for _, j := range cfg.IgnoreFailingJobs {
				if strings.EqualFold(j, val) {
					fmt.Printf("%q already in ignoreFailingJobs\n", val)
					return nil
				}
			}
			cfg.IgnoreFailingJobs = append(cfg.IgnoreFailingJobs, val)
		default:
			return fmt.Errorf("unknown key %q", key)
		}
		if err := cfg.Save(); err != nil {
			return err
		}
		fmt.Printf("set ignoreFailingJobs += %q\n", val)
		return nil
	},
}

var configUnsetCmd = &cobra.Command{
	Use:   "unset <key> <value>",
	Short: "Remove a value from a config list",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		key, val := args[0], args[1]
		cfg := config.Load()
		switch key {
		case "ignoreFailingJobs":
			var kept []string
			found := false
			for _, j := range cfg.IgnoreFailingJobs {
				if strings.EqualFold(j, val) {
					found = true
				} else {
					kept = append(kept, j)
				}
			}
			if !found {
				return fmt.Errorf("%q not in ignoreFailingJobs", val)
			}
			cfg.IgnoreFailingJobs = kept
		default:
			return fmt.Errorf("unknown key %q", key)
		}
		if err := cfg.Save(); err != nil {
			return err
		}
		fmt.Printf("unset ignoreFailingJobs -= %q\n", val)
		return nil
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "Show current configuration",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := config.Load()
		if len(cfg.IgnoreFailingJobs) == 0 {
			fmt.Println("ignoreFailingJobs: (none)")
		} else {
			fmt.Println("ignoreFailingJobs:")
			for _, j := range cfg.IgnoreFailingJobs {
				fmt.Printf("  - %s\n", j)
			}
		}
		return nil
	},
}

func init() {
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configUnsetCmd)
	configCmd.AddCommand(configListCmd)
}
