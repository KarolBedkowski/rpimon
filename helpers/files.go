package helpers

import (
	"bufio"
	l "k.prv/rpimon/helpers/logging"
	"os"
	"strconv"
	"strings"
)

// Read one line from given file
func ReadLineFromFile(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		l.Warn("ReadLineFromFile Error", filename, err)
		return "", err
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	line, err := reader.ReadString('\n')
	if err == nil {
		line = strings.Trim(line, " \n")
	}
	return line, err
}

// Read first line from givern file and return value as int.
func ReadIntFromFile(filename string) int {
	line, err := ReadLineFromFile(filename)
	if err != nil {
		l.Warn("ReadIntFromFile Error", filename, err)
		return 0
	}
	if len(line) == 0 {
		return 0
	}
	res, err := strconv.Atoi(line)
	if err != nil {
		l.Warn("ReadIntFromFile Error", filename, err)
		return 0
	}
	return res
}
