const STORAGE_KEYS = {
  THEME: "app_theme",
  LANG: "ceylon_news_lang",
  CUSTOM_FONT: "use_custom_font",
  SELECTED_SOURCE_IDS: "selected_source_ids",
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

function getSelectedSourceIds() {
  // TODO: Remove this backward compatibility code once all users are on the new API version
  // This migration code can be removed after 30 days or next major release
  const oldSources = localStorage.getItem("selected_sources");
  if (oldSources) {
    localStorage.setItem(STORAGE_KEYS.SELECTED_SOURCE_IDS, oldSources);
    localStorage.removeItem("selected_sources");
    return JSON.parse(oldSources);
  }
  const sourceIds = localStorage.getItem(STORAGE_KEYS.SELECTED_SOURCE_IDS);
  return sourceIds ? JSON.parse(sourceIds) : [];
}

function setSelectedSourceIds(sourceIds) {
  localStorage.setItem(STORAGE_KEYS.SELECTED_SOURCE_IDS, JSON.stringify(sourceIds));
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

function getReadHistory() {
  return readHistory.getAll();
}

function addToReadHistory(article) {
  return readHistory.add(article);
}

function clearReadHistory() {
  return readHistory.clear();
}
