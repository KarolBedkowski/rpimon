/* jshint strict: true */
/* jshint undef: true, unused: true */
/* global jQuery: false */
/* global window: false */
/* global RPI: false */


var RPI = RPI || {};

RPI.utils = (function(self, $) {
	"use strict";
	var token = "";

	function action(url) {
		RPI.showLoadingMsg();
		$.ajax({
			url: url,
			method: "POST",
			data: {"BasePageContext.CsrfToken": token}
		}).always(function(result) {
			RPI.hideLoadingMsg();
		}).done(function(data) {
			RPI.showFlash("success", data, 5);
		}).fail(function(jqXHR, result) {
			RPI.alert(result).open();
		});
	}

	self.init = function initF(params) {
		token = params.token || "";
		$("a.list-group-item").on("click", function(evt) {
			evt.preventDefault();
			action(this.href);
		});
	};

	return self;
}(RPI.utils || {}, jQuery));
