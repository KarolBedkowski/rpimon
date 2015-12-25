/* jshint strict: true */
/* jshint undef: true, unused: true */
/* global window: false */
/* global jQuery: false */

"use strict";

var RPI = RPI || {};

RPI.net = (function(self, $) {
	var urls = {
			"net-serv-info": "/net/serv/info",
			"net-action": "/net/action"
		},
		contextMenu = null,
		contextMenuIface = null;

	function getHistory() {
		$.ajax({
			url: urls["net-serv-info"],
			cache: false,
			dataType: 'json'
		}).done(function(msg) {
			msg.ifaces.forEach(function(iface) {
				var name = iface.Name;
				$('#addr4-'+name).text(iface.Address);
				$('#addr6-'+name).text(iface.Address6);
				$('#state-'+name).text(iface.State);
			});
			var name;
			for (name in msg.netusage) {
				var usage = msg.netusage[name],
					inp = usage.Input,
					out = usage.Output;
				$('#chart-in-'+name).text(inp || '0').change();
				$('#chart-out-'+name).text(out || '0').change();
				if (inp) {
					$("#net-down-" + name).text(Math.round(inp[inp.length-1] / 1024) + " kB/s");
				}
				if (out) {
					$("#net-up-"+name).text(Math.round(out[out.length-1] / 1024) + " kB/s");
				}
			}
			window.setTimeout(getHistory, 5000);
		}).fail(function(jqXHR, textStatus) {
			window.setTimeout(getHistory, 10000);
		});
	}

	function onActionLink(event) {
		event.preventDefault();
		contextMenu.hide();
		if (!contextMenuIface) {
			return
		}
		var a = $(this),
			action = a.data("action");
		RPI.confirmDialog("Please confirm " + action + " action.", {
			title: "Confirm action",
			btnSuccess: "Continue",
			btnSuccessClass: "btn-warning",
			onSuccess: function() {
				RPI.showLoadingMsg();
				$.ajax({
					url: urls["net-action"],
					data: {
						"action": action,
						"iface": contextMenuIface
					}
				}).fail(function(msg) {
					RPI.hideLoadingMsg();
					if (window.console && window.console.log) { window.console.log(msg); }
					RPI.alert(msg.responseText || "Error").open();
					contextMenuIface = null;
				}).done(function(msg) {
					RPI.hideLoadingMsg();
					RPI.showFlash("success", msg, 5);
					selectPath(currentPath);
					contextMenuIface = null;
				});
			}
		}).open();
	}

	self.init = function init(params) {
		urls = $.extend({}, urls, params.urls || {});
		contextMenu = $("#contextMenu");
		$("span.chart-line").peity("line");
		$("a.iface-menu").on("click", function(event) {
			event.preventDefault();
			var pos = $(this).offset();
			contextMenuIface = $(this).data("iface");
			contextMenu.css({
				left: pos.left,
				top: pos.top + $(this).height()
			}).show();
			return false;
		});
		$(document).click(function () {
			contextMenu.hide();
		});
		$("#contextMenu a.iface-menu-item").on("click", onActionLink);
		getHistory();

	};

	return self;
}(RPI.net || {}, jQuery));
