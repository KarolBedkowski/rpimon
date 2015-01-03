/* jshint strict: true */
/* jshint undef: true, unused: true */
/* global jQuery: false */
/* global window: false */
/* global RPI: false */


var MPD = MPD || {};

MPD.search = (function(self, $) {
	"use strict";

	var urls = {
			"mpd-service-song-info": "/mpd/service/song-info",
			"mpd-file": "/mnd/file"
		},
		table = null;

	function action(action, uri) {
	   RPI.showLoadingMsg();
	   $.ajax({
	   	url: urls["mpd-file"],
	   	method: "PUT",
	   	data: {
	   		"action": action,
	   		"uri": uri
	   	}
	   }).always(function(result) {
	   	RPI.hideLoadingMsg();
	   }).fail(function(jqXHR, result) {
	   	alert(result);
	   });
	}

	function fileInfo(event) {
		event.preventDefault();
		var uri = $(this).closest('tr').data("uri");
		$.ajax({
			url: urls["mpd-service-song-info"],
			type: "GET",
			data: {
				uri: uri
			}
		}).done(function(data) {
			RPI.confirmDialog(data, {
				title: "Song info",
				btnSuccess: "none"
			}).open();
		});
	}

	self.init = function initF(params) {
		urls = $.extend(urls, params.urls || {});

		table = $('table').dataTable({
			"bAutoWidth": false,
			"bStateSave": true,
			//"sPaginationType": "bootstrap",
			"iDisplayLength": 25,
			"sDom": "<'row'<'col-xs-12 col-sm-6'l><'col-xs-12 col-sm-6'f>r>" + "t"+
				"<'row'<'col-xs-12 col-sm-6'i><'col-xs-12 col-sm-6'p>>"
		});
		$("a.add-file-action").on("click", function(event){
			event.preventDefault();
			action("add", $(this).closest('tr').data("uri"));
		});
		$("a.info-file-action").on("click", fileInfo);

	};

	return self;
}(MPD.library || {}, jQuery));
