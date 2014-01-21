/* jshint strict: true */
/* jshint undef: true, unused: true */
/* global Messi: false */
/* global jQuery: false */
/* global window: false */
/* global RPI: false */


var MPD = MPD || {};

MPD.library = (function(self, $) {
	"use strict";

	var mpdControlUrl = null,
		mpdServiceInfoUrl = null;

	self.init = function initF(mpdControlUrl_, mpdServiceInfoUrl_) {
		mpdControlUrl = mpdControlUrl_;
		mpdServiceInfoUrl = mpdServiceInfoUrl_;

		$('table').dataTable({
			"bAutoWidth": false,
			"bStateSave": true,
			"sPaginationType": "bootstrap",		
			"aoColumnDefs": [{
				"aTargets": [1],
				"bSortable": false,
			}],
			"sDom": "<'row'<'col-xs-12 col-sm-6'l><'col-xs-12 col-sm-6'f>r>" + "t"+
				"<'row'<'col-xs-12 col-sm-6'i><'col-xs-12 col-sm-6'p>>",
		});


		$("a.ajax-action").on("click", function(event) {
			event.preventDefault();
			var link = $(this),
				lmessage = new Messi('Adding...', {
					closeButton: false,
					modal: true,
					width: 'auto',
				});
			$.ajax({
				type: "PUT",
				data: {
					a: link.data("action"),
					u: link.data("uri"),
				}
			}).done(function() {
				lmessage.hide();
			}).fail(function(jqXHR, textStatus) {
				window.console.log(textStatus);
				lmessage.hide();
				window.alert(textStatus);
			});
		});

		$("a.action-info").on("click", function(event) {
			event.preventDefault();
			$.ajax({
				url: '/mpd/service/song-info',
				type: "GET",
				data: {
					uri: $(this).data("uri"),
				},
			}).done(function(data) {
				RPI.confirmDialog(data, {
					title: "Song info",
					btnSuccess: "none",
				}).open();
			});
		});
	};

	return self;
}(MPD.library || {}, jQuery));

