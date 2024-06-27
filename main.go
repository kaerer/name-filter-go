package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"slices"
	"strings"
	"sync"

	"cmp"
)

const WORKER_AMOUNT = 1

type nameCount struct {
	Name  string
	Count int
}

func main() {
	argsWithoutProg := os.Args[1:]
	fmt.Println(len(os.Args), os.Args)
	if len(argsWithoutProg) == 0 {
		log.Fatalf("File name needed")
		os.Exit(1)
	}

	nameChan := make(chan string)
	aggregationMutex := &sync.Mutex{}
	var aggregationResults = make(map[string]int)
	var inputFileName string = argsWithoutProg[0]
	f, error := os.Open(inputFileName)
	if error != nil {
		log.Fatalf("Can not open file: %+v\n", inputFileName)
		os.Exit(1)
	}
	defer f.Close()

	var wg sync.WaitGroup
	for i := 0; i < WORKER_AMOUNT; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()

			// Aggregate
			for name := range nameChan {
				aggregationMutex.Lock()
				aggregationResults[name]++
				aggregationMutex.Unlock()
			}
		}(&wg)
	}

	fileReader := csv.NewReader(f)
	for {
		line, err := fileReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("Loop error %+v\n", err)
		}
		//fmt.Printf("%+v\n\n", line)

		// Distribute
		name := strings.ToLower(line[2])
		nameChan <- name
	}
	close(nameChan)

	wg.Wait()

	// Convert the map to a slice of pairs for sorting
	var nameCounts []nameCount
	for name, count := range aggregationResults {
		nameCounts = append(nameCounts, nameCount{name, count})
	}

	// Sort by count in descending order
	slices.SortFunc(nameCounts, func(i, j nameCount) int {
		return -cmp.Compare(i.Count, j.Count)
	})

	// Print the top 10 names
	fmt.Println("Top 10 names:")
	for i := 0; i < 10 && i < len(nameCounts); i++ {
		fmt.Printf("%d: %s with %d occurrences\n", i+1, nameCounts[i].Name, nameCounts[i].Count)
	}
}
