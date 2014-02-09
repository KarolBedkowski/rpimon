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

// ReadFile read last n lines from file
func ReadFile(filename string, limit int) (string, error) {
	if limit < 0 {
		lines, err := ioutil.ReadFile(filename)
		return string(lines), err
	}
	file, err := os.Open(filename)
	if err != nil {
		l.Warn("helpers.ReadFile %s error %s", filename, err)
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

// ReadCommand read result command
func ReadCommand(name string, arg ...string) string {
	l.Debug("helpers.ReadCommand %s %v", name, arg)
	out, err := exec.Command(name, arg...).CombinedOutput()
	if err != nil {
		l.Warn("helpers.ReadCommand error %s, %v, %s", name, arg, err.Error())
		return err.Error()
	}
	return string(out)
}

// AppendToFile add given data do file
func AppendToFile(filename, data string) error {
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND, 0660)
	if err != nil {
		f, err = os.Create(filename)
	}
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(data)
	return err
}
