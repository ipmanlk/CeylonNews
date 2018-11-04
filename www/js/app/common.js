document.addEventListener("backbutton", onBackKeyDown, false);

function onBackKeyDown(e) {
  // prevent back button exit
  e.preventDefault();

  // handle accordingly
  switch (currentPage) {
    case "post":
    goToNewsList();
    break;
    default:
    ons.notification.confirm('Do you really want to close the app?') // Ask for confirmation
    .then(function(index) {
      if (index === 1) { // OK button
        exitApp(); // Close the app
      }
    });
  }
}

function exitApp() {
  navigator.app.exitApp();
}