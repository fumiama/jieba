# jieba

[![GoDoc](https://godoc.org/github.com/fumiama/jieba?status.svg)](https://godoc.org/github.com/fumiama/jieba)

[结巴分词](https://github.com/fxsjy/jieba) 是由 [@fxsjy](https://github.com/fxsjy) 使用 Python 编写的中文分词组件，本仓库是结巴分词的 Golang 语言实现，修改于[jiebago](https://github.com/wangbin/jiebago)，优化了速度与性能，增加了从`fs.File`加载字典等功能。


## 使用

```
go get -d github.com/fumiama/jieba
```

## 示例

```
package main

import (
        "fmt"

        "github.com/fumiama/jieba"
)

func main() {
	seg, err := LoadDictionaryAt("dict.txt")
	if err != nil {
		panic(err)
	}

	fmt.Print("【全模式】：")
	fmt.Println(seg.CutAll("我来到北京清华大学"))

	fmt.Print("【精确模式】：")
	fmt.Println(seg.Cut("我来到北京清华大学", false))

	fmt.Print("【新词识别】：")
	fmt.Println(seg.Cut("他来到了网易杭研大厦", true))

	fmt.Print("【搜索引擎模式】：")
	fmt.Println(seg.CutForSearch("小明硕士毕业于中国科学院计算所，后在日本京都大学深造", true))
}
```
输出结果：

```
【全模式】：[我 来到 北京 清华 清华大学 华大 大学]
【精确模式】：[我 来到 北京 清华大学]
【新词识别】：[他 来到 了 网易 杭研 大厦]
【搜索引擎模式】：[小明 硕士 毕业 于 中国 科学 学院 科学院 中国科学院 计算 计算所 ， 后 在 日本 京都 大学 日本京都大学 深造]
```

更多信息请参考[文档](https://godoc.org/github.com/fumiama/jieba)。

## 分词速度
```c
goos: darwin
goarch: amd64
pkg: github.com/fumiama/jieba
cpu: Intel(R) Core(TM) i5-8265U CPU @ 1.60GHz
BenchmarkCutNoHMM-8            	   50101	     22889 ns/op	   4.67 MB/s	   24492 B/op	     148 allocs/op
BenchmarkCut-8                 	   47473	     25152 ns/op	   4.25 MB/s	   31310 B/op	     185 allocs/op
BenchmarkCutAll-8              	   81760	     14286 ns/op	   7.49 MB/s	   22746 B/op	      75 allocs/op
BenchmarkCutForSearchNoHMM-8   	   49009	     24371 ns/op	   4.39 MB/s	   26421 B/op	     157 allocs/op
BenchmarkCutForSearch-8        	   44643	     26597 ns/op	   4.02 MB/s	   33224 B/op	     194 allocs/op
PASS
ok  	github.com/fumiama/jieba	8.769s
```
