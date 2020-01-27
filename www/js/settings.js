const saveSettings = () => {
    localStorage.setItem("settings", JSON.stringify(settings));
    localStorage.setItem("data", JSON.stringify(data));
}

const loadSettings = () => {
    if (localStorage.getItem("settings") && localStorage.getItem("data")) {
        window.settings = JSON.parse(localStorage.getItem("settings"));
        window.data = JSON.parse(localStorage.getItem("data"));
    } else {
        setDefaultSettings();
    }
}

const setDefaultSettings = () => {
    window.settings = getDefaultSettings();

    window.data = {

    };

    saveSettings();
}

const getDefaultSettings = () => {
    return {
        "st-darkmode": false,
        "st-sinhalafont": true,
        "st-news-list-justify": false,
        "st-news-post-title-justify": false,
        "st-news-post-body-justify": true,
        "st-news-post-body-lgfont": false,
        "st-news-list-autoload": true
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
    // apply theme
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

    if (page == "news-list") {
        if (settings["st-news-list-justify"]) {
            $(".list-item").addClass("text-justify");
        }

        if (settings["st-sinhalafont"]) {
            $(".list-item__title").addClass("sinhala-font");
        }
    }

    if (page == "news-post") {
        if (settings["st-news-post-title-justify"]) {
            $("#lbl-news-post-title").addClass("text-justify");
        }

        if (settings["st-news-post-body-justify"]) {
            $("#lbl-news-post-body").addClass("text-justify");
        }

        if (settings["st-sinhalafont"]) {
            $("#lbl-news-post-title").addClass("sinhala-font");
            $("#lbl-news-post-body").addClass("sinhala-font");
        }

        const postBodyFontSize = settings["st-news-post-body-lgfont"] ? "21px" : "17px";
        $("#lbl-news-post-body").css("font-size", postBodyFontSize);
    }
}

const changeSetting = (id, isChecked) => {
    settings[id] = isChecked;
    saveSettings();
    applySettings();
}