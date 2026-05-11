package cmd

import "testing"

func TestStripCronBlock(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no marker",
			input: "0 * * * * /usr/bin/something\n",
			want:  "0 * * * * /usr/bin/something\n",
		},
		{
			name:  "marker with entry removed",
			input: "other job\n" + cronMarker + "\n0 8-18 * * 1-5 /usr/bin/gh next status\n",
			want:  "other job\n",
		},
		{
			name:  "only marker block",
			input: cronMarker + "\n0 8-18 * * 1-5 /usr/bin/gh next status\n",
			want:  "",
		},
		{
			name:  "marker at end without entry",
			input: "other\n" + cronMarker + "\n",
			want:  "other\n",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := stripCronBlock(tc.input)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestShowCronLine(t *testing.T) {
	cases := []struct {
		name    string
		crontab string
		want    string
	}{
		{
			name:    "present",
			crontab: cronMarker + "\n0 8-18 * * 1-5 /usr/bin/gh next status\n",
			want:    "0 8-18 * * 1-5 /usr/bin/gh next status",
		},
		{
			name:    "absent",
			crontab: "something else\n",
			want:    "",
		},
		{
			name:    "empty",
			crontab: "",
			want:    "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := showCronLine(tc.crontab)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}
