/* jshint strict: true */
/* jshint undef: true, unused: true */
/* global window: false */
/* global Messi: false */
/* global jQuery: false */

"use strict";
var MPD = MPD || {};

MPD.status = (function(self, $) {
	var changingPos = false,
		changingVol = false,
		lastState = {
			Status: {},
			Current: {},
		},
		mpdControlUrl = "/mpd/control",
		mpdServiceInfoUrl = "/mpd/service/info",
		connectingMessage = null;

	function onError(errormsg) {
		$("div.mpd-buttons-sect").hide();
		$("div.mpd-info-section").hide();
		new Messi(errormsg, {
			title: 'Error',
			titleClass: 'anim warning',
			buttons: [{
				"id": 1, 
				"label": "Reconnect", 
				"val": "R", 
				"class": 'btn-success'
			}],
			callback: function(val) {
				if (val == "R") {
					refresh();
				}
			},
		});
	}

	function ts2str(ts) {
		ts = Math.floor(parseFloat(ts));
		if (ts > 0) {
			var seconds = Math.floor(ts % 60);
			ts = Math.floor(ts / 60);
			var minutes = ts % 60,
				hours = Math.floor(ts / 60);
			return [hours,
				(minutes < 10) ? ("0" + minutes) : minutes,
				(seconds < 10) ? ("0" + seconds) : seconds].join(":");
			}
		return "";
	}

	function refresh() {
		$.ajax({
			url: mpdServiceInfoUrl,
			cache: false,
			dataType: 'json'
		}).done(function(msg) {
			if (msg.Error) {
				onError(msg.Error);
			} else {
				$("div.mpd-buttons-sect").show();
				$("div.mpd-info-section").show();
				var current = msg.Current,
					status = msg.Status,
					volume = status.volume,
					songTime = current.Time;
				if (lastState.Current.Id != current.Id) {
					$('#curr-name').text(current.Name);
					$('#curr-artist').text(current.Artist);
					$('#curr-title').text(current.Title);
					$('#curr-album').text(current.Album);
					$('#curr-track').text(current.Track);
					$('#curr-date').text(current.Date);
					$('#curr-genre').text(current.Genre);
					$('#curr-file').text(current.file);
				}
				$("#st-time").text(ts2str(status.elapsed));
				$("#st-audio").text(status.audio);
				$("#st-bitrate").text(status.bitrate);
				if (status.random != lastState.Status.random) {
					if (status.random == "1") {
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
				if (status.repeat != lastState.Status.repeat) {
					if (status.repeat == "1") {
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
				$("#st-playlistlength").text(status.playlistlength);
				$("#st-state").text(status.state);
				$("#st-error").text(status.error);
				$("#st-volume").text(volume);
				if (!changingVol) {
					if ($("#slider-volume").slider("value") != volume) {
						$("#slider-volume").slider("value", volume);
					}
				}
				$("#curr-time").text(ts2str(songTime));
				if (!changingPos) {
					if (songTime) {
						songTime = parseInt(songTime);
						$("#slider-song-pos").slider("option", "disabled", false);
						$("#slider-song-pos").slider("option", "max", songTime);
						$("#slider-song-pos").slider("value", parseInt(status.elapsed));
					} else {
						$("#slider-song-pos").slider("value", 0);
						$("#slider-song-pos").slider("option", "disabled", true);
					}
				}
				if (status.state != lastState.Status.state) {
					if (status.state == "play") {
						$('a[data-action="play"]').hide();
						$('a[data-action="pause"]').show();
					} else {
						$('a[data-action="play"]').show();
						$('a[data-action="pause"]').hide();
					}
					if (status.state == "stop") {
						$('a[data-action="stop"]').addClass("active");
					} else {
						$('a[data-action="stop"]').removeClass("active");
					}
				}
				lastState = msg;
				if (connectingMessage) {
					connectingMessage.hide();
					connectingMessage = null;
				}
				window.setTimeout(refresh, 1000);
			}
		}).fail(function(jqXHR, textStatus) {
			onError(textStatus);
		});
	}

	function doAction(event) {
		event.preventDefault();
		var act = $(this).data("action");
		$.get(mpdControlUrl + "/" + act);
	}

	function setVolume(value) {
		$.get(mpdControlUrl + "/volume", {vol: value});
	}

	function seek(value) {
		$.get(mpdControlUrl + "/seek", {time: value});
	}

	self.init = function(mpdControlUrl_, mpdServiceInfoUrl_) {
		mpdControlUrl = mpdControlUrl_;
		mpdServiceInfoUrl = mpdServiceInfoUrl_;
		connectingMessage = new Messi('Connecting...', {
			closeButton: false,
			width: 'auto',
		});
		$("div.mpd-buttons-sect").hide();
		$("div.mpd-info-section").hide();
		$("a.ajax-action").on("click", doAction);
		$("#slider-volume").slider({
			min: 0,
			max: 100,
			range: "min",
			// slide
			start: function() {//event, ui) {
				changingVol = true;
			},
			stop: function(event, ui) {
				changingVol = false;
				setVolume(ui.value);
			}
		});
		$("#slider-song-pos").slider({
			disabled: true,
			min: 0,
			range: "min",
			start: function() { // event, ui) {
				changingPos = true;
			},
			stop: function(event, ui) {
				changingPos = false;
				seek(ui.value);
			}
		});

		$("a#action-info").on("click", function(event) {
			event.preventDefault();
			if (lastState.Current && lastState.Current.file) {
				var opt = {params: {
						uri: lastState.Current.file,
					},
				};
				Messi.load('/mpd/service/song-info', opt);
			}
		});

		window.setTimeout(refresh, 50);
	};

	return self;
}(MPD.status || {}, jQuery));
