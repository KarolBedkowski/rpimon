package mpd

import (
	//"code.google.com/p/gompd/mpd"
	"errors"
	"github.com/fhs/gompd/mpd"
	l "k.prv/rpimon/logging"
	"k.prv/rpimon/model"
	n "k.prv/rpimon/modules/notepad"
	"sort"
	"strconv"
	"strings"
	"time"
)

type mpdStatus struct {
	Status  map[string]string
	Current map[string]string
	Error   string
}

const poolSize = 3

var (
	host           string
	logSongToNotes string
	lastSong       string
	watcher        = Watcher{
		end: make(chan bool),
	}
	connPool = ConnectionPool{}
)

// Init MPD configuration
func initConnector(conf map[string]string) {
	host = conf["host"]
	logSongToNotes = conf["log to notes"]
	if !watcher.active {
		watcher.Connect()
	}
	connPool.InitPool(poolSize)
}

// Close MPD connection
func closeConnector() {
	watcher.Close()
	connPool.Close()
}

// Watcher client
type Watcher struct {
	watcher *mpd.Watcher
	end     chan bool
	active  bool
}

func (m *Watcher) watch() (err error) {
	l.Info("MPD: Starting mpd watcher...")
	m.watcher, err = mpd.NewWatcher("tcp", host, "")
	if err != nil {
		l.Error("MPD: %s", err.Error())
		return
	}
	logSong()
	for {
		select {
		case subsystem := <-m.watcher.Event:
			l.Debug("MPD: changed subsystem:", subsystem)
			l.Debug("MPD: changed subsystem:", watcher)
			switch subsystem {
			case "player":
				logSong()
			}
		case err := <-m.watcher.Error:
			l.Info("MPD watcher error: %s", err.Error())
			m.watcher.Close()
			m.watcher = nil
			return err
		case _ = <-m.end:
			l.Info("MPD watcher stopping")
			m.watcher.Close()
			m.watcher = nil
			return
		}
	}
	return nil
}

// Connect to mpd daemon
func (m *Watcher) Connect() (err error) {

	go func() {
		defer func() {
			l.Info("mpd.watch: closing")
			if m.watcher != nil {
				m.watcher.Close()
				m.watcher = nil
			}
		}()

		m.active = true

		for m.active {
			if err = m.watch(); err != nil {
				l.Info("mpd.watch: start watch error: %s", err.Error())
			}
			time.Sleep(5 * time.Second)
		}
	}()
	return
}

// Close MPD client
func (m *Watcher) Close() {
	l.Info("mpd.Close")
	if m.active {
		m.active = false
		if m.watcher != nil {
			m.end <- true
		}
	}
}

type mpdConnection struct {
	conn *mpd.Client
}

func (m *mpdConnection) release() {
	connPool.PutBack(m)
}

type ConnectionPool struct {
	conn        chan *mpdConnection
	initialized bool
}

func (p *ConnectionPool) InitPool(size int) {
	if !p.initialized {
		p.conn = make(chan *mpdConnection, poolSize)
		for x := 0; x < size; x++ {
			p.conn <- &mpdConnection{}
		}
		p.initialized = true
	}
}

func (p *ConnectionPool) Close() {
	p.initialized = false
	close(p.conn)
}

func (p *ConnectionPool) get() (c *mpdConnection, err error) {
	if !p.initialized {
		return nil, errors.New("Closed")
	}
	c = <-p.conn
	if c.conn != nil {
		if err = c.conn.Ping(); err != nil {
			c.conn.Close()
			c.conn = nil
		}
	}
	if c.conn == nil {
		c.conn, err = mpd.Dial("tcp", host)
	}
	return c, err
}

func (p *ConnectionPool) PutBack(c *mpdConnection) {
	if p.initialized {
		p.conn <- c
	}
}

