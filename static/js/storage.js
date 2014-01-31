/* jshint strict: true */
/* jshint undef: true, unused: true */
/* global jQuery: false */
/* global window: false */
/* global RPI: false */

var RPI = RPI || {};

RPI.storage = (function(self, $) {
	"use strict";

	var table = null,
		urls = {
			'storage-umount': ""
		};

	self.init = function initF(params) {
		urls = $.extend(urls, params.urls || {});
		table = $('table').dataTable({
			"bAutoWidth": false,
			"bStateSave": true,
			"bFilter": false,
			"sPaginationType": "bootstrap",
			"iDisplayLength": 15,
			"bLengthChange": false,
			"aoColumnDefs": [{
				"aTargets": [4],
				"bSortable": false
			}]
		});

		$("a.umount-action").on("click", function(event) {
			event.preventDefault();
			var fs = $(this).data('fs');
			window.RPI.confirmDialog("Umount " + fs + "?", {
				title: "Confirm umount",
				btnSuccess: "Umount",
				btnSuccessClass: "btn-warning",
				onSuccess: function() {
					window.location.href = urls['storage-umount'] + "?fs=" + fs;
				}
			}).open();
		});
	};

	return self;
}(RPI.storage || {}, jQuery));
