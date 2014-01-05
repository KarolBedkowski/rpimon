package mpd

import (
	"code.google.com/p/gompd/mpd"
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
	default:
		l.Warn("page.mpd mpdAction: wrong action ", action)
	}
	return nil
}

func mpdPlaylistInfo() (playlist []mpd.Attrs, err error, currentSong string) {
	conn, err := mpd.Dial("tcp", host)
	if err != nil {
		return
	}
	defer conn.Close()
	playlist, err = conn.PlaylistInfo(-1, -1)
	if err != nil {
		l.Error(err.Error())
	}
	stat, err := conn.Status()
	if err != nil {
		l.Error(err.Error())
	} else {
		currentSong = stat["songid"]
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

	default:
		l.Warn("page.mpd mpdAction: wrong action ", action)
	}
	return nil
}

func mpdGetPlaylists() (playlists []mpd.Attrs, err error) {
	conn, err := mpd.Dial("tcp", host)
	if err != nil {
		return
	}
	defer conn.Close()
	playlists, err = conn.ListPlaylists()
	if err != nil {
		l.Error(err.Error())
	}
	return
}

func mpdPlaylistAction(playlist, action string) error {
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

func getFiles(path string) (folders []string, files []string, err error) {
	conn, err := mpd.Dial("tcp", host)
	if err != nil {
		return nil, nil, err
	}
	defer conn.Close()
	mpdFiles, err := conn.GetFiles()
	if err != nil {
		return nil, nil, err
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
			l.Debug(fname)
			files = append(files, fname)
		}
	}
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
	return conn.Add(uri)
}
