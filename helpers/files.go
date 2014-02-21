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
	l.Debug("helpers.ReadLineFromFile %s", filename)
	file, err := os.Open(filename)
	if err != nil {
		l.Warn("helpers.ReadLineFromFile Error %s: %s", filename, err)
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
func ReadIntFromFile(filename string) (int, error) {
	l.Debug("helpers.ReadIntFromFile %s", filename)
	line, err := ReadLineFromFile(filename)
	if err != nil {
		l.Warn("helpers.ReadIntFromFile %s error: %s", filename, err.Error())
		return 0, err
	}
	if len(line) == 0 {
		return 0, nil
	}
	res, err := strconv.Atoi(line)
	if err != nil {
		l.Warn("helpers.ReadIntFromFile Error %s: %s", filename, err.Error())
		return 0, err
	}
	return res, nil
}

// ReadFile read last n lines from file
func ReadFile(filename string, limit int) (string, error) {
	l.Debug("helpers.ReadLineFromFile %s, %d", filename, limit)
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
func ReadCommand(command string, arg ...string) string {
	l.Debug("helpers.ReadCommand %s %v", command, arg)
	out, err := exec.Command(command, arg...).CombinedOutput()
	outstr := string(out)
	if err != nil {
		l.Warn("helpers.ReadCommand error %s, %v, %s", command, arg, err.Error())
		outstr += "\n\n" + err.Error()
	}
	return outstr
}

// AppendToFile add given data do file
func AppendToFile(filename, data string) error {
	l.Debug("helpers.AppendToFile %s data_len=%d", filename, len(data))
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
