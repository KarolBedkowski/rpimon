
var MPD = MPD || {
	changingPos: false,
	changingVol: false,
	lastState: {
		Status: {},
		Current: {},
	},
	mpdControlUrl: "/mpd/control",
	mpdServiceInfoUrl: "/mpd/service/info",
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
		$("div.message").show();
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
		})
	});
};

