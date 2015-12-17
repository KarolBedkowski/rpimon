package model

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
	"io"
	h "k.prv/rpimon/helpers"
	l "k.prv/rpimon/logging"
	"os"
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

func (s *Song) MarshalJSON() ([]byte, error) {
	type Alias Song
	return json.Marshal(&struct {
		DateStr string `json:"DateStr"`
		*Alias
	}{
		DateStr: s.Date.Format("2006-01-02 15:04:05"),
		Alias:   (*Alias)(s),
	})
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

func GetSongsRange(offset, end int) (songs []*Song, total int) {
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
			if total >= offset && total < end {
				songs = append(songs, decodeSong(value))
			}
			total += 1
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

func DeleteOldSongs(maxAge time.Time) {
	en, _, err := db.db.Seek(mpdSongPrefix)
	if err != nil {
		return
	}
	var toDel [][]byte
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
	if len(toDel) == 0 {
		l.Info("model.DeleteSongs none found")
		return
	}
	l.Info("model.DeleteSongs delete: %i", len(toDel))
	if err := db.db.BeginTransaction(); err != nil {
		l.Error("model.DeleteSongs begin transaction error: %s", err)
		return
	}
	for _, key := range toDel {
		if err := db.db.Delete(key); err != nil {
			l.Error("model.DeleteSongs delete error: %s", err)
			db.db.Rollback()
			return
		}
	}
	if err := db.db.Commit(); err != nil {
		l.Error("model.DeleteSongs commit transaction error: %s", err)
		return
	}
}

func DumpOldSongsToFile(maxAge time.Time, filename string, delete bool) {
	l.Info("model.DumpOldSongsToFile maxAge=%s filename=%s delete=%q",
		maxAge, filename, delete)
	en, _, err := db.db.Seek(mpdSongPrefix)
	if err != nil {
		return
	}
	var songs []*Song
	for {
		key, value, err := en.Next()
		if err == io.EOF || !bytes.HasPrefix(key, mpdSongPrefix) {
			break
		}
		if err == nil {
			song := decodeSong(value)
			if maxAge.After(song.Date) {
				songs = append(songs, song)
			}
		}
	}
	if len(songs) == 0 {
		l.Info("model.DumpOldSongsToFile old songs not found")
		return
	}
	if err = DumpToFile(filename, songs); err != nil {
		l.Error("model.DumpOldSongsToFile open file %s error: %s", filename, err)
		return
	}
	if delete {
		l.Info("model.DumpOldSongsToFile delete: %i", len(songs))
		if err := db.db.BeginTransaction(); err != nil {
			l.Error("model.DumpOldSongsToFile begin transaction error: %s", err)
			return
		}
		for _, song := range songs {
			if err := db.db.Delete(getKey(song.ID)); err != nil {
				l.Error("model.DumpOldSongsToFile delete error: %s", err)
				db.db.Rollback()
				return
			}
		}
		if err := db.db.Commit(); err != nil {
			l.Error("model.DumpOldSongsToFile commit transaction error: %s", err)
			return
		}
	}
	l.Info("model.DumpOldSongsToFile finished")
}

func writeNonEmptyString(f *os.File, prefix, value string) {
	if value != "" {
		f.WriteString(prefix + value + "\n")
	}
}

func DumpToFile(filename string, songs []*Song) (err error) {
	var f *os.File
	if h.FileExists(filename) {
		f, err = os.OpenFile(filename, os.O_RDWR|os.O_APPEND, 0660)
	} else {
		f, err = os.Create(filename)
	}
	if err != nil {
		return
	}
	defer f.Close()
	for _, song := range songs {
		writeNonEmptyString(f, "Date: ", song.Date.String())
		writeNonEmptyString(f, "Track: ", song.Track)
		writeNonEmptyString(f, "Name: ", song.Name)
		writeNonEmptyString(f, "Album: ", song.Album)
		writeNonEmptyString(f, "Artist: ", song.Artist)
		writeNonEmptyString(f, "Title: ", song.Title)
		writeNonEmptyString(f, "File: ", song.File)
		f.WriteString("---------------------\n\n")
	}
	return
}
