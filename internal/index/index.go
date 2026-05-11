package index

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/maastrich/gh-next/internal/state"
)

type Entry struct {
	GithubURL string   `json:"github_url"`
	Paths     []string `json:"paths"`
}

type Index struct {
	IndexedAt string  `json:"indexed_at"`
	Entries   []Entry `json:"entries"`
}

var sshRemote = regexp.MustCompile(`^git@github\.com:(.+?)(?:\.git)?$`)
var httpsRemote = regexp.MustCompile(`^https://github\.com/(.+?)(?:\.git)?$`)

func normalizeGitHubURL(remote string) string {
	remote = strings.TrimSpace(remote)
	if m := sshRemote.FindStringSubmatch(remote); len(m) == 2 {
		return "https://github.com/" + m[1]
	}
	if m := httpsRemote.FindStringSubmatch(remote); len(m) == 2 {
		return "https://github.com/" + m[1]
	}
	return ""
}

func remoteURL(repoPath string) string {
	cmd := exec.Command("git", "-C", repoPath, "remote", "get-url", "origin")
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return normalizeGitHubURL(string(out))
}

// Build walks root, finds all git repos with a GitHub origin, and returns an Index.
func Build(root string) (*Index, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("resolve path: %w", err)
	}

	byURL := map[string][]string{}

	err = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable dirs
		}
		if !d.IsDir() {
			return nil
		}
		// skip hidden dirs (other than .git check below) and common noise
		name := d.Name()
		if name != root && strings.HasPrefix(name, ".") {
			return filepath.SkipDir
		}
		if name == "node_modules" || name == "vendor" {
			return filepath.SkipDir
		}

		gitDir := filepath.Join(path, ".git")
		if _, err := os.Stat(gitDir); err != nil {
			return nil // not a git root, keep walking
		}

		url := remoteURL(path)
		if url != "" {
			byURL[url] = append(byURL[url], path)
		}

		return filepath.SkipDir // don't recurse into a repo
	})
	if err != nil {
		return nil, fmt.Errorf("walk %s: %w", root, err)
	}

	idx := &Index{IndexedAt: time.Now().UTC().Format(time.RFC3339)}
	for url, paths := range byURL {
		idx.Entries = append(idx.Entries, Entry{GithubURL: url, Paths: paths})
	}

	return idx, nil
}

func Write(idx *Index) error {
	if err := os.MkdirAll(state.Dir(), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(state.IndexPath(), data, 0644)
}

func Read() (*Index, error) {
	data, err := os.ReadFile(state.IndexPath())
	if err != nil {
		return nil, err
	}
	var idx Index
	return &idx, json.Unmarshal(data, &idx)
}

// PathsFor returns local filesystem paths for a given GitHub repo URL.
func PathsFor(ghURL string) []string {
	idx, err := Read()
	if err != nil {
		return nil
	}
	ghURL = strings.TrimSuffix(strings.TrimSpace(ghURL), ".git")
	for _, e := range idx.Entries {
		if e.GithubURL == ghURL {
			return e.Paths
		}
	}
	return nil
}
