package model

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"io"
	l "k.prv/rpimon/helpers/logging"
	"sync"
	"time"
)

type Task struct {
	ID       int
	Label    string
	Command  string
	Dir      string
	Params   string
	Started  *time.Time
	Finished *time.Time
	LogFile  string
	Error    string
	Multi    bool

	mu sync.RWMutex
}

func (t *Task) Clone() *Task {
	return &Task{
		Label:    t.Label,
		Command:  t.Command,
		Dir:      t.Dir,
		Params:   t.Params,
		Started:  t.Started,
		Finished: t.Finished,
		LogFile:  t.LogFile,
		Error:    t.Error,
		Multi:    t.Multi,
	}
}

func (t *Task) Validate() error {
	if t.Command == "" {
		return errors.New("Missing command")
	} else if t.Label == "" {
		t.Label = t.Command
	}
	return nil
}

func GetTask(id int) (u *Task) {
	v, err := db.db.Get(nil, taskID2key(id))
	if err != nil {
		l.Warn("model.GetTask error: %s", err)
	}
	if v == nil {
		l.Debug("model.GetTask user not found login=%s", id)
		return nil
	}
	return decodeTask(v)
}

func taskID2key(id int) []byte {
	key := make([]byte, 8)
	binary.PutVarint(key, int64(id))
	return append(taskPrefix, key...)
}

func decodeTask(buff []byte) (t *Task) {
	t = &Task{}
	r := bytes.NewBuffer(buff)
	dec := gob.NewDecoder(r)
	if err := dec.Decode(t); err == nil {
		return t
	} else {
		l.Warn("model.decodeTask decode error: %s", err)
	}
	return nil
}

func SaveTask(t *Task) (err error) {
	l.Info("model.SaveTask %#v", t)
	if t.ID == 0 {
		var newID int64
		newID, err = db.db.Inc(tasksIDkey, 1)
		if err != nil {
			l.Error("model.SaveTask get key error: %s", err)
			return
		}
		t.ID = int(newID)
		l.Debug("model.SaveTask new id=%#v", t.ID)
	}
	r := new(bytes.Buffer)
	enc := gob.NewEncoder(r)
	if err = enc.Encode(t); err != nil {
		l.Warn("model.SaveTask encode error: %s - %s", t, err)
		return
	}
	if err = db.db.Set(taskID2key(t.ID), r.Bytes()); err != nil {
		l.Warn("model.SaveTask set error %s: %s", t, err)
	}
	return
}

func GetTasks() (tasks []*Task) {
	en, _, err := db.db.Seek(taskPrefix)
	if err != nil {
		return
	}
	for {
		key, value, err := en.Next()
		if err == io.EOF || !bytes.HasPrefix(key, taskPrefix) {
			break
		}
		if err == nil {
			tasks = append(tasks, decodeTask(value))
		} else {
			l.Error("model.GetTasks next error: %s", err)
		}
	}
	return
}

// DeleteUser from database by login
func DeleteTask(id int) error {
	return db.db.Delete(taskID2key(id))
}

func DeleteOldTasks(maxAge time.Time) {
	en, _, err := db.db.Seek(taskPrefix)
	if err != nil {
		return
	}
	toDel := make([][]byte, 0)
	for {
		key, value, err := en.Next()
		if err == io.EOF || !bytes.HasPrefix(key, taskPrefix) {
			break
		}
		if err == nil {
			task := decodeTask(value)
			if task.Finished != nil && maxAge.After(*task.Finished) {
				toDel = append(toDel, key)
			}
		} else {
			l.Error("model.DeleteOldTasks next error: %s", err)
		}
	}
	if len(toDel) > 0 {
		l.Info("model.DeleteOldTasks delete: %i", len(toDel))
		for _, key := range toDel {
			db.db.Delete(key)
		}
	}
}
