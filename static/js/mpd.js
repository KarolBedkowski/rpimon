/* jshint strict: true */
/* jshint undef: true, unused: true */
/* global window: false */
/* global jQuery: false */

var MPD = MPD || {};

MPD.status = (function(self, $) {
	"use strict";

	var changingPos = false,
		changingVol = false,
		lastState = {
			Status: {},
			Current: {},
		},
		mpdControlUrl = "/mpd/control",
		mpdServiceInfoUrl = "/mpd/service/info",
		a_toggle_random = $('a[data-action="toggle_random"]'),
		a_toggle_random_label = $('a[data-action="toggle_random"] span.button-label'),
		a_toggle_repeat = $('a[data-action="toggle_repeat"]'),
		a_toggle_repeat_label = $('a[data-action="toggle_repeat"] span.button-label'),
		a_play = $('a[data-action="play"]'),
		a_pause = $('a[data-action="pause"]'),
		a_stop = $('a[data-action="stop"]');

	function onError(errormsg) {
		$("div.mpd-buttons-sect").hide();
		$("div.mpd-info-section").hide();
		$("#main-alert-error").text(errormsg);
		$("#main-alert").show();
		RPI.hideLoadingMsg();
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
					$("#curr-time").text(ts2str(songTime));
					if (songTime) {
						songTime = parseInt(songTime);
						$("#slider-song-pos").slider("enable").slider("option", "max", songTime);
					} else {
						$("#slider-song-pos").slider("disable").slider("value", 0);
					}
				}
				//$("#st-time").text(ts2str(status.elapsed));
				$("#st-audio").text(status.audio);
				$("#st-bitrate").text(status.bitrate);
				if (status.random != lastState.Status.random) {
					if (status.random == "1") {
						a_toggle_random.addClass("active").attr("title", "Random ON");
						a_toggle_random_label.text("ON");
					} else {
						a_toggle_random.removeClass("active").attr("title", "Random OFF");
						a_toggle_random_label.text("off");
					}
				}
				if (status.repeat != lastState.Status.repeat) {
					if (status.repeat == "1") {
						a_toggle_repeat.addClass("active").attr("title", "Repeat ON");
						a_toggle_repeat_label.text("ON");
					} else {
						a_toggle_repeat.removeClass("active").attr("title", "Repeat OFF");
						a_toggle_repeat_label.text("off");
					}
				}
				$("#st-playlistlength").text(status.song + "/" + status.playlistlength);
				$("#st-state").text(status.state);
				$("#st-error").text(status.error);
				//$("#st-volume").text(volume);
				if (!changingVol) {
					if ($("#slider-volume").slider("value") != volume) {
						$("#slider-volume").slider("value", volume);
					}
				}
				if (!changingPos) {
					$("#slider-song-pos").slider("value", parseInt(status.elapsed));
				}
				if (status.state != lastState.Status.state) {
					if (status.state == "play") {
						a_play.hide();
						a_pause.show();
					} else {
						a_play.show();
						a_pause.hide();
					}
					if (status.state == "stop") {
						a_stop.addClass("active");
					} else {
						a_stop.removeClass("active");
					}
				}
				lastState = msg;
				RPI.hideLoadingMsg();
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
		RPI.showLoadingMsg();
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
			},
			slide:  function(event, ui) {
				$("#st-volume").text(ui.value);
			},
			change: function(event, ui) {
				if (!changingVol) {
					$("#st-volume").text(ui.value);
				}
			},
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
			},
			slide:  function(event, ui) {
				$("#st-time").text(ts2str(ui.value));
			},
			change: function(event, ui) {
				if (!changingPos) {
					$("#st-time").text(ts2str(ui.value));
				}
			},
		});

		$("a#action-info").on("click", function(event) {
			event.preventDefault();
			if (lastState.Current && lastState.Current.file) {
				$.ajax({
					url: '/mpd/service/song-info',
					type: "GET",
					data: {
						uri: lastState.Current.file,
					},
				}).done(function(data) {
					RPI.confirmDialog(data, {
						title: "Song info",
						btnSuccess: "none",
					}).open();
				});
			}
		});

		window.setTimeout(refresh, 50);
	};

	return self;
}(MPD.status || {}, jQuery));
