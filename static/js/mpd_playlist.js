
var MPD = MPD || {};

MPD.plist = (function(self, $) {
	var message = null;
	var table = null;

	function processPlaylist(response) {
		var tbody = $("#playlist-tbody");
		var current = response.stat.songid;
		var playlist = response.playlist;
		var resLen = playlist.length;
		for (var i=0; i<resLen; i++) {
			var item = playlist[i];
			var tr = $("<tr>").attr("data-songid", item.Id).append(
				$("<td>").text(item.Album),
				$("<td>").text(item.Artist),
				$("<td>").text(item.Track));
			if ("Title" in item) {
				tr.append($("<td>").text(item.Title));
			} else {
				tr.append($("<td>").text(item.file));
			}
			tr.append(
				$("<td>").html('<a href="#" class="play-song-action"><span class="glyphicon glyphicon-play" title="Play"></span></a>&nbsp;<a href="#" class="remove-song-action"><span class="glyphicon glyphicon-remove" title="Remove"></span></a>')
			);
			if (item.Id == current) {
				tr.addClass("playlist-current-song active");
			}
			tbody.append(tr);
		};
		table = $('table').dataTable({
			"bAutoWidth": false,
			"bStateSave": true,
			"iDisplayLength": 15,
			"aLengthMenu": [[15, 25, 50, 100, -1], [15, 25, 50, 100, "All"]],
		});
		$("a.play-song-action").on("click", playSong);
		$("a.remove-song-action").on("click", removeSong);
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
		$.ajax({
			url: "/mpd/playlist/serv/info",
		}).done(function(response) {
			$("#playlist-tbody").text("");
			if (response.error == null) {
				processPlaylist(response);
				message.hide()
			} else {
				message.hide()
				showError(response.error);
			}
		}).fail(function(jqXHR, message) {
			showError(message);
		});
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
				table.fnDeleteRow(tr[0], function() {
					$("tr.active").removeClass("active").removeClass("playlist-current-song");
					var newSongId = result.Status.songid;
					$('tr[data-songid='+newSongId+']').addClass("playlist-current-song active");
				}, true);
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
