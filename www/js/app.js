// Theme Manager
function initTheme() {
  const savedTheme = localStorage.getItem("theme") || "light";
  document.documentElement.setAttribute("data-theme", savedTheme);
}

function toggleTheme() {
  const current = document.documentElement.getAttribute("data-theme");
  const next = current === "light" ? "dark" : "light";
  document.documentElement.setAttribute("data-theme", next);
  localStorage.setItem("theme", next);
}

// Language Manager
function checkLanguage() {
  // If we are on index.html (Language selection), and language is already set, go to home
  if (
    window.location.pathname.includes("index.html") ||
    window.location.pathname === "/"
  ) {
    if (localStorage.getItem("ceylon_news_lang")) {
      window.location.href = "home.html";
    }
  }
}

function setLanguage(lang) {
  localStorage.setItem("ceylon_news_lang", lang);
  window.location.href = "home.html";
}

// Initialize
document.addEventListener("DOMContentLoaded", () => {
  initTheme();
  // Only run check on index page to prevent loops
  if (document.getElementById("lang-select-page")) {
    checkLanguage();
  }
});

// Cordova deviceready event handler
document.addEventListener(
  "deviceready",
  function () {
    console.log(
      "Cordova is now initialized. Platform: " +
        cordova.platformId +
        ", Version: " +
        cordova.version,
    );
  },
  false,
);
