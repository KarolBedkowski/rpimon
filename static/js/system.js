
var SYSTEM = (function(self) {
	var infoUrl = "/main/info";
	var connectingMessage = null;

	function getHistory() {
		$.ajax({
			url: infoUrl,
			cache: false,
			dataType: 'json'
		}).done(function(msg) {
			$('#load-chart').text(msg['load']).change()
			$('#cpu-chart').text(msg['cpu']).change()
			$('#mem-chart').text(msg['mem']).change()
			var meminfo = msg['meminfo'];
			$('#meminfo-used').text(meminfo['UsedPerc']);
			$('#meminfo-buff').text(meminfo['BuffersPerc']);
			$('#meminfo-cach').text(meminfo['CachePerc']);
			$('#meminfo-swap').text(meminfo['SwapUsedPerc']);
			var cpuusage = msg['cpuusage'];
			$('#cpuusage-user').text(cpuusage['User']);
			$('#cpuusage-system').text(cpuusage['System']);
			$('#cpuusage-iowait').text(cpuusage['IoWait']);
			var cpuinfo = msg['cpuinfo'];
			$('#cpuinfo-freq').text(cpuinfo['Freq']);
			$('#cpuinfo-temp').text(cpuinfo['Temp']);
			var load = msg['loadinfo'];
			$('#load-load1').text(load["Load1"]);
			$('#load-load5').text(load["Load5"]);
			$('#load-load15').text(load["Load15"]);
			// network
			var nettablebody = $('tbody#network-interfaces-table');
			nettablebody.text("");
			msg["iface"].forEach(function(entry) {
				nettablebody.append(["<tr><td>", entry["Name"], "</td><td>",
					entry["Address"], "</td></tr>"].join(""));
			});
			// fs
			var fstablebody = $('tbody#fs-table');
			fstablebody.text("");
			msg["fs"].forEach(function(entry) {
				fstablebody.append(["<tr><td>", entry["MountPoint"], "</td><td>",
					"<span class=\"pie\" data-diameter=\"32\" data-colours='[\"red\", \"#f0f0f0\"]'>", 
					entry["UsedPerc"], "/100</span></td><td>",
					entry["UsedPerc"], "%</td></tr>"].join(""));

			});
			var uptime = msg["uptime"];
			$('#uptime-uptime').text(uptime['Uptime'])
			$('#uptime-users').text(uptime['Users'])
			$("span.pie").peity("pie")
			connectingMessage.hide();
			setTimeout(getHistory, 5000);
		}).fail(function(jqXHR, textStatus) {
			connectingMessage.show();
			setTimeout(getHistory, 10000);
		});
	};

	self.init = function init(infoUrl_) {
		infoUrl = infoUrl_
		connectingMessage = new Messi('Connecting...', {
			closeButton: false,
			width: 'auto',
		});
		getHistory();
	};

	return self;
}(SYSTEM || {}));
