function settingsLoadDefault() {
	var settings = {
		"darkMode": false,
		"sinhalaFont": true,
		"newsListJustify": true,
		"postTitleJustify": true,
		"postBodyJustify": true,
		"imgLoad": true
	}

	localStorage.setItem("settings", JSON.stringify(settings));
}

function settingsCheck() {
	if (!localStorage.getItem("settings")) {
		settingsLoadDefault();
	}
}

function settingsHandlersSet() {
	$('ons-switch').change(function () {
		settingsSet(this.id, this.checked);
	});
}

function settingsSet(setting, isSet) {
	var settings = JSON.parse(localStorage.getItem("settings"));
	settings[setting] = isSet;
	localStorage.setItem("settings", JSON.stringify(settings));
}
