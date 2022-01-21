package gh

import (
	"reflect"
	"testing"

	"github.com/shurcooL/githubv4"
)

func equal(x, y interface{}) bool {
	return reflect.DeepEqual(x, y)
}

func notEqual(x, y interface{}) bool {
	return !equal(x, y)
}

func Test_userProfileQuery_toUserProfile(t *testing.T) {
	q := &userProfileQuery{
		User: struct {
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
		}{
			Login: "foo",
			Name:  "foo bar",
			Bio:   "bar",
			Followers: struct {
				TotalCount githubv4.Int
			}{
				TotalCount: 123,
			},
			Following: struct {
				TotalCount githubv4.Int
			}{
				TotalCount: 456,
			},
			Location:   "japan",
			Company:    "baz",
			WebsiteUrl: "http://example.com/qux",
			AvatarUrl:  "http://example.com/foo",
		},
	}
	want := &UserProfile{
		Login:      "foo",
		Name:       "foo bar",
		Bio:        "bar",
		Followers:  123,
		Following:  456,
		Location:   "japan",
		Company:    "baz",
		WebsiteUrl: "http://example.com/qux",
		AvatarUrl:  "http://example.com/foo",
	}
	got := q.toUserProfile()
	if notEqual(got, want) {
		t.Errorf("got: %v, want: %v", got, want)
	}
}
