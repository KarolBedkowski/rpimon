/* jshint strict: true */
/* jshint undef: true, unused: true */
/* global window: false */
/* global jQuery: false */


var SYSTEM = (function(self, $) {
	"use strict";

	var urls = {
		"main-serv-status": ""
	};

	function getHistory() {
		$.ajax({
			url: urls["main-serv-status"],
			cache: false,
			dataType: 'json'
		}).done(function(msg) {
			var meminfo = msg.meminfo,
				cpuusage = msg.cpuusage,
				cpuinfo = msg.cpuinfo,
				load = msg.loadinfo,
				nettablebody = $('tbody#network-interfaces-table'),
				fstablebody = $('tbody#fs-table'),
				uptime = msg.uptime,
				netuseInput = msg.netusage.Input || [],
				netuseOutput = msg.netusage.Output || [];
			$('#load-chart').text(msg.load).change();
			$('#cpu-chart').text(msg.cpu).change();
			$('#mem-chart').text(msg.mem).change();
			$("#net-in-chart").text(netuseInput.join(",")).change();
			$("#net-out-chart").text(netuseOutput.join(",")).change();
			if (netuseInput) {
				$("#network-download").text(Math.round(netuseInput[netuseInput.length-1] / 1024) + " kB/s");
			}
			if (netuseOutput) {
				$("#network-upload").text(Math.round(netuseOutput[netuseOutput.length-1] / 1024) + " kB/s");
			}
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
				if (entry.State != "UP") {
					return
				}
				var row = ["<tr><td>", entry.Name, "</td><td>"];
				if (entry.Address && entry.Address6) {
					row.push(entry.Address + "<br/>"+ entry.Address6);
				} else if (entry.Address) {
					row.push(entry.Address);
				} else {
					row.push(entry.Address6);
				}
				row.push("</td></tr>");
				nettablebody.append(row.join(""));
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

	self.init = function init(params) {
		urls = $.extend({}, urls, params.urls || {});
		RPI.showLoadingMsg();
		getHistory();
	};

	return self;
}(SYSTEM || {}, jQuery));
