package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"time"
)

var (
	decodeFlag = flag.Bool("d", false, "Decode mode")
	helpFlag   = flag.Bool("h", false, "Show help")
	widthFlag  = flag.Int("w", 0, "Number of encoded characters per line (0 for no wrapping)")
)

const bufferSize = 1024 * 1024 // 1MB buffer

func usage() {
	fmt.Fprintf(os.Stderr, "Encode binary data to Cyrillic characters and back.\n\n")
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] < infile > outfile\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Options:\n")
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if *helpFlag {
		flag.Usage()
		os.Exit(0)
	}

	encodeMap, decodeMap := createMaps()

	reader := bufio.NewReaderSize(os.Stdin, bufferSize)
	writer := bufio.NewWriterSize(os.Stdout, bufferSize)

	start := time.Now()
	var err error
	if *decodeFlag {
		err = decode(reader, writer, decodeMap)
	} else {
		err = encode(reader, writer, encodeMap, *widthFlag)
	}
	duration := time.Since(start)

	if err != nil {
		fmt.Fprintf(os.Stderr, "\nError: %v\n", err)
		os.Exit(1)
	}

	if err := writer.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "\nError flushing output: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "\nOperation completed in %v\n", duration)
}

func createMaps() (map[byte]rune, map[rune]byte) {
	encodeMap := make(map[byte]rune)
	decodeMap := make(map[rune]byte)
	
	ranges := []struct{start, end rune}{
		{0x0400, 0x045F},
		{0x0460, 0x0481},
		{0x0490, 0x04FF},
		{0x0500, 0x052F},
	}
	
	i := 0
	for _, r := range ranges {
		for c := r.start; c <= r.end && i < 256; c++ {
			encodeMap[byte(i)] = c
			decodeMap[c] = byte(i)
			i++
		}
	}
	
	return encodeMap, decodeMap
}

func encode(reader *bufio.Reader, writer *bufio.Writer, encodeMap map[byte]rune, width int) error {
	totalBytes := 0
	lineBuffer := make([]rune, 0, width)

	for {
		b, err := reader.ReadByte()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading input: %w", err)
		}

		r := encodeMap[b]
		lineBuffer = append(lineBuffer, r)

		if width > 0 && len(lineBuffer) == width {
			if _, err := writer.WriteString(string(lineBuffer) + "\r\n"); err != nil {
				return fmt.Errorf("error writing output: %w", err)
			}
			lineBuffer = lineBuffer[:0]
		}

		totalBytes++
		if totalBytes%bufferSize == 0 {
			fmt.Fprintf(os.Stderr, "\rProcessed: %d MB", totalBytes/1024/1024)
		}
	}

	// Write any remaining data
	if len(lineBuffer) > 0 {
		if _, err := writer.WriteString(string(lineBuffer)); err != nil {
			return fmt.Errorf("error writing final output: %w", err)
		}
	}

	fmt.Fprint(os.Stderr, "\n")
	return nil
}

func decode(reader *bufio.Reader, writer *bufio.Writer, decodeMap map[rune]byte) error {
	totalBytes := 0

	for {
		r, _, err := reader.ReadRune()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("error reading input: %w", err)
		}

		if b, ok := decodeMap[r]; ok {
			if err := writer.WriteByte(b); err != nil {
				return fmt.Errorf("error writing output: %w", err)
			}
			totalBytes++
			if totalBytes%bufferSize == 0 {
				fmt.Fprintf(os.Stderr, "\rProcessed: %d MB", totalBytes/1024/1024)
			}
		}
		// Ignore line breaks and unknown characters
	}

	fmt.Fprint(os.Stderr, "\n")
	return nil
}