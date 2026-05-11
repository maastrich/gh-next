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
  font-family: -apple-system, BlinkMacSystemFont, "Segoe UI Variable Display",
               "Segoe UI", system-ui, sans-serif;
  background: oklch(13%% 0.008 250);
  color: var(--t1);
  padding: 32px 24px 64px;
  max-width: 820px;
  margin: 0 auto;
  font-size: 13px;
  line-height: 1.5;
}

:root {
  --bg:      oklch(13%% 0.008 250);
  --surface: oklch(17%% 0.01 250);
  --border:  oklch(25%% 0.01 250);
  --t1: oklch(91%% 0.005 250);
  --t2: oklch(57%% 0.008 250);
  --t3: oklch(39%% 0.007 250);
  --mono: "SFMono-Regular", "SF Mono", "Fira Code", Consolas, "Liberation Mono", monospace;
}

header {
  font-size: 0.75rem;
  color: var(--t3);
  letter-spacing: .04em;
  margin-bottom: 36px;
  font-family: var(--mono);
}

.group { margin-bottom: 40px; }

.group-theirs {
  --t1: oklch(66%% 0.007 250);
  --t2: oklch(44%% 0.006 250);
  --t3: oklch(32%% 0.005 250);
}
.group-parked {
  --t1: oklch(46%% 0.006 250);
  --t2: oklch(33%% 0.005 250);
  --t3: oklch(24%% 0.004 250);
}

h2 {
  font-size: 0.8rem;
  font-weight: 600;
  letter-spacing: .05em;
  text-transform: uppercase;
  color: var(--t1);
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  gap: 8px;
}

h2 .g-count {
  font-size: 0.72rem;
  font-weight: 500;
  background: oklch(22%% 0.012 250);
  color: var(--t2);
  border-radius: 20px;
  padding: 1px 8px;
  letter-spacing: 0;
  text-transform: none;
}

.subgroup { margin-bottom: 20px; }

h3 {
  font-size: 0.68rem;
  font-weight: 600;
  color: var(--t3);
  text-transform: uppercase;
  letter-spacing: .07em;
  margin-bottom: 6px;
  display: flex;
  align-items: center;
  gap: 6px;
}

.sg-dot {
  width: 6px; height: 6px;
  border-radius: 1px;
  flex-shrink: 0;
  display: inline-block;
}

h3 .sg-count {
  font-size: 0.65rem;
  font-weight: 500;
  color: var(--t3);
  letter-spacing: 0;
  text-transform: none;
  background: oklch(20%% 0.01 250);
  padding: 0px 5px;
  border-radius: 10px;
}

ul { list-style: none; display: flex; flex-direction: column; gap: 3px; }

li.item {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 5px;
  padding: 8px 12px;
  display: flex;
  flex-direction: column;
  gap: 3px;
  transition: background .1s, border-color .1s;
}
li.item:hover {
  background: oklch(20%% 0.012 250);
  border-color: oklch(30%% 0.012 250);
}

.row-a {
  display: flex;
  align-items: baseline;
  gap: 7px;
}

.kind {
  font-family: var(--mono);
  font-size: 0.6rem;
  font-weight: 600;
  letter-spacing: .04em;
  padding: 1px 4px;
  border-radius: 3px;
  white-space: nowrap;
  flex-shrink: 0;
  position: relative;
  top: -1px;
}
.kind-pr         { background: oklch(20%% 0.06 240); color: oklch(72%% 0.12 240); }
.kind-issue      { background: oklch(20%% 0.06 300); color: oklch(72%% 0.12 300); }
.kind-discussion { background: oklch(20%% 0.06 155); color: oklch(72%% 0.12 155); }

