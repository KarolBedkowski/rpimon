/* jshint strict: true */
/* jshint undef: true, unused: true */
/* global Messi: false */
/* global jQuery: false */

"use strict";

var FILES = FILES || {};

FILES.browser = (function(self, $) {
	var msg_loading = null,
		table = null;

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

	function selectPath(path) {
		showLoadingMessage();
		$('input[name=p]').val(path);
		$.ajax({
			url: "serv/files",
			data: {
				id: path,
			},
			cache: true,
			dataType: "json"
		}).done(function(msg) {
			var new_location = "?p="+path;
			window.history.pushState({ path: new_location }, window.title, new_location);
			table.fnClearTable();
			table.fnAddData(msg);
			hideLoadingMessage();
		});
	}

	self.init = function initF() {
		showLoadingMessage();

		table = $('table').dataTable({
			"bAutoWidth": false,
			"bStateSave": true,
			"sPaginationType": "bootstrap",		
			"bFilter": false,
			"iDisplayLength": 50,
			"bLengthChange": false,
			"aoColumnDefs": [
				{
					"aTargets": [0],
					"mData": 0,
					"mRender": function(data, type, full) {
						if (data == 'file') {
							return '<span class="glyphicon glyphicon-file"></span>';
						} else {
							return '<span class="glyphicon glyphicon-folder-close"></span>';
						}
					},
				},
				{
					"aTargets": [1],
					"mData": 1,
					"mRender": function(data, type, full) {
						if (full[0] == 'file') {
							return ['<a href="?p=', full[4], '">', data, '</a>'].join("");
						} else {
							return ['<a href="#" data-kind="' + full[0] + '" data-p="', full[4], '">', data, '</a>'].join("");
						}
					},
				},
			],
			"fnDrawCallback": function() { //oSettings) {
				$("table a[href=#]").on("click", function(event) {
					event.preventDefault();
					var path = $(this).data("p");
					selectPath(path);
				});
			},
		});

		$(window).bind('popstate', function(event) {	
			var location = window.location.search;
			if (location && location.startsWith("?p=")) {
				location = location.substr(3, location.length);
				location = decodeURIComponent(location);
				selectPath(location || ".");
			} 
		});

		var location = window.location.search;
		if (location && location.startsWith("?p=")) {
			location = location.substr(3, location.length);
			location = decodeURIComponent(location);
		} 
		selectPath(location || ".");
	};

	return self;
}(FILES.browser || {}, jQuery));
