package cmd

import (
	"fmt"
	"os"

	"github.com/cli/go-gh/v2/pkg/api"
	"github.com/maastrich/gh-next/internal/classify"
	"github.com/maastrich/gh-next/internal/fetch"
	"github.com/maastrich/gh-next/internal/notify"
	"github.com/maastrich/gh-next/internal/render"
	"github.com/maastrich/gh-next/internal/state"
	"github.com/spf13/cobra"
)

var staleDays int

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Fetch and classify your GitHub items",
	RunE: func(cmd *cobra.Command, args []string) error {
		restClient, err := api.DefaultRESTClient()
		if err != nil {
			return fmt.Errorf("auth: %w", err)
		}
		var userResp struct {
			Login string `json:"login"`
		}
		if err := restClient.Get("user", &userResp); err != nil {
			return fmt.Errorf("get user: %w", err)
		}
		user := userResp.Login

		fmt.Fprintf(os.Stderr, "Fetching as %s...\n", user)
		data, err := fetch.Fetch(user)
		if err != nil {
			return fmt.Errorf("fetch: %w", err)
		}

		prev := state.ReadPrevState()
		items := classify.Run(data, staleDays)
		summary := classify.Group(items)
		summary.UpdatedAt = data.FetchedAt

		notify.Diff(prev, items)
		notify.Summary(summary)

		if err := state.WriteSummary(summary); err != nil {
			return fmt.Errorf("write summary: %w", err)
		}
		if err := state.WritePrevState(items, summary.YourCount, data.FetchedAt); err != nil {
			return fmt.Errorf("write state: %w", err)
		}

		if err := render.HTML(summary); err != nil {
			fmt.Fprintf(os.Stderr, "warning: HTML report failed: %v\n", err)
		}

		render.Summary(summary)
		return nil
	},
}

func init() {
	statusCmd.Flags().IntVar(&staleDays, "stale-days", 7, "Days without activity before an item is considered stale")
}
