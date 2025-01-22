package github

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

const APIBaseUrl = "https://api.github.com/"

// GitHub contains the functions necessary for interacting with GitHub release
// objects
type GitHub interface {
	GetRelease(ctx context.Context, tag string) (*github.RepositoryRelease, error)
}

// Client is the client for interacting with the GitHub API
type Client struct {
	Owner, Repo string
	*github.Client
}

// NewClient creates and initializes a new GitHubClient
func NewClient(owner, repo, token, urlStr string) (GitHub, error) {
	var client *github.Client

	if len(owner) == 0 {
		return nil, fmt.Errorf("missing GitHub repository owner")
	}

	if len(repo) == 0 {
		return nil, fmt.Errorf("missing GitHub repository name")
	}

	baseURL, err := url.ParseRequestURI(urlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Github API URL: %v", err)
	}

	if len(token) > 0 {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		tc := oauth2.NewClient(context.TODO(), ts)
		client = github.NewClient(tc)
	} else {
		client = github.NewClient(&http.Client{})
	}

	client.BaseURL = baseURL

	return &Client{Owner: owner, Repo: repo, Client: client}, nil
}

// GetRelease queries the GitHub API for a specified release object
func (c *Client) GetRelease(ctx context.Context, tag string) (*github.RepositoryRelease, error) {
	// Check Release whether already exists or not
	errHandler := func(release *github.RepositoryRelease, res *github.Response, err error) (*github.RepositoryRelease, error) {
		if err != nil {
			if res == nil {
				return nil, fmt.Errorf("failed to get release version for repository: %s", c.Repo)
			}

			switch {
			case res.StatusCode == http.StatusUnauthorized:
				return nil, fmt.Errorf("wrong token is specified or there is no permission, invalid status: %s", res.Status)
			case res.StatusCode != http.StatusNotFound:
				return nil, fmt.Errorf("repository %s not found, invalid status: %s", c.Repo, res.Status)
			}

			return nil, fmt.Errorf("release version %s not found for repository: %s", tag, c.Repo)
		}

		return release, nil
	}

	if len(tag) == 0 {
		return errHandler(c.Repositories.GetLatestRelease(ctx, c.Owner, c.Repo))
	} else {
		return errHandler(c.Repositories.GetReleaseByTag(ctx, c.Owner, c.Repo, tag))
	}
}
