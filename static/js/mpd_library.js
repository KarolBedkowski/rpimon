
var MPD = MPD || {};

MPD.library = (function(self, $) {
	var message = null;
	var mpdControlUrl = null;
	var mpdServiceInfoUrl = null;

	self.init = function initF(mpdControlUrl_, mpdServiceInfoUrl_) {
		mpdControlUrl = mpdControlUrl_
		mpdServiceInfoUrl = mpdServiceInfoUrl_

		$('table').dataTable({
			"bAutoWidth": false,
			"bStateSave": true,
			"sPaginationType": "full_numbers",
			"iDisplayLength": 15,
			"aLengthMenu": [[15, 25, 50, 100, -1], [15, 25, 50, 100, "All"]],
			"aoColumnDefs": [{
				"aTargets": [1],
				"bSortable": false,
			}],
		});


		$("a.action").on("click", function(event) {
			event.preventDefault();
			var link = $(this);
			var p = link.data("path");
			var lmessage = new Messi('Adding...', {
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
			}).done(function(msg) {
				lmessage.hide()
			}).fail(function(jqXHR, textStatus) {
				console.log(textStatus);
				lmessage.hide()
				new Messi(textStatus, {
					title: 'Error',
					titleClass: 'anim warning',
					buttons: [{
						id: 0, label: 'Close', val: 'X'
					}]
				});
			});
		});
	};

	return self;
}(MPD.library || {}, jQuery));

