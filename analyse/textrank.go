package analyse

import (
	"hash/crc64"
	"math"
	"sort"

	"github.com/fumiama/jieba/posseg"
	"github.com/fumiama/jieba/util/helper"
)

const dampingFactor = 0.85

var (
	defaultAllowPOS = [...]string{"ns", "n", "vn", "v"}
)

type edge struct {
	weight uint64
	start  string
	end    string
}

type edges []*edge

func (es edges) Len() int {
	return len(es)
}

func (es edges) Less(i, j int) bool {
	return es[i].weight < es[j].weight
}

func (es edges) Swap(i, j int) {
	es[i], es[j] = es[j], es[i]
}

type undirectWeightedGraph struct {
	graph map[string]edges
	keys  sort.StringSlice
}

func newUndirectWeightedGraph() *undirectWeightedGraph {
	return &undirectWeightedGraph{
		graph: make(map[string]edges, 256),
		keys:  make(sort.StringSlice, 0, 256),
	}
}

func (u *undirectWeightedGraph) addEdge(start, end string, weight uint64) {
	if _, ok := u.graph[start]; !ok {
		u.keys = append(u.keys, start)
		u.graph[start] = edges{&edge{start: start, end: end, weight: weight}}
	} else {
		u.graph[start] = append(u.graph[start], &edge{start: start, end: end, weight: weight})
	}

	if _, ok := u.graph[end]; !ok {
		u.keys = append(u.keys, end)
		u.graph[end] = edges{&edge{start: end, end: start, weight: weight}}
	} else {
		u.graph[end] = append(u.graph[end], &edge{start: end, end: start, weight: weight})
	}
}

func (u *undirectWeightedGraph) rank() Segments {
	if !sort.IsSorted(u.keys) {
		sort.Sort(u.keys)
	}

	ws := make(map[string]float64, len(u.graph)*2)
	outSum := make(map[string]uint64, len(u.graph)*2)

	wsdef := 1.0
	if len(u.graph) > 0 {
		wsdef /= float64(len(u.graph))
	}
	for n, out := range u.graph {
		ws[n] = wsdef
		sum := uint64(0)
		for _, e := range out {
			sum += e.weight
		}
		outSum[n] = sum
	}

	for x := 0; x < 10; x++ {
		for _, n := range u.keys {
			s := 0.0
			inedges := u.graph[n]
			for _, e := range inedges {
				s += float64(e.weight) * ws[e.end] / float64(outSum[e.end])
			}
			ws[n] = (1 - dampingFactor) + dampingFactor*s
		}
	}
	minRank := math.MaxFloat64
	maxRank := math.SmallestNonzeroFloat64
	for _, w := range ws {
		if w < minRank {
			minRank = w
		} else if w > maxRank {
			maxRank = w
		}
	}
	result := make(Segments, len(ws))
	i := 0
	for n, w := range ws {
		result[i].text = n
		result[i].weight = (w - minRank/10.0) / (maxRank - minRank/10.0)
		i++
	}
	sort.Sort(sort.Reverse(result))
	return result
}

// TextRankWithPOS extracts keywords from sentence using TextRank algorithm.
// Parameter allowPOS allows a customized pos list.
func (t *TextRanker) TextRankWithPOS(sentence string, topK int, allowPOS []string) Segments {
	posFilt := make(map[string]int, len(allowPOS)*2)
	for _, pos := range allowPOS {
		posFilt[pos] = 1
	}
	g := newUndirectWeightedGraph()
	cm := make(map[uint64]uint64, 256)
	hm := make(map[uint64][2]string, 256)
	gethash := func(a, b string) uint64 {
		h := crc64.New(crc64.MakeTable(crc64.ISO))
		h.Write(helper.StringToBytes(a))
		h.Write([]byte("\t"))
		h.Write(helper.StringToBytes(b))
		return h.Sum64()
	}
	span := 5
	var pairs []posseg.Segment
	for pair := range (*posseg.Segmenter)(t).Cut(sentence, true) {
		pairs = append(pairs, pair)
	}
	for i := range pairs {
		if _, ok := posFilt[pairs[i].Pos()]; ok {
			for j := i + 1; j < i+span && j <= len(pairs); j++ {
				if _, ok := posFilt[pairs[j].Pos()]; !ok {
					continue
				}
				h := gethash(pairs[i].Text(), pairs[j].Text())
				if _, ok := cm[h]; !ok {
					cm[h] = 1
					hm[h] = [2]string{pairs[i].Text(), pairs[j].Text()}
				} else {
					cm[h]++
				}
			}
		}
	}
	for h, weight := range cm {
		startEnd := hm[h]
		g.addEdge(startEnd[0], startEnd[1], weight)
	}
	tags := g.rank()
	if topK > 0 && len(tags) > topK {
		tags = tags[:topK]
	}
	return tags
}

// TextRank extract keywords from sentence using TextRank algorithm.
// Parameter topK specify how many top keywords to be returned at most.
func (t *TextRanker) TextRank(sentence string, topK int) Segments {
	return t.TextRankWithPOS(sentence, topK, defaultAllowPOS[:])
}

// TextRanker is used to extract tags from sentence.
type TextRanker posseg.Segmenter

// NewTextRanker reads a given file and create a new dictionary file for Textranker.
func NewTextRanker(fileName string) (TextRanker, error) {
	seg := posseg.Segmenter{}
	return TextRanker(seg), seg.LoadDictionary(fileName)
}
