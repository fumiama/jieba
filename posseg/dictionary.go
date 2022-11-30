package posseg

import (
	"io/fs"
	"math"
	"sync"

	"github.com/fumiama/jieba/dictionary"
)

// A Dictionary represents a thread-safe dictionary used for word segmentation.
type Dictionary struct {
	total, logTotal float64
	freqMap         map[string]float64
	posMap          map[string]string
	sync.RWMutex
}

// Load loads all tokens from given channel
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
	runes := []rune(token.Text())
	n := len(runes)
	for i := 0; i < n; i++ {
		frag := string(runes[:i+1])
		if _, ok := d.freqMap[frag]; !ok {
			d.freqMap[frag] = 0.0
		}
	}
	if len(token.Pos()) > 0 {
		d.posMap[token.Text()] = token.Pos()
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

// Pos returns the POS and existence of give word
func (d *Dictionary) Pos(key string) (string, bool) {
	d.RLock()
	pos, ok := d.posMap[key]
	d.RUnlock()
	return pos, ok
}

func (d *Dictionary) loadDictionary(file fs.File) error {
	return dictionary.LoadDictionary(d, file)
}

func (d *Dictionary) loadDictionaryAt(file string) error {
	return dictionary.LoadDictionaryAt(d, file)
}
