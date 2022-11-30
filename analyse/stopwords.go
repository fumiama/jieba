package analyse

import (
	"io/fs"
	"sync"

	"github.com/fumiama/jieba/dictionary"
)

// DefaultStopWordMap contains some stop words.
var DefaultStopWordMap = map[string]int{
	"the":   1,
	"of":    1,
	"is":    1,
	"and":   1,
	"to":    1,
	"in":    1,
	"that":  1,
	"we":    1,
	"for":   1,
	"an":    1,
	"are":   1,
	"by":    1,
	"be":    1,
	"as":    1,
	"on":    1,
	"with":  1,
	"can":   1,
	"if":    1,
	"from":  1,
	"which": 1,
	"you":   1,
	"it":    1,
	"this":  1,
	"then":  1,
	"at":    1,
	"have":  1,
	"all":   1,
	"not":   1,
	"one":   1,
	"has":   1,
	"or":    1,
}

// StopWord is a thread-safe dictionary for all stop words.
type StopWord struct {
	sync.RWMutex
	stopWordMap map[string]int
}

// AddToken adds a token into StopWord dictionary.
func (s *StopWord) AddToken(token dictionary.Token) {
	s.Lock()
	s.stopWordMap[token.Text()] = 1
	s.Unlock()
}

// NewStopWord create a new StopWord with default stop words.
func NewStopWord() *StopWord {
	m := make(map[string]int, len(DefaultStopWordMap)*2)
	for k, v := range DefaultStopWordMap {
		m[k] = v
	}
	return &StopWord{
		stopWordMap: m,
	}
}

// IsStopWord checks if a given word is stop word.
func (s *StopWord) IsStopWord(word string) bool {
	s.RLock()
	_, ok := s.stopWordMap[word]
	s.RUnlock()
	return ok
}

// Load loads all tokens into StopWord dictionary.
func (s *StopWord) Load(tokens ...dictionary.Token) {
	s.Lock()
	for _, token := range tokens {
		s.stopWordMap[token.Text()] = 1
	}
	s.Unlock()
}

func (s *StopWord) loadDictionary(file fs.File) error {
	return dictionary.LoadDictionary(s, file)
}

func (s *StopWord) loadDictionaryAt(file string) error {
	return dictionary.LoadDictionaryAt(s, file)
}
