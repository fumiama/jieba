// Package posseg is the Golang implementation of Jieba's posseg module.
package posseg

import (
	"io"
	"math"
	"regexp"

	"github.com/fumiama/jieba/util"
)

var (
	reHanDetail    = regexp.MustCompile(`(\p{Han}+)`)
	reSkipDetail   = regexp.MustCompile(`([[\.[:digit:]]+|[:alnum:]]+)`)
	reEng          = regexp.MustCompile(`[[:alnum:]]`)
	reNum          = regexp.MustCompile(`[\.[:digit:]]+`)
	reEng1         = regexp.MustCompile(`[[:alnum:]]$`)
	reHanInternal  = regexp.MustCompile(`([\p{Han}+[:alnum:]+#&\._]+)`)
	reSkipInternal = regexp.MustCompile(`(\r\n|\s)`)
)

// Segment represents a word with it's POS
type Segment struct {
	text, pos string
}

// Text returns the Segment's text.
func (s Segment) Text() string {
	return s.text
}

// Pos returns the Segment's POS.
func (s Segment) Pos() string {
	return s.pos
}

// Segmenter is a Chinese words segmentation struct.
type Segmenter Dictionary

// LoadDictionary loads dictionary from given file name.
// Everytime LoadDictionaryAt is called, previously loaded dictionary will be cleard.
func LoadDictionary(file io.Reader) (*Segmenter, error) {
	dict := &Dictionary{freqMap: make(map[string]float64), posMap: make(map[string]string)}
	err := dict.loadDictionary(file)
	if err != nil {
		return nil, err
	}
	return (*Segmenter)(dict), nil
}

// LoadDictionaryAt loads dictionary from given file name.
// Everytime LoadDictionaryAt is called, previously loaded dictionary will be cleard.
func LoadDictionaryAt(file string) (*Segmenter, error) {
	dict := &Dictionary{freqMap: make(map[string]float64), posMap: make(map[string]string)}
	err := dict.loadDictionaryAt(file)
	if err != nil {
		return nil, err
	}
	return (*Segmenter)(dict), nil
}

// LoadUserDictionary loads a user specified dictionary, it must be called
// after LoadDictionary, and it will not clear any previous loaded dictionary,
// instead it will override exist entries.
func (seg *Segmenter) LoadUserDictionary(file io.Reader) error {
	return (*Dictionary)(seg).loadDictionary(file)
}

// LoadUserDictionaryAt loads a user specified dictionary, it must be called
// after LoadDictionary, and it will not clear any previous loaded dictionary,
// instead it will override exist entries.
func (seg *Segmenter) LoadUserDictionaryAt(fileName string) error {
	return (*Dictionary)(seg).loadDictionaryAt(fileName)
}

func (seg *Segmenter) cutDetailInternal(sentence string) (results []Segment) {
	runes := []rune(sentence)
	posList := viterbi(runes)
	begin := 0
	next := 0
	for i, char := range runes {
		pos := posList[i]
		switch pos.position() {
		case "B":
			begin = i
		case "E":
			results = append(results, Segment{string(runes[begin : i+1]), pos.pos()})
			next = i + 1
		case "S":
			results = append(results, Segment{string(char), pos.pos()})
			next = i + 1
		}
	}
	if next < len(runes) {
		results = append(results, Segment{string(runes[next:]), posList[next].pos()})
	}
	return
}

func (seg *Segmenter) cutDetail(sentence string) (results []Segment) {
	for _, blk := range util.RegexpSplit(reHanDetail, sentence, -1) {
		if reHanDetail.MatchString(blk) {
			results = append(results, seg.cutDetailInternal(blk)...)
			continue
		}
		for _, x := range util.RegexpSplit(reSkipDetail, blk, -1) {
			if len(x) == 0 {
				continue
			}
			switch {
			case reNum.MatchString(x):
				results = append(results, Segment{x, "m"})
			case reEng.MatchString(x):
				results = append(results, Segment{x, "eng"})
			default:
				results = append(results, Segment{x, "x"})
			}
		}
	}
	return
}

func (seg *Segmenter) dag(runes []rune) [][]int {
	n := len(runes)
	dag := make([][]int, n)
	for k := 0; k < n; k++ {
		dag[k] = make([]int, 0, 64)
		i := k
		frag := runes[k : k+1]
		for {
			freq, ok := (*Dictionary)(seg).Frequency(string(frag))
			if !ok {
				break
			}
			if freq > 0.0 {
				dag[k] = append(dag[k], i)
			}
			i++
			if i >= n {
				break
			}
			frag = runes[k : i+1]
		}
		if len(dag[k]) == 0 {
			dag[k] = append(dag[k], k)
		}
	}
	return dag
}

type route struct {
	frequency float64
	index     int
}

