package main

import (
	"fmt"
	"os"
	"sort"
	"sync"
	"time"
)

func trackTime(start time.Time, name string) {
	elapsed := time.Since(start)
	fmt.Fprintf(os.Stderr, "timed execution finished: name=%s, elapsed=%s\n", name, elapsed)
}

func sortKeys[V any](m map[string]V) []string {
	i := 0
	keys := make([]string, len(m))
	for key := range m {
		keys[i] = key
		i += 1
	}

	sort.Strings(keys)
	return keys
}

func mergeChannels[T any](cs ...<-chan T) <-chan T {
	var wg sync.WaitGroup
	out := make(chan T)

	output := func(c <-chan T) {
		defer wg.Done()
		for n := range c {
			out <- n
		}
	}

	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func printStations(stations map[string]*StationStats) {
	sorted := sortKeys(stations)
	fmt.Print("{")
	for i, key := range sorted {
		stats := stations[key]
		if i != 0 {
			fmt.Print(", ")
		}

		fmt.Printf("%s=%.1f/%.1f/%.1f", key, stats.Min, stats.Sum/float64(stats.Count), stats.Max)
	}
	fmt.Println("}")
}
