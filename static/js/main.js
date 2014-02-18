
var RPI = RPI || {};

RPI.main = (function(self, $) {
	"use strict";

	var urls = {
		"main-serv-alerts": "/main/serv/alerts"
	};

	function getAlarms() {
		$.get(urls["main-serv-alerts"]).success(function(data) {
			var datawarn = data.warnings || {},
				warnings = datawarn.Warnings || [],
				errors = datawarn.Errors || [],
				infos = datawarn.Infos || [],
				ddst = $(".dropdown-alerts");
			if (!warnings && !errors && !infos) {
				$("#nav-alerts-dropdown").hide();
				return
			}
			ddst.html("");
			if (errors.length > 0) {
				errors.forEach(function(warn) {
					$('<li>').append($('<a href="#">').append($("<div>")).append($('<span>').append('<span class="label label-danger"><span class="glyphicon glyphicon-exclamation-sign"></span></span> ').append(warn))).appendTo(ddst);
				});
				$("#nav-errors-cnt").text(errors.length).show();
				$('<li class="divider"></li>').appendTo(ddst);
			} else {
				$("#nav-errors-cnt").hide();
			}
			if (warnings.length > 0) {
				warnings.forEach(function(warn) {
					$('<li>').append($('<a href="#">').append($("<div>")).append($('<span>').append('<span class="label label-warning"><span class="glyphicon glyphicon-warning-sign"></span></span> ').append(warn))).appendTo(ddst);
				});
				$("#nav-warns-cnt").text(warnings.length).show();
				$('<li class="divider"></li>').appendTo(ddst);
			} else {
				$("#nav-warns-cnt").hide();
			}
			if (infos.length > 0) {
				infos.forEach(function(warn) {
					$('<li>').append($('<a href="#">').append($("<div>")).append($('<span>').append('<span class="label label-info"><span class="glyphicon glyphicon-info-sign"></span></span> ').append(warn))).appendTo(ddst);
				});
				$("#nav-infos-cnt").text(infos.length).show();
			} else {
				$("#nav-infos-cnt").hide();
			}
			$("#nav-alerts-dropdown").show();
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
