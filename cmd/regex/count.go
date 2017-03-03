package regex

import "strings"

type Counter struct {
	Total   int
	Matches map[string]*Match
	Actual  map[string]string
}

type Match struct {
	Count int
	Files map[string]int
}

func NewCounter() *Counter {
	return &Counter{
		Total:   0,
		Matches: make(map[string]*Match),
		Actual:  make(map[string]string),
	}
}

func NewMatch() *Match {
	return &Match{0, make(map[string]int)}
}

func (counter *Counter) Add(file string, match string, ignoreCase, ignoreSpace bool) {
	canon := match
	if ignoreCase {
		canon = strings.ToLower(canon)
	}
	if ignoreSpace {
		canon = strings.Replace(canon, " ", "", -1)
		canon = strings.Replace(canon, "\t", "", -1)
		canon = strings.Replace(canon, "\n", "", -1)
	}

	if _, ok := counter.Actual[canon]; !ok {
		counter.Actual[canon] = match
		counter.Matches[canon] = NewMatch()
	}

	counter.Total++
	counter.Matches[canon].Add(file)
}

func (m *Match) Add(file string) {
	m.Count++
	m.Files[file]++
}

func (counter *Counter) Merge(other *Counter) {
	counter.Total += other.Total
	for canon, actual := range other.Actual {
		if _, ok := counter.Actual[canon]; !ok {
			counter.Actual[canon] = actual
			counter.Matches[canon] = NewMatch()
		}
	}

	for canon, match := range other.Matches {
		counter.Matches[canon].Merge(match)
	}
}

func (m *Match) Merge(other *Match) {
	m.Count += other.Count
	for file, count := range other.Files {
		m.Files[file] += count
	}
}
