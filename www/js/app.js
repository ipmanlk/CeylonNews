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

function getSavedArticles() {
  const saved = localStorage.getItem("saved_articles");
  return saved ? JSON.parse(saved) : [];
}

function saveArticle(article) {
  const saved = getSavedArticles();
  const exists = saved.find((a) => a.id === article.id);
  if (!exists) {
    saved.push(article);
    localStorage.setItem("saved_articles", JSON.stringify(saved));
  }
}

function removeArticle(articleId) {
  let saved = getSavedArticles();
  saved = saved.filter((a) => a.id !== articleId);
  localStorage.setItem("saved_articles", JSON.stringify(saved));
}

function isArticleSaved(articleId) {
  const saved = getSavedArticles();
  return saved.some((a) => a.id === articleId);
}

function toggleSavedArticle(article) {
  if (isArticleSaved(article.id)) {
    removeArticle(article.id);
    return false;
  } else {
    saveArticle(article);
    return true;
  }
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
