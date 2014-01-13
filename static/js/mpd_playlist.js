
var MPD = MPD || {};

MPD.plist = (function(self, $) {
	var message = null;
	var table = null;
	var currentSongId = -1;

	function processServerData(sSource, aoData, fnCallback) {
		$.ajax({
			url: sSource,
			data: aoData || {},
		}).done(function(response) {
			message.hide();
			if (response.error) {
				showError(message);
			}
			else {
				currentSongId = response.stat.songid;
				fnCallback(response);
			}
		}).fail(function(jqXHR, message) {
			message.hide()
			showError(message);
		});
	};

	function showLoadingMessage() {
		message = new Messi('Loading...', {
			closeButton: false,
			modal: true,
			width: 'auto',
		});
	};

	function showError(message) {
		console.log(message);
		message.hide()
		new Messi(message, {
			title: 'Error',
			titleClass: 'anim warning',
			buttons: [{
				id: 0, label: 'Close', val: 'X'
			}]
		});
	};

	self.refresh = function refreshF() {
		table = $('table').dataTable({
			"bAutoWidth": false,
			"bStateSave": true,
			"bSort": false,
			"iDisplayLength": 15,
			"aLengthMenu": [[15, 25, 50, 100, -1], [15, 25, 50, 100, "All"]],
			"sPaginationType": "full_numbers",
			"bProcessing": true,
			"bServerSide": true,
			"sAjaxSource": "/mpd/playlist/serv/info",
			"fnServerData": processServerData,
			"aoColumns": [
				{"mData": "Album"},
				{"mData": "Artist"},
				{"mData": "Track"},
				{"mData": "Title"},
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
				$(row).data("songid", aData.Id);
				if (aData.Id == currentSongId) {
					// mark current song
					$(row).addClass("playlist-current-song active");
				}
			},
			"fnDrawCallback": function( oSettings ) {
				$("a.play-song-action").on("click", playSong);
				$("a.remove-song-action").on("click", removeSong);
			},
		});
		message.hide();
		return
	};

	function playSong(event) {
		event.preventDefault()
		var tr = $(this).closest('tr')
		var id = tr.data("songid");
		showLoadingMessage();
		$.ajax({
			url: "/mpd/song/" + id  + "/play",
			method: "PUT"
		}).done(function(result) {
			message.hide()
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
			} else {
				showError(result.error);
			}
		}).fail(function(jqXHR, message) {
			message.hide()
			showError(message);
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
			message.hide()
			if (result.Error == "") {
				// redraw table on success
				table.fnDraw();
			} else {
				showError(result.error);
			}
		}).fail(function(jqXHR, message) {
			message.hide()
			showError(message);
		});
	};

	self.init = function initF() {
		showLoadingMessage();
		self.refresh();
	};

	return self;
}(MPD.plist || {}, jQuery));
