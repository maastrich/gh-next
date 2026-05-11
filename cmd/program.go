package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/maastrich/gh-next/internal/state"
	"github.com/spf13/cobra"
)

const cronMarker = "# gh-next"
const defaultCron = "0 8-18 * * 1-5"

var (
	removeScheduleFlag bool
	showScheduleFlag   bool
)

var programCmd = &cobra.Command{
	Use:   "program [cron-expression]",
	Short: "Schedule recurring status refresh",
	Long: `Set up a cron job to run 'gh next status' on a recurring schedule.

Default: every hour from 8am to 6pm, Monday–Friday (` + defaultCron + `)

Examples:
  gh next program                     use default schedule
  gh next program "0 * * * *"         every hour, all day
  gh next program "0 9-17 * * 1-5"    9am–5pm weekdays
  gh next program --show               show current schedule
  gh next program --remove             remove the cron job`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if showScheduleFlag {
			return showCron()
		}
		if removeScheduleFlag {
			return removeCron()
		}
		expr := defaultCron
		if len(args) > 0 {
			expr = args[0]
		}
		return setCron(expr)
	},
}

func init() {
	programCmd.Flags().BoolVar(&removeScheduleFlag, "remove", false, "Remove the scheduled job")
	programCmd.Flags().BoolVar(&showScheduleFlag, "show", false, "Show current schedule")
}

func readCrontab() string {
	out, err := exec.Command("crontab", "-l").Output()
	if err != nil {
		return ""
	}
	return string(out)
}

func writeCrontab(content string) error {
	cmd := exec.Command("crontab", "-")
	cmd.Stdin = strings.NewReader(content)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func ghAuthToken() string {
	out, err := exec.Command("gh", "auth", "token").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func setCron(expr string) error {
	ghBin, err := exec.LookPath("gh")
	if err != nil {
		return fmt.Errorf("gh not found in PATH: %w", err)
	}

	token := ghAuthToken()
	if token == "" {
		return fmt.Errorf("gh not authenticated — run: gh auth login")
	}

	home, _ := os.UserHomeDir()
	logPath := state.LogPath()
	// Inject HOME and GH_TOKEN so cron's minimal env doesn't break gh auth
	block := fmt.Sprintf("%s\n%s HOME=%s GH_TOKEN=%s %s next status >> %s 2>&1",
		cronMarker, expr, home, token, ghBin, logPath)

	current := readCrontab()
	cleaned := stripCronBlock(current)
	if cleaned != "" && !strings.HasSuffix(cleaned, "\n") {
		cleaned += "\n"
	}
	cleaned += block + "\n"

	if err := writeCrontab(cleaned); err != nil {
		return fmt.Errorf("write crontab: %w", err)
	}

	fmt.Printf("Scheduled: %s\n", expr)
	fmt.Printf("Command:   %s next status\n", ghBin)
	fmt.Printf("Log:       %s\n", logPath)
	fmt.Printf("Auth:      GH_TOKEN injected\n")
	return nil
}

func removeCron() error {
	current := readCrontab()
	if !strings.Contains(current, cronMarker) {
		fmt.Println("No schedule found.")
		return nil
	}
	cleaned := stripCronBlock(current)
	if err := writeCrontab(cleaned); err != nil {
		return fmt.Errorf("write crontab: %w", err)
	}
	fmt.Println("Schedule removed.")
	return nil
}

func showCron() error {
	current := readCrontab()
	lines := strings.Split(current, "\n")
	for i, line := range lines {
		if line == cronMarker && i+1 < len(lines) {
			fmt.Println(lines[i+1])
			return nil
		}
	}
	fmt.Println("No schedule configured. Run: gh next program")
	return nil
}

// stripCronBlock removes the marker line and the cron entry that follows it.
func stripCronBlock(crontab string) string {
	lines := strings.Split(crontab, "\n")
	out := make([]string, 0, len(lines))
	skip := false
	for _, line := range lines {
		if line == cronMarker {
			skip = true
			continue
		}
		if skip {
			skip = false
			continue
		}
		out = append(out, line)
	}
	result := strings.Join(out, "\n")
	// Trim trailing blank lines but keep a single trailing newline
	result = strings.TrimRight(result, "\n")
	if result != "" {
		result += "\n"
	}
	return result
}
