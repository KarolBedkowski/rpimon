/* jshint strict: true */
/* jshint undef: true, unused: true */
/* global jQuery: false */
/* global window: false */


var MPD = MPD || {};
var RPI = RPI || {};

MPD.plist = (function(self, $) {
	"use strict";

	var msg_loading = null,
		table = null,
		currentSong = "",
		currentSongId = -1,
		urls = {
			'mpd-pl-serv-info': '',
			'mpd-song-action-play': '',
			'mpd-song-action-remove': '',
			"mpd-service-song-info": ''
		};

	function processServerData(sSource, aoData, fnCallback) {
		$.ajax({
			url: sSource,
			data: aoData || {}
		}).done(function(response) {
			RPI.hideLoadingMsg();
			if (response.error) {
				showError(response.error);
			}
			else {
				currentSong = response.stat.song;
				currentSongId = response.stat.songid;
				fnCallback(response);
			}
		}).fail(function(jqXHR, result) {
			showError(result);
		});
	}

	function showError(errormsg) {
		RPI.hideLoadingMsg();
		if (window.console && window.console.log) { window.console.log(errormsg); }
		$("#main-alert-error").text(errormsg);
		$("#main-alert").show();
		$("div.playlist-data").hide();
	}

	self.refresh = function refreshF() {
		table = $('table').dataTable({
			"bAutoWidth": false,
			"bSort": false,
			//"sPaginationType": "bootstrap",
			"bProcessing": true,
			"bServerSide": true,
			"sAjaxSource": urls['mpd-pl-serv-info'],
			"fnServerData": processServerData,
			"aoColumns": [
				{},
				{"mData": null}
			],
			"aoColumnDefs": [
				{
					"aTargets": [0],
					"mData": null,
					"mRender": function(data, type, full) {
						var title = full[3] || full[5];
						return ['<div class="row"><span class="col-title col-sm-6 col-xs-12 col-md-5"><a href="#" class="play-song-action">', title, '</a></span>' +
							'<span class="col-artist col-sm-6 col-xs-12 col-md-3">', full[1], '</span>',
							'<span class="col-track col-sm-2 col-xs-3 col-md-1">', full[2], '</span>',
							'<span class="col-album col-sm-10 col-xs-9 col-md-3">', full[0], '</span></div>'].join("");
					}
				},
				{
					"aTargets": [1],
					"mData": null,
					"bSortable": false,
					"mRender": function(data, type, full) {
						return '<a href="#" class="play-song-action"><span class="glyphicon glyphicon-play" title="Play"></span></a> <a href="#" class="remove-song-action"><span class="glyphicon glyphicon-remove" title="Remove"></span></a> ' +
							'<a href="#" class="action-info" data-uri="' + full[5] + '"><span class="glyphicon glyphicon-info-sign" title="Info"></a>';
					}
				}
			],
			"fnRowCallback": function(row, aData) { //, iDisplayIndex, iDisplayIndexFull) {
				$(row).data("songid", aData[4]);
				if (aData[4] == currentSongId) {
					// mark current song
					$(row).addClass("playlist-current-song active");
				}
			},
			"fnDrawCallback": function() { //oSettings) {
				$("a.play-song-action").on("click", playSong);
				$("a.remove-song-action").on("click", removeSong);
				$("a.action-info").on("click", songInfo);
			},
			"sDom": "<'row'<'col-xs-12 col-sm-6'l><'col-xs-12 col-sm-6'f>r>t<'row'<'col-xs-12 col-sm-6'i><'col-xs-12 col-sm-6'p>>"
		});
		return;
	};

	function playSong(event) {
		event.preventDefault();
		var tr = $(this),
			id = tr.data("songid");
		if (!id) {
			tr = tr.closest('tr');
			id = tr.data("songid");
		}
		RPI.showLoadingMsg();
		$.ajax({
			url: urls["mpd-song-action-play"].replace("000", id),
			method: "PUT"
		}).done(function(result) {
			if (result.Error === "") {
				$("tr.active").removeClass("active").removeClass("playlist-current-song");
				currentSongId = result.Status.songid;
				if (currentSongId != id) {
					// $('tr[data-songid=... not work on dynamic created data
					tr = $('tr').filter(function() {
						return $(this).data('songid') == currentSongId;
					}).first();
				}
				tr.addClass("playlist-current-song active");
				RPI.hideLoadingMsg();
			} else {
				showError(result.error);
			}
		}).fail(function(jqXHR, result) {
			showError(result);
		});
	}

	function removeSong(event) {
		event.preventDefault();
		if (!RPI.confirm()) {
			return;
		}
		var tr = $(this).closest('tr'),
			id = tr.data("songid");
		RPI.showLoadingMsg();
		$.ajax({
			url: urls["mpd-song-action-remove"].replace("000", id),
			method: "PUT"
		}).done(function(result) {
			RPI.hideLoadingMsg();
			if (result.Error === "") {
				// redraw table on success
				table.fnDraw();
			} else {
				showError(result.error);
			}
		}).fail(function(jqXHR, result) {
			showError(result);
		});
	}

	function songInfo(event) {
		event.preventDefault();
		$.ajax({
			url: urls["mpd-service-song-info"],
			type: "GET",
			data: {
				uri: $(this).data("uri")
			}
		}).done(function(data) {
			RPI.confirmDialog(data, {
				title: "Song info",
				btnSuccess: "none"
			}).open();
		});
	}

	function playlistAjaxAction(dialog, event, refresh) {
		event.preventDefault();
		$('button[type="submit"]', dialog).button('loading');
		var form = $("form", dialog);
		$.ajax({
			method: "POST",
			url: form.attr("action"),
			data: form.serialize()
		}).always(function() {
			dialog.modal("hide");
			$('button[type="submit"]', dialog).button('reset');
		}).done(function(msg) {
			RPI.showFlash("success", msg, 1);
			$('input[type="text"]', dialog).val("");
			if (refresh) {
				table.fnDraw();
			}
		}).fail(function(msg) {
			RPI.alert(msg.responseText).open();
		});
	}

	self.init = function initF(params) {
		RPI.showLoadingMsg();

		urls = $.extend({}, urls, params.urls || {});

		$('div.modal').on('shown.bs.modal', function() {
			var inputs = $('input:first-of-type');
			if (inputs) {
				inputs.focus();
			}
		});

		$("#save-playlist-dlg form").submit(function(event) {
			playlistAjaxAction($("#save-playlist-dlg"), event);
		});
		$("#add-custom-dlg form").submit(function(event) {
			playlistAjaxAction($("#add-custom-dlg"), event, true);
		});

		self.refresh();
	};

	self.gotoCurrentSong = function gotoCurrentSongF() {
		if (currentSong) {
			var page = Math.floor(parseInt(currentSong) / table.fnSettings()._iDisplayLength);
			table.fnPageChange(page);
		}
	};

	return self;
}(MPD.plist || {}, jQuery));
