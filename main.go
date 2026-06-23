package main

import (
	"bufio"
	"flag"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
)

func main() {
	inPtr := flag.String("in", "input.txt", "input file path")
	outPtr := flag.String("out", "", "output file path (empty for stdout)")
	sepPtr := flag.String("sep", ",", "separator for ranges")
	flag.Parse()

	file, err := os.Open(*inPtr)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// Pre-allocate memory if the approximate dataset volume is known (e.g., 100k lines)
	realIPs := make([]uint32, 0, 100000)

	// Read the input file line by line using a buffered scanner
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		ipString := scanner.Text() // Note: scanner.Bytes() can be used for further optimization
		if ipString == "" {
			continue
		}

		if sep := strings.Index(ipString, *sepPtr); sep > -1 {
			start := ip2LongZeroAlloc(ipString[:sep])
			end := ip2LongZeroAlloc(ipString[sep+1:])
			for i := start; i <= end; i++ {
				realIPs = append(realIPs, i)
			}
		} else {
			realIPs = append(realIPs, ip2LongZeroAlloc(ipString))
		}
	}

	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	// Fast in-place sorting of primitive integers
	sort.Slice(realIPs, func(i, j int) bool {
		return realIPs[i] < realIPs[j]
	})

	// Initialize the output stream (open the target file once or fall back to stdout)
	var writer *bufio.Writer
	if *outPtr != "" {
		outFile, err := os.Create(*outPtr)
		if err != nil {
			log.Fatal(err)
		}
		defer outFile.Close()
		writer = bufio.NewWriter(outFile)
		defer writer.Flush() // Ensure the buffer is fully flushed to disk at the end
	} else {
		writer = bufio.NewWriter(os.Stdout)
		defer writer.Flush()
	}

	if len(realIPs) == 0 {
		return
	}

	// Range aggregation algorithm
	lastIP := realIPs[0]
	deltaIP := realIPs[0]

	for i := 1; i < len(realIPs); i++ {
		ip := realIPs[i]

		if ip == deltaIP {
			continue // Skip exact duplicates
		}

		if ip != deltaIP+1 {
			// The consecutive sequence broke; write the accumulated range
			writeRange(writer, lastIP, deltaIP)
			lastIP = ip
		}
		deltaIP = ip
	}
	// Write the final remaining tail sequence
	writeRange(writer, lastIP, deltaIP)
}

// writeRange format and write data to the buffered stream
func writeRange(w *bufio.Writer, start, end uint32) {
	if start != end {
		w.WriteString(backtoIP4(int64(start)) + "-" + backtoIP4(int64(end)) + "\n")
	} else {
		w.WriteString(backtoIP4(int64(start)) + "\n")
	}
}

// ip2LongZeroAlloc converts an IPv4 string to uint32 with zero heap allocations and no reflection
func ip2LongZeroAlloc(ip string) uint32 {
	var res, currentByte uint32
	for i := 0; i < len(ip); i++ {
		if ip[i] == '.' {
			res = (res << 8) | currentByte
			currentByte = 0
		} else {
			currentByte = currentByte*10 + uint32(ip[i]-'0')
		}
	}
	return (res << 8) | currentByte
}

// backtoIP4 performs ultra-fast conversion from int64 back to dotted-decimal string representation
func backtoIP4(ipInt int64) string {
	b0 := strconv.FormatInt((ipInt>>24)&0xff, 10)
	b1 := strconv.FormatInt((ipInt>>16)&0xff, 10)
	b2 := strconv.FormatInt((ipInt>>8)&0xff, 10)
	b3 := strconv.FormatInt((ipInt & 0xff), 10)
	return b0 + "." + b1 + "." + b2 + "." + b3
}