a.title {
  color: var(--t1);
  text-decoration: none;
  font-weight: 500;
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
a.title:hover { color: oklch(72%% 0.12 240); }

.ts {
  font-family: var(--mono);
  font-size: 0.68rem;
  color: var(--t3);
  white-space: nowrap;
  flex-shrink: 0;
  margin-left: auto;
  padding-left: 12px;
}

.row-b {
  display: flex;
  align-items: center;
  gap: 5px;
  padding-left: 25px;
  flex-wrap: wrap;
  min-width: 0;
}

a.repo {
  font-family: var(--mono);
  font-size: 0.68rem;
  color: var(--t2);
  text-decoration: none;
  white-space: nowrap;
}
a.repo:hover { color: var(--t1); }

.num {
  font-family: var(--mono);
  font-size: 0.65rem;
  color: var(--t3);
  white-space: nowrap;
}

.d-sep  { color: var(--t3); padding: 0 1px; }
.d-ci-fail { color: oklch(66%% 0.18 25); font-size: 0.68rem; }
.d-ci-ok   { color: oklch(68%% 0.15 145); font-size: 0.68rem; }
.d-ci-pend { color: oklch(76%% 0.13 85); font-size: 0.68rem; }
.d-checks  { color: oklch(55%% 0.01 250); font-size: 0.65rem; font-family: var(--mono); }
.d-rev-rej { color: oklch(66%% 0.18 25); font-size: 0.68rem; }
.d-rev-ok  { color: oklch(68%% 0.15 145); font-size: 0.68rem; }
.d-meta    { color: var(--t2); font-size: 0.68rem; }
</style>
</head>
<body>`)

	w(`<header>gh next &mdash; %s</header>`, html.EscapeString(s.UpdatedAt))

	type sgDef struct {
		status string
		label  string
		dot    string // oklch color
	}
	subgroups := map[string][]sgDef{
		"your_turn": {
			{"merge_conflict", "Merge conflict", "oklch(66% 0.2 25)"},
			{"ready_to_merge", "Ready to merge", "oklch(70% 0.17 145)"},
			{"changes_needed", "Changes needed", "oklch(74% 0.17 50)"},
			{"awaiting_answer", "Awaiting answer", "oklch(70% 0.12 240)"},
			{"needs_reply", "Needs reply", "oklch(70% 0.12 240)"},
			{"review_requested", "Review requested", "oklch(78% 0.14 85)"},
			{"re_review_requested", "Re-review requested", "oklch(78% 0.14 85)"},
		},
		"their_turn": {
			{"awaiting_merge", "Awaiting merge", "oklch(42% 0.008 250)"},
			{"ci_running", "CI running", "oklch(42% 0.008 250)"},
			{"awaiting_review", "Awaiting review", "oklch(42% 0.008 250)"},
			{"awaiting_response", "Awaiting response", "oklch(42% 0.008 250)"},
			{"awaiting_action", "Awaiting action", "oklch(42% 0.008 250)"},
			{"team_review_requested", "Team review requested", "oklch(42% 0.008 250)"},
		},
		"parked": {
			{"draft", "Draft", "oklch(42% 0.008 250)"},
			{"answered", "Answered", "oklch(48% 0.1 145)"},
			{"stale", "Stale", "oklch(30% 0.007 250)"},
		},
	}

	type htmlGroup struct {
		key   string
		label string
		cls   string
		items []state.Item
	}
	groups := []htmlGroup{
		{"your_turn", "Your turn", "group-yours", s.YourTurn},
		{"their_turn", "Their turn", "group-theirs", s.TheirTurn},
		{"parked", "Parked", "group-parked", s.Parked},
	}

	kindLabel := map[string]string{"pr": "PR", "issue": "ISS", "discussion": "DISC"}

	for _, g := range groups {
		if len(g.items) == 0 {
			continue
		}
		w(`<section class="group %s">`, g.cls)
		w(`  <h2>%s <span class="g-count">%d</span></h2>`, html.EscapeString(g.label), len(g.items))

		byStatus := map[string][]state.Item{}
		for _, item := range g.items {
			byStatus[item.Status] = append(byStatus[item.Status], item)
		}

		for _, sg := range subgroups[g.key] {
			items := byStatus[sg.status]
			if len(items) == 0 {
				continue
			}
			w(`  <div class="subgroup">`)
			w(`    <h3><span class="sg-dot" style="background:%s"></span>%s<span class="sg-count">%d</span></h3>`,
				sg.dot, html.EscapeString(sg.label), len(items))
			w(`    <ul>`)
			for _, item := range items {
				kl := kindLabel[item.Kind]
				repoURL := "https://github.com/" + item.Repo
				detail := itemDetail(item)
				w(`      <li class="item">`)
				w(`        <div class="row-a">`)
				w(`          <span class="kind kind-%s">%s</span>`, item.Kind, kl)
				w(`          <a class="title" href="%s" target="_blank">%s</a>`, item.URL, html.EscapeString(item.Title))
				w(`          <span class="ts" data-ts="%s"></span>`, item.UpdatedAt)
				w(`        </div>`)
				w(`        <div class="row-b">`)
				w(`          <a class="repo" href="%s" target="_blank">%s</a>`, repoURL, html.EscapeString(item.Repo))
				w(`          <span class="num">#%d</span>`, item.Number)
				if detail != "" {
					w(`          %s`, detail)
				}
				w(`        </div>`)
				w(`      </li>`)
			}
			w(`    </ul>`)
			w(`  </div>`)
		}
		w(`</section>`)
	}

	w(`<script>
document.querySelectorAll('.ts[data-ts]').forEach(el => {
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

func itemDetail(item state.Item) string {
	if item.Kind != "pr" {
		return ""
	}
	var parts []string

	// CI failure — blocking
	switch {
	case len(item.FailedChecks) > 0:
		names := item.FailedChecks
		if len(names) > 3 {
			names = names[:3]
		}
		escaped := make([]string, len(names))
		for i, n := range names {
			if len(n) > 24 {
				n = n[:24] + "…"
			}
			escaped[i] = html.EscapeString(n)
		}
		parts = append(parts, fmt.Sprintf(
			`<span class="d-ci-fail">CI ✗</span> <span class="d-checks">%s</span>`,
			strings.Join(escaped, " · "),
		))
	case item.CIState == "FAILURE" || item.CIState == "ERROR":
		parts = append(parts, `<span class="d-ci-fail">CI ✗</span>`)
	case item.CIState == "PENDING" || item.CIState == "IN_PROGRESS" || item.CIState == "EXPECTED":
		parts = append(parts, `<span class="d-ci-pend">CI ⋯</span>`)
	}

	// Changes requested — blocking
	if len(item.ChangesRequestedBy) > 0 {
		reviewers := item.ChangesRequestedBy
		if len(reviewers) > 2 {
			reviewers = reviewers[:2]
		}
		names := make([]string, len(reviewers))
		for i, r := range reviewers {
			names[i] = html.EscapeString(r)
		}
		suffix := " requested changes"
		if len(item.ChangesRequestedBy) > 2 {
			suffix = fmt.Sprintf(" +%d requested changes", len(item.ChangesRequestedBy)-2)
		}
		parts = append(parts, fmt.Sprintf(`<span class="d-rev-rej">%s%s</span>`, strings.Join(names, ", "), suffix))
	}

	if len(parts) == 0 {
		return ""
	}
	return `<span class="d-sep">·</span> ` + strings.Join(parts, ` <span class="d-sep">·</span> `)
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
