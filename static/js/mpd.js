
var MPD = MPD || {
	changingPos: false,
	changingVol: false,
	lastState: {
		Status: {},
		Current: {},
	},
	mpdControlUrl: "/mpd/control",
	mpdServiceInfoUrl: "/mpd/service/info",
	plist: {},
};

function ts2str(ts) {
	ts = Math.floor(parseFloat(ts));
	if (ts > 0) {
		var seconds = Math.floor(ts % 60);
		ts = Math.floor(ts / 60);
		var minutes = ts % 60;
		var hours = Math.floor(ts / 60);
		return [hours,
			(minutes < 10) ? ("0" + minutes) : minutes,
			(seconds < 10) ? ("0" + seconds) : seconds].join(":");
		}
	return "";
};


MPD.onError = function onErrorF(errormsg) {
	$("div.mpd-buttons-sect").hide();
	$("div.mpd-info-section").hide();
	$("#error-msg").text(errormsg);
	$("#error-msg-box").show();
	setTimeout(MPD.refresh, 5000);
};

MPD.refresh = function refreshF() {
	$.ajax({
		url: MPD.mpdServiceInfoUrl,
		cache: false,
		dataType: 'json'
	}).done(function(msg) {
		if (msg["Error"]) {
			MPD.onError(msg["Error"]);
		} else {
			$("#error-msg-box").hide();
			$("div.mpd-buttons-sect").show();
			$("div.mpd-info-section").show();
			var current = msg["Current"];
			$('#curr-name').text(current["Name"]);
			$('#curr-artist').text(current["Artist"]);
			$('#curr-title').text(current["Title"]);
			$('#curr-album').text(current["Album"]);
			$('#curr-track').text(current["Track"]);
			$('#curr-date').text(current["Date"]);
			$('#curr-genre').text(current["Genre"]);
			$('#curr-file').text(current["file"]);
			var status = msg["Status"];
			$("#st-time").text(ts2str(status["elapsed"]));
			$("#st-audio").text(status["audio"]);
			$("#st-bitrate").text(status["bitrate"]);
			if (status["random"] != MPD.lastState["Status"]["random"]) {
				if (status["random"] == "1") {
					$('a[data-action="toggle_random"]')
						.addClass("active")
						.attr("title", "Random ON");
					$('a[data-action="toggle_random"] span.button-label')
						.text("ON");
				} else {
					$('a[data-action="toggle_random"]')
						.removeClass("active")
						.attr("title", "Random OFF");
					$('a[data-action="toggle_random"] span.button-label')
						.text("off");
				}
			}
			if (status["repeat"] != MPD.lastState["Status"]["repeat"]) {
				if (status["repeat"] == "1") {
					$('a[data-action="toggle_repeat"]')
						.addClass("active")
						.attr("title", "Repeat ON");
					$('a[data-action="toggle_repeat"] span.button-label')
						.text("ON");
				} else {
					$('a[data-action="toggle_repeat"]')
						.removeClass("active")
						.attr("title", "Repeat OFF");
					$('a[data-action="toggle_repeat"] span.button-label')
						.text("off");
				}
			}
			$("#st-playlistlength").text(status["playlistlength"]);
			$("#st-state").text(status["state"]);
			var volume = status["volume"];
			$("#st-volume").text(volume);
			if (!MPD.changingVol) {
				var currVol = $("#slider-volume").slider("value");
				if (currVol != volume) {
					$("#slider-volume").slider("value", volume);
				}
			}
			var songTime = current["Time"];
			$("#curr-time").text(ts2str(songTime));
			if (!MPD.changingPos) {
				if (songTime) {
					songTime = parseInt(songTime);
					var pos = parseInt(status["elapsed"])
					$("#slider-song-pos").slider("option", "disabled", false);
					$("#slider-song-pos").slider("option", "max", songTime);
					$("#slider-song-pos").slider("value", pos);
				} else {
					$("#slider-song-pos").slider("value", 0);
					$("#slider-song-pos").slider("option", "disabled", true);
				}
			}
			if (status["state"] != MPD.lastState["Status"]["state"]) {
				if (status["state"] == "play") {
					$('a[data-action="play"]').hide();
					$('a[data-action="pause"]').show();
				} else {
					$('a[data-action="play"]').show();
					$('a[data-action="pause"]').hide();
				}
				if (status["state"] == "stop") {
					$('a[data-action="stop"]').addClass("active");
				} else {
					$('a[data-action="stop"]').removeClass("active");
				}
			}
			MPD.lastState = msg;
			MPD.connectingMessage.hide()
			setTimeout(MPD.refresh, 1000);
		}
	}).fail(function(jqXHR, textStatus) {
		MPD.onError(textStatus);
	});
};

MPD.doAction = function doActionF(event) {
	event.preventDefault();
	var btn = $(this);
	var act = btn.data("action");
	$.get(MPD.mpdControlUrl + "/" + act)
};


MPD.setVolume = function setVolumeF(value) {
	$.get(MPD.mpdControlUrl + "/volume", {vol: value});
};


MPD.seek = function seekF(value) {
	$.get(MPD.mpdControlUrl + "/seek", {time: value});
};


