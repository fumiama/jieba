package jieba

import (
	"io"
	"math"
	"sync"

	"github.com/fumiama/jieba/dictionary"
)

// A Dictionary represents a thread-safe dictionary used for word segmentation.
type Dictionary struct {
	sync.RWMutex
	total, logTotal float64
	freqMap         map[string]float64
}

// Load loads all tokens
func (d *Dictionary) Load(tokens ...dictionary.Token) {
	d.Lock()
	for _, token := range tokens {
		d.addToken(token)
	}
	d.Unlock()
	d.updateLogTotal()
}

// AddToken adds one token
func (d *Dictionary) AddToken(token dictionary.Token) {
	d.Lock()
	d.addToken(token)
	d.Unlock()
	d.updateLogTotal()
}

func (d *Dictionary) addToken(token dictionary.Token) {
	d.freqMap[token.Text()] = token.Frequency()
	d.total += token.Frequency()
	for i := range token.Text() {
		if _, ok := d.freqMap[token.Text()[:i]]; i > 0 && !ok {
			d.freqMap[token.Text()[:i]] = 0.0
		}
	}
}

func (d *Dictionary) updateLogTotal() {
	d.logTotal = math.Log(d.total)
}

// Frequency returns the frequency and existence of give word
func (d *Dictionary) Frequency(key string) (float64, bool) {
	d.RLock()
	freq, ok := d.freqMap[key]
	d.RUnlock()
	return freq, ok
}

func (d *Dictionary) loadDictionary(file io.Reader) error {
	return dictionary.LoadDictionary(d, file)
}

func (d *Dictionary) loadDictionaryAt(file string) error {
	return dictionary.LoadDictionaryAt(d, file)
}
