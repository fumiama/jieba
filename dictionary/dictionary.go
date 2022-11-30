// Package dictionary contains a interface and wraps all io related work.
// It is used by jieba module to read/write files.
package dictionary

import (
	"bufio"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// DictLoader is the interface that could add one token or load
// tokens from channel.
type DictLoader interface {
	Load(...Token)
	AddToken(Token)
}

func loadDictionary(file *os.File) (tokens []Token, err error) {
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
func LoadDictionary(dl DictLoader, fileName string) error {
	filePath, err := dictPath(fileName)
	if err != nil {
		return err
	}
	dictFile, err := os.Open(filePath)
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

func dictPath(dictFileName string) (string, error) {
	if filepath.IsAbs(dictFileName) {
		return dictFileName, nil
	}
	var dictFilePath string
	cwd, err := os.Getwd()
	if err != nil {
		return dictFilePath, err
	}
	dictFilePath = filepath.Clean(filepath.Join(cwd, dictFileName))
	return dictFilePath, nil
}
