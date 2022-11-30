// Package jieba is the Golang implemention of [Jieba](https://github.com/fxsjy/jieba), Python Chinese text segmentation module.
package jieba

import (
	"io/fs"
	"math"
	"regexp"
	"strings"

	"github.com/fumiama/jieba/dictionary"
	"github.com/fumiama/jieba/finalseg"
	"github.com/fumiama/jieba/util"
)

var (
	reEng         = regexp.MustCompile(`[[:alnum:]]`)
	reHanCutAll   = regexp.MustCompile(`(\p{Han}+)`)
	reSkipCutAll  = regexp.MustCompile(`[^[:alnum:]+#\n]`)
	reHanDefault  = regexp.MustCompile(`([\p{Han}+[:alnum:]+#&\._]+)`)
	reSkipDefault = regexp.MustCompile(`(\r\n|\s)`)
)

// Segmenter is a Chinese words segmentation struct.
type Segmenter Dictionary

// Frequency returns a word's frequency and existence
func (seg *Segmenter) Frequency(word string) (float64, bool) {
	return (*Dictionary)(seg).Frequency(word)
}

// AddWord adds a new word with frequency to dictionary
func (seg *Segmenter) AddWord(word string, frequency float64) {
	(*Dictionary)(seg).AddToken(dictionary.NewToken(word, frequency, ""))
}

// DeleteWord removes a word from dictionary
func (seg *Segmenter) DeleteWord(word string) {
	(*Dictionary)(seg).AddToken(dictionary.NewToken(word, 0.0, ""))
}

/*
SuggestFrequency returns a suggested frequncy of a word or a long word
cutted into several short words.

This method is useful when a word in the sentence is not cutted out correctly.

If a word should not be further cutted, for example word "石墨烯" should not be
cutted into "石墨" and "烯", SuggestFrequency("石墨烯") will return the maximu
frequency for this word.

If a word should be further cutted, for example word "今天天气" should be
further cutted into two words "今天" and "天气",  SuggestFrequency("今天", "天气")
should return the minimum frequency for word "今天天气".
*/
func (seg *Segmenter) SuggestFrequency(words ...string) float64 {
	frequency := 1.0
	if len(words) > 1 {
		for _, word := range words {
			if freq, ok := (*Dictionary)(seg).Frequency(word); ok {
				frequency *= freq
			}
			frequency /= (*Dictionary)(seg).total
		}
		frequency, _ = math.Modf(frequency * (*Dictionary)(seg).total)
		wordFreq := 0.0
		if freq, ok := (*Dictionary)(seg).Frequency(strings.Join(words, "")); ok {
			wordFreq = freq
		}
		if wordFreq < frequency {
			frequency = wordFreq
		}
		return frequency
	}
	word := words[0]
	for _, segment := range seg.Cut(word, false) {
		if freq, ok := (*Dictionary)(seg).Frequency(segment); ok {
			frequency *= freq
		}
		frequency /= (*Dictionary)(seg).total
	}
	frequency, _ = math.Modf(frequency * (*Dictionary)(seg).total)
	frequency += 1.0
	wordFreq := 1.0
	if freq, ok := (*Dictionary)(seg).Frequency(word); ok {
		wordFreq = freq
	}
	if wordFreq > frequency {
		frequency = wordFreq
	}
	return frequency
}

// LoadDictionary loads dictionary from given file name. Everytime
// LoadDictionary is called, previously loaded dictionary will be cleard.
func LoadDictionary(file fs.File) (*Segmenter, error) {
	d := &Dictionary{freqMap: make(map[string]float64)}
	err := d.loadDictionary(file)
	return (*Segmenter)(d), err
}

// LoadDictionaryAt loads dictionary from given file name. Everytime
// LoadDictionaryAt is called, previously loaded dictionary will be cleard.
func LoadDictionaryAt(file string) (*Segmenter, error) {
	d := &Dictionary{freqMap: make(map[string]float64)}
	err := d.loadDictionaryAt(file)
	return (*Segmenter)(d), err
}

// LoadUserDictionary loads a user specified dictionary, it must be called
// after LoadDictionary, and it will not clear any previous loaded dictionary,
// instead it will override exist entries.
func (seg *Segmenter) LoadUserDictionary(file fs.File) error {
	return (*Dictionary)(seg).loadDictionary(file)
}

