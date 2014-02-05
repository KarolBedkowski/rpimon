/* jshint strict: true */
/* jshint undef: true, unused: true */
/* global window: false */
/* global jQuery: false */

"use strict";

var SYSTEM = (function(self, $) {
	var infoUrl = "/main/info";

	function getHistory() {
		$.ajax({
			url: infoUrl,
			cache: false,
			dataType: 'json'
		}).done(function(msg) {
			var meminfo = msg.meminfo,
				cpuusage = msg.cpuusage,
				cpuinfo = msg.cpuinfo,
				load = msg.loadinfo,
				nettablebody = $('tbody#network-interfaces-table'),
				fstablebody = $('tbody#fs-table'),
				uptime = msg.uptime;
			$('#load-chart').text(msg.load).change();
			$('#cpu-chart').text(msg.cpu).change();
			$('#mem-chart').text(msg.mem).change();
			$("#net-in-chart").text(msg.netusage.Input.join(",")).change();
			$("#net-out-chart").text(msg.netusage.Output.join(",")).change();
			$('#mem-chart').text(msg.mem).change();
			$('#meminfo-used').text(meminfo.UsedPerc);
			$('#meminfo-buff').text(meminfo.BuffersPerc);
			$('#meminfo-cach').text(meminfo.CachePerc);
			$('#meminfo-swap').text(meminfo.SwapUsedPerc);
			$('#cpuusage-user').text(cpuusage.User);
			$('#cpuusage-system').text(cpuusage.System);
			$('#cpuusage-iowait').text(cpuusage.IoWait);
			$('#cpuinfo-freq').text(cpuinfo.Freq);
			$('#cpuinfo-temp').text(cpuinfo.Temp);
			$('#load-load1').text(load.Load1);
			$('#load-load5').text(load.Load5);
			$('#load-load15').text(load.Load15);
			// network
			nettablebody.text("");
			msg.iface.forEach(function(entry) {
				nettablebody.append(["<tr><td>", entry.Name, "</td><td>",
					entry.Address, "</td></tr>"].join(""));
			});
			// fs
			fstablebody.text("");
			msg["fs"].forEach(function(entry) {
				fstablebody.append(["<tr><td>", entry["MountPoint"], "</td><td>",
					"<span class=\"pie\" data-diameter=\"32\" data-colours='[\"red\", \"#f0f0f0\"]'>",
					entry["UsedPerc"], "/100</span></td><td>",
					entry["UsedPerc"], "%</td></tr>"].join(""));

			});
			$('#uptime-uptime').text(uptime.Uptime);
			$('#uptime-users').text(uptime.Users);
			$("span.pie").peity("pie");
			RPI.hideLoadingMsg();
			window.setTimeout(getHistory, 5000);
		}).fail(function(jqXHR, textStatus) {
			RPI.showLoadingMsg();
			window.setTimeout(getHistory, 10000);
		});
	}

	self.init = function init(infoUrl_) {
		infoUrl = infoUrl_;
		RPI.showLoadingMsg();
		getHistory();
	};

	return self;
}(SYSTEM || {}, jQuery));
