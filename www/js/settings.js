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
		"newsListAutoLoad": false,
		"notificationShow": false,
		"backgroundMode": false
	};
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
	// custom behaviour for some settings
	customBehaviourApply(setting, isSet);
}

function customBehaviourApply(setting, isSet) {
	// custom behaviour for background mode
	if (setting == "backgroundMode" && !isSet) {
		ons.notification.alert('App need to restart in order to apply this setting.')
			.then(function () {
				navigator.app.exitApp();
			});
	}
}

function settingsApply() {
	// page specific settigns
	if (currentPage == "news-list") {
		if (settings.newsListJustify) {
			$(".list-item").addClass("justify");
		}
	}
	if (currentPage == "post") {

		if (settings.postTitleJustify) {
			$("#postTitle").addClass("justify");
		}

		if (!settings.postBodyJustify) {
			$("#postBody").removeClass("justify");
		}

		var postBodyFontSize = settings.postBodyBigFontSize ? "21px" : "17px";
		$("#postBody").css("font-size", postBodyFontSize);
	}

	// global settings
	// sinhala font
	if (!settings.sinhalaFont) {
		$("*").removeClass("sinhala");
	}

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

	// background mode (run)
	if (settings.backgroundMode) {
		cordova.plugins.backgroundMode.enable();
		cordova.plugins.backgroundMode.overrideBackButton();
		cordova.plugins.backgroundMode.excludeFromTaskList();
	}
}

// check if arrays equal
function isArraysEqual(a, b) {
	return JSON.stringify(a) == JSON.stringify(b);
}