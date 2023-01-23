package rxpmap

import (
	"fmt"
	"strings"

	"github.com/becheran/wildmatch-go"
)

type query struct {
	staticPrefix []byte
	ws           map[int][]*wildmatch.WildMatch
	data         [][]string
}

func parseKey(key string) []string {
	return strings.Split(key, ".")
}

func parseQuery(unparsed string) (query, error) {
	q := &query{
		ws: make(map[int][]*wildmatch.WildMatch),
	}
	gotFirstWildcard := false
	prefix := []string{}
	for index, part := range strings.Split(unparsed, ".") {
		if part[0] == '[' {
			if part[len(part)-1] != ']' {
				return *q, fmt.Errorf("invalid query: expected ']' but got '%v'", part[len(part)-1])
			}
			gotFirstWildcard = true
			q.parseWildcards(index, part)
		} else {
			p := strings.TrimSpace(part)
			if !gotFirstWildcard {
				prefix = append(prefix, p)
			}
			q.data = append(q.data, []string{p})
		}
	}
	q.staticPrefix = []byte(strings.Join(prefix, "."))
	return *q, nil
}

func (q *query) parseWildcards(index int, unparsed string) {
	ws := []string{}
	for _, wildcard := range strings.Split(unparsed[1:len(unparsed)-1], ",") {
		w := strings.TrimSpace(wildcard)
		q.ws[index] = append(q.ws[index], wildmatch.NewWildMatch(w))
		ws = append(ws, w)
	}
	q.data = append(q.data, ws)
}

func (q query) Matches(key []string) bool {
	keylen := len(key)
	if len(q.data) > keylen {
		return false
	}
	for i := 0; i < keylen; i++ {
		if ws, ok := q.ws[i]; ok {
			matches := false
			for _, w := range ws {
				if w.IsMatch(key[i]) {
					matches = true
					break
				}
			}
			if !matches {
				return false
			}
		} else {
			if key[i] != q.data[i][0] {
				return false
			}
		}
	}
	return true
}
