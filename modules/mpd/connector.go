package mpd

import (
	//"code.google.com/p/gompd/mpd"
	"errors"
	"github.com/turbowookie/gompd/mpd"
	h "k.prv/rpimon/helpers"
	l "k.prv/rpimon/helpers/logging"
	n "k.prv/rpimon/modules/notepad"
	"strconv"
	"strings"
	"time"
)

type mpdStatus struct {
	Status  map[string]string
	Current map[string]string
	Error   string
}

var (
	host              string
	logSongToNotes	 string
	watcher           *mpd.Watcher
	playlistCache     = h.NewSimpleCache(600)
	mpdListFilesCache = h.NewSimpleCache(600)
	mpdLibraryCache   = h.NewKeyCache(600)
	connection        *mpd.Client
	lastSong          string
)

// Init MPD configuration
func initConnector(conf map[string]string) {
	host = conf["host"]
	logSongToNotes = conf["log to notes"]
	connectWatcher()
}

// Close MPD connection
func closeConnector() {
	playlistCache.Clear()
	mpdLibraryCache.Clear()
	mpdListFilesCache.Clear()
	if watcher != nil {
		watcher.Close()
		watcher = nil
	}
	if connection != nil {
		connection.Close()
		connection = nil
	}
}

func connectWatcher() {
	if watcher != nil {
		return
	}
	l.Info("MPD: Starting mpd watcher... %#v", watcher)
	var err error
	watcher, err = mpd.NewWatcher("tcp", host, "")
	if err != nil {
		l.Error("MPD: %s", err.Error())
		return
	}
	go func() {
		for {
			select {
			case subsystem := <-watcher.Event:
				l.Debug("MPD: changed subsystem:", subsystem)
				l.Debug("MPD: changed subsystem:", watcher)
				switch subsystem {
				case "player":
					logSong()
				case "playlist":
					playlistCache.Clear()
				case "database":
					mpdListFilesCache.Clear()
					mpdLibraryCache.Clear()
				}
			case err := <-watcher.Error:
				l.Info("MPD; watcher error: %s", err.Error())
				watcher.Close()
				watcher = nil
				return
			}
		}
	}()
	logSong()
}

func connect() (client *mpd.Client, err error) {
	if connection != nil {
		if err = connection.Ping(); err != nil {
			connection.Close()
			connection = nil
		}
	}
	if connection == nil {
		connection, err = mpd.Dial("tcp", host)
		if err != nil {
			l.Error("MPD: connect error: %s", err.Error())
			closeConnector()
			return nil, err
		}
		connectWatcher()
	}
	return connection, nil
}

func getStatus() (status *mpdStatus) {
	status = new(mpdStatus)
	if _, err := connect(); err != nil {
		status.Error = err.Error()
		return
	}
	if stat, err := connection.Status(); err != nil {
		status.Error = err.Error()
	} else {
		status.Status = stat
	}
	if song, err := connection.CurrentSong(); err != nil {
		status.Error = err.Error()
	} else {
		status.Current = song
	}
	return
}

func mpdAction(action string) (err error) {
	if _, err := connect(); err != nil {
		return err
	}

	var stat mpd.Attrs
	switch action {
	case "play":
		err = connection.Play(-1)
	case "stop":
		err = connection.Stop()
	case "pause":
		if stat, err = connection.Status(); err == nil {
			err = connection.Pause(stat["state"] != "pause")
		}
	case "next":
		err = connection.Next()
	case "prev":
		err = connection.Previous()
	case "toggle_random":
		if stat, err = connection.Status(); err == nil {
			err = connection.Random(stat["random"] == "0")
		}
	case "toggle_repeat":
		if stat, err = connection.Status(); err == nil {
			err = connection.Repeat(stat["repeat"] == "0")
		}
	case "update":
		_, err = connection.Update("")
	default:
		l.Warn("page.mpd mpdAction: wrong action ", action)
	}
	return nil
}

func mpdActionUpdate(uri string) (err error) {
	if _, err = connect(); err == nil {
		var jobid int
		jobid, err = connection.Update(uri)
		l.Debug("mpdActionUpdate jobid: %d, err: %v", jobid, err)
	}
	return err
}

// current playlist & status
func mpdPlaylistInfo() (playlist []mpd.Attrs, err error, status mpd.Attrs) {
	cachedPlaylist := playlistCache.Get(func() h.Value {
		if _, err := connect(); err == nil {
			plist, err := connection.PlaylistInfo(-1, -1)
			if err != nil {
				l.Error(err.Error())
			}
			return plist
		}
		return nil
	})
	if _, err := connect(); err == nil {
		playlist = cachedPlaylist.([]mpd.Attrs)
		status, err = connection.Status()
		if err != nil {
			l.Error(err.Error())
		}
	}
	return
}

func mpdSongAction(songID int, action string) (err error) {
	if _, err := connect(); err == nil {
		switch action {
		case "play":
			err = connection.PlayId(songID)
		case "remove":
			err = connection.DeleteId(songID)
		default:
			l.Warn("page.mpd mpdAction: wrong action ", action)
		}
	}
	return
}

