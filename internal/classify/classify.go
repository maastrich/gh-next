package classify

import (
	"time"

	"github.com/maastrich/gh-next/internal/fetch"
	"github.com/maastrich/gh-next/internal/state"
)

const defaultStaleDays = 7

func Run(data *fetch.RawData, staleDays int) []state.Item {
	if staleDays <= 0 {
		staleDays = defaultStaleDays
	}
	staleThreshold := time.Duration(staleDays) * 24 * time.Hour

	var items []state.Item

	authoredURLs := map[string]bool{}
	for _, pr := range data.AuthoredPRs {
		authoredURLs[pr.URL] = true
		items = append(items, classifyPR(pr, "mine", data.User, staleThreshold))
	}

	for _, pr := range data.ReviewPRs {
		if authoredURLs[pr.URL] {
			continue
		}
		items = append(items, classifyPR(pr, "review", data.User, staleThreshold))
	}

	for _, issue := range data.Issues {
		items = append(items, classifyIssue(issue, data.User, staleThreshold))
	}

	for _, disc := range data.Discussions {
		items = append(items, classifyDiscussion(disc, data.User, staleThreshold))
	}

	return items
}

func parseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}
	}
	return t
}

func isStale(updatedAt string, threshold time.Duration) bool {
	t := parseTime(updatedAt)
	if t.IsZero() {
		return false
	}
	return time.Since(t) > threshold
}

func classifyPR(pr fetch.PR, role, user string, staleThreshold time.Duration) state.Item {
	item := state.Item{
		Number:    pr.Number,
		Title:     pr.Title,
		URL:       pr.URL,
		Repo:      pr.Repository.NameWithOwner,
		Kind:      "pr",
		UpdatedAt: pr.UpdatedAt,
	}

	lastCommitDate := ""
	ciState := "NONE"
	if len(pr.Commits.Nodes) > 0 {
		lastCommitDate = pr.Commits.Nodes[len(pr.Commits.Nodes)-1].Commit.CommittedDate
		if sc := pr.Commits.Nodes[len(pr.Commits.Nodes)-1].Commit.StatusCheckRollup; sc != nil {
			ciState = sc.State
		}
	}

	if role == "review" {
		var myLastReview time.Time
		for _, r := range pr.Reviews.Nodes {
			if r.Author.Login == user {
				t := parseTime(r.SubmittedAt)
				if t.After(myLastReview) {
					myLastReview = t
				}
			}
		}
		lastCommit := parseTime(lastCommitDate)
		if myLastReview.IsZero() {
			item.Status = "review_requested"
			item.Icon = "👀"
			item.Group = "your_turn"
		} else if lastCommit.After(myLastReview) {
			item.Status = "re_review_requested"
			item.Icon = "🔁"
			item.Group = "your_turn"
		} else {
			item.Status = "awaiting_action"
			item.Icon = "⏳"
			item.Group = "their_turn"
		}
		return item
	}

	// authored PR
	switch {
	case pr.IsDraft:
		item.Status = "draft"
		item.Icon = "📝"
		item.Group = "parked"

	case pr.Mergeable == "CONFLICTING":
		item.Status = "merge_conflict"
		item.Icon = "⚔️"
		item.Group = "your_turn"

	case pr.ReviewDecision == "CHANGES_REQUESTED":
		item.Status = "changes_needed"
		item.Icon = "🔧"
		item.Group = "your_turn"

	case ciState == "PENDING" || ciState == "IN_PROGRESS" || ciState == "EXPECTED":
		item.Status = "ci_running"
		item.Icon = "⚙️"
		item.Group = "their_turn"

	case pr.ReviewDecision == "APPROVED":
		if ciState == "FAILURE" || ciState == "ERROR" {
			item.Status = "changes_needed"
			item.Icon = "🔧"
			item.Group = "your_turn"
		} else if pr.ViewerCanUpdate {
			item.Status = "ready_to_merge"
			item.Icon = "✅"
			item.Group = "your_turn"
		} else {
			item.Status = "awaiting_merge"
			item.Icon = "⏳"
			item.Group = "their_turn"
		}

	default:
		lastCommit := parseTime(lastCommitDate)
		hasNewComments := false
		for _, r := range pr.Reviews.Nodes {
			if r.Author.Login != user {
				t := parseTime(r.SubmittedAt)
				if t.After(lastCommit) {
					hasNewComments = true
					break
				}
			}
		}
		if hasNewComments {
			item.Status = "awaiting_answer"
			item.Icon = "💬"
			item.Group = "your_turn"
		} else if ciState == "FAILURE" || ciState == "ERROR" {
			item.Status = "changes_needed"
			item.Icon = "🔧"
			item.Group = "your_turn"
		} else {
			item.Status = "awaiting_review"
			item.Icon = "🔍"
			item.Group = "their_turn"
		}
	}

	if item.Group != "your_turn" && isStale(pr.UpdatedAt, staleThreshold) {
		item.Status = "stale"
		item.Icon = "🕸️"
		item.Group = "parked"
	}

	return item
}

