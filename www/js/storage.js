const STORAGE_KEYS = {
  THEME: "app_theme",
  LANG: "ceylon_news_lang",
  CUSTOM_FONT: "use_custom_font",
  SELECTED_SOURCES: "selected_sources",
  RECENT_SEARCHES: "recent_searches",
};

function getLanguage() {
  return localStorage.getItem(STORAGE_KEYS.LANG) || "en";
}

function setLanguage(lang) {
  localStorage.setItem(STORAGE_KEYS.LANG, lang);
  if (localStorage.getItem(STORAGE_KEYS.CUSTOM_FONT) === null) {
    localStorage.setItem(STORAGE_KEYS.CUSTOM_FONT, "true");
  }
}

function isCustomFontEnabled() {
  return localStorage.getItem(STORAGE_KEYS.CUSTOM_FONT) === "true";
}

function setCustomFontEnabled(enabled) {
  localStorage.setItem(STORAGE_KEYS.CUSTOM_FONT, enabled.toString());
}

function getSavedArticles() {
  return savedArticles.getAll();
}

function saveArticle(article) {
  return savedArticles.save(article);
}

function removeArticle(articleId) {
  return savedArticles.remove(articleId);
}

function isArticleSaved(articleId) {
  return savedArticles.has(articleId);
}

function toggleSavedArticle(article) {
  return savedArticles.toggle(article);
}

function getSelectedSources() {
  const sources = localStorage.getItem(STORAGE_KEYS.SELECTED_SOURCES);
  return sources ? JSON.parse(sources) : [];
}

function setSelectedSources(sources) {
  localStorage.setItem(STORAGE_KEYS.SELECTED_SOURCES, JSON.stringify(sources));
}

function getRecentSearches() {
  const searches = localStorage.getItem(STORAGE_KEYS.RECENT_SEARCHES);
  return searches ? JSON.parse(searches) : [];
}

function saveRecentSearch(query) {
  let searches = getRecentSearches();
  searches = searches.filter((s) => s.toLowerCase() !== query.toLowerCase());
  searches.unshift(query);
  searches = searches.slice(0, 5);
  localStorage.setItem(STORAGE_KEYS.RECENT_SEARCHES, JSON.stringify(searches));
}

function clearRecentSearches() {
  localStorage.removeItem(STORAGE_KEYS.RECENT_SEARCHES);
}

function resetApp() {
  localStorage.clear();
  window.location.href = "index.html";
}