// LoadUserDictionaryAt loads a user specified dictionary, it must be called
// after LoadDictionary, and it will not clear any previous loaded dictionary,
// instead it will override exist entries.
func (seg *Segmenter) LoadUserDictionaryAt(file string) error {
	return (*Dictionary)(seg).loadDictionaryAt(file)
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

// ratio words and letters in an article commonly
const (
	RatioLetterWord     float32 = 1.5
	RatioLetterWordFull float32 = 1
)

type cutFunc func(sentence string) []string

func (seg *Segmenter) cutDAG(sentence string) []string {
	result := make([]string, 0, int(float32(len(sentence))/RatioLetterWord)+1)
	runes := []rune(sentence)
	routes := seg.calc(runes)
	buf := make([]rune, 0, 256)
	for x := 0; x < len(runes); {
		y := routes[x].index + 1
		frag := runes[x:y]
		if y-x == 1 {
			buf = append(buf, frag...)
		} else {
			if len(buf) > 0 {
				bufString := string(buf)
				if len(buf) == 1 {
					result = append(result, bufString)
				} else {
					if v, ok := (*Dictionary)(seg).Frequency(bufString); !ok || v == 0.0 {
						result = append(result, finalseg.Cut(bufString)...)
					} else {
						for _, elem := range buf {
							result = append(result, string(elem))
						}
					}
				}
				buf = buf[:0]
			}
			result = append(result, string(frag))
		}
		x = y
	}

	if len(buf) > 0 {
		bufString := string(buf)
		if len(buf) == 1 {
			result = append(result, bufString)
		} else {
			if v, ok := (*Dictionary)(seg).Frequency(bufString); !ok || v == 0.0 {
				result = append(result, finalseg.Cut(bufString)...)
			} else {
				for _, elem := range buf {
					result = append(result, string(elem))
				}
			}
		}
	}

	return result
}

func (seg *Segmenter) cutDAGNoHMM(sentence string) []string {
	result := make([]string, 0, int(float32(len(sentence))/RatioLetterWord)+1)
	runes := []rune(sentence)
	routes := seg.calc(runes)
	buf := make([]rune, 0, 256)
	for x := 0; x < len(runes); {
		y := routes[x].index + 1
		frag := runes[x:y]
		if reEng.MatchString(string(frag)) && len(frag) == 1 {
			buf = append(buf, frag...)
			x = y
			continue
		}
		if len(buf) > 0 {
			result = append(result, string(buf))
			buf = buf[:0]
		}
		result = append(result, string(frag))
		x = y
	}
	if len(buf) > 0 {
		result = append(result, string(buf))
	}

	return result
}

// Cut cuts a sentence into words using accurate mode.
// Parameter hmm controls whether to use the Hidden Markov Model.
// Accurate mode attempts to cut the sentence into the most accurate
// segmentations, which is suitable for text analysis.
func (seg *Segmenter) Cut(sentence string, hmm bool) []string {
	result := make([]string, 0, int(float32(len(sentence))/RatioLetterWord)+1)
	var cut cutFunc
	if hmm {
		cut = seg.cutDAG
	} else {
		cut = seg.cutDAGNoHMM
	}

	for _, block := range util.RegexpSplit(reHanDefault, sentence, -1) {
		if len(block) == 0 {
			continue
		}
		if reHanDefault.MatchString(block) {
			result = append(result, cut(block)...)
			continue
		}
		for _, subBlock := range util.RegexpSplit(reSkipDefault, block, -1) {
			if reSkipDefault.MatchString(subBlock) {
				result = append(result, subBlock)
				continue
			}
			for _, r := range subBlock {
				result = append(result, string(r))
			}
		}
	}

	return result
}

func (seg *Segmenter) cutAll(sentence string) []string {
	result := make([]string, 0, int(float32(len(sentence))/RatioLetterWord)+1)
	runes := []rune(sentence)
	dag := seg.dag(runes)
	start := -1
	for k := 0; k < len(dag); k++ {
		l := dag[k]
		if len(l) == 1 && k > start {
			result = append(result, string(runes[k:l[0]+1]))
			start = l[0]
			continue
		}
		for _, j := range l {
			if j > k {
				result = append(result, string(runes[k:j+1]))
				start = j
			}
		}
	}

	return result
}

// CutAll cuts a sentence into words using full mode.
// Full mode gets all the possible words from the sentence.
// Fast but not accurate.
func (seg *Segmenter) CutAll(sentence string) []string {
	result := make([]string, 0, int(float32(len(sentence))/RatioLetterWordFull)+1)

	for _, block := range util.RegexpSplit(reHanCutAll, sentence, -1) {
		if len(block) == 0 {
			continue
		}
		if reHanCutAll.MatchString(block) {
			result = append(result, seg.cutAll(block)...)
			continue
		}
		result = append(result, reSkipCutAll.Split(block, -1)...)
	}

	return result
}

// CutForSearch cuts sentence into words using search engine mode.
// Search engine mode, based on the accurate mode, attempts to cut long words
// into several short words, which can raise the recall rate.
// Suitable for search engines.
func (seg *Segmenter) CutForSearch(sentence string, hmm bool) []string {
	result := make([]string, 0, int(float32(len(sentence))/RatioLetterWordFull)+1)

	for _, word := range seg.Cut(sentence, hmm) {
		runes := []rune(word)
		for _, increment := range []int{2, 3} {
			if len(runes) <= increment {
				continue
			}
			for i := 0; i < len(runes)-increment+1; i++ {
				gram := string(runes[i : i+increment])
				if v, ok := (*Dictionary)(seg).Frequency(gram); ok && v > 0.0 {
					result = append(result, gram)
				}
			}
		}
		result = append(result, word)
	}

	return result
}
