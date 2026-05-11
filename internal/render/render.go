package render

import (
	"fmt"
	"html"
	"os"
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

func HTML(s *state.Summary) error {
	f, err := os.Create(state.HTMLPath())
	if err != nil {
		return err
	}
	defer f.Close()

	w := func(format string, args ...any) {
		fmt.Fprintf(f, format+"\n", args...)
	}

	w(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta http-equiv="refresh" content="300">
<title>gh next</title>
<style>
*, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
body {
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
  background: #0d1117; color: #e6edf3;
  padding: 28px 24px; max-width: 860px; margin: 0 auto; font-size: 14px;
}
h1 { font-size: 0.85rem; color: #484f58; font-weight: 400; margin-bottom: 28px; letter-spacing: .02em; }
.group { margin-bottom: 36px; }
h2 { font-size: 1rem; font-weight: 600; margin-bottom: 14px; display: flex; align-items: center; gap: 8px; }
.subgroup { margin-bottom: 18px; }
h3 { font-size: 0.78rem; font-weight: 600; color: #8b949e; text-transform: uppercase;
     letter-spacing: .06em; margin-bottom: 6px; display: flex; align-items: center; gap: 6px; }
.count { background: #21262d; border-radius: 20px; padding: 1px 8px;
         font-size: 0.75rem; color: #8b949e; font-weight: 500; }
ul { list-style: none; display: flex; flex-direction: column; gap: 4px; }
.sg-merge_conflict      ul { --accent: #f85149; }
.sg-ready_to_merge      ul { --accent: #3fb950; }
.sg-changes_needed      ul { --accent: #d29922; }
.sg-awaiting_answer     ul,
.sg-needs_reply         ul { --accent: #58a6ff; }
.sg-review_requested    ul,
.sg-re_review_requested ul { --accent: #e3b341; }
.sg-awaiting_merge      ul,
.sg-ci_running          ul,
.sg-awaiting_review     ul,
.sg-awaiting_response   ul { --accent: #484f58; }
.sg-draft               ul { --accent: #484f58; }
.sg-stale               ul { --accent: #30363d; }
.sg-answered            ul { --accent: #1a4a2e; }
li.item {
  background: #161b22; border: 1px solid #21262d;
  border-left: 3px solid var(--accent, #30363d);
  border-radius: 0 6px 6px 0;
  padding: 9px 12px; display: flex; flex-direction: column; gap: 4px;
  transition: border-color .15s;
}
li.item:hover { border-left-color: var(--accent, #58a6ff); background: #1c2128; }
.item-top { display: flex; align-items: flex-start; gap: 8px; }
.kind-badge {
  font-size: 0.6rem; font-weight: 700; letter-spacing: .05em;
  padding: 2px 5px; border-radius: 3px; white-space: nowrap; flex-shrink: 0; margin-top: 1px;
}
.kind-badge.pr         { background: #1f4b8e; color: #79c0ff; }
.kind-badge.issue      { background: #3d1f8e; color: #d2a8ff; }
.kind-badge.discussion { background: #1f6b5e; color: #56d364; }
a.title { color: #e6edf3; text-decoration: none; font-weight: 500; line-height: 1.4; flex: 1; }
a.title:hover { color: #58a6ff; }
.time { font-size: 0.75rem; color: #484f58; white-space: nowrap; flex-shrink: 0; margin-left: auto; padding-left: 8px; }
.item-meta { display: flex; align-items: center; gap: 6px; padding-left: 32px; }
a.repo { font-size: 0.75rem; color: #6e7681; text-decoration: none; }
a.repo:hover { color: #8b949e; text-decoration: underline; }
.num { font-size: 0.72rem; color: #484f58; }
</style>
</head>
<body>`)

	w(`<h1>gh next &mdash; %s</h1>`, html.EscapeString(s.UpdatedAt))

	type htmlGroup struct {
		label string
		color string
		items []state.Item
	}
	groups := []htmlGroup{
		{"🟢 Your turn", "#3fb950", s.YourTurn},
		{"🟡 Their turn", "#d29922", s.TheirTurn},
		{"⚪ Parked", "#8b949e", s.Parked},
	}

	type subgroup struct {
		status string
		label  string
	}
	subgroups := map[string][]subgroup{
		"your_turn": {
			{"merge_conflict", "⚔️ Merge conflict"},
			{"ready_to_merge", "✅ Ready to merge"},
			{"changes_needed", "🔧 Changes needed"},
			{"awaiting_answer", "💬 Awaiting answer"},
			{"needs_reply", "💬 Needs reply"},
			{"review_requested", "👀 Review requested"},
			{"re_review_requested", "🔁 Re-review requested"},
		},
		"their_turn": {
			{"awaiting_merge", "⏳ Awaiting merge"},
			{"ci_running", "⚙️ CI running"},
			{"awaiting_review", "🔍 Awaiting review"},
			{"awaiting_response", "🔍 Awaiting response"},
			{"awaiting_action", "⏳ Awaiting action"},
		},
		"parked": {
			{"draft", "📝 Draft"},
			{"answered", "✅ Answered"},
			{"stale", "🕸️ Stale"},
		},
	}

	groupKey := map[string]string{
		"🟢 Your turn":  "your_turn",
		"🟡 Their turn": "their_turn",
		"⚪ Parked":     "parked",
	}

	for _, g := range groups {
		if len(g.items) == 0 {
			continue
		}
		w(`<section class="group">`)
		w(`<h2 style="color:%s">%s <span class="count">%d</span></h2>`, g.color, g.label, len(g.items))

		byStatus := map[string][]state.Item{}
		for _, item := range g.items {
			byStatus[item.Status] = append(byStatus[item.Status], item)
		}

		key := groupKey[g.label]
		for _, sg := range subgroups[key] {
			items := byStatus[sg.status]
			if len(items) == 0 {
				continue
			}
			w(`  <div class="subgroup sg-%s">`, sg.status)
			w(`    <h3>%s <span class="count">%d</span></h3>`, sg.label, len(items))
			w(`    <ul>`)
			for _, item := range items {
				kindLabel := map[string]string{"pr": "PR", "issue": "ISS", "discussion": "DISC"}[item.Kind]
				repoURL := "https://github.com/" + item.Repo
				w(`      <li class="item %s">`, item.Kind)
				w(`        <div class="item-top">`)
				w(`          <span class="kind-badge %s">%s</span>`, item.Kind, kindLabel)
				w(`          <a class="title" href="%s" target="_blank">%s</a>`, item.URL, html.EscapeString(item.Title))
				w(`          <span class="time" data-ts="%s"></span>`, item.UpdatedAt)
				w(`        </div>`)
				w(`        <div class="item-meta">`)
				w(`          <a class="repo" href="%s" target="_blank">%s</a>`, repoURL, html.EscapeString(item.Repo))
				w(`          <span class="num">#%d</span>`, item.Number)
				w(`        </div>`)
				w(`      </li>`)
			}
			w(`    </ul>`)
			w(`  </div>`)
		}
		w(`</section>`)
	}

	w(`<script>
document.querySelectorAll('.time[data-ts]').forEach(el => {
  const s = (Date.now() - new Date(el.dataset.ts)) / 1000;
  if      (s < 3600)   el.textContent = Math.floor(s/60)    + 'm ago';
  else if (s < 86400)  el.textContent = Math.floor(s/3600)  + 'h ago';
  else if (s < 604800) el.textContent = Math.floor(s/86400) + 'd ago';
  else                 el.textContent = Math.floor(s/604800) + 'w ago';
});
</script>
</body></html>`)

	return nil
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
