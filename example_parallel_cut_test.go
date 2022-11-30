package jieba

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

type line struct {
	number int
	text   string
}

var (
	segmenter  *Segmenter
	numThreads = runtime.NumCPU()
	task       = make(chan line, numThreads)
	result     = make(chan line, numThreads)
)

func worker() {
	for l := range task {
		segments := segmenter.Cut(l.text, true)

		l.text = fmt.Sprintf("%s\n", strings.Join(segments, " / "))
		result <- l
	}
}

func Example_parallelCut() {
	// Set the number of goroutines
	runtime.GOMAXPROCS(numThreads)

	// open file for segmentation
	file, err := os.Open("README.md")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Load dictionary
	segmenter, err = LoadDictionaryAt("dict.txt")
	if err != nil {
		log.Fatal(err)
	}

	// start worker routines
	for i := 0; i < numThreads; i++ {
		go worker()
	}

	var length, size int
	scanner := bufio.NewScanner(file)

	t0 := time.Now()

	lines := make([]string, 0)

	// Read lines
	for scanner.Scan() {
		t := scanner.Text()
		size += len(t)
		lines = append(lines, t)
	}
	length = len(lines)

	// Segmentation
	go func() {
		for i := 0; i < length; i++ {
			task <- line{number: i, text: lines[i]}
		}
		close(task)
	}()

	// Make sure the segmentation result contains same line as original file
	for i := 0; i < length; i++ {
		l := <-result
		lines[l.number] = l.text
	}

	t1 := time.Now()

	// Write the segments into a file for verify
	outputFile, _ := os.OpenFile("parallelCut.log", os.O_CREATE|os.O_WRONLY, 0600)
	defer outputFile.Close()
	writer := bufio.NewWriter(outputFile)
	for _, l := range lines {
		writer.WriteString(l)
	}
	writer.Flush()

	log.Printf("Time cousumed: %v", t1.Sub(t0))
	log.Printf("Segmentation speed: %f MB/s", float64(size)/t1.Sub(t0).Seconds()/(1024*1024))
}
