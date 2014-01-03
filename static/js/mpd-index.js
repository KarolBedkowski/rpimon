
function ts2str(ts) {
	ts = parseInt(ts);
	var date = new Date(ts * 1000);
	var hours = date.getHours();
	var minutes = date.getMinutes();
	var seconds = date.getSeconds();
	return [(hours < 10) ? ("0" + hours) : hours,
		(minutes < 10) ? ("0" + minutes) : minutes,
		(seconds < 10) ? ("0" + seconds) : seconds].join(":");
}

function refresh() {
	$.ajax({
		url: '/mpd/service/info',
		cache: false,
		dataType: 'json'
	}).done(function(msg) {
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
		$("#st-time").text(ts2str(status["time"]));
		$("#st-audio").text(status["audio"]);
		$("#st-bitrate").text(status["bitrate"]);
		$("#st-random").text(status["random"] == "1" ? "YES" : "NO");
		$("#st-repeat").text(status["repeat"] == "1" ? "YES" : "NO");
		$("#st-volume").text(status["volume"]);
		setTimeout(refresh, 1000);
	}).fail(function(jqXHR, textStatus) {
		setTimeout(refresh, 5000);
	});
 }