func getStatus() (status *mpdStatus) {
	status = new(mpdStatus)
	c, err := connPool.get()
	defer c.release()
	if err != nil {
		status.Error = err.Error()
		return
	}
	if stat, err := c.conn.Status(); err != nil {
		status.Error = err.Error()
	} else {
		status.Status = stat
	}
	if song, err := c.conn.CurrentSong(); err != nil {
		status.Error = err.Error()
	} else {
		status.Current = song
	}
	return
}

func mpdAction(action string) (err error) {
	c, err := connPool.get()
	defer c.release()
	if err != nil {
		return err
	}

	var stat mpd.Attrs
	switch action {
	case "play":
		err = c.conn.Play(-1)
	case "stop":
		err = c.conn.Stop()
	case "pause":
		if stat, err = c.conn.Status(); err == nil {
			err = c.conn.Pause(stat["state"] != "pause")
		}
	case "next":
		err = c.conn.Next()
	case "prev":
		err = c.conn.Previous()
	case "toggle_random":
		if stat, err = c.conn.Status(); err == nil {
			err = c.conn.Random(stat["random"] == "0")
		}
	case "toggle_repeat":
		if stat, err = c.conn.Status(); err == nil {
			err = c.conn.Repeat(stat["repeat"] == "0")
		}
	case "update":
		_, err = c.conn.Update("")
	default:
		l.Warn("page.mpd mpdAction: wrong action ", action)
	}
	return nil
}

func mpdActionUpdate(uri string) (err error) {
	c, err := connPool.get()
	defer c.release()
	if err == nil {
		var jobid int
		jobid, err = c.conn.Update(uri)
		l.Debug("mpdActionUpdate jobid: %d, err: %v", jobid, err)
	}
	return err
}

// current playlist & status
func mpdPlaylistInfo() (playlist []mpd.Attrs, status mpd.Attrs, err error) {
	c, err := connPool.get()
	defer c.release()
	if err != nil {
		return
	}
	playlist, err = c.conn.PlaylistInfo(-1, -1)
	if err != nil {
		return
	}
	status, err = c.conn.Status()
	return
}

func mpdSongAction(songID int, action string) (err error) {
	c, err := connPool.get()
	defer c.release()
	if err == nil {
		switch action {
		case "play":
			err = c.conn.PlayID(songID)
		case "remove":
			err = c.conn.DeleteID(songID)
		default:
			l.Warn("page.mpd mpdAction: wrong action ", action)
		}
	}
	return err
}

func mpdGetPlaylists() (playlists []mpd.Attrs, err error) {
	c, err := connPool.get()
	defer c.release()
	if err == nil {
		playlists, err = c.conn.ListPlaylists()
	}
	return
}

func mpdGetPlaylistContent(playlist string) (files []mpd.Attrs, err error) {
	c, err := connPool.get()
	defer c.release()
	if err == nil {
		files, err = c.conn.PlaylistContents(playlist)
	}
	return
}

func mpdPlaylistsAction(playlist, action string) (result string, err error) {
	c, err := connPool.get()
	defer c.release()
	if err == nil {
		return "", err
	}

	switch action {
	case "play":
		c.conn.Clear()
		c.conn.PlaylistLoad(playlist, -1, -1)
		c.conn.Play(-1)
		return "Plylist loaded", nil
	case "add":
		c.conn.PlaylistLoad(playlist, -1, -1)
		c.conn.Play(-1)
		return "Plylist added", nil
	case "remove":
		err := c.conn.PlaylistRemove(playlist)
		if err == nil {
			return "Plylist removed", nil
		}
		return "", err
	default:
		l.Warn("page.mpd mpdAction: wrong action ", action)
	}
	return "", errors.New("invalid action")
}

func setVolume(volume int) (err error) {
	c, err := connPool.get()
	defer c.release()
	if err == nil {
		err = c.conn.SetVolume(volume)
	}
	return err
}

