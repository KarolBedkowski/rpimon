package worker

import (
	"errors"
	"sync"
	"time"
)

type Task struct {
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

type workerDb struct {
	mu    sync.RWMutex
	tasks []*Task
}

func (w *workerDb) getTasks() []*Task {
	w.mu.RLock()
	defer w.mu.RUnlock()

	return w.tasks[:]
}

func (w *workerDb) putTask(task *Task) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.tasks = append(w.tasks, task)
}