MPD.initIndexPage = function initIndexPageF(mpdControlUrl, mpdServiceInfoUrl) {
	MPD.mpdControlUrl = mpdControlUrl
	MPD.mpdServiceInfoUrl = mpdServiceInfoUrl
	$("#error-msg-box").hide();
	MPD.connectingMessage = new Messi('Connecting...', {
		closeButton: false,
		width: 'auto',
	});
	$("div.mpd-buttons-sect").hide();
	$("div.mpd-info-section").hide();
	$("a.btn").on("click", MPD.doAction);
	$("a.ajax-action").on("click", MPD.doAction);
	$("#slider-volume").slider({
		min: 0,
		max: 100,
		// slide
		start: function(event, ui) {
			MPD.changingVol = true;
		},
		stop: function(event, ui) {
			MPD.changingVol = false;
			MPD.setVolume(ui.value);
		}
	});
	$("#slider-song-pos").slider({
		disabled: true,
		min: 0,
		start: function(event, ui) {
			MPD.changingPos = true;
		},
		stop: function(event, ui) {
			MPD.changingPos = false;
			MPD.seek(ui.value);
		}
	});
	setTimeout(MPD.refresh, 50);
};

MPD.initLibraryPage = function initLibraryPageF(mpdControlUrl, mpdServiceInfoUrl) {
	MPD.mpdControlUrl = mpdControlUrl
	MPD.mpdServiceInfoUrl = mpdServiceInfoUrl
	$("a.action").on("click", function(event) {
		event.preventDefault();
		var link = $(this);
		var p = link.data("path");
		var message = new Messi('Adding...', {
			closeButton: false,
			modal: true,
			width: 'auto',
		});
		$.ajax({
			type: "PUT",
			data: {
				a: link.data("action"),
				u: link.data("uri"),
			}
		}).done(function(msg) {
			message.hide()
		}).fail(function(jqXHR, textStatus) {
			console.log(textStatus);
			message.hide()
			new Messi(textStatus, {
				title: 'Error',
				titleClass: 'anim warning',
				buttons: [{
					id: 0, label: 'Close', val: 'X'
				}]
			});
		});
	});
};


MPD.plist = (function() {
	var self = {},
		message = null;

	function processPlaylist(response) {
		var tbody = $("#playlist-tbody");
		var current = response.stat.songid;
		$.each(response["playlist"], function(i, item) {
			var tr = $("<tr>").attr("data-songid", item.Id).append(
				$("<td>").text(i + 1),
				$("<td>").text(item.Album),
				$("<td>").text(item.Artist),
				$("<td>").text(item.Track));
			if ("Title" in item) {
				tr.append($("<td>").text(item.Title));
			} else {
				tr.append($("<td>").text(item.file));
			}
			tr.append(
				$("<td>").html('<a href="#" class="play-song-action"><span class="glyphicon glyphicon-play" title="Play"></span></a>'+
					'&nbsp;<a href="#" class="remove-song-action"><span class="glyphicon glyphicon-remove" title="Remove"></span></a>')
			);
			if (item.Id == current) {
				tr.addClass("playlist-current-song active");
			}
			tbody.append(tr);
		});
		$('table').tablesorter();
		$("a.play-song-action").on("click", self.playSong);
		$("a.remove-song-action").on("click", self.removeSong);
	};

	function showLoadingMessage() {
		message = new Messi('Loading...', {
			closeButton: false,
			modal: true,
			width: 'auto',
		});
	};

	function showError(message) {
		console.log(message);
		message.hide()
		new Messi(message, {
			title: 'Error',
			titleClass: 'anim warning',
			buttons: [{
				id: 0, label: 'Close', val: 'X'
			}]
		});
	};

	self.refresh = function refreshF() {
		$.ajax({
			url: "/mpd/playlist/serv/info",
		}).done(function(response) {
			$("#playlist-tbody").text("");
			if (response.error == null) {
				processPlaylist(response);
				message.hide()
			} else {
				message.hide()
				showError(response.error);
			}
		}).fail(function(jqXHR, message) {
			showError(message);
		});
	};

	self.playSong = function playSongF(event) {
		event.preventDefault()
		var id = $(this).closest('tr').data("songid");
		showLoadingMessage();
		$.ajax({
			url: "/mpd/song/" + id  + "/play",
			method: "PUT"
		}).done(function(result) {
			message.hide()
			if (result.Error == "") {
				$("tr.active").removeClass("active").removeClass("playlist-current-song");
				var newSongId = result.Status.songid;
				$('tr[data-songid='+newSongId+']').addClass("playlist-current-song active");
			} else {
				showError(result.error);
			}
		}).fail(function(jqXHR, message) {
			message.hide()
			showError(message);
		});
	};

	self.removeSong = function removeSongF(event) {
		event.preventDefault()
		if (!RPI.confirm()) {
			return
		}
		var tr = $(this).closest('tr')
		var id = tr.data("songid");
		showLoadingMessage();
		$.ajax({
			url: "/mpd/song/" + id  + "/remove",
			method: "PUT"
		}).done(function(result) {
			message.hide()
			if (result.Error == "") {
				tr.slideUp(100, function() {
					$("tr.active").removeClass("active").removeClass("playlist-current-song");
					var newSongId = result.Status.songid;
					$('tr[data-songid='+newSongId+']').addClass("playlist-current-song active");
					tr.remove();
				});
			} else {
				showError(result.error);
			}
		}).fail(function(jqXHR, message) {
			message.hide()
			showError(message);
		});
	};

	self.init = function initF() {
		showLoadingMessage();
		self.refresh();
	};

	return self;
})();
