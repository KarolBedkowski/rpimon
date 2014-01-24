/* jshint strict: true */
/* jshint undef: true, unused: true */
/* global jQuery: false */
/* global window: false */
/* global RPI: false */



var FILES = FILES || {};

FILES.browser = (function(self, $) {
	"use strict";

	var table = null,
		dlgDirTreeSelection = null,
		currentPath = null,
		urls = {
			"service-dirs": "serv/dirs",
			"service-files": "serv/files",
			"file-action": "action",
		}
		;


	function gotoAction(event) {
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
			return;
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
		$("#breadcrumb a").on("click", gotoAction);
	}

	function selectPath(path) {
		currentPath = path;
		RPI.showLoadingMsg();
		$('input[name=p]').val(path);
		$.ajax({
			url: urls["service-files"],
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
			RPI.hideLoadingMsg();
		});
	}

	function deleteAction(event) {
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
				RPI.showLoadingMsg();
				$.ajax({
					method: "PUT",
					url: urls["file-action"],
					data: {
						"action": "delete",
						"p": p,
					},
				}).fail(function(msg) {
					RPI.hideLoadingMsg();
					window.console.log(msg);
					RPI.alert(msg.responseText).open();
				}).done(function(msg) {
					RPI.hideLoadingMsg();
					window.console.log(msg);
					selectPath(currentPath);
				});
			}
		}).open();
	}

	function moveAction(event) {
		event.preventDefault();
		var p = $(this).closest('tr').data("p"),
			dlg = $("div#dialog-dirtree");
		if (!p) {
			return;
		}
		createTree();
		$("#dialog-dirtree-label", dlg).html("Move destination");
		$("#dialog-dirtree-msg", dlg).html("Move " + p + " to:");
		dlg.modal("show");
		$("#dialog-dirtree-success", dlg).on("click", function() {
			dlg.modal("hide");
			if (p != dlgDirTreeSelection) {
				RPI.showLoadingMsg();
				$.ajax({
					method: "PUT",
					url: urls["file-action"],
					data: {
						"action": "move",
						"p": p,
						"d": dlgDirTreeSelection,
					},
				}).fail(function(msg) {
					RPI.hideLoadingMsg();
					window.console.log(msg);
					RPI.alert(msg.responseText).open();
				}).done(function(msg) {
					RPI.hideLoadingMsg();
					window.console.log(msg);
					selectPath(currentPath);
				});
			}
		});
	}

	function createDirectory(event) {
		event.preventDefault();
		$('#create-folder-dlg button[type="submit"]').button('loading');
		$.ajax({
			method: "POST",
			url: $(this).attr("action"),
			data: $(this).serialize(),
		}).done(function(msg) {
			$("#create-folder-dlg").modal("hide");
			window.console.log(msg);
			selectPath(currentPath);
			$('#create-folder-dlg input[name="name"]').val("");
			$('#create-folder-dlg button[type="submit"]').button('reset');
		}).fail(function(msg) {
			window.console.log(msg);
			$("#create-folder-dlg").modal("hide");
			$('#create-folder-dlg button[type="submit"]').button('reset');
			RPI.alert(msg.responseText);
		});
	}

	function createTree() {
		$('#dialog-dirtree #dialog-dirtree-tree').jstree({
			'core' : {
				'data' : {
					'url' : function () {
						return urls["service-dirs"];
					},
					'data' : function (node) {
						return { 'id' : node.id };
					}
				},
				"themes" : {
					"variant": "small",
					"responsive": false,
				},
			}
		}).on("select_node.jstree", function (e, data) {
			var path = data.selected[0];
			dlgDirTreeSelection = (path == "dt--root") ? "." : decodeURIComponent(path.substr(3, path.length));
		}).on("loaded.jstree", function() {
		}).on("loading.jstree", function() {
			RPI.showLoadingMsg();
		}).on("ready.jstree", function() {
			RPI.hideLoadingMsg();
		});
	}

	$.fn.dataTableExt.oSort['data-asc']  = function(a,b) {
		var x = parseInt($(a).data("sortval"));
		var y = parseInt($(b).data("sortval"));
		return ((x < y) ? -1 : ((x > y) ?  1 : 0));
	};
	$.fn.dataTableExt.oSort['data-desc']  = function(a,b) {
		var x = parseInt($(a).data("sortval"));
		var y = parseInt($(b).data("sortval"));
		return ((x < y) ? 1 : ((x > y) ?  -1 : 0));
	};

	self.init = function initF(params) {
		RPI.showLoadingMsg();
		urls = $.extend(urls, params.urls || {});
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
							return '<a class="ajax-action-open" href="#" data-sortval="1"><span class="glyphicon glyphicon-file"></span></a>';
						} else {
							if (full[1] == '..') {
								return '<a class="ajax-action-open" href="#" data-sortval="-1"><span class="glyphicon glyphicon-folder-close"></span></a>';
							} else {
								return '<a class="ajax-action-open" href="#" data-sortval="0"><span class="glyphicon glyphicon-folder-close"></span></a>';
							}
						}
					},
					"sType": "data",
					"aDataSort": [ 0, 1 ],
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
					"aDataSort": [ 1, 0 ],
				},
				{
					"aTargets": [4],
					"mData": 1,
					"bSortable": false,
					"mRender": function(data, type, full) {
						if (data != "..") {
							return '<a href="#" class="ajax-action-del"><span class="glyphicon glyphicon-remove" title="Remove"></span></a>'+
							' <a href="#" class="ajax-action-move"><span class="glyphicon glyphicon-share-alt" title="Move"></span></a>';
						}
						return "";
					},
				},
			],
			"aaSorting": [[0,'asc'], [1,'asc']],
			"fnRowCallback": function(row, aData) { //, iDisplayIndex, iDisplayIndexFull) {
				$(row).data("p", aData[4]); //.on("click", gotoAction);
			},
			"fnDrawCallback": function() { //oSettings) {
				$("table a.ajax-action-open").on("click", gotoAction);
				$("table a.ajax-action-del").on("click", deleteAction);
				$("table a.ajax-action-move").on("click", moveAction);
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
		$('#create-folder-dlg').on('shown.bs.modal', function() {
			$('#folder-name').focus();
		});
		$('#upload-dlg').on('shown.bs.modal', function() {
			$('#upload-dlg input[type="file"]').focus();
		});
		$("#create-folder-dlg form").submit(createDirectory);

		var location = window.location.search;
		if (location && location.startsWith("?p=")) {
			location = location.substr(3, location.length);
			location = decodeURIComponent(location);
		}
		selectPath(location || ".");
	};

	return self;
}(FILES.browser || {}, jQuery));
