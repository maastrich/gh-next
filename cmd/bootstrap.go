package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/maastrich/gh-next/internal/state"
	"github.com/spf13/cobra"
)

var bootstrapCmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "Ensure all dependencies and helper files are set up",
	Long: `Checks and installs required dependencies, then writes helper files
needed for notifications to work correctly.

Run this once after installation, or again if notifications stop working.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ok := func(msg string) { fmt.Printf("✓ %s\n", msg) }
		warn := func(msg string) { fmt.Printf("⚠  %s\n", msg) }
		fail := func(msg string) error { return fmt.Errorf("✗ %s", msg) }

		fmt.Println("=== gh next bootstrap ===")
		fmt.Println()

		// terminal-notifier
		if _, err := exec.LookPath("terminal-notifier"); err != nil {
			warn("terminal-notifier not found — installing via brew...")
			if err := run("brew", "install", "terminal-notifier"); err != nil {
				return fail("brew install terminal-notifier failed: " + err.Error())
			}
		}
		out, _ := exec.Command("terminal-notifier", "-version").Output()
		ok("terminal-notifier " + string(out))

		// state dir
		if err := os.MkdirAll(state.Dir(), 0755); err != nil {
			return fail("create state dir: " + err.Error())
		}
		ok("state dir: " + state.Dir())

		// cron check
		fmt.Println()
		entry := getCronEntry()
		if entry == "" {
			warn("no schedule found — run: gh next program")
		} else {
			ok("cron: " + entry)
		}

		fmt.Println()
		fmt.Println("Bootstrap complete. Run: gh next status")
		return nil
	},
}

func run(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func getCronEntry() string {
	current := readCrontab()
	return showCronLine(current)
}

func showCronLine(crontab string) string {
	lines := splitLines(crontab)
	for i, line := range lines {
		if line == cronMarker && i+1 < len(lines) {
			return lines[i+1]
		}
	}
	return ""
}

func splitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\n' {
			lines = append(lines, s[start:i])
			start = i + 1
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

func init() {
	Root.AddCommand(bootstrapCmd)
}
