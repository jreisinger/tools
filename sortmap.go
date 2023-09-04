package tools

import "sort"

type Pair struct {
	Key   string
	Value int
}

type Pairs []Pair

func (p Pairs) Len() int           { return len(p) }
func (p Pairs) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
func (p Pairs) Less(i, j int) bool { return p[i].Value < p[j].Value }

func SortMapByValue(m map[string]int, descending bool) Pairs {
	p := make(Pairs, len(m))
	i := 0
	for k, v := range m {
		p[i] = Pair{k, v}
		i++
	}
	if descending {
		sort.Sort(sort.Reverse(p))
	} else {
		sort.Sort(p)
	}
	return p
}
