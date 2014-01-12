
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
			currentSongId = response.stat.songid;
			var aaData = [];
			var playlist = response.playlist
			var plistlen = playlist.length;
			for (idx=0; idx < plistlen; ++idx) {
				var item = playlist[idx];
				if (item != null) {
					if (item.Album == null) {
						item.Album = "";
					}
					if (item.Artist == null) {
						item.Artist = "";
					}
					if (item.Track == null) {
						item.Track = "";
					}
					if (!("Title" in item) || item.Title == null) {
						item.Title = item.file;
					}
					aaData.push(item);
				}
			}
			var playlist = {
				"iTotalDisplayRecords": parseInt(response.stat.playlistlength),
				"iTotalRecords": parseInt(response.stat.playlistlength),
				"aaData": aaData,
				"sEcho": response.echo,
			};
			fnCallback(playlist);
			message.hide();
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
		var id = $(this).closest('tr').data("songid");
		showLoadingMessage();
		$.ajax({
			url: "/mpd/song/" + id  + "/play",
			method: "PUT"
		}).done(function(result) {
			message.hide()
			if (result.Error == "") {
				$("tr.active").removeClass("active").removeClass("playlist-current-song");
				var newSongId = result.Status.songid;
				$('tr[data-songid='+newSongId+']').addClass("playlist-current-song active");
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
