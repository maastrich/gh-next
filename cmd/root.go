package cmd

import (
	"fmt"
	"os"

	"github.com/maastrich/gh-next/internal/render"
	"github.com/maastrich/gh-next/internal/state"
	"github.com/spf13/cobra"
)

var Root = &cobra.Command{
	Use:   "gh-next",
	Short: "What needs your attention on GitHub",
	Long:  "Show cached status. Run `gh next status` to refresh.",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := state.ReadSummary()
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Fprintln(os.Stderr, "No data yet. Run: gh next status")
				return nil
			}
			return err
		}
		render.Summary(s)
		return nil
	},
}

func init() {
	Root.AddCommand(statusCmd)
	Root.AddCommand(exploreCmd)
	Root.AddCommand(programCmd)
}
