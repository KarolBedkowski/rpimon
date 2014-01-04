
var MPD = MPD || {
	changingPos: false,
	changingVol: false,
	lastState: {
		Status: {},
		Current: {},
	},
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
}


MPD.onError = function onErrorF(errormsg) {
	$("div.mpd-buttons-sect").hide();
	$("div.mpd-info-section").hide();
	$("#error-msg").text(errormsg);
	$("#error-msg-box").show();
	setTimeout(MPD.refresh, 5000);
}

MPD.refresh = function refreshF() {
	$.ajax({
		url: '/mpd/service/info',
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
						.removeClass("ui-state-default")
						.addClass("ui-state-active")
						.attr("title", "Random OFF");
				} else {
					$('a[data-action="toggle_random"]')
						.removeClass("ui-state-active")
						.addClass("ui-state-default")
						.attr("title", "Random ON");
				}
			}
			if (status["repeat"] != MPD.lastState["Status"]["repeat"]) {
				if (status["repeat"] == "1") {
					$('a[data-action="toggle_repeat"]')
						.removeClass("ui-state-default")
						.addClass("ui-state-active")
						.attr("title", "Repeat OFF");
				} else {
					$('a[data-action="toggle_repeat"]')
						.removeClass("ui-state-active")
						.addClass("ui-state-default")
						.attr("title", "Repeat ON");
				}
			}
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
			}
			MPD.lastState = msg;
			setTimeout(MPD.refresh, 1000);
		}
	}).fail(function(jqXHR, textStatus) {
		MPD.onError(textStatus);
	});
 }

MPD.doAction = function doActionF(t) {
	var btn = $(this);
	var act = btn.data("action");
	$.get("/mpd/action/" + act)
 }


MPD.setVolume = function setVolumeF(value) {
	$.get("/mpd/action/volume", {vol: value});
 }


MPD.seek = function seekF(value) {
	$.get("/mpd/action/seek", {time: value});
 }


 MPD.init = function initF() {
	$("div.mpd-buttons-sect").hide();
	$("div.mpd-info-section").hide();
	$("a.pure-button").on("click", MPD.doAction);
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
 }
