const onBackKeyDown = (e) => {
	e.preventDefault();
	switch (vars.currentPage) {
		case "news-post":
			showNewsList();
			break;
		case "news-list":
			exitApp();
			break;
		case "settings":
			showNewsList();
			break;
		default:
			exitApp();
	}
}

const exitApp = () => {
	ons.notification.confirm('Do you really want to close the app?')
		.then((index) => {
			if (index === 1) { // OK button
				navigator.app.exitApp();
			}
		});
}
