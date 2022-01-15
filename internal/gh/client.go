package gh

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/emirpasic/gods/maps/linkedhashmap"
	"github.com/shurcooL/githubv4"
	"golang.org/x/oauth2"
)

const (
	githubBaseUrl = "https://github.com"
)

func repositoryUrl(owner, repo string) string {
	return fmt.Sprintf("%s/%s/%s", githubBaseUrl, owner, repo)
}

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

type UserPullRequests struct {
	TotalCount int
	Owners     []*UserPullRequestsOwner
}

type UserPullRequestsOwner struct {
	Name         string
	Repositories []*UserPullRequestsRepository
}

type UserPullRequestsRepository struct {
	Name         string
	Description  string
	Url          string
	Watchers     int
	Stars        int
	Forks        int
	LangName     string
	LangColor    string
	PullRequests []*UserPullRequestsPullRequest
}

type UserPullRequestsPullRequest struct {
	Title     string
	State     string
	Number    int
	Url       string
	Additions int
	Deletions int
	Comments  int
	CretaedAt time.Time
	ClosedAt  time.Time
}

type userPullRequestsQuery struct {
	Search struct {
		IssueCount githubv4.Int
		Edges      []userPullRequestsQueryEdge
	} `graphql:"search(query:$searchQuery,type:ISSUE,first:$first,after:$after)"`
}

type userPullRequestsQueryEdge struct {
	Cursor githubv4.String
	Node   struct {
		PullRequest struct {
			Title     githubv4.String
			State     githubv4.String
			Number    githubv4.Int
			Url       githubv4.String
			Additions githubv4.Int
			Deletions githubv4.Int
			Comments  struct {
				TotalCount githubv4.Int
			}
			Reviews struct {
				TotalCount githubv4.Int
			}
			CreatedAt  githubv4.DateTime
			ClosedAt   githubv4.DateTime
			Repository userPullRequestsQueryRepository
		} `graphql:"... on PullRequest"`
	}
}

func newEmptyUserPullRequestsQuery() *userPullRequestsQuery {
	return &userPullRequestsQuery{
		Search: struct {
			IssueCount githubv4.Int
			Edges      []userPullRequestsQueryEdge
		}{
			IssueCount: 0,
			Edges:      make([]userPullRequestsQueryEdge, 0),
		},
	}
}

func (q *userPullRequestsQuery) merge(qq *userPullRequestsQuery) {
	q.Search.IssueCount = qq.Search.IssueCount
	q.Search.Edges = append(q.Search.Edges, qq.Search.Edges...)
}

type userPullRequestsQueryRepository struct {
	Name        githubv4.String
	Description githubv4.String
	Owner       struct {
		Login githubv4.String
	}
	PrimaryLanguage struct {
		Name  githubv4.String
		Color githubv4.String
	}
	Stargazers struct {
		TotalCount githubv4.Int
	}
	Watchers struct {
		TotalCount githubv4.Int
	}
	ForkCount githubv4.Int
}

func (q *userPullRequestsQuery) toUserPullRequests() *UserPullRequests {
	repoNodesMap := linkedhashmap.New()
	for _, edge := range q.Search.Edges {
		repo := edge.Node.PullRequest.Repository
		ownerName := string(repo.Owner.Login)
		repoName := string(repo.Name)
		key := fmt.Sprintf("%s/%s", ownerName, repoName)
		if _, ok := repoNodesMap.Get(key); !ok {
			repoNodesMap.Put(key, repo)
		}
	}

	ownerMap := linkedhashmap.New()
	for _, edge := range q.Search.Edges {
		pn := edge.Node.PullRequest
		ownerName := string(pn.Repository.Owner.Login)
		repoName := string(pn.Repository.Name)
		repoMap, ok := ownerMap.Get(ownerName)
		if !ok {
			repoMap = linkedhashmap.New()
			ownerMap.Put(ownerName, repoMap)
		}
		prs, ok := repoMap.(*linkedhashmap.Map).Get(repoName)
		if !ok {
			prs = make([]*UserPullRequestsPullRequest, 0)
		}
		pullRequest := &UserPullRequestsPullRequest{
			Title:     string(pn.Title),
			State:     string(pn.State),
			Number:    int(pn.Number),
			Url:       string(pn.Url),
			Additions: int(pn.Additions),
			Deletions: int(pn.Deletions),
			Comments:  int(pn.Comments.TotalCount),
			CretaedAt: pn.CreatedAt.Time,
			ClosedAt:  pn.ClosedAt.Time,
		}
		pullRequests := prs.([]*UserPullRequestsPullRequest)
		pullRequests = append(pullRequests, pullRequest)
		repoMap.(*linkedhashmap.Map).Put(repoName, pullRequests)
	}

	owners := make([]*UserPullRequestsOwner, 0)
	for _, ownerName := range ownerMap.Keys() {
		repositories := make([]*UserPullRequestsRepository, 0)
		repoMap, _ := ownerMap.Get(ownerName)
		for _, repoName := range repoMap.(*linkedhashmap.Map).Keys() {
			key := fmt.Sprintf("%s/%s", ownerName, repoName)
			repoNode, _ := repoNodesMap.Get(key)
			rn := repoNode.(userPullRequestsQueryRepository)
			prs, _ := repoMap.(*linkedhashmap.Map).Get(repoName)
			repository := &UserPullRequestsRepository{
				Name:         string(rn.Name),
				Description:  string(rn.Description),
				Url:          repositoryUrl(ownerName.(string), repoName.(string)),
				Watchers:     int(rn.Watchers.TotalCount),
				Stars:        int(rn.Stargazers.TotalCount),
				Forks:        int(rn.ForkCount),
				LangName:     string(rn.PrimaryLanguage.Name),
				LangColor:    string(rn.PrimaryLanguage.Color),
				PullRequests: prs.([]*UserPullRequestsPullRequest),
			}
			repositories = append(repositories, repository)
		}
		owner := &UserPullRequestsOwner{
			Name:         ownerName.(string),
			Repositories: repositories,
		}
		owners = append(owners, owner)
	}
	ret := &UserPullRequests{
		TotalCount: int(q.Search.IssueCount),
		Owners:     owners,
	}
	return ret
}

func (c *GitHubClient) QueryUserPullRequests(id string) (*UserPullRequests, error) {
	q := newEmptyUserPullRequestsQuery()
	issueCount := math.MaxInt32
	total := 0
	cursor := ""
	for total < issueCount {
		qq, err := c.queryUserPullRequests(id, cursor)
		if err != nil {
			return nil, err
		}
		issueCount = int(qq.Search.IssueCount)
		edges := qq.Search.Edges
		cursor = string(edges[len(edges)-1].Cursor)
		total += len(edges)
		q.merge(qq)
	}
	return q.toUserPullRequests(), nil
}

func (c *GitHubClient) queryUserPullRequests(id string, cursorAfter string) (*userPullRequestsQuery, error) {
	searchQuery := fmt.Sprintf("author:%s -user:%s is:pr sort:created-desc", id, id)
	var query userPullRequestsQuery
	variables := map[string]interface{}{
		"searchQuery": githubv4.String(searchQuery),
		"first":       githubv4.Int(50),
	}
	if cursorAfter == "" {
		variables["after"] = (*githubv4.String)(nil)
	} else {
		variables["after"] = githubv4.String(cursorAfter)
	}
	if err := c.client.Query(context.Background(), &query, variables); err != nil {
		return nil, err
	}
	return &query, nil
}
