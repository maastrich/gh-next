package cmd

import (
	"os/exec"

	"github.com/maastrich/gh-next/internal/state"
	"github.com/spf13/cobra"
)

var openURL string

var openCmd = &cobra.Command{
	Use:    "open",
	Short:  "Open the HTML status report in the default browser",
	Hidden: true, // called internally by terminal-notifier -execute
	RunE: func(cmd *cobra.Command, args []string) error {
		target := state.HTMLPath()
		if openURL != "" {
			target = openURL
		}
		return exec.Command("/usr/bin/open", target).Run()
	},
}

func init() {
	openCmd.Flags().StringVar(&openURL, "url", "", "Open a specific URL instead of the HTML report")
	Root.AddCommand(openCmd)
}
