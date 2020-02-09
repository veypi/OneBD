package router

import (
	"testing"
)

var ts = trie{}
var allPaths = make([]string, 0, 100)

func init() {
	for _, api := range githubAPi {
		for _, m := range api.methods {
			p := "/" + m + api.path
			e := ts.Add(p, p)
			allPaths = append(allPaths, p)
			if e != nil {
				panic(e)
			}
		}
	}
	logger.Info().Int("sum", len(allPaths)).Msg("github api")
}

func TestTrie(t *testing.T) {
	t.Log(ts.String())
	var p *trie
	for _, api := range allPaths {
		p = ts.Match(api)
		if p == nil || p.handler != api {
			t.Errorf("request %s but recieve %+v", api, p.handler)
		} else {
			//t.Log(api)
		}
	}
}

func BenchmarkTrie_GitHub_ALL(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, api := range allPaths {
			ts.Match(api)
		}
	}
}

func BenchmarkTrie_GitHub_Static(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ts.Match("/POST/markdown/raw")
	}
}

func BenchmarkTrie_GitHub_Param1(b *testing.B) {
	tempPath := "/GET/teams/aasd/repos/asd"
	for i := 0; i < b.N; i++ {
		ts.Match(tempPath)
	}
}
