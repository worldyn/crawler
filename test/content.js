var system = require('system');
var args = system.args;
var webPage = require('webpage');
var page = webPage.create();


page.onLoadFinished = function (status) {
	setTimeout(function() {
		var content = page.content;
		console.log('Content: ' + content);
		phantom.exit();
	}, 1000);
};

page.open(args[1]);
