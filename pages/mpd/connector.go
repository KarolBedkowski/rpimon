package mpd

import (
	"code.google.com/p/gompd/mpd"
	"errors"
	h "k.prv/rpimon/helpers"
	l "k.prv/rpimon/helpers/logging"
	"strconv"
	"strings"
)

type mpdStatus struct {
	Status  map[string]string
	Current map[string]string
	Error   string
}

var (
	host              string
	watcher           *mpd.Watcher
	playlistCache     = h.NewSimpleCache(600)
	mpdListFilesCache = h.NewSimpleCache(600)
	mpdLibraryCache   = h.NewKeyCache(600)
)

// Init MPD configuration
func Init(mpdHost string) {
	host = mpdHost
	connectWatcher()
}

func Close() {
	playlistCache.Clear()
	mpdLibraryCache.Clear()
	mpdListFilesCache.Clear()
	if watcher != nil {
		watcher.Close()
		watcher = nil
	}
}

func connectWatcher() {
	if watcher != nil {
		return
	}
	l.Info("Starting mpd watcher... %#v", watcher)
	var err error
	watcher, err = mpd.NewWatcher("tcp", host, "")
	if err != nil {
		l.Error(err.Error())
		return
	}
	go func() {
		for {
			select {
			case subsystem := <-watcher.Event:
				l.Debug("MPD: changed subsystem:", subsystem)
				switch subsystem {
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
}

func connect() (client *mpd.Client, err error) {
	client, err = mpd.Dial("tcp", host)
	if err != nil {
		l.Error("Mpd connect error: %s", err.Error())
		Close()
		return
	}
	connectWatcher()
	return
}

func getStatus() (status *mpdStatus) {
	status = new(mpdStatus)
	conn, err := connect()
	if err != nil {
		status.Error = err.Error()
		return
	}
	defer conn.Close()

	stat, err := conn.Status()
	if err != nil {
		status.Error = err.Error()
		l.Error(err.Error())
		return
	}
	song, err := conn.CurrentSong()
	if err != nil {
		status.Error = err.Error()
		l.Error(err.Error())
		return
	}
	status.Status = stat
	status.Current = song
	return
}

func mpdAction(action string) error {
	conn, err := connect()
	if err != nil {
		return err
	}
	defer conn.Close()
	stat, err := conn.Status()
	if err != nil {
		l.Error(err.Error())
		return err
	}

	switch action {
	case "play":
		conn.Play(-1)
	case "stop":
		conn.Stop()
	case "pause":
		conn.Pause(stat["state"] != "pause")
	case "next":
		conn.Next()
	case "prev":
		conn.Previous()
	case "toggle_random":
		conn.Random(stat["random"] == "0")
	case "toggle_repeat":
		conn.Repeat(stat["repeat"] == "0")
	case "update":
		conn.Update("")
	default:
		l.Warn("page.mpd mpdAction: wrong action ", action)
	}
	return nil
}

func mpdActionUpdate(uri string) error {
	conn, err := connect()
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = conn.Update(uri)
	return err
}

// current playlist & status
func mpdPlaylistInfo() (playlist []mpd.Attrs, err error, status mpd.Attrs) {
	conn, err := connect()
	if err != nil {
		return
	}
	defer conn.Close()
	cachedPlaylist := playlistCache.Get(func() h.Value {
		plist, err := conn.PlaylistInfo(-1, -1)
		if err != nil {
			l.Error(err.Error())
		}
		return plist
	})
	playlist = cachedPlaylist.([]mpd.Attrs)
	status, err = conn.Status()
	if err != nil {
		l.Error(err.Error())
	}
	return
}

func mpdSongAction(songID int, action string) error {
	conn, err := connect()
	if err != nil {
		return err
	}
	defer conn.Close()

	switch action {
	case "play":
		conn.PlayId(songID)
	case "remove":
		conn.DeleteId(songID)
	default:
		l.Warn("page.mpd mpdAction: wrong action ", action)
	}
	return nil
}

func mpdGetPlaylists() (playlists []mpd.Attrs, err error) {
	conn, err := connect()
	if err != nil {
		l.Warn("mpdGetPlaylists error ", err.Error)
		return
	}
	defer conn.Close()
	playlists, err = conn.ListPlaylists()
	if err != nil {
		l.Error(err.Error())
	}
	return
}

func mpdPlaylistsAction(playlist, action string) (string, error) {
	conn, err := connect()
	if err != nil {
		return "", err
	}
	defer conn.Close()

	switch action {
	case "play":
		conn.Clear()
		conn.PlaylistLoad(playlist, -1, -1)
		conn.Play(-1)
		return "Plylist loaded", nil
	case "add":
		conn.PlaylistLoad(playlist, -1, -1)
		conn.Play(-1)
		return "Plylist added", nil
	case "remove":
		err := conn.PlaylistRemove(playlist)
		if err == nil {
			return "Plylist removed", nil
		}
		return "", err
	default:
		l.Warn("page.mpd mpdAction: wrong action ", action)
	}
	return "", errors.New("invalid action")
}

func setVolume(volume int) error {
	conn, err := connect()
	if err != nil {
		return err
	}
	defer conn.Close()
	return conn.SetVolume(volume)
}

func seekPos(pos, time int) error {
	conn, err := connect()
	if err != nil {
		return err
	}
	defer conn.Close()
	song, err := conn.CurrentSong()
	sid, err := strconv.Atoi(song["Id"])
	return conn.SeekId(sid, time)
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
		conn, err := connect()
		if err != nil {
			return nil, nil, err
		}
		defer conn.Close()
		// FIXME: mpd - zmianic na ls
		mpdFiles, err = conn.GetFiles()
		if err != nil {
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

func addFileToPlaylist(uri string, clearPlaylist bool) error {
	conn, err := connect()
	if err != nil {
		return err
	}
	defer conn.Close()
	if clearPlaylist {
		conn.Clear()
	}
	err = conn.Add(uri)
	if err == nil {
		conn.Play(-1)
	}
	return err
}

func playlistAction(action string) (err error) {
	conn, err := connect()
	if err != nil {
		return err
	}
	defer conn.Close()
	switch action {
	case "clear":
		return conn.Clear()
	}
	return
}

func playlistSave(name string) (err error) {
	conn, err := connect()
	if err != nil {
		return err
	}
	defer conn.Close()
	return conn.PlaylistSave(name)
}

func addToPlaylist(uri string) (err error) {
	conn, err := connect()
	if err != nil {
		return err
	}
	defer conn.Close()
	return conn.Add(uri)
}

var mpdShortStatusCache = h.NewSimpleCache(5)

// GetShortStatus return cached MPD status
func GetShortStatus() (map[string]string, error) {
	if result, ok := mpdShortStatusCache.GetValue(); ok {
		cachedValue := result.(mpd.Attrs)
		return map[string]string(cachedValue), nil
	}
	conn, err := connect()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	status, err := conn.Status()
	if err == nil {
		mpdShortStatusCache.SetValue(status)
	}
	return status, err
}

func getSongInfo(uri string) (info []mpd.Attrs, err error) {
	conn, err := connect()
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return conn.ListAllInfo(uri)
}
