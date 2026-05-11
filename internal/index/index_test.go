package index

import "testing"

func TestNormalizeGitHubURL(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"git@github.com:org/repo.git", "https://github.com/org/repo"},
		{"git@github.com:org/repo", "https://github.com/org/repo"},
		{"https://github.com/org/repo.git", "https://github.com/org/repo"},
		{"https://github.com/org/repo", "https://github.com/org/repo"},
		{"git@github.com:org/repo.git\n", "https://github.com/org/repo"},
		{"https://gitlab.com/org/repo.git", ""},
		{"not-a-remote", ""},
		{"", ""},
	}

	for _, tc := range cases {
		got := normalizeGitHubURL(tc.input)
		if got != tc.want {
			t.Errorf("normalizeGitHubURL(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
