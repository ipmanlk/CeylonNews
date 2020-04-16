const saveSettings = () => {
    localStorage.setItem("settings", JSON.stringify(settings));
    localStorage.setItem("data", JSON.stringify(data));
}

const loadSettings = () => {
    if (localStorage.getItem("settings") && localStorage.getItem("data")) {
        // parse current settings from local storage
        const currentSettings = JSON.parse(localStorage.getItem("settings"));

        // check for setting conflicts
        const defaultSettings = getDefaultSettings();
        if (JSON.stringify(Object.keys(currentSettings)) !== JSON.stringify(Object.keys(defaultSettings))) {
          setDefaultSettings();
          showTimedToast("Settings has been reset to avoid possible conflicts.", 3000);
          return;
        }

        // load settings
        window.settings = currentSettings;
        window.data = JSON.parse(localStorage.getItem("data"));
    } else {
        setDefaultSettings();
    }
}

const setDefaultSettings = () => {
    window.settings = getDefaultSettings();
    window.data = {};
    saveSettings();
}

const getDefaultSettings = () => {
    return {
        "st-darkmode": false,
        "st-news-list-justify": false,
        "st-news-post-title-justify": false,
        "st-news-post-body-justify": true,
        "st-news-post-body-lgfont": false,
        "st-news-list-autoload": true,
        "st-news-list-card-ui": false,
        "st-searchbar": false
    };
}

const showSettingsPage = () => {
    fn.loadPage("./views/settings.html").then(() => {
        updateSettingsUI();

        // event listener to handle setting switching
        $('ons-switch').change((e) => {
            changeSetting(e.target.id, e.target.checked);
        });

        vars.currentPage = "settings";
    });
}

const updateSettingsUI = () => {
    Object.keys(settings).forEach(setting => {
        if (settings[setting]) {
            $(`#${setting}`).attr('checked', true);
        }
    });
}

const applySettings = () => {
    const page = vars.currentPage;

    // theme: start
    if (settings["st-darkmode"]) {
        const darkCssPath = "./lib/css/dark-onsen-css-components.min.css";
        if ($("#app-theme").attr("href") !== darkCssPath) {
            $("#app-theme").attr("href", darkCssPath);
        }
        if (page == "news-post") {
            $("#lbl-news-post-body").removeClass("text-black");
            $("#lbl-news-post-body").addClass("text-white");
        }
    } else {
        const lightCssPath = "./lib/css/onsen-css-components.min.css";
        if ($("#app-theme").attr("href") !== lightCssPath) {
            $("#app-theme").attr("href", lightCssPath);
        }
        if (page == "news-post") {
            $("#lbl-news-post-body").removeClass("text-white");
            $("#lbl-news-post-body").addClass("text-black");
        }
    }
    // theme: end

    // news-list: start
    if (page == "news-list") {
        if (settings["st-news-list-justify"]) {
            if (settings["st-news-list-card-ui"]) {
                $(".news-list-card").addClass("text-justify");
            } else {
                $(".list-item").addClass("text-justify");
            }
        }


        // change font: start
        if (settings["st-news-list-card-ui"]) {
            $(".news-list-card-title").addClass(`${data.lang}-font`);
        } else {
            $(".list-item__title").addClass(`${data.lang}-font`);
        }

        if (settings["st-news-list-card-ui"]) {
            $(".news-list-card-title").addClass(`${data.lang}-font`);
        } else {
            $(".list-item__title").addClass(`${data.lang}-font`);
        }
        // change font: end


        if (settings["st-searchbar"]) {
            $("#txtNewsSearch").show();
        } else {
            $("#txtNewsSearch").hide();
        }
    }   
    // news-list: end

    // news-post: start
    if (page == "news-post") {
        if (settings["st-news-post-title-justify"]) {
            $("#lbl-news-post-title").addClass("text-justify");
        }

        if (settings["st-news-post-body-justify"]) {
            $("#lbl-news-post-body").addClass("text-justify");
        }

        // change font: start
        $("#lbl-news-post-title").addClass(`${data.lang}-font`);
        $("#lbl-news-post-body").addClass(`${data.lang}-font`);
        // change font: end

        const postBodyFontSize = settings["st-news-post-body-lgfont"] ? "21px" : "17px";
        $("#lbl-news-post-body").css("font-size", postBodyFontSize);
    }
    // news-post: end
}

const changeSetting = (id, isChecked) => {
    if (settings[id] !== undefined) {
        settings[id] = isChecked;
        saveSettings();
        applySettings();
    }
}

const resetSettings = () => {
    ons.notification.confirm("Do you really want to reset settings?.")
        .then((index) => {
            if (index === 1) {
                localStorage.removeItem("settings");
                localStorage.removeItem("data");
                localStorage.removeItem("settings");
                ons.notification.alert("Settings has been reset!. Please restart the app.");
            }
        });
}