package posseg_test

import (
	"fmt"

	"github.com/fumiama/jieba/posseg"
)

func Example() {
	seg, err := posseg.LoadDictionaryAt("../dict.txt")
	if err != nil {
		panic(err)
	}

	for _, segment := range seg.Cut("我爱北京天安门", true) {
		fmt.Printf("%s %s\n", segment.Text(), segment.Pos())
	}
	// Output:
	// 我 r
	// 爱 v
	// 北京 ns
	// 天安门 ns
}
