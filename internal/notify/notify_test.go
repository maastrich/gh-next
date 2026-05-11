package notify

import (
	"testing"

	"github.com/maastrich/gh-next/internal/state"
)

func TestBuildSummaryParts(t *testing.T) {
	cases := []struct {
		name  string
		items []state.Item
		want  []string
	}{
		{
			name:  "empty",
			items: nil,
			want:  nil,
		},
		{
			name:  "single merge conflict",
			items: []state.Item{{Status: "merge_conflict"}},
			want:  []string{"⚔️ 1"},
		},
		{
			name: "multiple statuses ordered",
			items: []state.Item{
				{Status: "ready_to_merge"},
				{Status: "merge_conflict"},
				{Status: "changes_needed"},
				{Status: "needs_reply"},
			},
			want: []string{"⚔️ 1", "✅ 1", "🔧 1", "💬 1"},
		},
		{
			name: "awaiting_answer and needs_reply share emoji, sum separately",
			items: []state.Item{
				{Status: "awaiting_answer"},
				{Status: "needs_reply"},
				{Status: "awaiting_answer"},
			},
			want: []string{"💬 2", "💬 1"},
		},
		{
			name:  "review_requested",
			items: []state.Item{{Status: "review_requested"}},
			want:  []string{"👀 1"},
		},
		{
			name:  "re_review_requested",
			items: []state.Item{{Status: "re_review_requested"}},
			want:  []string{"🔁 1"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := buildSummaryParts(tc.items)
			if len(got) != len(tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("[%d] got %q, want %q", i, got[i], tc.want[i])
				}
			}
		})
	}
}
