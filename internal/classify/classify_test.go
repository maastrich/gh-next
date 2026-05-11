package classify

import (
	"fmt"
	"testing"
	"time"

	"github.com/maastrich/gh-next/internal/fetch"
	"github.com/maastrich/gh-next/internal/state"
)

const user = "alice"

func ago(d time.Duration) string {
	return time.Now().UTC().Add(-d).Format(time.RFC3339)
}

func basePR(number int) fetch.PR {
	return fetch.PR{
		Number:    number,
		Title:     fmt.Sprintf("PR %d", number),
		URL:       fmt.Sprintf("https://github.com/org/repo/pull/%d", number),
		UpdatedAt: ago(1 * time.Hour),
		Repository: struct {
			NameWithOwner string "json:\"nameWithOwner\""
		}{NameWithOwner: "org/repo"},
	}
}

func withCommit(pr fetch.PR, committedAt string, ciState string) fetch.PR {
	node := struct {
		Commit struct {
			CommittedDate     string             `json:"committedDate"`
			StatusCheckRollup *fetch.CheckRollup `json:"statusCheckRollup"`
		} `json:"commit"`
	}{}
	node.Commit.CommittedDate = committedAt
	if ciState != "" {
		node.Commit.StatusCheckRollup = &fetch.CheckRollup{State: ciState}
	}
	pr.Commits.Nodes = append(pr.Commits.Nodes, node)
	return pr
}

func withRollup(pr fetch.PR, committedAt string, rollup *fetch.CheckRollup) fetch.PR {
	node := struct {
		Commit struct {
			CommittedDate     string             `json:"committedDate"`
			StatusCheckRollup *fetch.CheckRollup `json:"statusCheckRollup"`
		} `json:"commit"`
	}{}
	node.Commit.CommittedDate = committedAt
	node.Commit.StatusCheckRollup = rollup
	pr.Commits.Nodes = append(pr.Commits.Nodes, node)
	return pr
}

func withReview(pr fetch.PR, login, submittedAt, reviewState string) fetch.PR {
	pr.Reviews.Nodes = append(pr.Reviews.Nodes, struct {
		Author struct {
			Login string "json:\"login\""
		} "json:\"author\""
		State       string "json:\"state\""
		SubmittedAt string "json:\"submittedAt\""
	}{
		Author: struct {
			Login string "json:\"login\""
		}{Login: login},
		State:       reviewState,
		SubmittedAt: submittedAt,
	})
	return pr
}

func staleUpdatedAt() string { return ago(10 * 24 * time.Hour) }
func freshUpdatedAt() string { return ago(1 * time.Hour) }

