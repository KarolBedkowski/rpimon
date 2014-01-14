
var MPD = MPD || {};

MPD.plist = (function(self, $) {
	var msg_loading = null;
	var table = null;
	var currentSongId = -1;

	function processServerData(sSource, aoData, fnCallback) {
		$.ajax({
			url: sSource,
			data: aoData || {},
		}).done(function(response) {
			hideLoadingMessage();
			if (response.error) {
				showError(response.error);
			}
			else {
				currentSongId = response.stat.songid;
				fnCallback(response);
			}
		}).fail(function(jqXHR, result) {
			showError(result);
		});
	};

	function showLoadingMessage() {
		if (msg_loading) {
			return;
		}
		msg_loading = new Messi('Loading...', {
			closeButton: false,
			modal: true,
			width: 'auto',
		});
	};

	function hideLoadingMessage() {
		if (msg_loading) {
			msg_loading.hide();
			msg_loading = null;
		}
	};

	function showError(errormsg) {
		hideLoadingMessage();
		console.log(errormsg);
		new Messi(errormsg, {
			title: 'Error',
			titleClass: 'anim warning',
			buttons: [
				{id: 1, label: "Reload", val: "R", class: 'btn-success'},
			],
			callback: function(val) {
				if (val == "R") {
					location.reload();
				}
			},	
		});
	};

	self.refresh = function refreshF() {
		table = $('table').dataTable({
			"bAutoWidth": false,
			"bStateSave": true,
			"bSort": false,
			"iDisplayLength": 15,
			"aLengthMenu": [[15, 25, 50, 100, -1], [15, 25, 50, 100, "All"]],
			"sPaginationType": "bootstrap",		
			"bProcessing": true,
			"bServerSide": true,
			"sAjaxSource": "/mpd/playlist/serv/info",
			"fnServerData": processServerData,
			"aoColumns": [
				{"sTitle": "Album"},
				{"sTitle": "Artist"},
				{"sTitle": "Track"},
				{"sTitle": "Title"},
				{"mData": null},
			],
			"aoColumnDefs": [{
				"aTargets": [4],
				"mData": null,
				"bSortable": false,
				"mRender": function(data, type, full) {
					return '<a href="#" class="play-song-action"><span class="glyphicon glyphicon-play" title="Play"></span></a>&nbsp;<a href="#" class="remove-song-action"><span class="glyphicon glyphicon-remove" title="Remove"></span></a>';
				},
			}],
			"fnRowCallback": function(row, aData, iDisplayIndex, iDisplayIndexFull) {
				$(row).data("songid", aData[4]);
				if (aData[4] == currentSongId) {
					// mark current song
					$(row).addClass("playlist-current-song active");
				}
			},
			"fnDrawCallback": function( oSettings ) {
				$("a.play-song-action").on("click", playSong);
				$("a.remove-song-action").on("click", removeSong);
				$("tr").on("click",  playSong);
			},
			"sDom": "t"+
				"<'row'<'col-xs-12 col-sm-6'i><'col-xs-12 col-sm-6'p>>" + 
				"<'row'<'col-xs-12 col-sm-6'l><'col-xs-12 col-sm-6'f>r>"
		});
		return
	};

	function playSong(event) {
		event.preventDefault()
		var tr = $(this);
		var id = tr.data("songid");
		if (!id) {
			tr = tr.closest('tr');
			id = tr.data("songid");
		}
		showLoadingMessage();
		$.ajax({
			url: "/mpd/song/" + id  + "/play",
			method: "PUT"
		}).done(function(result) {
			if (result.Error == "") {
				$("tr.active").removeClass("active").removeClass("playlist-current-song");
				currentSongId = result.Status.songid;
				if (currentSongId != id) {
					// $('tr[data-songid=... not work on dynamic created data
					tr = $('tr').filter(function() {
					    return $(this).data('songid') == currentSongId;
					}).first();
				}
				tr.addClass("playlist-current-song active");
				hideLoadingMessage()
			} else {
				showError(result.error);
			}
		}).fail(function(jqXHR, result) {
			showError(result);
		});
	};

	function removeSong(event) {
		event.preventDefault()
		if (!RPI.confirm()) {
			return
		}
		var tr = $(this).closest('tr')
		var id = tr.data("songid");
		showLoadingMessage();
		$.ajax({
			url: "/mpd/song/" + id  + "/remove",
			method: "PUT"
		}).done(function(result) {
			hideLoadingMessage()
			if (result.Error == "") {
				// redraw table on success
				table.fnDraw();
			} else {
				showError(result.error);
			}
		}).fail(function(jqXHR, result) {
			showError(result);
		});
	};

	self.init = function initF() {
		showLoadingMessage();
		self.refresh();
	};

	return self;
}(MPD.plist || {}, jQuery));
