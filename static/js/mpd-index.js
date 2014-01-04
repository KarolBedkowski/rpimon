
var MPD = MPD || {
	changingPos: false,
	changingVol: false,
};

function ts2str(ts) {
	ts = parseFloat(ts);
	if (ts > 0) {
		var date = new Date(ts * 1000.0);
		var hours = date.getHours() - 1;
		var minutes = date.getMinutes();
		var seconds = date.getSeconds();
		return [(hours < 10) ? ("0" + hours) : hours,
			(minutes < 10) ? ("0" + minutes) : minutes,
			(seconds < 10) ? ("0" + seconds) : seconds].join(":");
		}
	return "";
}


MPD.onError = function onErrorF(errormsg) {
	$("#buttons-sect").hide();
	$("#currsong-sect").hide();
	$("#status-sect").hide();
	$("#actions-sect").hide();
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
			$("#buttons-sect").show();
			$("#currsong-sect").show();
			$("#status-sect").show();
			$("#actions-sect").show();
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
			$("#st-state").text(status["state"]);
			$("#st-time").text(ts2str(status["elapsed"]));
			$("#st-audio").text(status["audio"]);
			$("#st-bitrate").text(status["bitrate"]);
			$("#st-random").text(status["random"] == "1" ? "YES" : "NO");
			$("#st-repeat").text(status["repeat"] == "1" ? "YES" : "NO");
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
			setTimeout(MPD.refresh, 1000);
		}
	}).fail(function(jqXHR, textStatus) {
		MPD.onError(msg["Error"]);
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
	$("#buttons-sect").hide();
	$("#currsong-sect").hide();
	$("#status-sect").hide();
	$("#actions-sect").hide();
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
