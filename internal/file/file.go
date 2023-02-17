package file

import (
	"bufio"
	"os"
	"strings"
)

// ReadFile line by line into a string array
func ReadFile(path string) (lines []string, err error) {
	// Open file
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read file
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if text := scanner.Text(); strings.ReplaceAll(text, " ", "") != "" {
			lines = append(lines, scanner.Text())
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return
}
