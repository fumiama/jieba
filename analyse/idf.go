package analyse

import (
	"io"
	"sort"
	"sync"

	"github.com/fumiama/jieba/dictionary"
)

// Idf represents a thread-safe dictionary for all words with their
// IDFs(Inverse Document Frequency).
type Idf struct {
	sync.RWMutex
	median  float64
	freqMap map[string]float64
	freqs   []float64
}

// AddToken adds a new word with IDF into it's dictionary.
func (i *Idf) AddToken(token dictionary.Token) {
	i.Lock()
	i.freqMap[token.Text()] = token.Frequency()
	i.freqs = append(i.freqs, token.Frequency())
	sort.Float64s(i.freqs)
	i.median = i.freqs[len(i.freqs)/2]
	i.Unlock()
}

// Load loads all tokens into it's dictionary.
func (i *Idf) Load(tokens ...dictionary.Token) {
	i.Lock()
	for _, token := range tokens {
		i.freqMap[token.Text()] = token.Frequency()
		i.freqs = append(i.freqs, token.Frequency())
	}
	sort.Float64s(i.freqs)
	i.median = i.freqs[len(i.freqs)/2]
	i.Unlock()
}

func (i *Idf) loadDictionary(file io.Reader) error {
	return dictionary.LoadDictionary(i, file)
}

func (i *Idf) loadDictionaryAt(fileName string) error {
	return dictionary.LoadDictionaryAt(i, fileName)
}

// Frequency returns the IDF of given word.
func (i *Idf) Frequency(key string) (float64, bool) {
	i.RLock()
	freq, ok := i.freqMap[key]
	i.RUnlock()
	return freq, ok
}

// NewIdf creates a new Idf instance.
func NewIdf() *Idf {
	return &Idf{freqMap: make(map[string]float64, 256), freqs: make([]float64, 0, 256)}
}
