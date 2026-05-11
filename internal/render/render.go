package render

import (
	"fmt"
	"strings"
	"time"

	"github.com/maastrich/gh-next/internal/state"
)

const (
	colorReset  = "\033[0m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorGray   = "\033[37m"
	colorDim    = "\033[2m"
)

func Summary(s *state.Summary) {
	fmt.Println()
	fmt.Printf("gh next — %s\n", formatTime(s.UpdatedAt))
	fmt.Println(strings.Repeat("─", 60))

	printGroup("🟢 Your turn", colorGreen, s.YourTurn)
	printGroup("🟡 Their turn", colorYellow, s.TheirTurn)
	printGroup("⚪ Parked", colorGray, s.Parked)

	fmt.Println()
	if s.YourCount > 0 {
		fmt.Printf(colorGreen+"%d item(s) need your attention."+colorReset+"\n\n", s.YourCount)
	} else {
		fmt.Printf(colorGreen + "All clear." + colorReset + "\n\n")
	}
}

func printGroup(label, color string, items []state.Item) {
	if len(items) == 0 {
		return
	}
	fmt.Println()
	fmt.Printf("%s%s (%d)%s\n", color, label, len(items), colorReset)
	for _, item := range items {
		title := item.Title
		if len(title) > 70 {
			title = title[:70] + "…"
		}
		fmt.Printf("  %s %s  %s[%s]%s\n", item.Icon, title, colorDim, item.Repo, colorReset)
		fmt.Printf("     %s%s%s\n", colorDim, item.URL, colorReset)
	}
}

func formatTime(ts string) string {
	t, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return ts
	}
	since := time.Since(t)
	switch {
	case since < time.Minute:
		return "just now"
	case since < time.Hour:
		return fmt.Sprintf("%dm ago", int(since.Minutes()))
	case since < 24*time.Hour:
		return fmt.Sprintf("%dh ago", int(since.Hours()))
	default:
		return fmt.Sprintf("%dd ago", int(since.Hours()/24))
	}
}
