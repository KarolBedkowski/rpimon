package logs

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
)

type logsDef struct {
	Name     string `json:"name"`
	Filename string `json:"filename, omitempty"`
	Dir      string `json:"dir,omitempty"`
	Prefix   string `json:"prefix,omitempty"`
	Limit    int    `json:"limit,omitempty"`
	Command  string `json:"command,omitempty"`
}

type logsGroup struct {
	Name string    `json:"name"`
	Logs []logsDef `json:"logs"`
}

type configuration struct {
	Groups []logsGroup `json:"groups"`
}

var config configuration

// Init utils pages
func Init(filename string) error {
	log.Print("pages.files Loading configuration file ", filename)
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatal("Error: ", err.Error())
		return err
	}
	err = json.Unmarshal(file, &config)
	if err != nil {
		log.Fatal("Error: ", err.Error())
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
