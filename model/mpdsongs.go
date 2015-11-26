package model

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"io"
	l "k.prv/rpimon/logging"
	"time"
)

type Song struct {
	ID     int64
	Date   time.Time
	Track  string
	Name   string
	Album  string
	Artist string
	Title  string
	File   string
}

func getKey(id int64) []byte {
	key := make([]byte, 8)
	binary.PutVarint(key, id)
	return append(mpdSongPrefix, key...)
}

func decodeSong(buff []byte) (s *Song) {
	s = &Song{}
	r := bytes.NewBuffer(buff)
	dec := gob.NewDecoder(r)
	if err := dec.Decode(s); err != nil {
		l.Warn("model.decodeSong decode error: %s", err)
		return nil
	}
	return
}

func (s *Song) Save() (err error) {
	l.Info("model.Song.Save %#v", s)
	if s.ID == 0 {
		s.ID, err = db.db.Inc(mpdSongIDkey, 1)
		if err != nil {
			l.Error("model.Song.Save get key error: %s", err)
			return
		}
		l.Debug("model.Song.Save new id=%#v", s.ID)
	}
	r := new(bytes.Buffer)
	enc := gob.NewEncoder(r)
	if err = enc.Encode(s); err != nil {
		l.Warn("model.Song.Save encode error: %s - %s", s, err)
		return
	}
	if err = db.db.Set(getKey(s.ID), r.Bytes()); err != nil {
		l.Warn("model.Song.Save set error %s: %s", s, err)
	}
	return
}

func GetSongs() (songs []*Song) {
	en, _, err := db.db.Seek(mpdSongPrefix)
	if err != nil {
		return
	}
	for {
		key, value, err := en.Next()
		if err == io.EOF || !bytes.HasPrefix(key, mpdSongPrefix) {
			break
		}
		if err == nil {
			songs = append(songs, decodeSong(value))
		} else {
			l.Error("model.GetSongs next error: %s", err)
		}
	}
	return
}

// DeleteUser from database by login
func DeleteSong(id int64) error {
	return db.db.Delete(getKey(id))
}

func DeleteSongs(maxAge time.Time) {
	en, _, err := db.db.Seek(mpdSongPrefix)
	if err != nil {
		return
	}
	toDel := make([][]byte, 0)
	for {
		key, value, err := en.Next()
		if err == io.EOF || !bytes.HasPrefix(key, mpdSongPrefix) {
			break
		}
		if err == nil {
			song := decodeSong(value)
			if maxAge.After(song.Date) {
				toDel = append(toDel, key)
			}
		} else {
			l.Error("model.DeleteSongs next error: %s", err)
		}
	}
	if len(toDel) > 0 {
		l.Info("model.DeleteSongs delete: %i", len(toDel))
		for _, key := range toDel {
			db.db.Delete(key)
		}
	}
}
