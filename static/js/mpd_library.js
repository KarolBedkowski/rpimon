/* jshint strict: true */
/* jshint undef: true, unused: true */
/* global jQuery: false */
/* global window: false */
/* global RPI: false */


var MPD = MPD || {};

MPD.library = (function(self, $) {
	"use strict";

	var urls = {
			"mpd-service-song-info": "",
			"mpd-library-action": ""
		};

	self.init = function initF(params) {
		urls = $.extend(urls, params.urls || {});

		$('table').dataTable({
			"bAutoWidth": false,
			"bStateSave": true,
			"sPaginationType": "bootstrap",
			"iDisplayLength": 25,
			"aoColumnDefs": [{
				"aTargets": [1],
				"bSortable": false
			}],
			"sDom": "<'row'<'col-xs-12 col-sm-6'l><'col-xs-12 col-sm-6'f>r>" + "t"+
				"<'row'<'col-xs-12 col-sm-6'i><'col-xs-12 col-sm-6'p>>"
		});

		$("a.ajax-action").on("click", function(event) {
			event.preventDefault();
			var link = $(this),
				uri = link.closest('tr').data("uri");
			RPI.showLoadingMsg();
			$.ajax({
				url: urls["mpd-library-action"],
				type: "PUT",
				data: {a: link.data("action"), u: uri}
			}).always(function() {
				RPI.hideLoadingMsg();
			}).done(function(res) {
				RPI.showFlash("success", res, 2);
			}).fail(function(jqXHR, textStatus) {
				RPI.alert(textStatus, {
					title: "Error"
				}).open();
			});
		});

		$("a.action-info").on("click", function(event) {
			var uri = $(this).closest('tr').data("uri");
			event.preventDefault();
			$.ajax({
				url: urls["mpd-service-song-info"],
				type: "GET",
				data: {uri: uri}
			}).done(function(data) {
				RPI.confirmDialog(data, {
					title: "Song info",
					btnSuccess: "none"
				}).open();
			});
		});

		$("a#action-update").on("click", function(event) {
			event.preventDefault();
			var url = $(this).attr("href"),
				uri = $(this).data("uri");
			RPI.confirmDialog("Start updating " + (uri ? "folder?" : "library?"), {
				title: "Library",
				btnSuccess: "Update",
				onSuccess: function() {
					RPI.showLoadingMsg();
					$.get(url, {uri: uri
					}).always(function() {
						RPI.hideLoadingMsg();
					}).done(function() {
						RPI.showFlash("success", "Library update started", 5);
					}).fail(function(jqXHR, textStatus) {
						RPI.showFlash("error", textStatus);
					});
				}
			}).open();
		});
	};

	return self;
}(MPD.library || {}, jQuery));
