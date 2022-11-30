// Package analyse is the Golang implementation of Jieba's analyse module.
package analyse

import (
	"io/fs"
	"sort"
	"strings"
	"unicode/utf8"

	jieba "github.com/fumiama/jieba"
)

// Segment represents a word with weight.
type Segment struct {
	text   string
	weight float64
}

// Text returns the segment's text.
func (s Segment) Text() string {
	return s.text
}

// Weight returns the segment's weight.
func (s Segment) Weight() float64 {
	return s.weight
}

// Segments represents a slice of Segment.
type Segments []Segment

func (ss Segments) Len() int {
	return len(ss)
}

func (ss Segments) Less(i, j int) bool {
	if ss[i].weight == ss[j].weight {
		return ss[i].text < ss[j].text
	}

	return ss[i].weight < ss[j].weight
}

func (ss Segments) Swap(i, j int) {
	ss[i], ss[j] = ss[j], ss[i]
}

// TagExtracter is used to extract tags from sentence.
type TagExtracter struct {
	seg      *jieba.Segmenter
	idf      *Idf
	stopWord *StopWord
}

// LoadDictionary reads the given filename and create a new dictionary.
func (t *TagExtracter) LoadDictionary(file fs.File) (err error) {
	t.stopWord = NewStopWord()
	t.seg, err = jieba.LoadDictionary(file)
	return
}

// LoadDictionaryAt reads the given filename and create a new dictionary.
func (t *TagExtracter) LoadDictionaryAt(file string) (err error) {
	t.stopWord = NewStopWord()
	t.seg, err = jieba.LoadDictionaryAt(file)
	return
}

// LoadIdf reads the given file and create a new Idf dictionary.
func (t *TagExtracter) LoadIdf(file fs.File) error {
	t.idf = NewIdf()
	return t.idf.loadDictionary(file)
}

// LoadIdfAt reads the given file and create a new Idf dictionary.
func (t *TagExtracter) LoadIdfAt(fileName string) error {
	t.idf = NewIdf()
	return t.idf.loadDictionaryAt(fileName)
}

// LoadStopWords reads the given file and create a new StopWord dictionary.
func (t *TagExtracter) LoadStopWords(file fs.File) error {
	t.stopWord = NewStopWord()
	return t.stopWord.loadDictionary(file)
}

// LoadStopWordsAt reads the given file and create a new StopWord dictionary.
func (t *TagExtracter) LoadStopWordsAt(file string) error {
	t.stopWord = NewStopWord()
	return t.stopWord.loadDictionaryAt(file)
}

// ExtractTags extracts the topK key words from sentence.
func (t *TagExtracter) ExtractTags(sentence string, topK int) (tags Segments) {
	freqMap := make(map[string]uint64, 256)

	for _, w := range t.seg.Cut(sentence, true) {
		w = strings.TrimSpace(w)
		if utf8.RuneCountInString(w) < 2 {
			continue
		}
		if t.stopWord.IsStopWord(w) {
			continue
		}
		if v, ok := freqMap[w]; ok {
			freqMap[w] = v + 1
		} else {
			freqMap[w] = 1
		}
	}
	total := uint64(0)
	for _, freq := range freqMap {
		total += freq
	}
	ws := make(Segments, len(freqMap))
	i := 0
	for k, v := range freqMap {
		ws[i].text = k
		if freq, ok := t.idf.Frequency(k); ok {
			ws[i].weight = freq * float64(v) / float64(total)
		} else {
			ws[i].weight = t.idf.median * float64(v) / float64(total)
		}
		i++
	}
	sort.Sort(sort.Reverse(ws))
	if len(ws) > topK {
		tags = ws[:topK]
	} else {
		tags = ws
	}
	return tags
}
