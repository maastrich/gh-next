package fetch

import (
	"fmt"
	"time"

	"github.com/cli/go-gh/v2/pkg/api"
)

type RawData struct {
	User        string
	FetchedAt   string
	AuthoredPRs []PR
	ReviewPRs   []PR
	Issues      []Issue
	Discussions []Discussion
}

type CheckContext struct {
	// CheckRun fields
	Conclusion string `json:"conclusion"`
	DetailsUrl string `json:"detailsUrl"`
	// StatusContext fields
	State       string `json:"state"`
	TargetUrl   string `json:"targetUrl"`
	Description string `json:"description"`
}

type CheckRollup struct {
	State    string `json:"state"`
	Contexts struct {
		Nodes []CheckContext `json:"nodes"`
	} `json:"contexts"`
}

type PR struct {
	Number          int    `json:"number"`
	Title           string `json:"title"`
	URL             string `json:"url"`
	IsDraft         bool   `json:"isDraft"`
	UpdatedAt       string `json:"updatedAt"`
	Mergeable       string `json:"mergeable"`
	ReviewDecision  string `json:"reviewDecision"`
	ViewerCanUpdate bool   `json:"viewerCanUpdate"`
	ViewerDidAuthor bool   `json:"viewerDidAuthor"`
	Repository      struct {
		NameWithOwner string `json:"nameWithOwner"`
	} `json:"repository"`
	Commits struct {
		Nodes []struct {
			Commit struct {
				CommittedDate     string       `json:"committedDate"`
				StatusCheckRollup *CheckRollup `json:"statusCheckRollup"`
			} `json:"commit"`
		} `json:"nodes"`
	} `json:"commits"`
	Reviews struct {
		Nodes []struct {
			Author struct {
				Login    string `json:"login"`
				Typename string `json:"__typename"`
			} `json:"author"`
			State       string `json:"state"`
			SubmittedAt string `json:"submittedAt"`
			Comments    struct {
				TotalCount int `json:"totalCount"`
			} `json:"comments"`
		} `json:"nodes"`
	} `json:"reviews"`
	ReviewRequests struct {
		Nodes []struct {
			RequestedReviewer struct {
				Login string `json:"login"`
				Slug  string `json:"slug"`
			} `json:"requestedReviewer"`
		} `json:"nodes"`
	} `json:"reviewRequests"`
}

type Issue struct {
	Number     int    `json:"number"`
	Title      string `json:"title"`
	URL        string `json:"url"`
	UpdatedAt  string `json:"updatedAt"`
	Repository struct {
		NameWithOwner string `json:"nameWithOwner"`
	} `json:"repository"`
	Comments struct {
		Nodes []struct {
			Author struct {
				Login string `json:"login"`
			} `json:"author"`
			CreatedAt string `json:"createdAt"`
		} `json:"nodes"`
	} `json:"comments"`
}

type Discussion struct {
	Number     int    `json:"number"`
	Title      string `json:"title"`
	URL        string `json:"url"`
	UpdatedAt  string `json:"updatedAt"`
	IsAnswered bool   `json:"isAnswered"`
	Repository struct {
		NameWithOwner string `json:"nameWithOwner"`
	} `json:"repository"`
	Comments struct {
		Nodes []struct {
			Author struct {
				Login string `json:"login"`
			} `json:"author"`
			CreatedAt string `json:"createdAt"`
		} `json:"nodes"`
	} `json:"comments"`
}

const prFragment = `
... on PullRequest {
  number title url isDraft updatedAt mergeable
  reviewDecision viewerCanUpdate viewerDidAuthor
  repository { nameWithOwner }
  commits(last: 1) {
    nodes { commit { committedDate statusCheckRollup { state contexts(last: 20) { nodes {
      ... on CheckRun { conclusion detailsUrl }
      ... on StatusContext { state targetUrl description }
    } } } } }
  }
  reviews(last: 30) {
    nodes {
      author { login __typename }
      state submittedAt
      comments { totalCount }
    }
  }
  reviewRequests(first: 10) {
    nodes { requestedReviewer { ... on User { login } ... on Team { slug } } }
  }
}`

