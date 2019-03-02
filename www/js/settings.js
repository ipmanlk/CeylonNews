function settingSetDefault() {
	settings = settingsGetDefault();
	localStorage.setItem("settings", JSON.stringify(settings));
}

function settingsGetDefault() {
	return {
		"darkMode": false,
		"sinhalaFont": true,
		"newsListJustify": false,
		"postTitleJustify": false,
		"postBodyJustify": true,
		"imgLoad": true
	}
}

function settingsCheck() {
	if (!localStorage.getItem("settings")) {
		settingSetDefault();
	} else {
	    settings = JSON.parse(localStorage.getItem("settings"));
		settingsApply();
	}
}

function settingHandlersReg() {
	$('ons-switch').change(function () {
		settingSet(this.id, this.checked);
	});
}

function settingsUIupdate() {
	var ids = Object.keys(settings);
	for (var i = 0; i < ids.length; i++) {
		var id = "#" + ids[i];
		$(id)[0].checked = settings[ids[i]];
	}
}

function settingSet(setting, isSet) {	
	settings[setting] = isSet;
	localStorage.setItem("settings", JSON.stringify(settings));
	settingsApply();
}

function settingsApply() {
	// page specific settigns
	if (currentPage == "news-list") {
		settings.newsListJustify ? $(".list-item").addClass("justify") : false;
	}
	if (currentPage == "post") {
		settings.postTitleJustify ? $("#postTitle").addClass("justify") : false;
		!settings.postBodyJustify ? $("#postBody").removeClass("justify") : false;
	}

	// global settings
	!settings.sinhalaFont ? $("*").removeClass("sinhala") : false;

	if (settings.darkMode) {
		$("#theme").attr("href", "./lib/css/dark-onsen-css-components.min.css");
		if (currentPage == "post") {
			$("#postBody").removeClass("black");
			$("#postBody").addClass("white");
		}
	} else {
		$("#theme").attr("href", "./lib/css/onsen-css-components.min.css");
		if (currentPage == "post") {
			$("#postBody").removeClass("white");
			$("#postBody").addClass("black");
		}
	}
}