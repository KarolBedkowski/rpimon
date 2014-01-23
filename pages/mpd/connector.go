package mpd

import (
	"code.google.com/p/gompd/mpd"
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

var host string

// Init MPD configuration
func Init(mpdHost string) {
	host = mpdHost
}

func getStatus() (status *mpdStatus) {
	status = new(mpdStatus)
	conn, err := mpd.Dial("tcp", host)
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
	conn, err := mpd.Dial("tcp", host)
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
		mpdListFilesCache.Clear()
		mpdLibraryCache.Clear()
	default:
		l.Warn("page.mpd mpdAction: wrong action ", action)
	}
	return nil
}

func mpdPlaylistInfo(start, end int) (playlist []mpd.Attrs, err error, stat mpd.Attrs) {
	conn, err := mpd.Dial("tcp", host)
	if err != nil {
		return
	}
	defer conn.Close()
	playlist, err = conn.PlaylistInfo(start, end)
	if err != nil {
		l.Error(err.Error())
	}
	stat, err = conn.Status()
	if err != nil {
		l.Error(err.Error())
	}
	return
}

func mpdSongAction(songID int, action string) error {
	conn, err := mpd.Dial("tcp", host)
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
	var conn *mpd.Client
	conn, err = mpd.Dial("tcp", host)
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

func mpdPlaylistsAction(playlist, action string) error {
	conn, err := mpd.Dial("tcp", host)
	if err != nil {
		return err
	}
	defer conn.Close()

	switch action {
	case "play":
		conn.Clear()
		conn.PlaylistLoad(playlist, -1, -1)
		conn.Play(-1)
	case "add":
		conn.PlaylistLoad(playlist, -1, -1)
		conn.Play(-1)
	case "remove":
		return conn.PlaylistRemove(playlist)
	default:
		l.Warn("page.mpd mpdAction: wrong action ", action)
	}
	return nil
}

func setVolume(volume int) error {
	conn, err := mpd.Dial("tcp", host)
	if err != nil {
		return err
	}
	defer conn.Close()
	return conn.SetVolume(volume)
}

func seekPos(pos, time int) error {
	conn, err := mpd.Dial("tcp", host)
	if err != nil {
		return err
	}
	defer conn.Close()
	song, err := conn.CurrentSong()
	sid, err := strconv.Atoi(song["Id"])
	return conn.SeekId(sid, time)
}

var mpdLibraryCache = h.NewKeyCache(300)

// LibraryDir keep subdirectories and files for one folder in library
type LibraryDir struct {
	Folders []string
	Files   []string
}

var mpdListFilesCache = h.NewSimpleCache(300)

func getFiles(path string) (folders []string, files []string, err error) {
	if cached, ok := mpdLibraryCache.GetValue(path); ok {
		cachedLD := cached.(LibraryDir)
		return cachedLD.Folders, cachedLD.Files, nil
	}

	var mpdFiles []string
	if filesC, ok := mpdListFilesCache.GetValue(); ok {
		mpdFiles = filesC.([]string)
	} else {
		conn, err := mpd.Dial("tcp", host)
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
	conn, err := mpd.Dial("tcp", host)
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
	conn, err := mpd.Dial("tcp", host)
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
	conn, err := mpd.Dial("tcp", host)
	if err != nil {
		return err
	}
	defer conn.Close()
	return conn.PlaylistSave(name)
}

func addToPlaylist(uri string) (err error) {
	conn, err := mpd.Dial("tcp", host)
	if err != nil {
		return err
	}
	defer conn.Close()
	return conn.Add(uri)
}

var mpdShortStatusCache = h.NewSimpleCache(5)

func GetShortStatus() (map[string]string, error) {
	if result, ok := mpdShortStatusCache.GetValue(); ok {
		cachedValue := result.(mpd.Attrs)
		return map[string]string(cachedValue), nil
	}
	conn, err := mpd.Dial("tcp", host)
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
	conn, err := mpd.Dial("tcp", host)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	return conn.ListAllInfo(uri)
}
