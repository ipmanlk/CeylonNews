let searchTimeout;
const currentLanguage = getLanguage();

let searchInput, searchClear, recentSearchesEl, resultsContainer;
let emptyState, searchResultsList, searchLoading, noResults, recentSearchesList;

function loadRecentSearchesUI() {
  const searches = getRecentSearches();
  recentSearchesList.innerHTML = "";

  if (searches.length === 0) {
    recentSearchesEl.classList.add("hidden");
    emptyState.classList.remove("hidden");
    return;
  }

  recentSearchesEl.classList.remove("hidden");
  emptyState.classList.add("hidden");

  searches.forEach(search => {
    const card = document.createElement("div");
    card.className = "card card-padded mb-4";
    card.innerHTML = `<div class="card-compact clickable" data-query="${search}">
      <i class="ph ph-clock-counter-clockwise text-muted icon-md"></i>
      <span class="flex-1 ml-3 font-medium">${search}</span>
      <i class="ph ph-x text-muted icon-sm"></i>
    </div>`;
    recentSearchesList.appendChild(card);
  });
}

function performSearch(query) {
  if (!query.trim()) return;

  clearTimeout(searchTimeout);
  searchTimeout = setTimeout(() => {
    searchLoading.classList.remove("hidden");
    recentSearchesEl.classList.add("hidden");
    resultsContainer.classList.remove("hidden");
    emptyState.classList.add("hidden");
    noResults.classList.add("hidden");
    searchResultsList.innerHTML = "";

    searchArticles({
      query: query,
      languages: currentLanguage,
      limit: 20
    })
      .then(response => {
        searchLoading.classList.add("hidden");
        if (response.data && response.data.length > 0) {
          renderSearchResults(response.data);
          saveRecentSearch(query);
        } else {
          noResults.classList.remove("hidden");
        }
      })
      .catch(error => {
        console.error("Search failed:", error);
        searchLoading.classList.add("hidden");
        noResults.classList.remove("hidden");
      });
  }, 300);
}

function renderSearchResults(articles) {
  const fontClass = getFontClass();
  articles.forEach(article => {
    const card = document.createElement("div");
    card.className = "card card-standard mb-3";
    const imgHTML = article.image_url ? `<img src="${article.image_url}" class="card-image" alt="${article.title}" />` : "";
    card.innerHTML = `${imgHTML}
      <div class="card-content">
        <h3 class="card-title ${fontClass}">${article.title}</h3>
        <div class="card-footer">
          <span class="badge">${article.source_name}</span>
          <span class="text-sm text-muted">${formatDate(article.published_at)}</span>
        </div>
      </div>`;
    card.onclick = () => window.location.href = `article.html?id=${article.id}`;
    searchResultsList.appendChild(card);
  });
}

function handleSearchInput() {
  const value = searchInput.value.trim();
  if (value) {
    searchClear.classList.remove("hidden");
  } else {
    searchClear.classList.add("hidden");
  }

  if (value.length > 0) {
    performSearch(value);
  } else {
    recentSearchesEl.classList.remove("hidden");
    resultsContainer.classList.add("hidden");
    emptyState.classList.add("hidden");
  }
}

function handleSearchClear() {
  searchInput.value = "";
  searchClear.classList.add("hidden");
  recentSearchesEl.classList.remove("hidden");
  resultsContainer.classList.add("hidden");
  emptyState.classList.add("hidden");
  searchInput.focus();
}

function handleRecentSearchClick(e) {
  const card = e.target.closest(".card-compact");
  if (card && !e.target.classList.contains("ph-x")) {
    searchInput.value = card.dataset.query;
    searchInput.dispatchEvent(new Event("input"));
  } else if (e.target.classList.contains("ph-x")) {
    const query = e.target.closest(".card-compact").dataset.query;
    let searches = getRecentSearches().filter(s => s !== query);
    localStorage.setItem(STORAGE_KEYS.RECENT_SEARCHES, JSON.stringify(searches));
    loadRecentSearchesUI();
  }
}

document.addEventListener("DOMContentLoaded", () => {
  searchInput = document.getElementById("search-input");
  searchClear = document.getElementById("search-clear");
  recentSearchesEl = document.getElementById("recent-searches");
  resultsContainer = document.getElementById("results-container");
  emptyState = document.getElementById("empty-state");
  searchResultsList = document.getElementById("search-results-list");
  searchLoading = document.getElementById("search-loading");
  noResults = document.getElementById("no-results");
  recentSearchesList = document.getElementById("recent-searches-list");

  loadRecentSearchesUI();

  searchInput.addEventListener("input", handleSearchInput);
  searchClear.addEventListener("click", handleSearchClear);
  document.getElementById("clear-all").addEventListener("click", () => {
    clearRecentSearches();
    loadRecentSearchesUI();
  });
  recentSearchesList.addEventListener("click", handleRecentSearchClick);
});
