//
// router_test.go
// Copyright (C) 2024 veypi <i@veypi.com>
// 2024-08-08 18:20
// Distributed under terms of the MIT license.
//

package rest

import (
	"net/http"
	"testing"

	"github.com/veypi/utils/logx"
)

const paramPrefix = "paramPrefix"
const checkPath = false
const twentyColon = "/:a/:b/:c/:d/:e/:f/:g/:h/:i/:j/:k/:l/:m/:n/:o/:p/:q/:r/:s/:t"
const twentyBrace = "/{a}/{b}/{c}/{d}/{e}/{f}/{g}/{h}/{i}/{j}/{k}/{l}/{m}/{n}/{o}/{p}/{q}/{r}/{s}/{t}"
const twentyRoute = "/a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p/q/r/s/t"

type fakeResponseWriter struct {
	d string
}

func (f *fakeResponseWriter) Header() http.Header {
	return nil
}

func (f *fakeResponseWriter) Write(p []byte) (int, error) {
	f.d = string(p)
	return len(p), nil
}

func (f *fakeResponseWriter) WriteHeader(statusCode int) {
	return
}

func githubRouter() Router {
	r := NewRouter()
	r.Use(func(x *X) error {
		logx.Info().Int("id", 1).Str("p", x.Request.URL.Path).Msg(x.Params[0][0])
		err := x.Next()
		logx.Info().Int("id", 10).Err(err).Str("p", x.Request.URL.Path).Msg(x.Params[0][1])
		return err
	})
	r.Use(func(x *X) error {
		logx.Info().Int("id", 2).Str("p", x.Request.URL.Path).Msg(x.Params.GetStr("sha"))
		return nil
	})
	for _, api := range githubAPi {
		for _, m := range api.methods {
			r.Set(api.path, m, func(x *X) error {
				logx.Info().Int("id", 0).Str("p", x.Request.URL.Path).Msg("0")
				x.Write([]byte(x.Request.URL.Path))
				return nil
			})
		}
	}
	r.Use(func(x *X) error {
		logx.Info().Int("id", 3).Str("p", x.Request.URL.Path).Msg("0")
		// return errors.New("123")
		return nil
	})
	r.Use(func(x *X) error {
		logx.Info().Int("id", 4).Str("p", x.Request.URL.Path).Msg("0")
		return nil
	})
	r.Use(func(x *X) error {
		logx.Info().Int("id", 5).Str("p", x.Request.URL.Path).Msg(x.Params.GetStr("owner"))
		return nil
	})
	return r
}

var testR Router

func init() {
	logx.SetLevel(logx.WarnLevel)
	testR = githubRouter()
}

var req *http.Request
var temPath = ""
var w = new(fakeResponseWriter)

func BenchmarkRoute_GitHub_ALL(b *testing.B) {
	req, _ = http.NewRequest("GET", "/", nil)
	for i := 0; i < b.N; i++ {
		for _, api := range githubAPi {
			req.URL.Path = api.path
			req.RequestURI = api.path
			for _, m := range api.methods {
				req.Method = m
				testR.ServeHTTP(w, req)
			}
		}
	}
}

func BenchmarkRoute_GitHub_Static(b *testing.B) {
	req, _ := http.NewRequest("POST", "/markdown/raw", nil)
	for i := 0; i < b.N; i++ {
		testR.ServeHTTP(w, req)
	}
}
func BenchmarkRoute_GitHub_Param1(b *testing.B) {
	temPath = "/teams/:id/repos"
	//temPath = strings.Replace(temPath, ":", paramPrefix, -1)
	req, _ := http.NewRequest("GET", temPath, nil)
	for i := 0; i < b.N; i++ {
		testR.ServeHTTP(w, req)
	}
}

// func TestRoute_ServeHTTP2(t *testing.T) {
// 	w := new(fakeResponseWriter)
// 	req, _ := http.NewRequest("GET", "/markdown/raw/?a=1", nil)
// 	testR.ServeHTTP(w, req)
// }

func TestRoute_ServeHTTP(t *testing.T) {
	w := new(fakeResponseWriter)
	req, _ := http.NewRequest(http.MethodGet, "/", nil)
	temPath := twentyRoute
	for _, api := range githubAPi[0:10] {
		temPath = api.path
		req.URL.Path = temPath
		req.RequestURI = temPath
		for _, m := range api.methods {
			req.Method = m
			w.d = ""
			testR.ServeHTTP(w, req)
			// t.Error(w.d)
			if w.d != temPath {
				t.Errorf("request %s(%s): but recive  %s;\n",
					api.path, m, w.d)
				return
			}
		}
	}
}

