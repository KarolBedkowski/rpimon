package helpers

import (
	"bufio"
	"io/ioutil"
	l "k.prv/rpimon/helpers/logging"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// ReadLineFromFile - Read one line from given file
func ReadLineFromFile(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		l.Warn("ReadLineFromFile Error %s: %s", filename, err)
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

// ReadIntFromFile Read first line from given file and return value as int.
func ReadIntFromFile(filename string) int {
	line, err := ReadLineFromFile(filename)
	if err != nil {
		l.Warn("ReadIntFromFile Error %s: %s", filename, err)
		return 0
	}
	if len(line) == 0 {
		return 0
	}
	res, err := strconv.Atoi(line)
	if err != nil {
		l.Warn("ReadIntFromFile Error %s: %s", filename, err)
		return 0
	}
	return res
}

// ReadFromFileLastLines read last n lines from file
func ReadFromFileLastLines(filename string, limit int) (string, error) {
	if limit < 0 {
		lines, err := ioutil.ReadFile(filename)
		return string(lines), err
	}
	file, err := os.Open(filename)
	if err != nil {
		l.Warn("ReadLineFromFile Error", filename, err)
		return "", err
	}
	defer file.Close()
	reader := bufio.NewReader(file)
	buff := make([]string, limit)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		if len(buff) == limit {
			buff = buff[1:]
		}
		buff = append(buff, line)
	}
	return strings.Join(buff, ""), err
}

// ReadFromCommand read result command
func ReadFromCommand(name string, arg ...string) string {
	l.Debug("ReadFromCommand %s %s", name, arg)
	out, err := exec.Command(name, arg...).CombinedOutput()
	if err != nil {
		l.Warn("helpers.ReadFromCommand Error %s, %s, %s", name, arg, err)
		return err.Error()
	}
	return string(out)
}
