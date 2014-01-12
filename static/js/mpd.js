
var MPD = MPD || {};

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

MPD.status = (function(self) {
	var changingPos = false,
		changingVol = false,
		lastState = {
			Status: {},
			Current: {},
		},
		mpdControlUrl = "/mpd/control",
		mpdServiceInfoUrl = "/mpd/service/info";
	var connectingMessage = null;

	function onError(errormsg) {
		$("div.mpd-buttons-sect").hide();
		$("div.mpd-info-section").hide();
		$("#error-msg").text(errormsg);
		$("#error-msg-box").show();
		setTimeout(refresh, 5000);
	};

	function refresh() {
		$.ajax({
			url: mpdServiceInfoUrl,
			cache: false,
			dataType: 'json'
		}).done(function(msg) {
			if (msg["Error"]) {
				onError(msg["Error"]);
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
				if (status["random"] != lastState["Status"]["random"]) {
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
				if (status["repeat"] != lastState["Status"]["repeat"]) {
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
				$("#st-error").text(status["error"]);
				var volume = status["volume"];
				$("#st-volume").text(volume);
				if (!changingVol) {
					var currVol = $("#slider-volume").slider("value");
					if (currVol != volume) {
						$("#slider-volume").slider("value", volume);
					}
				}
				var songTime = current["Time"];
				$("#curr-time").text(ts2str(songTime));
				if (!changingPos) {
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
				if (status["state"] != lastState["Status"]["state"]) {
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
				lastState = msg;
				connectingMessage.hide()
				setTimeout(refresh, 1000);
			}
		}).fail(function(jqXHR, textStatus) {
			onError(textStatus);
		});
	};

	self.doAction = function doActionF(event) {
		event.preventDefault();
		var btn = $(this);
		var act = btn.data("action");
		$.get(mpdControlUrl + "/" + act)
	};


	self.setVolume = function setVolumeF(value) {
		$.get(mpdControlUrl + "/volume", {vol: value});
	};


	self.seek = function seekF(value) {
		$.get(mpdControlUrl + "/seek", {time: value});
	};


	self.init = function initF(mpdControlUrl_, mpdServiceInfoUrl_) {
		mpdControlUrl = mpdControlUrl_
		mpdServiceInfoUrl = mpdServiceInfoUrl_
		$("#error-msg-box").hide();
		connectingMessage = new Messi('Connecting...', {
			closeButton: false,
			width: 'auto',
		});
		$("div.mpd-buttons-sect").hide();
		$("div.mpd-info-section").hide();
		$("a.btn").on("click", self.doAction);
		$("a.ajax-action").on("click", self.doAction);
		$("#slider-volume").slider({
			min: 0,
			max: 100,
			// slide
			start: function(event, ui) {
				changingVol = true;
			},
			stop: function(event, ui) {
				changingVol = false;
				self.setVolume(ui.value);
			}
		});
		$("#slider-song-pos").slider({
			disabled: true,
			min: 0,
			start: function(event, ui) {
				changingPos = true;
			},
			stop: function(event, ui) {
				self.changingPos = false;
				self.seek(ui.value);
			}
		});
		setTimeout(refresh, 50);
	};

	return self;
})(MPD.status || {});
