package model

import (
	"encoding/gob"
	"github.com/cznic/kv"
	"k.prv/rpimon/helpers"
	l "k.prv/rpimon/helpers/logging"
	"sync"
)

type (
	DB struct {
		dbFilename string
		db         *kv.DB

		mu sync.RWMutex
	}
)

var (
	userPrefix = []byte("U_")
	taskPrefix = []byte("T_")
	tasksIDkey = []byte("_KEY_TASK")
	db         = &DB{}
)

func init() {
	gob.Register(&User{})
	gob.Register(&Task{})
}

func Open(filename string) (err error) {
	l.Debug("globals.openDatabases START %s", filename)

	db.dbFilename = filename

	dbOpts := &kv.Options{
		VerifyDbBeforeOpen:  true,
		VerifyDbAfterOpen:   true,
		VerifyDbBeforeClose: true,
		VerifyDbAfterClose:  true,
	}

	if helpers.FileExists(db.dbFilename) {
		db.db, err = kv.Open(db.dbFilename, dbOpts)
	} else {
		db.db, err = kv.Create(db.dbFilename, dbOpts)
	}
	if err != nil {
		l.Error("DB.open db error: %v", err)
		panic("DB.open  db error " + err.Error())
	}
	if u := GetUserByLogin("admin"); u == nil {
		l.Info("DB.openDatabases creating 'admin' user with password 'admin'")
		admin := &User{
			Login:    "admin",
			Password: CreatePassword("admin"),
			Privs:    []string{"admin", "mpd", "files", "notepad", "worker"},
		}
		AddUser(admin)
	}
	if GetUserByLogin("admin") == nil {
		panic("missing admin")
	}
	l.Debug("DB.openDatabases DONE")
	return
}

func Close() error {
	l.Info("DB.Close")
	db.mu.Lock()
	defer db.mu.Unlock()
	if db.db != nil {
		db.db.Close()
		db.db = nil
	}
	l.Info("DB.Close DONE")
	return nil
}
