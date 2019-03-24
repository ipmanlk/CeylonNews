function updateCheck() {
    var currentVersion = "5.7.0";
    requestSend(
        "get",
        {
            action: "version",
            lang: lang
        },
        function (info) {
            if (info.version > currentVersion) {
                content.load("./views/update.html").then(function () {
                    $("#lblCurrentVersion").text(currentVersion);
                    $("#lblNewVersion").text(info.version);
                });
            }
        }
    );
}

function updateRun(repo) {
    if (repo == "PlayStore") {
        window.open('https://play.google.com/store/apps/details?id=xyz.navinda.ceylonnews&hl=en', '_system', 'location=yes');
    } else {
        window.open('https://github.com/ipmanlk/CeylonNews/releases/latest', '_system', 'location=yes');
    }
}