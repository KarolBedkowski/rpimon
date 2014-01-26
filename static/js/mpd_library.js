/* jshint strict: true */
/* jshint undef: true, unused: true */
/* global jQuery: false */
/* global window: false */
/* global RPI: false */


var MPD = MPD || {};

MPD.library = (function(self, $) {
	"use strict";

	var urls = {
			"mpd-service-song-info": ""
		};

	self.init = function initF(params) {
		urls = $.extend(urls, params.urls || {});

		$('table').dataTable({
			"bAutoWidth": false,
			"bStateSave": true,
			"sPaginationType": "bootstrap",
			"aoColumnDefs": [{
				"aTargets": [1],
				"bSortable": false
			}],
			"sDom": "<'row'<'col-xs-12 col-sm-6'l><'col-xs-12 col-sm-6'f>r>" + "t"+
				"<'row'<'col-xs-12 col-sm-6'i><'col-xs-12 col-sm-6'p>>"
		});

		$("a.ajax-action").on("click", function(event) {
			event.preventDefault();
			var link = $(this);
			RPI.showLoadingMsg();
			$.ajax({
				type: "PUT",
				data: {
					a: link.data("action"),
					u: link.data("uri")
				}
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
		});

		$("a#action-update").on("click", function(event) {
			event.preventDefault();
			var url = $(this).attr("href");
			RPI.confirmDialog("Start updating library?", {
				title: "Library",
				btnSuccess: "Update",
				onSuccess: function() {
					RPI.showLoadingMsg();
					$.get(url
					).always(function() {
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