const issueFragment = `
... on Issue {
  number title url updatedAt
  repository { nameWithOwner }
  comments(last: 10) {
    nodes { author { login } createdAt }
  }
}`

const discussionFragment = `
... on Discussion {
  number title url updatedAt isAnswered
  repository { nameWithOwner }
  comments(last: 10) {
    nodes { author { login } createdAt }
  }
}`

func Fetch(user string, includeArchived bool) (*RawData, error) {
	client, err := api.DefaultGraphQLClient()
	if err != nil {
		return nil, fmt.Errorf("graphql client: %w", err)
	}

	type result[T any] struct {
		items []T
		err   error
	}

	archivedFilter := " archived:false"
	if includeArchived {
		archivedFilter = ""
	}

	authored := make(chan result[PR], 1)
	reviews := make(chan result[PR], 1)
	issues := make(chan result[Issue], 1)
	discussions := make(chan result[Discussion], 1)

	go func() {
		items, err := searchPRs(client, fmt.Sprintf("is:open is:pr author:%s%s", user, archivedFilter))
		authored <- result[PR]{items, err}
	}()
	go func() {
		items, err := searchPRs(client, fmt.Sprintf("is:open is:pr review-requested:%s%s", user, archivedFilter))
		reviews <- result[PR]{items, err}
	}()
	go func() {
		items, err := searchIssues(client, fmt.Sprintf("is:open is:issue author:%s%s", user, archivedFilter))
		issues <- result[Issue]{items, err}
	}()
	go func() {
		items, err := searchDiscussions(client, fmt.Sprintf("is:open author:%s%s", user, archivedFilter))
		discussions <- result[Discussion]{items, err}
	}()

	ar := <-authored
	rr := <-reviews
	ir := <-issues
	dr := <-discussions

	for _, e := range []error{ar.err, rr.err, ir.err, dr.err} {
		if e != nil {
			return nil, e
		}
	}

	return &RawData{
		User:        user,
		FetchedAt:   time.Now().UTC().Format(time.RFC3339),
		AuthoredPRs: ar.items,
		ReviewPRs:   rr.items,
		Issues:      ir.items,
		Discussions: dr.items,
	}, nil
}

func searchPRs(client *api.GraphQLClient, query string) ([]PR, error) {
	gql := fmt.Sprintf(`{ search(query: %q, type: ISSUE, first: 100) { nodes { %s } } }`, query, prFragment)
	var resp struct {
		Search struct {
			Nodes []PR `json:"nodes"`
		} `json:"search"`
	}
	if err := client.Do(gql, nil, &resp); err != nil {
		return nil, err
	}
	var out []PR
	for _, n := range resp.Search.Nodes {
		if n.Number != 0 {
			out = append(out, n)
		}
	}
	return out, nil
}

func searchIssues(client *api.GraphQLClient, query string) ([]Issue, error) {
	gql := fmt.Sprintf(`{ search(query: %q, type: ISSUE, first: 100) { nodes { %s } } }`, query, issueFragment)
	var resp struct {
		Search struct {
			Nodes []Issue `json:"nodes"`
		} `json:"search"`
	}
	if err := client.Do(gql, nil, &resp); err != nil {
		return nil, err
	}
	var out []Issue
	for _, n := range resp.Search.Nodes {
		if n.Number != 0 {
			out = append(out, n)
		}
	}
	return out, nil
}

func searchDiscussions(client *api.GraphQLClient, query string) ([]Discussion, error) {
	gql := fmt.Sprintf(`{ search(query: %q, type: DISCUSSION, first: 100) { nodes { %s } } }`, query, discussionFragment)
	var resp struct {
		Search struct {
			Nodes []Discussion `json:"nodes"`
		} `json:"search"`
	}
	if err := client.Do(gql, nil, &resp); err != nil {
		return nil, err
	}
	var out []Discussion
	for _, n := range resp.Search.Nodes {
		if n.Number != 0 {
			out = append(out, n)
		}
	}
	return out, nil
}
