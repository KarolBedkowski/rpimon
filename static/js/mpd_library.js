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
			"mpd-library-action": "",
			"mpd-library-content": "/mpd/library/content"
		},
		currentPath = "",
		table = null;

	function updateBreadcrumb(path) {
		var bc = $("#breadcrumb");
		if (!path || path == "/") {
			bc.html("<li>[Library]</li>");
			return;
		}
		bc.html('<li class="active"><a href="#" data-uri="/">[Library]</a></li>');
		var lpath = "/", pathParts = path.split("/"), idx;
		for (idx = 1; idx < pathParts.length - 1; ++idx) {
			lpath = lpath + pathParts[idx];
			$(['<li class="active"><a href="#" data-uri="', lpath, '">',
				pathParts[idx], '</a></li>'].join('')).appendTo(bc);
			lpath = lpath + "/";
		}
		$(['<li>', pathParts[pathParts.length - 1], '</li>'].join('')).appendTo(bc);
		$("#breadcrumb a").on("click", gotoAction);
	}

	function gotoAction(event) {
		event.preventDefault();
		var obj = $(this),
			p = obj.data("uri");
		if (!p) {
			p = obj.closest('tr').data("uri");
		}
		if (p.charAt(p.length - 1) != '/') {
			p = p + "/"
		}
		selectPath(p);
	}

	function selectPath(path) {
		currentPath = path;
		RPI.showLoadingMsg();
		$.ajax({
			url: urls["mpd-library-content"],
			data: {p: path}
		}).done(function(response) {
			if (response.error) {
				showError(response.error);
			}
			else {
				currentPath = response.path;
				updateBreadcrumb(currentPath);
				var new_location = "?p="  + currentPath;
				window.history.pushState({ path: new_location }, window.title, new_location);
				table.fnClearTable();
				table.fnAddData(response.items || []);
			}
		}).fail(function(jqXHR, result) {
			showError(result);
		}).always(function() {
			RPI.hideLoadingMsg();
		});
	}

	function showError(errormsg) {
		RPI.hideLoadingMsg();
		if (window.console && window.console.log) { window.console.log(errormsg); }
		$("#main-alert-error").text(errormsg);
		$("#main-alert").show();
		$("div.playlist-data").hide();
	}

	self.init = function initF(params) {
		urls = $.extend(urls, params.urls || {});

		table = $('table').dataTable({
			"bAutoWidth": false,
			"bStateSave": true,
			"sPaginationType": "bootstrap",
			"iDisplayLength": 25,
			"sDom": "<'row'<'col-xs-12 col-sm-6'l><'col-xs-12 col-sm-6'f>r>" + "t"+
				"<'row'<'col-xs-12 col-sm-6'i><'col-xs-12 col-sm-6'p>>",
			"aoColumnDefs": [
				{
					"aTargets": [0],
					"mData": 0,
					"mRender": function(data, type, full) {
						if (data == "0") { // folder
							return '<a href="#" class="open-action"><span class="glyphicon glyphicon-folder-close"></span></a>';
						} // file
						return '<span class="glyphicon glyphicon-file"></span>';
					}
				},
				{
					"aTargets": [1],
					"mData": 1,
					"mRender": function(data, type, full) {
						if (full[0] == "0") { // folder
							return '<a href="#" class="open-action">' + data + '</a>';
						}
						return data;
					}
				},
				{
					"aTargets": [2],
					"mData": 0,
					"mRender": function(data, type, full) {
						var res = ['<a href="#" class="ajax-action" data-action="add"><span class="glyphicon glyphicon-plus" title="Add"></a>',
								'<a href="#" class="ajax-action" data-action="replace"><span class="glyphicon glyphicon-play" title="Replace"></a>'
						];
						if (data == "1") { // file
							res.push('<a href="#" class="action-info"><span class="glyphicon glyphicon-info-sign" title="Info"></a>')
						}
						return res.join('&nbsp;');
					}
				}
			],
			"fnRowCallback": function(row, aData) { //, iDisplayIndex, iDisplayIndexFull) {
				$(row).data("uri", currentPath + aData[1]);
			},
			"fnDrawCallback": function() { //oSettings) {
				$("a.open-action").on("click", gotoAction);

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
			}
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

		$(window).bind('popstate', function(event) {
			var location = window.location.search;
			if (location && location.startsWith("?p=")) {
				location = location.substr(3, location.length);
				location = decodeURIComponent(location);
				selectPath(location || "/");
			}
		});

		var location = window.location.search;
		if (location && location.startsWith("?p=")) {
			location = location.substr(3, location.length);
			location = decodeURIComponent(location);
		}
		selectPath(location || "/");
	};

	return self;
}(MPD.library || {}, jQuery));
