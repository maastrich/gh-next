package cmd

import (
	"fmt"
	"sort"

	"github.com/maastrich/gh-next/internal/index"
	"github.com/spf13/cobra"
)

var exploreCmd = &cobra.Command{
	Use:   "explore <path>",
	Short: "Index all git repositories under a directory",
	Long: `Walk <path> recursively, find every git repository with a GitHub origin,
and write a lookup index to ~/.gh-next/gh-index.json.

The index maps each GitHub repository URL to one or more local filesystem paths
(useful when you have multiple clones or forks of the same repo).`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		root := args[0]
		fmt.Printf("Indexing %s...\n", root)

		idx, err := index.Build(root)
		if err != nil {
			return err
		}

		if err := index.Write(idx); err != nil {
			return fmt.Errorf("write index: %w", err)
		}

		sort.Slice(idx.Entries, func(i, j int) bool {
			return idx.Entries[i].GithubURL < idx.Entries[j].GithubURL
		})

		fmt.Printf("Found %d repositories:\n\n", len(idx.Entries))
		for _, e := range idx.Entries {
			fmt.Printf("  %s\n", e.GithubURL)
			for _, p := range e.Paths {
				fmt.Printf("    → %s\n", p)
			}
		}
		fmt.Printf("\nIndex written to ~/.gh-next/gh-index.json\n")
		return nil
	},
}
