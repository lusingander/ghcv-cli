package gh

import (
	"context"

	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

type GitHubClient struct {
	client *githubv4.Client
}

func NewGitHubClient(cfg *GithubConfig) *GitHubClient {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: cfg.AccessToken},
	)
	httpClient := oauth2.NewClient(context.Background(), src)
	client := githubv4.NewClient(httpClient)
	return &GitHubClient{
		client: client,
	}
}

type UserProfile struct {
	Login      string
	Name       string
	Location   string
	Company    string
	WebsiteUrl string
	AvatarUrl  string
}

type userProfileQuery struct {
	User struct {
		Login      githubv4.String
		Name       githubv4.String
		Location   githubv4.String
		Company    githubv4.String
		WebsiteUrl githubv4.String
		AvatarUrl  githubv4.String
	} `graphql:"user(login: $login)"`
}

func (q *userProfileQuery) toUserProfile() *UserProfile {
	return &UserProfile{
		Login:      string(q.User.Login),
		Name:       string(q.User.Name),
		Location:   string(q.User.Location),
		Company:    string(q.User.Company),
		WebsiteUrl: string(q.User.WebsiteUrl),
		AvatarUrl:  string(q.User.AvatarUrl),
	}
}

func (c *GitHubClient) QueryUserProfile(id string) (*UserProfile, error) {
	var query userProfileQuery
	variables := map[string]interface{}{
		"login": githubv4.String(id),
	}
	if err := c.client.Query(context.Background(), &query, variables); err != nil {
		return nil, err
	}
	return query.toUserProfile(), nil
}