func TestClassifyPR_authored(t *testing.T) {
	threshold := 7 * 24 * time.Hour

	cases := []struct {
		name       string
		pr         fetch.PR
		wantStatus string
		wantGroup  string
	}{
		{
			name:       "draft",
			pr:         func() fetch.PR { p := basePR(1); p.IsDraft = true; return p }(),
			wantStatus: "draft",
			wantGroup:  "parked",
		},
		{
			name: "merge_conflict",
			pr: func() fetch.PR {
				p := basePR(2)
				p.Mergeable = "CONFLICTING"
				return p
			}(),
			wantStatus: "merge_conflict",
			wantGroup:  "your_turn",
		},
		{
			name: "changes_needed via review decision",
			pr: func() fetch.PR {
				p := basePR(3)
				p.ReviewDecision = "CHANGES_REQUESTED"
				return p
			}(),
			wantStatus: "changes_needed",
			wantGroup:  "your_turn",
		},
		{
			name:       "ci_running PENDING",
			pr:         withCommit(basePR(4), ago(30*time.Minute), "PENDING"),
			wantStatus: "ci_running",
			wantGroup:  "their_turn",
		},
		{
			name:       "ci_running IN_PROGRESS",
			pr:         withCommit(basePR(5), ago(30*time.Minute), "IN_PROGRESS"),
			wantStatus: "ci_running",
			wantGroup:  "their_turn",
		},
		{
			name: "ready_to_merge approved + can update",
			pr: func() fetch.PR {
				p := withCommit(basePR(6), ago(30*time.Minute), "SUCCESS")
				p.ReviewDecision = "APPROVED"
				p.ViewerCanUpdate = true
				return p
			}(),
			wantStatus: "ready_to_merge",
			wantGroup:  "your_turn",
		},
		{
			name: "awaiting_merge approved + cannot update",
			pr: func() fetch.PR {
				p := withCommit(basePR(7), ago(30*time.Minute), "SUCCESS")
				p.ReviewDecision = "APPROVED"
				p.ViewerCanUpdate = false
				return p
			}(),
			wantStatus: "awaiting_merge",
			wantGroup:  "their_turn",
		},
		{
			name: "changes_needed approved + CI failure",
			pr: func() fetch.PR {
				p := withCommit(basePR(8), ago(30*time.Minute), "FAILURE")
				p.ReviewDecision = "APPROVED"
				p.ViewerCanUpdate = true
				return p
			}(),
			wantStatus: "changes_needed",
			wantGroup:  "your_turn",
		},
		{
			name: "awaiting_answer reviewer commented after last commit",
			pr: func() fetch.PR {
				commitTime := ago(2 * time.Hour)
				p := withCommit(basePR(9), commitTime, "")
				p = withReview(p, "bob", ago(1*time.Hour), "COMMENTED")
				return p
			}(),
			wantStatus: "awaiting_answer",
			wantGroup:  "your_turn",
		},
		{
			name: "changes_needed CI failure no approval",
			pr: func() fetch.PR {
				return withCommit(basePR(10), ago(30*time.Minute), "FAILURE")
			}(),
			wantStatus: "changes_needed",
			wantGroup:  "your_turn",
		},
		{
			name: "auth-gate only CI failure, no approval → awaiting_review",
			pr: func() fetch.PR {
				r := &fetch.CheckRollup{State: "FAILURE"}
				r.Contexts.Nodes = []fetch.CheckContext{
					{State: "FAILURE", TargetUrl: "https://vercel.com/git/authorize?team=foo"},
				}
				return withRollup(basePR(11), ago(30*time.Minute), r)
			}(),
			wantStatus: "awaiting_review",
			wantGroup:  "their_turn",
		},
		{
			name: "auth-gate only CI failure, approved + can update → ready_to_merge",
			pr: func() fetch.PR {
				p := basePR(12)
				p.ReviewDecision = "APPROVED"
				p.ViewerCanUpdate = true
				r := &fetch.CheckRollup{State: "FAILURE"}
				r.Contexts.Nodes = []fetch.CheckContext{
					{State: "FAILURE", TargetUrl: "https://vercel.com/git/authorize?team=foo"},
				}
				return withRollup(p, ago(30*time.Minute), r)
			}(),
			wantStatus: "ready_to_merge",
			wantGroup:  "your_turn",
		},
		{
			name: "mixed auth-gate + real CI failure → changes_needed",
			pr: func() fetch.PR {
				r := &fetch.CheckRollup{State: "FAILURE"}
				r.Contexts.Nodes = []fetch.CheckContext{
					{State: "FAILURE", TargetUrl: "https://vercel.com/git/authorize?team=foo"},
					{State: "FAILURE", TargetUrl: "https://github.com/org/repo/runs/123"},
				}
				return withRollup(basePR(13), ago(30*time.Minute), r)
			}(),
			wantStatus: "changes_needed",
			wantGroup:  "your_turn",
		},
		{
			name:       "awaiting_review default",
			pr:         withCommit(basePR(14), ago(30*time.Minute), "SUCCESS"),
			wantStatus: "awaiting_review",
			wantGroup:  "their_turn",
		},
		{
			name: "stale overrides their_turn",
			pr: func() fetch.PR {
				p := withCommit(basePR(15), ago(30*time.Minute), "SUCCESS")
				p.UpdatedAt = staleUpdatedAt()
				return p
			}(),
			wantStatus: "stale",
			wantGroup:  "parked",
		},
		{
			name: "stale does not override your_turn",
			pr: func() fetch.PR {
				p := basePR(16)
				p.Mergeable = "CONFLICTING"
				p.UpdatedAt = staleUpdatedAt()
				return p
			}(),
			wantStatus: "merge_conflict",
			wantGroup:  "your_turn",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := classifyPR(tc.pr, "mine", user, threshold)
			if got.Status != tc.wantStatus {
				t.Errorf("status: got %q, want %q", got.Status, tc.wantStatus)
			}
			if got.Group != tc.wantGroup {
				t.Errorf("group: got %q, want %q", got.Group, tc.wantGroup)
			}
		})
	}
}

