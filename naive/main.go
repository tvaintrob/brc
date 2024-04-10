package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

var measurementsFile = flag.String(
	"measurements",
	"../1brc/measurements100.txt",
	"Path to the measurements file",
)

func main() {
	flag.Parse()
	defer trackTime(time.Now(), "main")

	file, err := os.Open(*measurementsFile)
	if err != nil {
		panic(err)
	}

	stations := map[string]*StationStats{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		split := strings.SplitN(scanner.Text(), ";", 2)
		station := split[0]
		measurement, err := strconv.ParseFloat(split[1], 32)
		if err != nil {
			panic(err)
		}

		stats, ok := stations[station]
		if !ok {
			stations[station] = &StationStats{
				Min:   measurement,
				Max:   measurement,
				Mean:  measurement,
				Count: 1,
			}
		} else {
			if measurement < stats.Min {
				stats.Min = measurement
			}

			if measurement > stats.Max {
				stats.Max = measurement
			}

			stats.Mean = (stats.Mean*float64(stats.Count) + measurement) / (float64(stats.Count) + 1)
			stats.Count += 1
		}
	}

	sorted := sortKeys(stations)
	fmt.Print("{")
	for i, key := range sorted {
		stats := stations[key]
		if i != 0 {
			fmt.Print(", ")
		}

		fmt.Printf("%s=%.1f/%.1f/%.1f", key, stats.Min, stats.Mean, stats.Max)
	}
	fmt.Println("}")
}

type StationStats struct {
	Min   float64
	Max   float64
	Mean  float64
	Count int
}

func trackTime(start time.Time, name string) {
	elapsed := time.Since(start)
	fmt.Fprintf(os.Stderr, "timed execution finished: name=%s, elapsed=%s\n", name, elapsed)
}

func sortKeys[V any](m map[string]V) []string {
	defer trackTime(time.Now(), "sortedKeys")

	i := 0
	keys := make([]string, len(m))
	for key := range m {
		keys[i] = key
		i += 1
	}

	sort.Strings(keys)
	return keys
}
