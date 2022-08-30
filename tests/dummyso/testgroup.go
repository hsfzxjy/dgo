package main

type testgroup struct {
	cases []Equalable
	index int
}

func group(cases []Equalable) testgroup {
	return testgroup{cases, 0}
}

func (g *testgroup) next() Equalable {
	ret := g.cases[g.index]
	g.index++
	return ret
}

func (g *testgroup) exhausted() int {
	if g.index == len(g.cases) {
		return 1
	} else {
		return 0
	}
}