func withDirectRequest(pr fetch.PR, login string) fetch.PR {
	pr.ReviewRequests.Nodes = append(pr.ReviewRequests.Nodes, struct {
		RequestedReviewer struct {
			Login string `json:"login"`
			Slug  string `json:"slug"`
		} `json:"requestedReviewer"`
	}{RequestedReviewer: struct {
		Login string `json:"login"`
		Slug  string `json:"slug"`
	}{Login: login}})
	return pr
}

func TestClassifyPR_review(t *testing.T) {
	threshold := 7 * 24 * time.Hour

	cases := []struct {
		name       string
		pr         fetch.PR
		wantStatus string
		wantGroup  string
	}{
		{
			name: "directly requested, no review yet",
			pr: func() fetch.PR {
				p := withCommit(basePR(1), ago(2*time.Hour), "")
				return withDirectRequest(p, user)
			}(),
			wantStatus: "review_requested",
			wantGroup:  "your_turn",
		},
		{
			name:       "team-only request, no review yet → their_turn",
			pr:         withCommit(basePR(2), ago(2*time.Hour), ""),
			wantStatus: "team_review_requested",
			wantGroup:  "their_turn",
		},
		{
			name: "team-only request, stale → parked",
			pr: func() fetch.PR {
				p := withCommit(basePR(3), ago(2*time.Hour), "")
				p.UpdatedAt = staleUpdatedAt()
				return p
			}(),
			wantStatus: "stale",
			wantGroup:  "parked",
		},
		{
			name: "reviewed, no new commits",
			pr: func() fetch.PR {
				p := withCommit(basePR(4), ago(2*time.Hour), "")
				p = withReview(p, user, ago(1*time.Hour), "APPROVED")
				return p
			}(),
			wantStatus: "awaiting_action",
			wantGroup:  "their_turn",
		},
		{
			name: "new commit after my review",
			pr: func() fetch.PR {
				p := withCommit(basePR(5), ago(30*time.Minute), "")
				p = withReview(p, user, ago(2*time.Hour), "APPROVED")
				return p
			}(),
			wantStatus: "re_review_requested",
			wantGroup:  "your_turn",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := classifyPR(tc.pr, "review", user, threshold)
			if got.Status != tc.wantStatus {
				t.Errorf("status: got %q, want %q", got.Status, tc.wantStatus)
			}
			if got.Group != tc.wantGroup {
				t.Errorf("group: got %q, want %q", got.Group, tc.wantGroup)
			}
		})
	}
}

func TestClassifyIssue(t *testing.T) {
	threshold := 7 * 24 * time.Hour

	makeIssue := func(n int) fetch.Issue {
		return fetch.Issue{
			Number:    n,
			Title:     fmt.Sprintf("Issue %d", n),
			URL:       fmt.Sprintf("https://github.com/org/repo/issues/%d", n),
			UpdatedAt: freshUpdatedAt(),
			Repository: struct {
				NameWithOwner string "json:\"nameWithOwner\""
			}{NameWithOwner: "org/repo"},
		}
	}
	addComment := func(issue fetch.Issue, login, createdAt string) fetch.Issue {
		issue.Comments.Nodes = append(issue.Comments.Nodes, struct {
			Author struct {
				Login string "json:\"login\""
			} "json:\"author\""
			CreatedAt string "json:\"createdAt\""
		}{
			Author: struct {
				Login string "json:\"login\""
			}{Login: login},
			CreatedAt: createdAt,
		})
		return issue
	}

	cases := []struct {
		name       string
		issue      fetch.Issue
		wantStatus string
		wantGroup  string
	}{
		{
			name:       "no comments",
			issue:      makeIssue(1),
			wantStatus: "awaiting_response",
			wantGroup:  "their_turn",
		},
		{
			name: "last comment by other user after mine",
			issue: func() fetch.Issue {
				i := makeIssue(2)
				i = addComment(i, user, ago(2*time.Hour))
				i = addComment(i, "bob", ago(1*time.Hour))
				return i
			}(),
			wantStatus: "needs_reply",
			wantGroup:  "your_turn",
		},
		{
			name: "last comment by me",
			issue: func() fetch.Issue {
				i := makeIssue(3)
				i = addComment(i, "bob", ago(2*time.Hour))
				i = addComment(i, user, ago(1*time.Hour))
				return i
			}(),
			wantStatus: "awaiting_response",
			wantGroup:  "their_turn",
		},
		{
			name: "stale",
			issue: func() fetch.Issue {
				i := makeIssue(4)
				i.UpdatedAt = staleUpdatedAt()
				return i
			}(),
			wantStatus: "stale",
			wantGroup:  "parked",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := classifyIssue(tc.issue, user, threshold)
			if got.Status != tc.wantStatus {
				t.Errorf("status: got %q, want %q", got.Status, tc.wantStatus)
			}
			if got.Group != tc.wantGroup {
				t.Errorf("group: got %q, want %q", got.Group, tc.wantGroup)
			}
		})
	}
}

