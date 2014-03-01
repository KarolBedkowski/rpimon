package logs

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	l "k.prv/rpimon/helpers/logging"
)

// One log definiton
type logsDef struct {
	// Name of log
	Name string `json:"name"`
	// Filename, if empty - used when no Dir and Command defined
	Filename string `json:"filename, omitempty"`
	// Directory with log files
	Dir string `json:"dir,omitempty"`
	// Prefix of filename when looking for logs in Dir
	Prefix string `json:"prefix,omitempty"`
	// Log lines limit
	Limit int `json:"limit,omitempty"`
	// Command to get log
	Command string `json:"command,omitempty"`
}

// Log group
type logsGroup struct {
	// Name of logs group
	Name string `json:"name"`
	// Logs definition
	Logs []logsDef `json:"logs"`
}

type configuration struct {
	// Logs groups
	Groups []logsGroup `json:"groups"`
}

var config configuration

// Init utils pages
func loadConfiguration(filename string) error {
	l.Info("pages.logs.Init configuration file: %s ", filename)

	if filename == "" {
		return errors.New("missing configuration")
	}
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		l.Error("pages.log.Init read file error: %s", err.Error())
		return err
	}
	err = json.Unmarshal(file, &config)
	if err != nil {
		l.Error("pages.log.Init unmarshal error: %s", err.Error())
	}
	return err
}

func findGroup(page, log string) (result logsDef, group logsGroup, err error) {
	for _, group := range config.Groups {
		if group.Name != page {
			continue
		}
		if log == "" {
			return group.Logs[0], group, nil
		}
		for _, logsdef := range group.Logs {
			if logsdef.Name == log {
				return logsdef, group, nil
			}
		}
	}
	err = errors.New("Invalid log " + page + " / " + log)
	return
}