func mpdGetPlaylists() (playlists []mpd.Attrs, err error) {
	if _, err = connect(); err == nil {
		playlists, err = connection.ListPlaylists()
	}
	return
}

func mpdPlaylistsAction(playlist, action string) (result string, err error) {
	if _, err := connect(); err != nil {
		return "", err
	}

	switch action {
	case "play":
		connection.Clear()
		connection.PlaylistLoad(playlist, -1, -1)
		connection.Play(-1)
		return "Plylist loaded", nil
	case "add":
		connection.PlaylistLoad(playlist, -1, -1)
		connection.Play(-1)
		return "Plylist added", nil
	case "remove":
		err := connection.PlaylistRemove(playlist)
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
	if _, err = connect(); err == nil {
		err = connection.SetVolume(volume)
	}
	return
}

func seekPos(pos, time int) (err error) {
	if _, err = connect(); err == nil {
		var song mpd.Attrs
		if song, err = connection.CurrentSong(); err == nil {
			var sid int
			if sid, err = strconv.Atoi(song["Id"]); err == nil {
				err = connection.SeekId(sid, time)
			}
		}
	}
	return err
}

// LibraryDir keep subdirectories and files for one folder in library
type LibraryDir struct {
	Folders []string
	Files   []string
}

func getFiles(path string) (folders []string, files []string, err error) {
	if cached, ok := mpdLibraryCache.GetValue(path); ok {
		cachedLD := cached.(LibraryDir)
		return cachedLD.Folders, cachedLD.Files, nil
	}

	var mpdFiles []string
	if filesC, ok := mpdListFilesCache.GetValue(); ok {
		mpdFiles = filesC.([]string)
	} else {
		if _, err = connect(); err != nil {
			return nil, nil, err
		}
		// FIXME: mpd - zmianic na ls
		if mpdFiles, err = connection.GetFiles(); err != nil {
			return nil, nil, err
		}
		mpdListFilesCache.SetValue(mpdFiles)
	}

	var prefixLen = 0
	if path != "" {
		prefixLen = len(path) + 1
	}

	loadedFolders := make(map[string]bool)

	for _, fname := range mpdFiles {
		if !strings.HasPrefix(fname, path) {
			continue
		}
		fname = fname[prefixLen:]
		slashIndex := strings.Index(fname, "/")
		if slashIndex > 0 {
			fname = fname[:slashIndex]
			if _, added := loadedFolders[fname]; !added {
				loadedFolders[fname] = true
				folders = append(folders, fname)
			}
		} else {
			files = append(files, fname)
		}
	}

	mpdLibraryCache.SetValue(path, LibraryDir{folders, files})
	return
}

func addFileToPlaylist(uri string, clearPlaylist bool) (err error) {
	if _, err = connect(); err == nil {
		if clearPlaylist {
			connection.Clear()
		}
		if err = connection.Add(uri); err == nil {
			connection.Play(-1)
		}
	}
	return err
}

func playlistAction(action string) (err error) {
	if _, err = connect(); err == nil {
		switch action {
		case "clear":
			return connection.Clear()
		}
	}
	return
}

func playlistSave(name string) (err error) {
	if _, err = connect(); err == nil {
		return connection.PlaylistSave(name)
	}
	return err
}

func addToPlaylist(uri string) (err error) {
	if _, err = connect(); err == nil {
		return connection.Add(uri)
	}
	return
}

var mpdShortStatusCache = h.NewSimpleCache(5)

// GetShortStatus return cached MPD status
func GetShortStatus() (status map[string]string, err error) {
	if !Module.Enabled() {
		return nil, nil
	}
	if result, ok := mpdShortStatusCache.GetValue(); ok && result != nil {
		if cachedValue, ok := result.(mpd.Attrs); ok {
			return map[string]string(cachedValue), nil
		}
	}
	if _, err = connect(); err == nil {
		if status, err = connection.Status(); err == nil {
			mpdShortStatusCache.SetValue(status)
		}
	}
	return
}

func getSongInfo(uri string) (info []mpd.Attrs, err error) {
	if _, err = connect(); err == nil {
		return connection.ListAllInfo(uri)
	}
	return nil, err
}

func find(query string) (data []mpd.Attrs, err error) {
	if _, err = connect(); err == nil {
		return connection.Find(query)
	}
	return nil, err
}

func mpdFileAction(uri, action string) (err error) {
	if _, err = connect(); err == nil {
		switch action {
		case "add":
			return connection.Add(uri)
		default:
			err = errors.New("nnvalid action")
		}
	}
	return
}

func logSong() {
	if logSongToNotes != "yes" {
		return
	}
	if connection == nil {
		if _, err := connect(); err != nil {
			return
		}
	}
	song, err := connection.CurrentSong()
	if err != nil {
		return
	}
	var data []string
	for key, val := range song {
		data = append(data, key + ": " + val + "\n")
	}
	strData := strings.Join(data, "")
	if lastSong == strData {
		return
	}
	lastSong = strData
	strData = time.Now().Format("2006-01-02 15:04:05") + "\n" + strData + "\n"
	n.AppendToNote("mpd_log.txt", strData)
}
