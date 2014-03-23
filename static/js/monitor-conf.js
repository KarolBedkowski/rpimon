/* jshint strict: true */
/* jshint undef: true, unused: true */
/* global window: false */
/* global jQuery: false */

var Monitor = Monitor || {};

Monitor.conf = (function(self, $) {
	"use strict";

	function onRemoveRow(event) {
		event.preventDefault();
		if (confirm("Remove row?")) {
			$(this).closest('tr').remove();
		}
	}

	function updateServTableRows(parent) {
		$('input[type="number"]', parent).each(function (i) {
			$(this).rules("add", {
				range: [1, 65535],
				required: true
			});
		});
		$('a.serv-delete', parent).on("click", onRemoveRow);
	};

	function updateHostsTableRows(parent) {
		$('input[type="text"]', parent).each(function (i) {
			$(this).rules("add", {required: true});
		});
		$('input[type="number"]', parent).each(function (i) {
			$(this).rules("add", {range: [0, 9999]});
		});
		$('a.serv-delete', parent).on("click", onRemoveRow);
	};


	self.init = function() {
		$("form").validate({
			rules: {
				UpdateInterval: {min: 0},
				LoadWarning: {min: 0},
				LoadError: {min: 0},
				RAMUsageWarning: {range: [0, 100]},
				SwapUsageWarning: {range: [0, 100]},
				DefaultFSUsageWarning: {range: [0, 100]},
				DefaultFSUsageError: {range: [0, 100]},
				CPUTempWarning: {min: 0},
				CPUTempError: {min: 0}
			}
		});

		updateServTableRows($("#table-services tbody tr"));
		updateHostsTableRows($("#table-hosts tbody tr"));

		$("#services-add-row").on("click", function(event) {
			event.preventDefault();
			var lastId = $("#table-services").data("numserv"),
				tmpl = $("#tmpl-services-row").html(),
				newRow = $(tmpl.replace(/\[\[idx\]\]/g, lastId));
			$('#table-services tbody').append(newRow);
			$("#table-services").data("numserv", lastId + 1);
			updateServTableRows(newRow);
		});

		$("#hosts-add-row").on("click", function(event) {
			event.preventDefault();
			var lastId = $("#table-hosts").data("numhosts"),
				tmpl = $("#tmpl-hosts-row").html(),
				newRow = $(tmpl.replace(/\[\[idx\]\]/g, lastId));
			$('#table-hosts tbody').append(newRow);
			$("#table-hosts").data("numhosts", lastId + 1);
			updateHostsTableRows(newRow);
		});
	};

	return self;
}(Monitor.conf || {}, jQuery));
