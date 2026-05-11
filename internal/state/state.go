package state

import (
	"encoding/json"
	"os"
	"path/filepath"
)

const (
	DirName     = ".gh-next"
	SummaryFile = "summary.json"
	StateFile   = "state.json"
	HTMLFile    = "summary.html"
	LogFile     = "run.log"
)

const IndexFile = "gh-index.json"

func Dir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, DirName)
}

func SummaryPath() string { return filepath.Join(Dir(), SummaryFile) }
func StatePath() string   { return filepath.Join(Dir(), StateFile) }
func HTMLPath() string    { return filepath.Join(Dir(), HTMLFile) }
func LogPath() string     { return filepath.Join(Dir(), LogFile) }
func IndexPath() string   { return filepath.Join(Dir(), IndexFile) }

type Item struct {
	Number             int      `json:"number"`
	Title              string   `json:"title"`
	URL                string   `json:"url"`
	Repo               string   `json:"repo"`
	Kind               string   `json:"kind"` // pr | issue | discussion
	Status             string   `json:"status"`
	Icon               string   `json:"icon"`
	Group              string   `json:"group"` // your_turn | their_turn | parked
	UpdatedAt          string   `json:"updatedAt"`
	CIState            string   `json:"ciState,omitempty"`            // SUCCESS | FAILURE | ERROR | PENDING | IN_PROGRESS
	Approvals          int      `json:"approvals,omitempty"`          // count of APPROVED reviews
	FailedChecks       []string `json:"failedChecks,omitempty"`       // names of actionable failed checks
	ChangesRequestedBy []string `json:"changesRequestedBy,omitempty"` // logins who requested changes
	RequestedBy        []string `json:"requestedBy,omitempty"`        // direct review requesters
}

type Summary struct {
	YourTurn  []Item `json:"your_turn"`
	TheirTurn []Item `json:"their_turn"`
	Parked    []Item `json:"parked"`
	YourCount int    `json:"your_count"`
	UpdatedAt string `json:"updated_at"`
}

type PrevState struct {
	Items     []Item `json:"prs"`
	YourCount int    `json:"your_count"`
	TS        string `json:"ts"`
}

func ReadSummary() (*Summary, error) {
	data, err := os.ReadFile(SummaryPath())
	if err != nil {
		return nil, err
	}
	var s Summary
	return &s, json.Unmarshal(data, &s)
}

func WriteSummary(s *Summary) error {
	if err := os.MkdirAll(Dir(), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(SummaryPath(), data, 0644)
}

func ReadPrevState() *PrevState {
	data, err := os.ReadFile(StatePath())
	if err != nil {
		return &PrevState{}
	}
	var ps PrevState
	if err := json.Unmarshal(data, &ps); err != nil {
		return &PrevState{}
	}
	return &ps
}

func WritePrevState(items []Item, yourCount int, ts string) error {
	ps := PrevState{Items: items, YourCount: yourCount, TS: ts}
	data, err := json.MarshalIndent(ps, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(StatePath(), data, 0644)
}
