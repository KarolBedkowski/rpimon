/* jshint strict: true */
/* jshint undef: true, unused: true */
/* global jQuery: false */
/* global window: false */


var MPD = MPD || {};
var RPI = RPI || {};

MPD.history = (function(self, $) {
	"use strict";

	var msg_loading = null,
		table = null,
		urls = {
			'mpd-hist-serv': ''
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
			"bProcessing": true,
			"bSort": false,
			"bServerSide": true,
			"sAjaxSource": urls['mpd-hist-serv'],
			"fnServerData": processServerData,
			"bFilter": false,
			"columns": [
				{ "data": "ID" },
				{ "data": "DateStr" },
				{ "data": "Title" },
				{ "data": "Artist" },
				{ "data": "Track" },
				{ "data": "Album" },
				{ "data": "Name" }
			]
		});
		return;
	};

	self.init = function initF(params) {
		RPI.showLoadingMsg();
		urls = $.extend({}, urls, params.urls || {});
		self.refresh();
	};

	return self;
}(MPD.history || {}, jQuery));
