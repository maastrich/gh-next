package notify

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/maastrich/gh-next/internal/state"
)

// launcherPath is a small shell script written alongside the HTML report.
// terminal-notifier -execute calls /bin/bash with this script — the only
// reliable way to open a file from the notification's sandboxed context.
func launcherPath() string {
	return state.Dir() + "/open-report.sh"
}

func writeLauncher() {
	htmlPath := state.HTMLPath()
	content := "#!/bin/bash\n/usr/bin/open " + htmlPath + "\n"
	_ = os.WriteFile(launcherPath(), []byte(content), 0755)
}

func Send(title, message, url string) {
	writeLauncher()

	if path, err := exec.LookPath("terminal-notifier"); err == nil {
		executeCmd := "/bin/bash " + launcherPath()
		if url != "" {
			// For PR-specific notifications open the PR URL in the browser
			executeCmd = fmt.Sprintf("/usr/bin/open %q", url)
		}
		args := []string{
			"-title", title,
			"-message", message,
			"-group", "gh-next",
			"-execute", executeCmd,
		}
		cmd := exec.Command(path, args...)
		cmd.Stdout = os.Stderr
		cmd.Stderr = os.Stderr
		_ = cmd.Run()
		return
	}

	safe := func(s string) string {
		return strings.ReplaceAll(s, `"`, `\"`)
	}
	script := fmt.Sprintf(`display notification "%s" with title "%s"`, safe(message), safe(title))
	_ = exec.Command("osascript", "-e", script).Run()
}

func Summary(s *state.Summary) {
	if s.YourCount > 0 {
		parts := buildSummaryParts(s.YourTurn)
		msg := strings.Join(parts, " · ")
		if msg == "" {
			msg = fmt.Sprintf("%d item(s) need your attention", s.YourCount)
		}
		label := "item"
		if s.YourCount > 1 {
			label = "items"
		}
		Send(fmt.Sprintf("🟢 %d %s need attention", s.YourCount, label), msg, "")
	} else {
		Send("gh next", "All clear ✓", "")
	}
}

func buildSummaryParts(items []state.Item) []string {
	counts := map[string]int{}
	for _, item := range items {
		counts[item.Status]++
	}
	var parts []string
	add := func(status, label string) {
		if n := counts[status]; n > 0 {
			parts = append(parts, fmt.Sprintf("%s %d", label, n))
		}
	}
	add("merge_conflict", "⚔️")
	add("ready_to_merge", "✅")
	add("changes_needed", "🔧")
	add("awaiting_answer", "💬")
	add("needs_reply", "💬")
	add("review_requested", "👀")
	add("re_review_requested", "🔁")
	return parts
}

func Diff(prev *state.PrevState, items []state.Item) {
	prevMap := map[string]string{}
	for _, p := range prev.Items {
		prevMap[p.URL] = p.Status
	}
	for _, item := range items {
		oldStatus := prevMap[item.URL]
		if oldStatus == "" || oldStatus == item.Status {
			continue
		}
		switch item.Status {
		case "ready_to_merge":
			Send("Ready to Merge "+item.Icon, item.Title, item.URL)
		case "changes_needed":
			Send("Changes Needed "+item.Icon, item.Title, item.URL)
		case "merge_conflict":
			Send("Merge Conflict "+item.Icon, item.Title, item.URL)
		case "stale":
			Send("Gone Stale "+item.Icon, item.Title, item.URL)
		}
	}
}
