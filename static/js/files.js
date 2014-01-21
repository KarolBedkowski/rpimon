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

	function gotoPath(event) {
		event.preventDefault();
		var obj = $(this),
			p = obj.data("p");
		if (!p) {
			p = obj.closest('tr').data("p");
		}
		selectPath(p);
	}

	function updateBreadcrumb(path) {
		var bc = $("#breadcrumb"),
			pathParts = path.split("/"),
			idx;
		if (!path || path == ".") {
			bc.html("<li>[Root]</li>");
			return
		}
		bc.html('<li class="active"><a href="#" data-p=".">[Root]</a></li>');
		var lpath = "";
		for (idx = 0; idx < pathParts.length - 1; ++idx) {
			if (lpath) {
				lpath = lpath + "/";
			}
			lpath = lpath + pathParts[idx];
			$(['<li class="active"><a href="#" data-p="', lpath, '">',
				pathParts[idx], '</a></li>'].join('')).appendTo(bc);
		}
		$(['<li>', pathParts[pathParts.length - 1], '</li>'].join('')).appendTo(bc);
		$("#breadcrumb a").on("click", gotoPath);
	};

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
			updateBreadcrumb(path);
			hideLoadingMessage();
		});
	}

	function removePath(event) {
		event.preventDefault();
		var p = $(this).closest('tr').data("p");
		if (!p) {
			return;
		}
		window.RPI.confirmDialog("Delete " + p + "?", {
			title: "Confirm delete",
			btnSuccess: "Delete",
			btnSuccessClass: "btn-warning",
			onSuccess: function() {
				window.location.href = "action?action=delete&p=" + p;
			}
		}).open();
	}

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
							return ['<a class="ajax-action-open" href="#">', data, '</a>'].join("");
						}
					},
				},
				{
					"aTargets": [4],
					"mData": 1,
					"mRender": function(data, type, full) {
						if (data != "..") {
							return '<a href="#" class="ajax-action-del"><span class="glyphicon glyphicon-remove" title="Remove"></span></a>';
				  		}
						return "";
					},
				},
			],
			"fnRowCallback": function(row, aData) { //, iDisplayIndex, iDisplayIndexFull) {
				$(row).data("p", aData[4]);
			},
			"fnDrawCallback": function() { //oSettings) {
				$("table a.ajax-action-open").on("click", gotoPath);
				$("table a.ajax-action-del").on("click", removePath);
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
