package notify

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/maastrich/gh-next/internal/state"
)

func Send(title, message, url string) {
	htmlURL := "file://" + state.HTMLPath()
	openURL := url
	if openURL == "" {
		openURL = htmlURL
	}

	if path, err := exec.LookPath("terminal-notifier"); err == nil {
		args := []string{
			"-title", title,
			"-message", message,
			"-group", "gh-next",
			"-open", openURL,
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
