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
		"postBodyBigFontSize": false,
		"imgLoad": true,
		"newsListAutoLoad": false
	}
}

function settingsCheck() {
	if (!localStorage.getItem("settings")) {
		settingSetDefault();
	} else {
		settings = JSON.parse(localStorage.getItem("settings"));
		// fix for side menu
		newsListAutoLoad = settings.newsListAutoLoad;
		// check for setting conflicts between app versions
		var keys = Object.keys(settings);
		var defaultKeys = Object.keys(settingsGetDefault());
		if (!isArraysEqual(keys, defaultKeys)) {
			settingSetDefault();
			toastToggle("Settings have been reset to recover from a possible conflict.", 4000);
		}
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

		var postBodyFontSize = settings.postBodyBigFontSize ? "21px" : "17px";
		$("#postBody").css("font-size", postBodyFontSize);
	}

	// global settings
	// sinhala font
	!settings.sinhalaFont ? $("*").removeClass("sinhala") : false;

	// handle dark mode
	if (settings.darkMode) {
		if ($("#theme").attr("href") !== "./lib/css/dark-onsen-css-components.min.css") {
			$("#theme").attr("href", "./lib/css/dark-onsen-css-components.min.css");
		}
		if (currentPage == "post") {
			$("#postBody").removeClass("black");
			$("#postBody").addClass("white");
		}
	} else {
		if ($("#theme").attr("href") !== "./lib/css/onsen-css-components.min.css") {
			$("#theme").attr("href", "./lib/css/onsen-css-components.min.css");
		}
		if (currentPage == "post") {
			$("#postBody").removeClass("white");
			$("#postBody").addClass("black");
		}
	}
}

// check if arrays equal
function isArraysEqual(a, b) {
	return JSON.stringify(a) == JSON.stringify(b);
}