var githubAPi = []struct {
	path    string
	methods []string
}{
	{twentyColon, []string{"GET"}},
	{"/", []string{"GET"}},
	{"/gitignore/templates", []string{"GET"}},
	{"/repos/:owner/:repo/commits/:sha", []string{"GET"}},
	{"/repos/:owner/:repo/issues/:number", []string{"GET"}},
	{"/applications/:client_id/tokens", []string{"DELETE"}},
	{"/users/:user/gists", []string{"GET"}},
	{"/notifications", []string{"GET", "PUT"}},
	{"/repos/:owner/:repo/hooks", []string{"GET", "POST"}},
	{"/repos/:owner/:repo/labels", []string{"GET", "POST"}},
	{"/repos/:owner/:repo/git/commits/:sha", []string{"GET"}},
	{"/users/:user/events", []string{"GET"}},
	{"/repos/:owner/:repo/pulls", []string{"GET", "POST"}},
	{"/repos/:owner/:repo/languages", []string{"GET"}},
	{"/gists/:id", []string{"GET", "DELETE"}},
	{"/repos/:owner/:repo/git/commits", []string{"POST"}},
	{"/orgs/:org/events", []string{"GET"}},
	{"/repos/:owner/:repo/stats/commit_activity", []string{"GET"}},
	{"/gists", []string{"GET", "POST"}},
	{"/repos/:owner/:repo/statuses/:ref", []string{"GET", "POST"}},
	{"/issues", []string{"GET"}},
	{"/rate_limit", []string{"GET"}},
	{"/orgs/:org/members", []string{"GET"}},
	{"/repos/:owner/:repo", []string{"GET", "DELETE"}},
	{"/repos/:owner/:repo/collaborators", []string{"GET"}},
	{"/user/starred/:owner/:repo", []string{"GET", "PUT", "DELETE"}},
	{"/markdown/raw", []string{"POST"}},
	{"/users/:user/repos", []string{"GET"}},
	{"/repos/:owner/:repo/keys", []string{"GET", "POST"}},
	{"/teams/:id/members", []string{"GET"}},
	{"/repos/:owner/:repo/releases/:id/assets", []string{"GET"}},
	{"/repos/:owner/:repo/milestones/:number/labels", []string{"GET"}},
	{"/repos/:owner/:repo/keys/:id", []string{"GET", "DELETE"}},
	{"/repos/:owner/:repo/git/tags", []string{"POST"}},
	{"/repos/:owner/:repo/teams", []string{"GET"}},
	{"/repos/:owner/:repo/issues/:number/events", []string{"GET"}},
	{"/repos/:owner/:repo/milestones", []string{"GET", "POST"}},
	{"/repos/:owner/:repo/notifications", []string{"GET", "PUT"}},
	{"/user/keys", []string{"GET", "POST"}},
	{"/emojis", []string{"GET"}},
	{"/search/issues", []string{"GET"}},
	{"/orgs/:org/issues", []string{"GET"}},
	{"/repos/:owner/:repo/commits/:sha/comments", []string{"GET", "POST"}},
	{"/search/code", []string{"GET"}},
	{"/meta", []string{"GET"}},
	{"/repos/:owner/:repo/git/blobs/:sha", []string{"GET"}},
	{"/notifications/threads/:id/subscription", []string{"GET", "PUT", "DELETE"}},
	{"/legacy/user/search/:keyword", []string{"GET"}},
	{"/user/orgs", []string{"GET"}},
	{"/repos/:owner/:repo/pulls/:number/files", []string{"GET"}},
	{"/users/:user/following", []string{"GET"}},
	{"/orgs/:org", []string{"GET"}},
	{"/search/users", []string{"GET"}},
	{"/user/teams", []string{"GET"}},
	{"/repos/:owner/:repo/stats/code_frequency", []string{"GET"}},
	{"/teams/:id/repos", []string{"GET"}},
	{"/events", []string{"GET"}},
	{"/orgs/:org/members/:user", []string{"GET", "DELETE"}},
	{"/repos/:owner/:repo/git/trees/:sha", []string{"GET"}},
	{"/users/:user/received_events", []string{"GET"}},
	{"/networks/:owner/:repo/events", []string{"GET"}},
	{"/repos/:owner/:repo/hooks/:id", []string{"GET", "DELETE"}},
	{"/repos/:owner/:repo/pulls/:number/comments", []string{"GET", "PUT"}},
	{"/user/following", []string{"GET"}},
	{"/gitignore/templates/:name", []string{"GET"}},
	{"/repos/:owner/:repo/tags", []string{"GET"}},
	{"/users/:user/events/orgs/:org", []string{"GET"}},
	{"/repos/:owner/:repo/releases/:id", []string{"GET", "DELETE"}},
	{"/gists/:id/star", []string{"PUT", "DELETE", "GET"}},
	{"/repos/:owner/:repo/collaborators/:user", []string{"GET", "PUT", "DELETE"}},
	{"/user/repos", []string{"GET", "POST"}},
	{"/repos/:owner/:repo/branches", []string{"GET"}},
	{"/notifications/threads/:id", []string{"GET"}},
	{"/repos/:owner/:repo/issues/:number/labels", []string{"GET", "POST", "PUT", "DELETE"}},
	{"/repos/:owner/:repo/contributors", []string{"GET"}},
	{"/orgs/:org/public_members", []string{"GET"}},
	{"/users/:user/received_events/public", []string{"GET"}},
	{"/repos/:owner/:repo/git/refs", []string{"GET", "POST"}},
	{"/user/subscriptions/:owner/:repo", []string{"GET", "PUT", "DELETE"}},
	{"/legacy/user/email/:email", []string{"GET"}},
	{"/repos/:owner/:repo/git/blobs", []string{"POST"}},
	{"/legacy/issues/search/:owner/:repository/:state/:keyword", []string{"GET"}},
	{"/repos/:owner/:repo/events", []string{"GET"}},
	{"/user/subscriptions", []string{"GET"}},
	{"/markdown", []string{"POST"}},
	{"/gists/:id/forks", []string{"POST"}},
	{"/repos/:owner/:repo/stargazers", []string{"GET"}},
	{"/users/:user", []string{"GET"}},
	{"/user/following/:user", []string{"GET", "PUT", "DELETE"}},
	{"/user/emails", []string{"GET", "POST", "DELETE"}},
	{"/repos/:owner/:repo/comments", []string{"GET"}},
	{"/teams/:id", []string{"GET", "DELETE"}},
	{"/repos/:owner/:repo/milestones/:number", []string{"GET", "DELETE"}},
	{"/repos/:owner/:repo/stats/contributors", []string{"GET"}},
	{"/teams/:id/repos/:owner/:repo", []string{"GET", "PUT", "DELETE"}},
	{"/repos/:owner/:repo/stats/punch_card", []string{"GET"}},
	{"/users/:user/keys", []string{"GET"}},
	{"/repos/:owner/:repo/hooks/:id/tests", []string{"POST"}},
	{"/users/:user/subscriptions", []string{"GET"}},
	{"/repos/:owner/:repo/assignees", []string{"GET"}},
	{"/user", []string{"GET"}},
	{"/authorizations/:id", []string{"GET", "DELETE"}},
	{"/orgs/:org/teams", []string{"GET", "POST"}},
	{"/repos/:owner/:repo/issues", []string{"GET", "POST"}},
	{"/repos/:owner/:repo/issues/:number/comments", []string{"GET", "POST"}},
	{"/applications/:client_id/tokens/:access_token", []string{"GET", "DELETE"}},
	{"/feeds", []string{"GET"}},
	{"/repos/:owner/:repo/comments/:id", []string{"GET", "DELETE"}},
	{"/repos/:owner/:repo/pulls/:number", []string{"GET"}},
	{"/repos/:owner/:repo/downloads/:id", []string{"GET", "DELETE"}},
	{"/users/:user/orgs", []string{"GET"}},
	{"/orgs/:org/repos", []string{"GET", "POST"}},
	{"/users/:user/following/:target_user", []string{"GET"}},
	{"/repos/:owner/:repo/readme", []string{"GET"}},
	{"/repos/:owner/:repo/forks", []string{"GET", "POST"}},
	{"/repos/:owner/:repo/issues/:number/labels/:name", []string{"DELETE"}},
	{"/legacy/repos/search/:keyword", []string{"GET"}},
	{"/repos/:owner/:repo/merges", []string{"POST"}},
	{"/repos/:owner/:repo/git/tags/:sha", []string{"GET"}},
	{"/search/repositories", []string{"GET"}},
	{"/user/starred", []string{"GET"}},
	{"/teams/:id/members/:user", []string{"GET", "PUT", "DELETE"}},
	{"/users", []string{"GET"}},
	{"/user/issues", []string{"GET"}},
	{"/repos/:owner/:repo/subscribers", []string{"GET"}},
	{"/repos/:owner/:repo/git/trees", []string{"POST"}},
	{"/users/:user/events/public", []string{"GET"}},
	{"/repos/:owner/:repo/pulls/:number/merge", []string{"GET", "PUT"}},
	{"/repos/:owner/:repo/assignees/:assignee", []string{"GET"}},
	{"/users/:user/starred", []string{"GET"}},
	{"/repos/:owner/:repo/labels/:name", []string{"GET", "DELETE"}},
	{"/user/followers", []string{"GET"}},
	{"/orgs/:org/public_members/:user", []string{"GET", "PUT", "DELETE"}},
	{"/authorizations", []string{"GET", "POST"}},
	{"/repos/:owner/:repo/downloads", []string{"GET"}},
	{"/repos/:owner/:repo/releases", []string{"GET", "POST"}},
	{"/user/keys/:id", []string{"GET", "DELETE"}},
	{"/repos/:owner/:repo/stats/participation", []string{"GET"}},
	{"/repos/:owner/:repo/subscription", []string{"GET", "PUT", "DELETE"}},
	{"/repositories", []string{"GET"}},
	{"/repos/:owner/:repo/branches/:branch", []string{"GET"}},
	{"/repos/:owner/:repo/pulls/:number/commits", []string{"GET"}},
	{"/users/:user/followers", []string{"GET"}},
	{"/repos/:owner/:repo/commits", []string{"GET"}},
}
