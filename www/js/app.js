// Theme Manager
function initTheme() {
  const savedTheme = localStorage.getItem("app_theme") || "light";
  document.documentElement.setAttribute("data-theme", savedTheme);
}

function toggleTheme() {
  const current = document.documentElement.getAttribute("data-theme");
  const next = current === "light" ? "dark" : "light";
  document.documentElement.setAttribute("data-theme", next);
  localStorage.setItem("app_theme", next);
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
  if (localStorage.getItem("use_custom_font") === null) {
    localStorage.setItem("use_custom_font", "true");
  }
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

function applyCustomFontToContent() {
  const lang = localStorage.getItem("ceylon_news_lang") || "en";
  const useCustomFont = localStorage.getItem("use_custom_font") === "true";
  const fontClass = useCustomFont ? `${lang}-font` : "";

  const titles = document.querySelectorAll('.card-title, .article-title');
  const bodies = document.querySelectorAll('.article-body, .card-content');

  titles.forEach(el => {
    el.classList.remove('en-font', 'si-font', 'ta-font');
    if (fontClass) el.classList.add(fontClass);
  });

  bodies.forEach(el => {
    el.classList.remove('en-font', 'si-font', 'ta-font');
    if (fontClass) el.classList.add(fontClass);
  });
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
