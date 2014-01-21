/* jshint strict: true */
/* jshint undef: true, unused: true */
/* global Messi: false */
/* global jQuery: false */
/* global window: false */


var MPD = MPD || {};
var RPI = RPI || {};

MPD.plist = (function(self, $) {
	"use strict";

	var msg_loading = null,
		table = null,
		currentSong = "",
		currentSongId = -1;

	function processServerData(sSource, aoData, fnCallback) {
		$.ajax({
			url: sSource,
			data: aoData || {},
		}).done(function(response) {
			hideLoadingMessage();
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

	function showLoadingMessage() {
		if (msg_loading) {
			return;
		}
		msg_loading = new Messi('Loading...', {
			closeButton: false,
			modal: true,
			width: 'auto',
		});
	}

	function hideLoadingMessage() {
		if (msg_loading) {
			msg_loading.hide();
			msg_loading = null;
		}
	}

	function showError(errormsg) {
		hideLoadingMessage();
		window.console.log(errormsg);
		$("#main-alert-error").text(errormsg);
		$("#main-alert").show();
		$("div.playlist-data").hide();
	}

	self.refresh = function refreshF() {
		table = $('table').dataTable({
			"bAutoWidth": false,
			"bSort": false,
			"sPaginationType": "bootstrap",		
			"bProcessing": true,
			"bServerSide": true,
			"sAjaxSource": "/mpd/playlist/serv/info",
			"fnServerData": processServerData,
			"aoColumns": [
				{},
				{"mData": null},
			],
			"aoColumnDefs": [
				{
					"aTargets": [0],
					"mData": null,
					"mRender": function(data, type, full) {
						var title = full[3] || full[5];
						return ['<div class="row"><span class="col-title col-sm-6 col-xs-12 col-md-5">', title, '</span>' +
							'<span class="col-artist col-sm-6 col-xs-12 col-md-4">', full[1], '</span>',
							'<span class="col-track col-sm-2 col-xs-3 col-md-1">', full[2], '</span>',
							'<span class="col-album col-sm-10 col-xs-9 col-md-2">', full[0], '</span></div>'].join("");
					},
				},
				{
					"aTargets": [1],
					"mData": null,
					"bSortable": false,
					"mRender": function(data, type, full) {
						return '<a href="#" class="play-song-action"><span class="glyphicon glyphicon-play" title="Play"></span></a> <a href="#" class="remove-song-action"><span class="glyphicon glyphicon-remove" title="Remove"></span></a> ' + 
							'<a href="#" class="action-info" data-uri="' + full[5] + '"><span class="glyphicon glyphicon-info-sign" title="Info"></a>';
					},
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
				$("tr").on("click",  playSong);
			},
			"sDom": "<'row'<'col-xs-12 col-sm-6'l><'col-xs-12 col-sm-6'f>r>t<'row'<'col-xs-12 col-sm-6'i><'col-xs-12 col-sm-6'p>>",
				
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
		showLoadingMessage();
		$.ajax({
			url: "/mpd/song/" + id  + "/play",
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
				hideLoadingMessage();
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
		showLoadingMessage();
		$.ajax({
			url: "/mpd/song/" + id  + "/remove",
			method: "PUT"
		}).done(function(result) {
			hideLoadingMessage();
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
			url: '/mpd/service/song-info',
			type: "GET",
			data: {
				uri: $(this).data("uri"),
			},
		}).done(function(data) {
			RPI.confirmDialog(data, {
				title: "Song info",
				btnSuccess: "none",
			}).open();
		});
	}

	self.init = function initF() {
		showLoadingMessage();
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