func seekPos(pos, time int) (err error) {
	c, err := connPool.get()
	defer c.release()
	if err != nil {
		return err
	}
	var song mpd.Attrs
	if song, err = c.conn.CurrentSong(); err == nil {
		var sid int
		if sid, err = strconv.Atoi(song["Id"]); err == nil {
			return c.conn.SeekID(sid, time)
		}
	}
	return
}

// LibraryDir keep subdirectories and files for one folder in library
type LibraryDir struct {
	Folders []string
	Files   []string
}

func getFiles(path string) (folders []string, files []string, err error) {
	c, err := connPool.get()
	defer c.release()
	if err != nil {
		return
	}
	var mpdFiles []mpd.Attrs
	if mpdFiles, err = c.conn.ListInfo(path); err != nil {
		return
	}
	prefixLen := len(path)
	if prefixLen > 1 {
		prefixLen += 1
	}
	for _, attr := range mpdFiles {
		if dir, ok := attr["directory"]; ok {
			folders = append(folders, dir[prefixLen:])
		} else if file, ok := attr["file"]; ok {
			files = append(files, file[prefixLen:])
		}
		// ignore others (i.e. playlists)
	}
	return
}

func addFileToPlaylist(uri string, clearPlaylist bool) (err error) {
	c, err := connPool.get()
	defer c.release()
	if err == nil {
		if clearPlaylist {
			c.conn.Clear()
		}
		if err = c.conn.Add(uri); err == nil {
			c.conn.Play(-1)
		}
	}
	return err
}

func playlistAction(action string) (err error) {
	c, err := connPool.get()
	defer c.release()
	if err == nil {
		switch action {
		case "clear":
			return c.conn.Clear()
		}
	}
	return err
}

func playlistSave(name string) (err error) {
	c, err := connPool.get()
	defer c.release()
	if err == nil {
		return c.conn.PlaylistSave(name)
	}
	return err
}

func addToPlaylist(uri string) (err error) {
	c, err := connPool.get()
	defer c.release()
	if err == nil {
		return c.conn.Add(uri)
	}
	return err
}

// GetShortStatus return cached MPD status
func GetShortStatus() (status map[string]string, err error) {
	if !Module.Enabled() {
		return
	}
	c, err := connPool.get()
	defer c.release()
	if err != nil {
		return
	}
	status, err = c.conn.Status()
	return status, err
}

func getSongInfo(uri string) (info []mpd.Attrs, err error) {
	c, err := connPool.get()
	defer c.release()
	if err == nil {
		return c.conn.ListAllInfo(uri)
	}
	return nil, err
}

func find(query string) (data []mpd.Attrs, err error) {
	c, err := connPool.get()
	defer c.release()
	if err == nil {
		return c.conn.Find(query)
	}
	return nil, err
}

func mpdFileAction(uri, action string) (err error) {
	c, err := connPool.get()
	defer c.release()
	if err != nil {
		return err
	}
	switch action {
	case "add":
		return c.conn.Add(uri)
	default:
		return errors.New("invalid action")
	}
	return
}

func logSong() {
	if logSongToNotes != "yes" {
		return
	}
	c, err := connPool.get()
	defer c.release()
	if err != nil {
		return
	}

	song, err := c.conn.CurrentSong()
	if err != nil {
		return
	}
	var data []string
	for key, val := range song {
		data = append(data, key+": "+val+"\n")
	}
	sort.Strings(data)
	strData := strings.Join(data, "")
	if lastSong == strData {
		return
	}
	lastSong = strData
	strData = time.Now().Format("2006-01-02 15:04:05") + "\n" + strData + "\n"
	n.AppendToNote("mpd_log.txt", strData)

	sl := &model.Song{
		Date:   time.Now(),
		Track:  getVal(song, "Track"),
		Name:   getVal(song, "Name"),
		Album:  getVal(song, "Album"),
		Artist: getVal(song, "Artist"),
		Title:  getVal(song, "Title"),
		File:   getVal(song, "file"),
	}
	sl.Save()
}

func getVal(dict map[string]string, key string) (val string) {
	val, _ = dict[key]
	return val
}
