const checkForUpdates = () => {
    //TODO: Update app version with each new release
    const appVersion = "v6.5.0";

    // Get latest version from Github releases and compare with app version
    const api = "https://api.github.com/repos/ipmanlk/CeylonNews/releases/latest";

    $.getJSON(api, (res) => {
        const githubVersion = res.tag_name;
        if (appVersion < githubVersion) {
            // show update window
            fn.loadPage("./views/update.html").then(() => {
                $("#lblCurrentVersion").text(appVersion);
                $("#lblNewVersion").text(githubVersion);
            });
        }
    }).fail((e) => {
        console.log(e);
    });
}

const runUpdate = (repo) => {
    if (repo == "PlayStore") {
        window.open('https://play.google.com/store/apps/details?id=xyz.navinda.ceylonnews&hl=en', '_system', 'location=yes');
    } else {
        window.open('https://github.com/ipmanlk/CeylonNews/releases/latest', '_system', 'location=yes');
    }
} 