
var RPI = RPI || {};

RPI.main = (function(self, $) {
	"use strict";

	var urls = {
		"main-serv-alerts": "/main/serv/alerts"
	};

	function getAlarms() {
		$.get(urls["main-serv-alerts"]).success(function(data) {
			var warnings = data.warnings || [],
				ddst = $(".dropdown-alerts");
			ddst.html("");
			warnings.forEach(function(warn) {
				$('<li>').append($('<a href="#">').append($("<div>")).append($('<span>').append('<span class="glyphicon glyphicon-info-sign"></span> ').append(warn))).appendTo(ddst);
			});
			if (warnings > 0) {
				$("#nav-alerts-dropdown").show();
				$("#nav-alerts-cnt").text(warnings.length).show();
			} else {
				$("#nav-alerts-dropdown").hide();
				$("#nav-alerts-cnt").text(warnings.length).hide();
			}
		}).always(function() {
			window.setTimeout(getAlarms, 10000);
		});
	}

	self.init = function(params) {
		urls = $.extend({}, urls, params.urls || {});
		getAlarms();
	};

	return self;
}(RPI.main || {}, jQuery));
