var RPI = (function(self, $) {
	"use strict";

	self.confirm = function confirmF() {
		return confirm('Are you sure?');
	};

	self.confirmDialog = function confirmDialogF(message, params) {
		var dlg = $("#dialog-confirm");
		$("#dialog-confirm .modal-body").html(message);
		$("#dialog-confirm .modal-title").html(params.title || "");
		if (params.btnCancel != "none") {
			$("#dialog-confirm #dialog-confirm-cancel")
				.html(params.btnCancel || "Close")
				.addClass(params.btnCancelClass || "btn-default");
		} else {
			$("#dialog-confirm #dialog-confirm-cancel").hide();
		}
		if (params.btnSuccess != "none") {
			$("#dialog-confirm #dialog-confirm-success")	
				.html(params.btnSuccess || "Yes")
				.addClass(params.btnSuccessClass || "btn-primary");
		} else {
			$("#dialog-confirm #dialog-confirm-success").hide();
		}
		$("#dialog-confirm-success").on("click", function(event) {
			dlg.modal("hide");
			if (params.onSuccess) {
				params.onSuccess(event);
			}
		});
		return {
			dlg: dlg,
			open: function() {
				dlg.modal('show');
				return dlg;
			},
		};
	};

	return self;
}(RPI || {}, jQuery));
