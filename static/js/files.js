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
		path = (path == "dt--root") ? "." : path.substring(3, path.length);
		$('input[name=p]').val(path);
		$.ajax({
			url: "serv/files",
			data: {
				id: path,
			},
			cache: true,
			dataType: "json"
		}).done(function(msg) {
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
			//"aLengthMenu": [[15, 25, 50, 100, -1], [15, 25, 50, 100, "All"]],
			"aoColumnDefs": [
				{
					"aTargets": [0],
					"mData": null,
					"mRender": function(data, type, full) {
						return ['<span class="glyphicon glyphicon-file"></span>',
							'<a href="?p=', full[3], '">', full[0], '</a>'].join("");
					},
				},
			],
		});

		$('#jstree_div').jstree({
			'core' : {
				'data' : {
					'url' : function () {
						return 'serv/dirs';
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
			selectPath(data.selected[0]);
		}).on("loaded.jstree", function() {
			selectPath("dt--root");
		}).on("loading.jstree", function() {
			showLoadingMessage();
		}).on("ready.jstree", function() {
			hideLoadingMessage();
		});
	};

	return self;
}(FILES.browser || {}, jQuery));
