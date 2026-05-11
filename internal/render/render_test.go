package render

import (
	"testing"
	"time"
)

func TestFormatTime(t *testing.T) {
	cases := []struct {
		age  time.Duration
		want string
	}{
		{30 * time.Second, "just now"},
		{5 * time.Minute, "5m ago"},
		{3 * time.Hour, "3h ago"},
		{2 * 24 * time.Hour, "2d ago"},
	}

	for _, tc := range cases {
		ts := time.Now().UTC().Add(-tc.age).Format(time.RFC3339)
		got := formatTime(ts)
		if got != tc.want {
			t.Errorf("formatTime(-%v) = %q, want %q", tc.age, got, tc.want)
		}
	}
}

func TestFormatTime_invalid(t *testing.T) {
	got := formatTime("not-a-time")
	if got != "not-a-time" {
		t.Errorf("expected passthrough for invalid ts, got %q", got)
	}
}
