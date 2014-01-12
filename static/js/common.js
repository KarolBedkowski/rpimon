var RPI = (function(self) {

	self.confirm = function confirmF() {
		return confirm('Are you sure?');
	};

	return self;
}(RPI || {}));