func classifyIssue(issue fetch.Issue, user string, staleThreshold time.Duration) state.Item {
	item := state.Item{
		Number:    issue.Number,
		Title:     issue.Title,
		URL:       issue.URL,
		Repo:      issue.Repository.NameWithOwner,
		Kind:      "issue",
		UpdatedAt: issue.UpdatedAt,
	}

	comments := issue.Comments.Nodes
	if len(comments) == 0 {
		item.Status = "awaiting_response"
		item.Icon = "🔍"
		item.Group = "their_turn"
	} else {
		last := comments[len(comments)-1]
		if last.Author.Login != "" && last.Author.Login != user {
			var myLastTs time.Time
			for _, c := range comments {
				if c.Author.Login == user {
					t := parseTime(c.CreatedAt)
					if t.After(myLastTs) {
						myLastTs = t
					}
				}
			}
			lastTs := parseTime(last.CreatedAt)
			if lastTs.After(myLastTs) {
				item.Status = "needs_reply"
				item.Icon = "💬"
				item.Group = "your_turn"
			} else {
				item.Status = "awaiting_response"
				item.Icon = "🔍"
				item.Group = "their_turn"
			}
		} else {
			item.Status = "awaiting_response"
			item.Icon = "🔍"
			item.Group = "their_turn"
		}
	}

	if item.Group != "your_turn" && isStale(issue.UpdatedAt, staleThreshold) {
		item.Status = "stale"
		item.Icon = "🕸️"
		item.Group = "parked"
	}

	return item
}

func classifyDiscussion(disc fetch.Discussion, user string, staleThreshold time.Duration) state.Item {
	item := state.Item{
		Number:    disc.Number,
		Title:     disc.Title,
		URL:       disc.URL,
		Repo:      disc.Repository.NameWithOwner,
		Kind:      "discussion",
		UpdatedAt: disc.UpdatedAt,
	}

	if disc.IsAnswered {
		item.Status = "answered"
		item.Icon = "✅"
		item.Group = "parked"
		return item
	}

	comments := disc.Comments.Nodes
	if len(comments) == 0 {
		item.Status = "awaiting_response"
		item.Icon = "🔍"
		item.Group = "their_turn"
	} else {
		last := comments[len(comments)-1]
		if last.Author.Login != "" && last.Author.Login != user {
			var myLastTs time.Time
			for _, c := range comments {
				if c.Author.Login == user {
					t := parseTime(c.CreatedAt)
					if t.After(myLastTs) {
						myLastTs = t
					}
				}
			}
			lastTs := parseTime(last.CreatedAt)
			if lastTs.After(myLastTs) {
				item.Status = "needs_reply"
				item.Icon = "💬"
				item.Group = "your_turn"
			} else {
				item.Status = "awaiting_response"
				item.Icon = "🔍"
				item.Group = "their_turn"
			}
		} else {
			item.Status = "awaiting_response"
			item.Icon = "🔍"
			item.Group = "their_turn"
		}
	}

	if item.Group != "your_turn" && isStale(disc.UpdatedAt, staleThreshold) {
		item.Status = "stale"
		item.Icon = "🕸️"
		item.Group = "parked"
	}

	return item
}

func Group(items []state.Item) *state.Summary {
	s := &state.Summary{}
	for _, item := range items {
		switch item.Group {
		case "your_turn":
			s.YourTurn = append(s.YourTurn, item)
		case "their_turn":
			s.TheirTurn = append(s.TheirTurn, item)
		case "parked":
			s.Parked = append(s.Parked, item)
		}
	}
	if s.YourTurn == nil {
		s.YourTurn = []state.Item{}
	}
	if s.TheirTurn == nil {
		s.TheirTurn = []state.Item{}
	}
	if s.Parked == nil {
		s.Parked = []state.Item{}
	}
	s.YourCount = len(s.YourTurn)
	return s
}
