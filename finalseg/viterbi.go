package finalseg

import (
	"fmt"
	"sort"
)

const minFloat = -3.14e100

var (
	prevStatus = map[byte][2]byte{
		'B': {'E', 'S'},
		'M': {'M', 'B'},
		'S': {'S', 'E'},
		'E': {'B', 'M'},
	}
	probStart = map[byte]float64{
		'B': -0.26268660809250016,
		'E': -3.14e+100,
		'M': -3.14e+100,
		'S': -1.4652633398537678,
	}
)

type probState struct {
	prob  float64
	state byte
}

func (p probState) String() string {
	return fmt.Sprintf("(%f, %x)", p.prob, p.state)
}

type probStates []*probState

func (ps probStates) Len() int {
	return len(ps)
}

func (ps probStates) Less(i, j int) bool {
	if ps[i].prob == ps[j].prob {
		return ps[i].state < ps[j].state
	}
	return ps[i].prob < ps[j].prob
}

func (ps probStates) Swap(i, j int) {
	ps[i], ps[j] = ps[j], ps[i]
}

func viterbi(obs []rune, states ...byte) (float64, []byte) {
	path := [256][]byte{}
	newPath := [256][]byte{}
	V := make([][256]float64, len(obs))
	for _, y := range states {
		if val, ok := probEmit[y][obs[0]]; ok {
			V[0][y] = val + probStart[y]
		} else {
			V[0][y] = minFloat + probStart[y]
		}
		path[y] = []byte{y}
	}
	for t := 1; t < len(obs); t++ {
		for _, y := range states {
			ps0 := make(probStates, 0, 2)
			var emP float64
			if val, ok := probEmit[y][obs[t]]; ok {
				emP = val
			} else {
				emP = minFloat
			}
			for _, y0 := range prevStatus[y] {
				var transP float64
				if tp, ok := probTrans[y0][y]; ok {
					transP = tp
				} else {
					transP = minFloat
				}
				prob0 := V[t-1][y0] + transP + emP
				ps0 = append(ps0, &probState{prob: prob0, state: y0})
			}
			sort.Sort(sort.Reverse(ps0))
			V[t][y] = ps0[0].prob
			pp := make([]byte, len(path[ps0[0].state]))
			copy(pp, path[ps0[0].state])
			newPath[y] = append(pp, y)
		}
		path = newPath
	}
	ps := probStates{
		&probState{V[len(obs)-1]['E'], 'E'},
		&probState{V[len(obs)-1]['S'], 'S'},
	}
	sort.Sort(sort.Reverse(ps))
	v := ps[0]
	return v.prob, path[v.state]
}
