/* jshint strict: true */
/* jshint undef: true, unused: true */
/* global jQuery: false */
/* global window: false */


var MPD = MPD || {};
var RPI = RPI || {};

MPD.plists = (function(self, $) {
	"use strict";

	var table = null,
		urls = {
			"mpd-playlists-serv-list": "/mpd/playlists/serv/list",
			"mpd-playlists-action": "/mpd-playlists/action"
		};

	function refresh() {
		RPI.showLoadingMsg();
		$.ajax(urls["mpd-playlists-serv-list"]
		).always(function() {
			RPI.hideLoadingMsg();
		}).done(function(response) {
			if (response.error || !response.items) {
				showError(response.error || "Data error");
			} else {
				table.fnClearTable();
				table.fnAddData(response.items || []);
			}
		}).fail(function(jqXHR, result) {
			showError(result);
		});
	};

	function action(action, playlist, refreshOnSuccess) {
		RPI.showLoadingMsg();
		$.ajax({
			url: urls["mpd-playlists-action"],
			type: "PUT",
			data: {
				"a": action,
				"p": playlist
			}
		}).done(function(msg) {
			RPI.hideLoadingMsg();
			RPI.showFlash("success", msg, 1);
			if (refreshOnSuccess) {
				refresh();
			}
		}).fail(function(jqXHR, textStatus) {
			RPI.hideLoadingMsg();
			RPI.alert(textStatus, {
				title: "Error"
			}).open();
		});
	}

	self.init = function initF(params) {
		urls = $.extend(urls, params.urls || {});

		table = $('table').dataTable({
			"bAutoWidth": false,
			"bStateSave": true,
			"bProcessing": true,
			"bFilter": false,
			//"sPaginationType": "bootstrap",
			"iDisplayLength": 15,
			"bLengthChange": false,
			"aoColumnDefs": [
				{"aTargets": [0], "mData": "playlist"},
				{
					"aTargets": [1],
					"mData": "Last-Modified",
					"mRender": function(data) {
						return data.replace("T", " ");
					}
				},
				{
					"aTargets": [2],
					"mData": null,
					"bSortable": false,
					"mRender": function() {
						return ('<td><a href="#" title="Play" class="ajax-action-play"><span class="glyphicon glyphicon-play">' +
							'<a href="#" title="Remove" class="ajax-action-remove"><span class="glyphicon glyphicon-remove"></a>');
					}
				}
			],
			"fnRowCallback": function(row, aData) { //, iDisplayIndex, iDisplayIndexFull) {
				$(row).data("name", aData.playlist);
			},
			"fnDrawCallback": function() { //oSettings) {
				$('a.ajax-action-remove').on("click", function(event) {
					event.preventDefault();
					var playlist = $(this).closest('tr').data("name");
					RPI.confirmDialog("Remove playlist " + playlist + "?", {
						title: "Playlists",
						btnSuccess: "Remove",
						onSuccess: function() {
							action("remove", playlist, true);
						}
					}).open();
				});

				$('a.ajax-action-play').on("click", function (event) {
					event.preventDefault();
					action("play", $(this).closest('tr').data("name"));
				});
			}
		});

		refresh();
	};

	return self;
}(MPD.plists || {}, jQuery));
