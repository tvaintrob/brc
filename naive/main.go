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

	defer trackTime(time.Now())
	file, err := os.Open(*measurementsFile)
	if err != nil {
		panic(err)
	}

	stations := map[string]*StationStats{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		split := strings.SplitN(scanner.Text(), ";", 2)
		station := split[0]
		measurement64, err := strconv.ParseFloat(split[1], 32)
		if err != nil {
			panic(err)
		}

		measurement := float32(measurement64)
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

			stats.Mean = (stats.Mean*float32(stats.Count) + measurement) / (float32(stats.Count) + 1)
			stats.Count += 1
		}
	}

	sorted := sortedKeys(stations)
	fmt.Print("{")
	for i, key := range sorted {
		stats := stations[key]
		if i == 0 {
			fmt.Printf("%s=%.1f/%.1f/%.1f", key, stats.Min, stats.Mean, stats.Max)
		} else {
			fmt.Printf(", %s=%.1f/%.1f/%.1f", key, stats.Min, stats.Mean, stats.Max)
		}
	}
	fmt.Println("}")
}

type StationStats struct {
	Min   float32
	Max   float32
	Mean  float32
	Count int
}

func trackTime(start time.Time) {
	elapsed := time.Since(start)
	fmt.Fprintf(os.Stderr, "timed execution finished: elapsed=%s\n", elapsed)
}

func sortedKeys[V any](m map[string]V) []string {
	i := 0
	keys := make([]string, len(m))
	for key := range m {
		keys[i] = key
		i += 1
	}

	sort.Strings(keys)
	return keys
}