func TestClassifyDiscussion(t *testing.T) {
	threshold := 7 * 24 * time.Hour

	makeDisc := func(n int) fetch.Discussion {
		return fetch.Discussion{
			Number:    n,
			Title:     fmt.Sprintf("Discussion %d", n),
			URL:       fmt.Sprintf("https://github.com/org/repo/discussions/%d", n),
			UpdatedAt: freshUpdatedAt(),
			Repository: struct {
				NameWithOwner string "json:\"nameWithOwner\""
			}{NameWithOwner: "org/repo"},
		}
	}
	addComment := func(disc fetch.Discussion, login, createdAt string) fetch.Discussion {
		disc.Comments.Nodes = append(disc.Comments.Nodes, struct {
			Author struct {
				Login string "json:\"login\""
			} "json:\"author\""
			CreatedAt string "json:\"createdAt\""
		}{
			Author: struct {
				Login string "json:\"login\""
			}{Login: login},
			CreatedAt: createdAt,
		})
		return disc
	}

	cases := []struct {
		name       string
		disc       fetch.Discussion
		wantStatus string
		wantGroup  string
	}{
		{
			name: "answered",
			disc: func() fetch.Discussion {
				d := makeDisc(1)
				d.IsAnswered = true
				return d
			}(),
			wantStatus: "answered",
			wantGroup:  "parked",
		},
		{
			name:       "no comments",
			disc:       makeDisc(2),
			wantStatus: "awaiting_response",
			wantGroup:  "their_turn",
		},
		{
			name: "needs reply",
			disc: func() fetch.Discussion {
				d := makeDisc(3)
				d = addComment(d, user, ago(2*time.Hour))
				d = addComment(d, "bob", ago(1*time.Hour))
				return d
			}(),
			wantStatus: "needs_reply",
			wantGroup:  "your_turn",
		},
		{
			name: "stale",
			disc: func() fetch.Discussion {
				d := makeDisc(4)
				d.UpdatedAt = staleUpdatedAt()
				return d
			}(),
			wantStatus: "stale",
			wantGroup:  "parked",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := classifyDiscussion(tc.disc, user, threshold)
			if got.Status != tc.wantStatus {
				t.Errorf("status: got %q, want %q", got.Status, tc.wantStatus)
			}
			if got.Group != tc.wantGroup {
				t.Errorf("group: got %q, want %q", got.Group, tc.wantGroup)
			}
		})
	}
}

func TestGroup(t *testing.T) {
	items := []state.Item{
		{Status: "ready_to_merge", Group: "your_turn"},
		{Status: "awaiting_review", Group: "their_turn"},
		{Status: "awaiting_review", Group: "their_turn"},
		{Status: "draft", Group: "parked"},
	}
	s := Group(items)
	if s.YourCount != 1 {
		t.Errorf("YourCount: got %d, want 1", s.YourCount)
	}
	if len(s.TheirTurn) != 2 {
		t.Errorf("TheirTurn len: got %d, want 2", len(s.TheirTurn))
	}
	if len(s.Parked) != 1 {
		t.Errorf("Parked len: got %d, want 1", len(s.Parked))
	}
}

func TestRun_deduplicatesReviewPRs(t *testing.T) {
	pr := basePR(1)
	pr.ViewerDidAuthor = true
	data := &fetch.RawData{
		User:        user,
		AuthoredPRs: []fetch.PR{pr},
		ReviewPRs:   []fetch.PR{pr},
	}
	items := Run(data, 7)
	if len(items) != 1 {
		t.Errorf("expected 1 item (dedup), got %d", len(items))
	}
}
