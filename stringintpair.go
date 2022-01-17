package main

type StringIntPair struct {
	Key   string
	Value int
}

type StringIntPairList []StringIntPair

func (p StringIntPairList) Len() int {
	return len(p)
}

func (p StringIntPairList) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

func (p StringIntPairList) Less(i, j int) bool {
	return p[i].Value < p[j].Value
}
