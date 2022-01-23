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

func (c *GitHubClient) ExistUser(id string) bool {
	var query struct {
		User struct {
			Login githubv4.String
		} `graphql:"user(login: $login)"`
	}
	variables := map[string]interface{}{
		"login": githubv4.String(id),
	}
	err := c.client.Query(context.Background(), &query, variables)
	return err == nil
}

type UserProfile struct {
	Login      string
	Name       string
	Bio        string
	Followers  int
	Following  int
	Location   string
	Company    string
	WebsiteUrl string
	AvatarUrl  string
	Url        string
}

type userProfileQuery struct {
	User struct {
		Login     githubv4.String
		Name      githubv4.String
		Bio       githubv4.String
		Followers struct {
			TotalCount githubv4.Int
		}
		Following struct {
			TotalCount githubv4.Int
		}
		Location   githubv4.String
		Company    githubv4.String
		WebsiteUrl githubv4.String
		AvatarUrl  githubv4.String
		Url        githubv4.String
	} `graphql:"user(login: $login)"`
}

func (q *userProfileQuery) toUserProfile() *UserProfile {
	return &UserProfile{
		Login:      string(q.User.Login),
		Name:       string(q.User.Name),
		Bio:        string(q.User.Bio),
		Followers:  int(q.User.Followers.TotalCount),
		Following:  int(q.User.Following.TotalCount),
		Location:   string(q.User.Location),
		Company:    string(q.User.Company),
		WebsiteUrl: string(q.User.WebsiteUrl),
		AvatarUrl:  string(q.User.AvatarUrl),
		Url:        string(q.User.Url),
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

func (p *UserPullRequests) Owner(owner string) *UserPullRequestsOwner {
	for _, o := range p.Owners {
		if o.Name == owner {
			return o
		}
	}
	return nil
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

type userPullRequestsQueryRepository struct {
	Name        githubv4.String
	Description githubv4.String
	Url         githubv4.String
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

type repoNodesMap struct {
	*linkedhashmap.Map
}

func newRepoNodesMap() *repoNodesMap {
	return &repoNodesMap{linkedhashmap.New()}
}

func (m *repoNodesMap) Exist(key string) bool {
	_, ok := m.Map.Get(key)
	return ok
}

func (m *repoNodesMap) Get(key string) userPullRequestsQueryRepository {
	node, _ := m.Map.Get(key)
	return node.(userPullRequestsQueryRepository)
}

func (m *repoNodesMap) Put(key string, value userPullRequestsQueryRepository) {
	m.Map.Put(key, value)
}

type ownerMap struct {
	*linkedhashmap.Map
}

func newOwnerMap() *ownerMap {
	return &ownerMap{linkedhashmap.New()}
}

func (m *ownerMap) Exist(key string) bool {
	_, ok := m.Map.Get(key)
	return ok
}

func (m *ownerMap) Get(key string) *repoMap {
	prs, _ := m.Map.Get(key)
	return prs.(*repoMap)
}

func (m *ownerMap) Put(key string, value *repoMap) {
	m.Map.Put(key, value)
}

func (m *ownerMap) Keys() []string {
	keys := m.Map.Keys()
	ret := make([]string, len(keys))
	for i, key := range keys {
		ret[i] = key.(string)
	}
	return ret
}

type repoMap struct {
	*linkedhashmap.Map
}

func newRepoMap() *repoMap {
	return &repoMap{linkedhashmap.New()}
}

func (m *repoMap) Exist(key string) bool {
	_, ok := m.Map.Get(key)
	return ok
}

func (m *repoMap) Get(key string) []*UserPullRequestsPullRequest {
	prs, _ := m.Map.Get(key)
	return prs.([]*UserPullRequestsPullRequest)
}

func (m *repoMap) Put(key string, value []*UserPullRequestsPullRequest) {
	m.Map.Put(key, value)
}

func (m *repoMap) Keys() []string {
	keys := m.Map.Keys()
	ret := make([]string, len(keys))
	for i, key := range keys {
		ret[i] = key.(string)
	}
	return ret
}

func (q *userPullRequestsQuery) toUserPullRequests() *UserPullRequests {
	rnMap := newRepoNodesMap()
	for _, edge := range q.Search.Edges {
		repo := edge.Node.PullRequest.Repository
		ownerName := string(repo.Owner.Login)
		repoName := string(repo.Name)
		key := fmt.Sprintf("%s/%s", ownerName, repoName)
		if !rnMap.Exist(key) {
			rnMap.Put(key, repo)
		}
	}

	ownerMap := newOwnerMap()
	for _, edge := range q.Search.Edges {
		pn := edge.Node.PullRequest
		ownerName := string(pn.Repository.Owner.Login)
		repoName := string(pn.Repository.Name)
		if !ownerMap.Exist(ownerName) {
			ownerMap.Put(ownerName, newRepoMap())
		}
		repoMap := ownerMap.Get(ownerName)
		if !repoMap.Exist(repoName) {
			repoMap.Put(repoName, make([]*UserPullRequestsPullRequest, 0))
		}
		pullRequests := repoMap.Get(repoName)
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
		pullRequests = append(pullRequests, pullRequest)
		repoMap.Put(repoName, pullRequests)
	}

	owners := make([]*UserPullRequestsOwner, 0)
	for _, ownerName := range ownerMap.Keys() {
		repositories := make([]*UserPullRequestsRepository, 0)
		repoMap := ownerMap.Get(ownerName)
		for _, repoName := range repoMap.Keys() {
			key := fmt.Sprintf("%s/%s", ownerName, repoName)
			rn := rnMap.Get(key)
			prs := repoMap.Get(repoName)
			repository := &UserPullRequestsRepository{
				Name:         string(rn.Name),
				Description:  string(rn.Description),
				Url:          string(rn.Url),
				Watchers:     int(rn.Watchers.TotalCount),
				Stars:        int(rn.Stargazers.TotalCount),
				Forks:        int(rn.ForkCount),
				LangName:     string(rn.PrimaryLanguage.Name),
				LangColor:    string(rn.PrimaryLanguage.Color),
				PullRequests: prs,
			}
			repositories = append(repositories, repository)
		}
		owner := &UserPullRequestsOwner{
			Name:         ownerName,
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
		if issueCount == 0 {
			break
		}
		edges := qq.Search.Edges
		cursor = string(edges[len(edges)-1].Cursor)
		total += len(edges)
		q.merge(qq)
	}
	return q.toUserPullRequests(), nil
}

func (c *GitHubClient) queryUserPullRequests(id, cursorAfter string) (*userPullRequestsQuery, error) {
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

func (c *GitHubClient) QueryUserRepositories(id string) (*UserRepositories, error) {
	q, err := c.queryUserRepositories(id, "")
	if err != nil {
		return nil, err
	}
	hasNext := bool(q.User.Repositories.PageInfo.HasNextPage)
	cursor := string(q.User.Repositories.PageInfo.EndCursor)
	for hasNext {
		qq, err := c.queryUserRepositories(id, cursor)
		if err != nil {
			return nil, err
		}
		hasNext = bool(qq.User.Repositories.PageInfo.HasNextPage)
		cursor = string(qq.User.Repositories.PageInfo.EndCursor)
		q.merge(qq)
	}
	return q.toUserRepositories(), nil
}

func (c *GitHubClient) queryUserRepositories(id, cursorAfter string) (*userRepositoriesQuery, error) {
	var query userRepositoriesQuery
	variables := map[string]interface{}{
		"login": githubv4.String(id),
		"first": githubv4.Int(50),
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

type UserRepositories struct {
	TotalCount   int
	Repositories []*UserRepository
}

type UserRepository struct {
	Name               string
	Description        string
	Url                string
	Watchers           int
	Stars              int
	Forks              int
	LangName           string
	LangColor          string
	OpenedIssues       int
	OpenedPullRequests int
	License            string
	CreatedAt          time.Time
	PushedAt           time.Time
}

type userRepositoriesQuery struct {
	User struct {
		Repositories struct {
			TotalCount githubv4.Int
			PageInfo   pageInfo
			Edges      []userRepositoriesQueryEdge
		} `graphql:"repositories(orderBy:{direction:DESC,field:STARGAZERS},privacy:PUBLIC,isFork:false,first:$first,after:$after)"`
	} `graphql:"user(login:$login)"`
}

func (q *userRepositoriesQuery) merge(qq *userRepositoriesQuery) {
	q.User.Repositories.TotalCount = qq.User.Repositories.TotalCount
	q.User.Repositories.PageInfo = qq.User.Repositories.PageInfo
	q.User.Repositories.Edges = append(q.User.Repositories.Edges, qq.User.Repositories.Edges...)
}

type pageInfo struct {
	EndCursor       githubv4.String
	HasNextPage     githubv4.Boolean
	HasPreviousPage githubv4.Boolean
	StartCursor     githubv4.String
}

type userRepositoriesQueryEdge struct {
	Cursor githubv4.String
	Node   struct {
		Name            githubv4.String
		Description     githubv4.String
		Url             githubv4.String
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
		Issues struct {
			TotalCount githubv4.Int
		} `graphql:"issues(states:OPEN)"`
		PullRequests struct {
			TotalCount githubv4.Int
		} `graphql:"pullRequests(states:OPEN)"`
		ForkCount   githubv4.Int
		LicenseInfo struct {
			Name   githubv4.String
			SpdxId githubv4.String // https://spdx.org/licenses
		}
		IsArchived githubv4.Boolean
		IsFork     githubv4.Boolean
		IsPrivate  githubv4.Boolean
		IsTemplate githubv4.Boolean
		PushedAt   githubv4.DateTime
		CreatedAt  githubv4.DateTime
	}
}

func (q *userRepositoriesQuery) toUserRepositories() *UserRepositories {
	repositories := make([]*UserRepository, 0)
	for _, edge := range q.User.Repositories.Edges {
		r := edge.Node
		repository := &UserRepository{
			Name:               string(r.Name),
			Description:        string(r.Description),
			Url:                string(r.Url),
			Watchers:           int(r.Watchers.TotalCount),
			Stars:              int(r.Stargazers.TotalCount),
			Forks:              int(r.ForkCount),
			LangName:           string(r.PrimaryLanguage.Name),
			LangColor:          string(r.PrimaryLanguage.Color),
			OpenedIssues:       int(r.Issues.TotalCount),
			OpenedPullRequests: int(r.PullRequests.TotalCount),
			License:            string(r.LicenseInfo.SpdxId),
			CreatedAt:          r.CreatedAt.Time,
			PushedAt:           r.PushedAt.Time,
		}
		repositories = append(repositories, repository)
	}
	return &UserRepositories{
		TotalCount:   int(q.User.Repositories.TotalCount),
		Repositories: repositories,
	}
}
