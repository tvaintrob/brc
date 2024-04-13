package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

var measurementsFile = flag.String(
	"measurements",
	"../1brc/measurements.txt",
	"Path to the measurements file",
)

type StationStats struct {
	Min   float64
	Max   float64
	Sum   float64
	Count int
}

var workerCount = runtime.NumCPU() * 10

func main() {
	flag.Parse()
	defer trackTime(time.Now(), "main")

	file, err := os.Open(*measurementsFile)
	if err != nil {
		panic(err)
	}

	chunks := splitToChunks(file, 1024*1024*50) // 500 mb chunks
	processors := make([]<-chan map[string]*StationStats, 0, runtime.NumCPU())
	for range workerCount {
		processors = append(processors, processChunks(chunks))
	}

	stations := map[string]*StationStats{}
	for result := range mergeChannels(processors...) {
		for key, val := range result {
			stats, ok := stations[key]
			if !ok {
				stations[key] = val
			} else {
				stats.Min = min(stats.Min, val.Min)
				stats.Max = max(stats.Max, val.Max)
				stats.Sum += val.Sum
				stats.Count += val.Count
			}
		}
	}

	printStations(stations)
}

func splitToChunks(r io.Reader, chunkSize int) <-chan []byte {
	chunks := make(chan []byte)

	go func() {
		buf := make([]byte, chunkSize)
		carry := make([]byte, 0, chunkSize)

		for {
			n, err := r.Read(buf)
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				}
				panic(err)
			}

			buf = buf[:n]
			chunk := make([]byte, n)
			copy(chunk, buf)

			lastNewLineIndex := bytes.LastIndexByte(buf, '\n')

			chunk = append(carry, buf[:lastNewLineIndex+1]...)
			carry = make([]byte, len(buf[lastNewLineIndex+1:]))
			copy(carry, buf[lastNewLineIndex+1:])

			chunks <- chunk
		}

		close(chunks)
	}()

	return chunks
}

func processChunks(chunks <-chan []byte) <-chan map[string]*StationStats {
	var wg sync.WaitGroup
	result := make(chan map[string]*StationStats)

	for chunk := range chunks {
		go func(chunk []byte) {
			wg.Add(1)
			stations := map[string]*StationStats{}
			scanner := bufio.NewScanner(bytes.NewReader(chunk))
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
						Sum:   measurement,
						Count: 1,
					}
				} else {
					stats.Min = min(measurement, stats.Min)
					stats.Max = max(measurement, stats.Max)
					stats.Sum += measurement
					stats.Count += 1
				}
			}

			result <- stations
			wg.Done()
		}(chunk)
	}

	go func() {
		wg.Wait()
		close(result)
	}()

	return result
}
