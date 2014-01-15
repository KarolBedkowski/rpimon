
var MPD = MPD || {};

MPD.plists = (function(self, $) {
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
			"bFilter": false,
			"sPaginationType": "bootstrap",		
			"iDisplayLength": 15,
			"bLengthChange": false,
			"aoColumnDefs": [{
				"aTargets": [1],
				"bSortable": false,
			}],
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
