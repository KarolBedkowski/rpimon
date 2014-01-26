/* jshint strict: true */
/* jshint undef: true, unused: true */
/* global jQuery: false */
/* global window: false */


var MPD = MPD || {};
var RPI = RPI || {};

MPD.plists = (function(self, $) {
	"use strict";

	function action(event) {
		event.preventDefault();
		RPI.showLoadingMsg();
		$.ajax({
			url: this.href,
			type: "PUT"
		}).done(function(msg) {
			RPI.hideLoadingMsg();
			RPI.showFlash("success", msg, 1);
		}).fail(function(jqXHR, textStatus) {
			RPI.hideLoadingMsg();
			RPI.alert(textStatus, {
				title: "Error"
			}).open();
		});
	}

	self.init = function initF() {
		$('table').dataTable({
			"bAutoWidth": false,
			"bStateSave": true,
			"bFilter": false,
			"sPaginationType": "bootstrap",
			"iDisplayLength": 15,
			"bLengthChange": false,
			"aoColumnDefs": [{
				"aTargets": [1],
				"bSortable": false
			}]
//			"sDom": "t"+
//				"<'row'<'col-xs-12 col-sm-6'i><'col-xs-12 col-sm-6'p>>"
		});
		$('a.action-confirm').on("click", function() {
			return RPI.confirm();
		});
		$('a.ajax-action').on("click", action);
	};

	return self;
}(MPD.plists || {}, jQuery));
