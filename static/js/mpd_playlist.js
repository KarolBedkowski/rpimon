
var MPD = MPD || {};

MPD.plist = (function(self) {
	var message = null;

	function processPlaylist(response) {
		var tbody = $("#playlist-tbody");
		var current = response.stat.songid;
		$.each(response["playlist"], function(i, item) {
			var tr = $("<tr>").attr("data-songid", item.Id).append(
				$("<td>").text(i + 1),
				$("<td>").text(item.Album),
				$("<td>").text(item.Artist),
				$("<td>").text(item.Track));
			if ("Title" in item) {
				tr.append($("<td>").text(item.Title));
			} else {
				tr.append($("<td>").text(item.file));
			}
			tr.append(
				$("<td>").html('<a href="#" class="play-song-action"><span class="glyphicon glyphicon-play" title="Play"></span></a>'+
					'&nbsp;<a href="#" class="remove-song-action"><span class="glyphicon glyphicon-remove" title="Remove"></span></a>')
			);
			if (item.Id == current) {
				tr.addClass("playlist-current-song active");
			}
			tbody.append(tr);
		});
		$('table').tablesorter();
		$("a.play-song-action").on("click", self.playSong);
		$("a.remove-song-action").on("click", self.removeSong);
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

	self.playSong = function playSongF(event) {
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

	self.removeSong = function removeSongF(event) {
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
				tr.slideUp(100, function() {
					$("tr.active").removeClass("active").removeClass("playlist-current-song");
					var newSongId = result.Status.songid;
					$('tr[data-songid='+newSongId+']').addClass("playlist-current-song active");
					tr.remove();
				});
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
})(MPD.plist || {});
