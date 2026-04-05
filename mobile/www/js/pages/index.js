function checkLanguageAndRedirect() {
  if (localStorage.getItem(STORAGE_KEYS.LANG)) {
    window.location.href = "home.html";
  }
}

function selectLanguage(lang) {
  setLanguage(lang);
  window.location.href = "home.html";
}

document.addEventListener("DOMContentLoaded", () => {
  checkLanguageAndRedirect();
  
  const themeToggle = document.getElementById("theme-toggle");
  if (themeToggle && localStorage.getItem("app_theme") === "dark") {
    themeToggle.checked = true;
  }
});
