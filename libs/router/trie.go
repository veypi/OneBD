package router

import (
	"github.com/lightjiang/OneBD/libs/oerr"
	"github.com/lightjiang/OneBD/utils"
	"strings"
)

type trie struct {
	// the word between //
	fragment string
	depth    uint
	// 记录变量偏移，以/计数
	params map[string]uint
	// params 和handler 同时存在
	handler  interface{}
	parent   *trie
	wildcard *trie
	colon    *trie
	// TODO:: pprof 显示map读取占据最多时间，每一级都有一次读操作，考虑是否使用slice降低每次读取时间，还是使用缓存，降低读取次数
	subTrie map[string]*trie
}

func (t *trie) String() string {
	return "\n/" + strings.Join(t.string(), "\n")
}

func (t *trie) string() []string {
	fc := func(res []string, subt *trie) []string {
		if subt != nil {
			for _, s := range subt.string() {
				res = append(res, "|...."+s)
			}
		}
		return res
	}
	res := make([]string, 0, 10)
	item := t.AbsPath()
	if t.handler != nil {
		item = "\033[32m" + item + "\033[0m"
	}
	res = append(res, item)
	for _, subT := range t.subTrie {
		res = fc(res, subT)
	}
	res = fc(res, t.colon)
	res = fc(res, t.wildcard)
	return res
}

func (t *trie) AbsPath() string {
	if t.parent != nil {
		return t.parent.AbsPath() + "/" + t.fragment
	}
	return t.fragment
}

func (t *trie) Add(url string, h interface{}) error {
	fragments := strings.Split(strings.TrimPrefix(url, "/"), "/")
	if len(fragments) == 0 || fragments[0] == "" {
		if t.handler != nil {
			return oerr.UrlDefinedDuplicate.AttachStr(url)
		}
		t.handler = h
		return nil
	}
	params := make(map[string]uint)
	return t.add(fragments, params, h)
}

func (t *trie) add(fragments []string, params map[string]uint, h interface{}) error {
	if len(fragments) == 0 {
		if t.handler != nil {
			return oerr.UrlDefinedDuplicate.AttachStr(t.AbsPath())
		}
		t.params = params
		t.handler = h
		return nil
	}
	f := fragments[0]
	next := &trie{
		fragment: f,
		depth:    t.depth + 1,
		parent:   t,
	}
	if f == "" {
	} else if f[0] == '*' {
		if t.wildcard == nil {
			t.wildcard = next
			t.wildcard.fragment = "*"
		}
		if len(fragments) > 1 {
			return oerr.UrlPatternNotSupport.AttachStr(t.wildcard.AbsPath()+"/"+strings.Join(fragments[1:], "/"), utils.CallPath(1))
		}
		params[f] = t.wildcard.depth
		return t.wildcard.add(fragments[1:], params, h)
	} else if f[0] == ':' {
		if t.colon == nil {
			t.colon = next
			t.colon.fragment = "@"
		}
		params[f] = t.colon.depth
		return t.colon.add(fragments[1:], params, h)
	}
	if t.subTrie == nil {
		t.subTrie = make(map[string]*trie)
	}
	if n := t.subTrie[f]; n != nil {
		next = n
	} else {
		t.subTrie[f] = next
	}
	return next.add(fragments[1:], params, h)
}

func (t *trie) Match(url string) *trie {
	if url == "" {
		return t
	}
	if url[0] == '/' {
		url = url[1:]
	}
	var res *trie
	for i, v := range url {
		if v == '/' {
			res = t.subMatch(url[:i])
			if res == nil {
				return nil
			} else if res.fragment[0] == '*' {
				return res
			}
			return res.Match(url[i:])
		}
	}
	return t.subMatch(url)
}

func (t *trie) subMatch(f string) *trie {
	if t.subTrie != nil && t.subTrie[f] != nil {
		return t.subTrie[f]
	}
	if t.colon != nil {
		return t.colon
	}
	return t.wildcard
}
