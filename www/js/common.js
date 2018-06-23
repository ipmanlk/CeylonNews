//loading spinner for events takes time
function showLoading(val) {
    if (val) {
        $.mobile.loading("show", {
            text: "Loading",
            textVisible: true,
            theme: "b"
        });
    } else {
        $.mobile.loading("hide");
    }
}

function goHome() {
    showLoading(true);
    location.replace('index.html')
}
