package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
)

var (
	decodeFlag = flag.Bool("d", false, "Decode mode")
	helpFlag   = flag.Bool("h", false, "Show help")
	widthFlag  = flag.Int("w", 0, "Wrap output at specified width (0 for no wrapping)")
)

func usage() {
	fmt.Fprintf(os.Stderr, "Encode data to Cyrillic characters and back.\n\n")
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

	encodeMap := make(map[byte]rune)
	decodeMap := make(map[rune]byte)
	for i := 0; i < 256; i++ {
		encodeMap[byte(i)] = rune(0x0400 + i)
		decodeMap[rune(0x0400+i)] = byte(i)
	}

	reader := bufio.NewReader(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)

	var err error
	if *decodeFlag {
		err = decode(reader, writer, decodeMap)
	} else {
		err = encode(reader, writer, encodeMap, *widthFlag)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := writer.Flush(); err != nil {
		fmt.Fprintf(os.Stderr, "Error flushing output: %v\n", err)
		os.Exit(1)
	}
}

func encode(reader *bufio.Reader, writer *bufio.Writer, encodeMap map[byte]rune, width int) error {
	buffer := make([]byte, 1024)
	count := 0

	for {
		n, err := reader.Read(buffer)
		if err != nil && err != io.EOF {
			return fmt.Errorf("error reading input: %w", err)
		}

		for i := 0; i < n; i++ {
			if _, err := writer.WriteRune(encodeMap[buffer[i]]); err != nil {
				return fmt.Errorf("error writing output: %w", err)
			}
			count++
			if width > 0 && count == width {
				if _, err := writer.WriteString("\r\n"); err != nil {
					return fmt.Errorf("error writing line break: %w", err)
				}
				count = 0
			}
		}

		if err == io.EOF {
			break
		}
	}

	if width > 0 && count > 0 {
		if _, err := writer.WriteString("\r\n"); err != nil {
			return fmt.Errorf("error writing final line break: %w", err)
		}
	}

	return nil
}

func decode(reader *bufio.Reader, writer *bufio.Writer, decodeMap map[rune]byte) error {
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
		}
	}
	return nil
}