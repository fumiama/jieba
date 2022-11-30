package finalseg

import (
	"math"
	"testing"
)

func TestViterbi(t *testing.T) {
	obs := "我们是程序员"
	prob, path := viterbi([]rune(obs), 'B', 'M', 'E', 'S')
	if math.Abs(prob+39.68824128493802) > 1e-10 {
		t.Fatal(prob)
	}
	for index, state := range []byte{'B', 'E', 'S', 'B', 'M', 'E'} {
		if path[index] != state {
			t.Fatal(path)
		}
	}
}

func TestCutHan(t *testing.T) {
	obs := "我们是程序员"
	result := cutHan(obs)
	if len(result) != 3 {
		t.Fatal(result)
	}
	if result[0] != "我们" {
		t.Fatal(result[0])
	}
	if result[1] != "是" {
		t.Fatal(result[1])
	}
	if result[2] != "程序员" {
		t.Fatal(result[2])
	}
}

func TestCut(t *testing.T) {
	sentence := "我们是程序员"
	result := Cut(sentence)
	if len(result) != 3 {
		t.Fatal(len(result))
	}
	if result[0] != "我们" {
		t.Fatal(result[0])
	}
	if result[1] != "是" {
		t.Fatal(result[1])
	}
	if result[2] != "程序员" {
		t.Fatal(result[2])
	}
	result2 := Cut("I'm a programmer!")
	if len(result2) != 8 {
		t.Fatal(result2)
	}
	result3 := Cut("程序员average年龄28.6岁。")
	if len(result3) != 6 {
		t.Fatal(result3)
	}

}