func (seg *Segmenter) calc(runes []rune) []*route {
	dag := seg.dag(runes)
	n := len(runes)
	rs := make([]*route, n+1)
	rs[n] = &route{frequency: 0.0, index: 0}
	for idx := n - 1; idx >= 0; idx-- {
		for _, i := range dag[idx] {
			var r *route
			if freq, ok := (*Dictionary)(seg).Frequency(string(runes[idx : i+1])); ok {
				r = &route{frequency: math.Log(freq) - (*Dictionary)(seg).logTotal + rs[i+1].frequency, index: i}
			} else {
				r = &route{frequency: math.Log(1.0) - (*Dictionary)(seg).logTotal + rs[i+1].frequency, index: i}
			}
			if v := rs[idx]; v == nil {
				rs[idx] = r
			} else {
				if v.frequency < r.frequency || (v.frequency == r.frequency && v.index < r.index) {
					rs[idx] = r
				}
			}
		}
	}
	return rs
}

func (seg *Segmenter) cutDAG(sentence string) (results []Segment) {
	runes := []rune(sentence)
	routes := seg.calc(runes)
	buf := make([]rune, 0, 256)
	for x := 0; x < len(runes); {
		y := routes[x].index + 1
		frag := runes[x:y]
		if y-x == 1 {
			buf = append(buf, frag...)
			x = y
			continue
		}
		if len(buf) > 0 {
			bufString := string(buf)
			if len(buf) == 1 {
				if tag, ok := (*Dictionary)(seg).Pos(bufString); ok {
					results = append(results, Segment{bufString, tag})
				} else {
					results = append(results, Segment{bufString, "x"})
				}
				buf = buf[:0]
				continue
			}
			if v, ok := (*Dictionary)(seg).Frequency(bufString); !ok || v == 0.0 {
				results = append(results, seg.cutDetail(bufString)...)
			} else {
				for _, elem := range buf {
					selem := string(elem)
					if tag, ok := (*Dictionary)(seg).Pos(selem); ok {
						results = append(results, Segment{selem, tag})
					} else {
						results = append(results, Segment{selem, "x"})
					}
				}
			}
			buf = buf[:0]
		}
		word := string(frag)
		if tag, ok := (*Dictionary)(seg).Pos(word); ok {
			results = append(results, Segment{word, tag})
		} else {
			results = append(results, Segment{word, "x"})
		}
		x = y
	}

	if len(buf) > 0 {
		bufString := string(buf)
		if len(buf) == 1 {
			if tag, ok := (*Dictionary)(seg).Pos(bufString); ok {
				results = append(results, Segment{bufString, tag})
			} else {
				results = append(results, Segment{bufString, "x"})
			}
			return
		}
		if v, ok := (*Dictionary)(seg).Frequency(bufString); !ok || v == 0.0 {
			results = append(results, seg.cutDetail(bufString)...)
			return
		}
		for _, elem := range buf {
			selem := string(elem)
			if tag, ok := (*Dictionary)(seg).Pos(selem); ok {
				results = append(results, Segment{selem, tag})
			} else {
				results = append(results, Segment{selem, "x"})
			}
		}
	}
	return
}

func (seg *Segmenter) cutDAGNoHMM(sentence string) (results []Segment) {
	runes := []rune(sentence)
	routes := seg.calc(runes)
	buf := make([]rune, 0, 256)
	for x := 0; x < len(runes); {
		y := routes[x].index + 1
		frag := runes[x:y]
		if reEng1.MatchString(string(frag)) && len(frag) == 1 {
			buf = append(buf, frag...)
			x = y
			continue
		}
		if len(buf) > 0 {
			results = append(results, Segment{string(buf), "eng"})
			buf = buf[:0]
		}
		word := string(frag)
		if tag, ok := (*Dictionary)(seg).Pos(word); ok {
			results = append(results, Segment{word, tag})
		} else {
			results = append(results, Segment{word, "x"})
		}
		x = y
	}
	if len(buf) > 0 {
		results = append(results, Segment{string(buf), "eng"})
	}
	return
}

// Cut cuts a sentence into words.
// Parameter hmm controls whether to use the Hidden Markov Model.
func (seg *Segmenter) Cut(sentence string, hmm bool) (results []Segment) {
	var cut func(sentence string) []Segment
	if hmm {
		cut = seg.cutDAG
	} else {
		cut = seg.cutDAGNoHMM
	}
	for _, blk := range util.RegexpSplit(reHanInternal, sentence, -1) {
		if reHanInternal.MatchString(blk) {
			results = append(results, cut(blk)...)
			continue
		}
		for _, x := range util.RegexpSplit(reSkipInternal, blk, -1) {
			if reSkipInternal.MatchString(x) {
				results = append(results, Segment{x, "x"})
				continue
			}
			for _, xx := range x {
				s := string(xx)
				switch {
				case reNum.MatchString(s):
					results = append(results, Segment{s, "m"})
				case reEng.MatchString(x):
					results = append(results, Segment{x, "eng"})
				default:
					results = append(results, Segment{s, "x"})
				}
			}
		}
	}
	return
}
