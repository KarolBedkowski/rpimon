package mpd

import (
	"code.google.com/p/gompd/mpd"
	l "k.prv/rpimon/helpers/logging"
)

type mpdStatus struct {
	Status  map[string]string
	Current map[string]string
	Error   string
}

var host string

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

func mpdAction(action string) (status *mpdStatus) {
	status = new(mpdStatus)
	conn, err := mpd.Dial("tcp", host)
	if err != nil {
		status.Error = err.Error()
		return status
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
	case "":
		// no action
		return
	default:
		l.Warn("page.mpd mpdAction: wrong action ", action)
		return
	}
	stat, err = conn.Status()
	if err != nil {
		status.Error = err.Error()
		l.Error(err.Error())
	}
	status.Status = stat
	return
}

func mpdPlaylistInfo(action string) (playlist []mpd.Attrs, err error, currentSong string) {
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
