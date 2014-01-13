
var MPD = MPD || {};

MPD.plists = (function(self, $) {
	var message = null;

	action = function(event) {
		event.preventDefault();
		var message = new Messi('Loading...', {
			closeButton: false,
			modal: true,
			width: 'auto',
		});
		$.ajax({
			url: this.href,
			type: "PUT",
		}).done(function(msg) {
			message.hide()
		}).fail(function(jqXHR, textStatus) {
			message.hide()
			new Messi(textStatus, {
				title: 'Error',
				titleClass: 'anim warning',
				buttons: [{
					id: 0, label: 'Close', val: 'X'
				}]
			});
		});
	};

	self.init = function initF() {
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
		$('a.action-confirm').on("click", function() {
			return RPI.confirm();
		});
		$('a.ajax-action').on("click", action);
	};

	return self;
}(MPD.plists || {}, jQuery));
