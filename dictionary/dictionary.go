// Package dictionary contains a interface and wraps all io related work.
// It is used by jieba module to read/write files.
package dictionary

import (
	"bufio"
	"io/fs"
	"os"
	"strconv"
	"strings"
)

// DictLoader is the interface that could add one token or load tokens
type DictLoader interface {
	Load(...Token)
	AddToken(Token)
}

func loadDictionary(file fs.File) (tokens []Token, err error) {
	scanner := bufio.NewScanner(file)
	var token Token
	var line string
	var fields []string
	for scanner.Scan() {
		line = scanner.Text()
		fields = strings.Split(line, " ")
		token.text = strings.TrimSpace(strings.Replace(fields[0], "\ufeff", "", 1))
		if length := len(fields); length > 1 {
			token.frequency, err = strconv.ParseFloat(fields[1], 64)
			if err != nil {
				return
			}
			if length > 2 {
				token.pos = strings.TrimSpace(fields[2])
			}
		}
		tokens = append(tokens, token)
	}

	if err = scanner.Err(); err != nil {
		return
	}
	return tokens, nil
}

// LoadDictionary reads the given file and passes all tokens to a DictLoader.
func LoadDictionary(dl DictLoader, file fs.File) error {
	tokens, err := loadDictionary(file)
	if err != nil {
		return err
	}
	dl.Load(tokens...)
	return nil
}

// LoadDictionaryAt reads the given file and passes all tokens to a DictLoader.
func LoadDictionaryAt(dl DictLoader, file string) error {
	dictFile, err := os.Open(file)
	if err != nil {
		return err
	}
	tokens, err := loadDictionary(dictFile)
	dictFile.Close()
	if err != nil {
		return err
	}
	dl.Load(tokens...)
	return nil
}
