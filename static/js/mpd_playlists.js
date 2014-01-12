
var MPD = MPD || {};

MPD.plists = (function(self) {
	var message = null;

	self.action = function(event) {
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

	return self;
})(MPD.plists || {});
