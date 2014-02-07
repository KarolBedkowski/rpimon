/* jshint strict: true */
/* jshint undef: true, unused: true */
/* global window: false */
/* global jQuery: false */

"use strict";

var RPI = RPI || {};

RPI.net = (function(self, $) {
	var infoUrl = "/net/serv/info";

	function getHistory() {
		$.ajax({
			url: infoUrl,
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
				var inp = msg.netusage[name].Input,
					out = msg.netusage[name].Output;
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

	self.init = function init(infoUrl_) {
		infoUrl = infoUrl_;
		$("span.chart-line").peity("line");
		getHistory();
	};

	return self;
}(RPI.net || {}, jQuery));
