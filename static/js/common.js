var RPI = (function(self, $) {
	"use strict";

	var alertTimers = {};

	self.confirm = function confirmF() {
		return confirm('Are you sure?');
	};

	self.confirmDialog = function confirmDialogF(message, params) {
		var dlg = $("#dialog-confirm");
		params = params || {};
		if (params.replace) {
			$("#dialog-confirm .modal-body").replaceWith(message);
		} else {
			$("#dialog-confirm .modal-body").html(message);
		}
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
			$("#dialog-confirm-success").off("click").on("click", function(event) {
				dlg.modal("hide");
				if (params.onSuccess) {
					params.onSuccess(event);
				}
			});
		} else {
			$("#dialog-confirm #dialog-confirm-success").hide();
		}
		return {
			dlg: dlg,
			open: function() {
				dlg.modal('show');
				return dlg;
			}
		};
	};

	self.alert = function alert(message, params) {
		var dlg = $("#dialog-alert");
		params = params || {};
		$("#dialog-alert #dialog-alert-head").html(message);
		$("#dialog-alert #dialog-alert-text").html(params.text || "");
		$("#dialog-alert .modal-title").html(params.title || "");
		if (params.btnCancel != "none") {
			$("#dialog-alert #dialog-alert-cancel")
				.html(params.btnCancel || "Close")
				.addClass(params.btnCancelClass || "btn-default");
		} else {
			$("#dialog-alert #dialog-alert-cancel").hide();
		}
		if (params.btnSuccess) {
			$("#dialog-alert #dialog-alert-success")
				.html(params.btnSuccess || "Yes")
				.addClass(params.btnSuccessClass || "btn-primary");
			$("#dialog-alert-success").off("click").on("click", function(event) {
				dlg.modal("hide");
				if (params.onSuccess) {
					params.onSuccess(event);
				}
			});
		} else {
			$("#dialog-alert #dialog-alert-success").hide();
		}
		return {
			dlg: dlg,
			open: function() {
				dlg.modal('show');
				return dlg;
			}
		};
	};

	self.showLoadingMsg = function showLoadingMsgF() {
		var dwidth = $(document).width(),
			dheight = $(document).height(),
			left = (dwidth - $("#loading-box .loading-wrapper").width()) / 2 +  $(window).scrollLeft();
		$("#loading-box .loading-wrapper").css("left", left + "px");
		$("#loading-box").css("z-index", 9900).css("height", dheight + "px").fadeTo(200, 0.3);
	};

	self.hideLoadingMsg = function hideLoadingMsgF() {
		$("#loading-box").fadeOut(300, function() {
			$("#loading-box").css("z-index", -990);
		});
	};

	self.showFlash = function showFlashF(kind, message, timeout) {
		if (!message) {
			return;
		}
		var div = $("#flash-" + kind),
			ul = $("ul", div),
			top = $(window).scrollTop();
		top = top + (top > 50 ? 20 : 70 - top);
		$("<li>").html(message).appendTo(ul);
		$("#flash-container").css("top", top + "px");
		$("#flash-" + kind).fadeIn(100, function() {
			if (timeout) {
				self.hideFlash(div, timeout);
			}
		});
	};

	self.hideFlash = function hideFlashF(div, timeout) {
		if (!div) {
			return;
		}
		var divid = div.prop("id"),
			timer = alertTimers[divid];
		if (timer) {
			window.clearTimeout(timer);
		}
		alertTimers[divid] = window.setTimeout(function() {
			if (div) {
				div.fadeOut(150, function() {
					$("ul", div).html("");
				});
			}
		}, (timeout || 5) * 1000);
	}

	self.hideLoadingMsg();

	$("#flash-container div.alert:visible").each(function(index, elem) {
		self.hideFlash($(elem), 2);
	});

	setTimeout(self.hideFlash, 2000);

	return self;
}(RPI || {}, jQuery));
