package posseg_test

import (
	"fmt"

	"github.com/fumiama/jieba/posseg"
)

func Example() {
	var seg posseg.Segmenter
	seg.LoadDictionaryAt("../dict.txt")

	for segment := range seg.Cut("我爱北京天安门", true) {
		fmt.Printf("%s %s\n", segment.Text(), segment.Pos())
	}
	// Output:
	// 我 r
	// 爱 v
	// 北京 ns
	// 天安门 ns
}
