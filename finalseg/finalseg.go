// Package finalseg is the Golang implementation of Jieba's finalseg module.
package finalseg

import (
	"regexp"
)

var (
	reHan  = regexp.MustCompile(`\p{Han}+`)
	reSkip = regexp.MustCompile(`(\d+\.\d+|[a-zA-Z0-9]+)`)
)

func cutHan(sentence string) []string {
	runes := []rune(sentence)
	result := make([]string, 0, len(runes))
	_, posList := viterbi(runes, 'B', 'M', 'E', 'S')
	begin, next := 0, 0
	for i, char := range runes {
		pos := posList[i]
		switch pos {
		case 'B':
			begin = i
		case 'E':
			result = append(result, string(runes[begin:i+1]))
			next = i + 1
		case 'S':
			result = append(result, string(char))
			next = i + 1
		}
	}
	if next < len(runes) {
		result = append(result, string(runes[next:]))
	}

	return result
}

// Cut cuts sentence into words using Hidden Markov Model with Viterbi
// algorithm. It is used by jieba for unknown words.
func Cut(s string) []string {
	result := make([]string, 0, len(s))
lop:
	for {
		hanLoc := reHan.FindStringIndex(s)
		if hanLoc == nil {
			if len(s) == 0 {
				break
			}
		} else if hanLoc[0] == 0 {
			hans := s[hanLoc[0]:hanLoc[1]]
			s = s[hanLoc[1]:]
			result = append(result, cutHan(hans)...)
			continue
		}
		nonhanLoc := reSkip.FindStringIndex(s)
		if nonhanLoc == nil {
			if len(s) == 0 {
				break
			}
		} else if nonhanLoc[0] == 0 {
			nonhans := s[nonhanLoc[0]:nonhanLoc[1]]
			s = s[nonhanLoc[1]:]
			if nonhans != "" {
				result = append(result, nonhans)
				continue
			}
		}
		var loc []int
		switch {
		case hanLoc == nil && nonhanLoc == nil:
			if len(s) > 0 {
				result = append(result, s)
				break lop
			}
		case hanLoc == nil:
			loc = nonhanLoc
		case nonhanLoc == nil:
			loc = hanLoc
		case hanLoc[0] < nonhanLoc[0]:
			loc = hanLoc
		default:
			loc = nonhanLoc
		}
		result = append(result, s[:loc[0]])
		s = s[loc[0]:]
	}

	return result
}
