// handle back key
document.addEventListener("backbutton", onBackKeyDown, false);
function onBackKeyDown(e) {
  e.preventDefault();
  // handle accordingly
  switch (currentPage) {
    case "post":
      newsListShow();
      break;
    case "settings":
      newsListShow();
      break;
    default:
      exitApp();
  }
}

function exitApp() {
  ons.notification.confirm('Do you really want to close the app?') // Ask for confirmation
    .then(function (index) {
      if (index === 1) { // OK button
        navigator.app.exitApp();
      }
    });
}
