
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


function onError(errormsg) {
	$("#buttons-sect").hide();
	$("#currsong-sect").hide();
	$("#status-sect").hide();
	$("#actions-sect").hide();
	$("#error-msg").text(errormsg);
	$("#error-msg-box").show();
	setTimeout(refresh, 5000);
}

function refresh() {
	$.ajax({
		url: '/mpd/service/info',
		cache: false,
		dataType: 'json'
	}).done(function(msg) {
		if (msg["Error"]) {
			onError(msg["Error"]);
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
			var currVol = $("#slider-volume").slider("value");
			if (currVol != volume) {
				$("#slider-volume").slider("value", volume);
			}
			var songTime = current["Time"];
			$("#curr-time").text(ts2str(songTime));
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
			setTimeout(refresh, 1000);
		}
	}).fail(function(jqXHR, textStatus) {
		onError(msg["Error"]);
	});
 }

function do_action(t) {
	var btn = $(this);
	var act = btn.data("action");
	$.get("/mpd/action/" + act, 
		function(data) {
		});
 }


 function setVolume(value) {
	$.get("/mpd/action/volume", {vol: value
 }).done(function(data) {
	});
 }

 function seek(value) {
	$.get("/mpd/action/seek", {time: value
 }).done(function(data) {
	});
 }
