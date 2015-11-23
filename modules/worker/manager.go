package worker

import (
	"io/ioutil"
	l "k.prv/rpimon/helpers/logging"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
)

var tskCntr int32 = 0

type worker struct {
	workerPool chan chan Job
	JobChannel chan Job
}

func (w worker) start() {
	go func() {
		for {
			w.workerPool <- w.JobChannel
			select {
			case job := <-w.JobChannel:
				job.run()
			}
		}
	}()
}

type Dispatcher struct {
	workerPool chan chan Job
	maxWorkers int
	JobQueue   chan Job
}

func NewDispatcher(maxWorkers int) *Dispatcher {
	return &Dispatcher{
		workerPool: make(chan chan Job, maxWorkers),
		maxWorkers: maxWorkers,
		JobQueue:   make(chan Job),
	}
}

func (d *Dispatcher) Run() {
	// starting n number of workers
	for i := 0; i < d.maxWorkers; i++ {
		worker := worker{
			workerPool: d.workerPool,
			JobChannel: make(chan Job),
		}
		worker.start()
	}

	go d.dispatch()
}

func (d *Dispatcher) Add(j Job) {
	d.JobQueue <- j
}

func (d *Dispatcher) dispatch() {
	for {
		select {
		case job := <-d.JobQueue:
			go func(job Job) {
				jobChannel := <-d.workerPool
				jobChannel <- job
			}(job)
		}
	}
}

type Job struct {
	task *Task
}

func (j *Job) run() {
	l.Info("job.run start %s (%s)", j.task.Label, j.task.Command)

	j.task.mu.Lock()
	now := time.Now()
	j.task.Started = &now
	cntr := atomic.AddInt32(&tskCntr, 1)
	j.task.LogFile = now.Format("2006-01-02_15-04-05.00000") +
		"-" + j.task.Label + "-" + strconv.Itoa(int(cntr)) + ".log"
	j.task.mu.Unlock()

	output, err := os.Create(path.Join(getLogsDir(), j.task.LogFile))
	if err != nil {
		l.Error("job.run creating output file error %s: %s", j.task.LogFile, err)
		j.task.mu.Lock()
		defer j.task.mu.Unlock()
		j.task.Error = err.Error()
		return
	}
	defer output.Close()

	err = execute(j.task, output)
	j.task.mu.Lock()
	defer j.task.mu.Unlock()
	if err != nil {
		j.task.Error = err.Error()
	}
	now = time.Now()
	j.task.Finished = &now
	l.Info("job.run finished %s %s", j.task.Label, j.task.Command)
}

func getLogsDir() (name string) {
	conf := Module.GetConfiguration()
	name, _ = filepath.Abs(conf["Logs_Dir"])
	os.MkdirAll(name, 0770)
	return
}

func execute(task *Task, output *os.File) (err error) {
	args := strings.Split(task.Params, "\n")
	cmd := exec.Command(task.Command, args...)
	cmd.Stderr = output
	cmd.Stdout = output
	cmd.Dir = task.Dir

	l.Info("worker.execute cmd: %#v", cmd)
	if err = cmd.Start(); err != nil {
		return err
	}
	err = cmd.Wait()

	return
}

func deleteOldLogs() error {
	time.Sleep(time.Duration(30) * time.Second)
	l.Info("Start deleteOldLogs")
	now := time.Now()
	now = now.Add(time.Duration(-7*24) * time.Hour)
	conf := Module.GetConfiguration()
	logdir, _ := filepath.Abs(conf["Logs_Dir"])
	if files, err := ioutil.ReadDir(logdir); err == nil {
		for _, file := range files {
			if file.IsDir() || !strings.HasSuffix(file.Name(), ".log") {
				continue
			}
			if now.After(file.ModTime()) {
				l.Info("Delete %s", file.Name())
				os.Remove(file.Name())
			}
		}
	} else {
		l.Error("deleteOldLogs error: %s", err.Error())
	}
	l.Info("deleteOldLogs finished")
	return nil
}
