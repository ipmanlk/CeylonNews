const homeState = {
  isLoading: false,
  currentPage: 0,
  perPage: 10,
  hasMore: true,
  currentSourceIds: getSelectedSourceIds(),
  currentLanguage: getLanguage(),
  moreStoriesAdded: false,
  articlesData: [],
  scrollPosition: 0,
  sourceMap: new Map()
};

const SESSION_STORAGE_KEY = 'home_page_state';

function loadArticles(loadMore) {
  if (homeState.isLoading) return;

  homeState.isLoading = true;
  const loadingEl = document.getElementById("loading-more");
  const endMessageEl = document.getElementById("end-message");

  if (!loadMore) {
    homeState.currentPage = 0;
    homeState.hasMore = true;
    homeState.moreStoriesAdded = false;
  } else if (loadingEl) {
    loadingEl.classList.remove("hidden");
  }

  const params = {
    limit: homeState.perPage,
    offset: homeState.currentPage * homeState.perPage,
    languages: homeState.currentLanguage
  };

  if (homeState.currentSourceIds.length > 0) {
    params.sourceIds = homeState.currentSourceIds;
  }

  getArticles(params)
    .then(response => {
      if (loadMore && loadingEl) loadingEl.classList.add("hidden");

      if (!loadMore) {
        const skeleton = document.getElementById("loading-skeleton");
        const content = document.getElementById("news-content");
        if (skeleton) skeleton.classList.add("hidden");
        if (content) {
          content.classList.remove("hidden");
          content.classList.add("animate-fade-in");
          content.innerHTML = "";
        }
      }

      if (response.data && response.data.length > 0) {
        renderArticles(response.data, loadMore);
        homeState.currentPage++;
        homeState.hasMore = response.has_next;
      } else {
        homeState.hasMore = false;
      }

      if (!homeState.hasMore && endMessageEl) endMessageEl.classList.remove("hidden");
    })
    .catch(error => {
      console.error("Failed to load articles:", error);
      if (!loadMore) {
        const skeleton = document.getElementById("loading-skeleton");
        const content = document.getElementById("news-content");
        if (skeleton) skeleton.classList.add("hidden");
        if (content) content.classList.remove("hidden");
      }
      if (loadingEl) loadingEl.classList.add("hidden");
    })
    .finally(() => {
      homeState.isLoading = false;
    });
}

function saveHomeState() {
  const state = {
    articles: homeState.articlesData,
    scrollPosition: window.scrollY,
    currentPage: homeState.currentPage,
    hasMore: homeState.hasMore,
    currentSourceIds: homeState.currentSourceIds,
    timestamp: Date.now()
  };
  sessionStorage.setItem(SESSION_STORAGE_KEY, JSON.stringify(state));
}

function getSavedHomeState() {
  const saved = sessionStorage.getItem(SESSION_STORAGE_KEY);
  if (!saved) return null;
  
  try {
    const state = JSON.parse(saved);
    // Only restore if less than 30 minutes old
    if (Date.now() - state.timestamp > 30 * 60 * 1000) {
      sessionStorage.removeItem(SESSION_STORAGE_KEY);
      return null;
    }
    return state;
  } catch (e) {
    return null;
  }
}

function clearHomeState() {
  sessionStorage.removeItem(SESSION_STORAGE_KEY);
}

function navigateToArticle(articleId) {
  saveHomeState();
  window.location.href = `article.html?id=${articleId}`;
}

function renderArticles(articles, append) {
  const newsContent = document.getElementById("news-content");
  const fontClass = getFontClass();

  // Build source map from articles for display names
  articles.forEach(article => {
    if (article.source_id && article.source_name) {
      homeState.sourceMap.set(article.source_id, article.source_name);
    }
  });

  if (!append) {
    homeState.articlesData = articles;
    let heroHTML = "";
    let standardHTML = "";
    let compactHTML = "";

    articles.forEach((article, index) => {
      const imgSrc = article.image_url || getFallbackImage();
      
      if (index === 0) {
        heroHTML = `<div class="card card-hero animate-fade-in-up stagger-1" onclick="navigateToArticle('${article.id}')">
          <img src="${imgSrc}" class="card-image" alt="${article.title}" onerror="this.src='${getFallbackImage()}'; this.onerror=null;" />
          <span class="card-source">${article.source_name}</span>
          <div class="card-overlay">
            <h2 class="card-title ${fontClass}">${article.title}</h2>
            <div class="card-meta"><span>${formatDate(article.published_at)}</span></div>
          </div>
        </div>`;
      } else if (index >= 1 && index <= 3) {
        standardHTML += `<div class="card card-standard animate-fade-in-up stagger-${index + 1}" onclick="navigateToArticle('${article.id}')">
          <img src="${imgSrc}" class="card-image" alt="${article.title}" onerror="this.src='${getFallbackImage()}'; this.onerror=null;" />
          <div class="card-content">
            <h3 class="card-title ${fontClass}">${article.title}</h3>
            <div class="card-footer">
              <span class="badge">${article.source_name}</span>
              <span class="text-sm text-muted">${formatDate(article.published_at)}</span>
            </div>
          </div>
        </div>`;
      } else {
        compactHTML += renderCompactCard(article, fontClass);
      }
    });

    newsContent.innerHTML = `${heroHTML}${standardHTML}
      <div class="section-header mt-5"><span class="section-title">More Stories</span></div>
      <div id="articles-list">${compactHTML}</div>
      <div id="loading-more" class="loading-indicator hidden">
        <div class="ptr-spinner"></div><span class="ptr-text">Loading more...</span>
      </div>
      <div id="end-message" class="loading-indicator hidden">
        <span class="text-muted">You're all caught up!</span>
      </div>`;

    homeState.moreStoriesAdded = true;
  } else {
    homeState.articlesData = homeState.articlesData.concat(articles);
    const articlesList = document.getElementById("articles-list");
    articles.forEach(article => {
      articlesList.insertAdjacentHTML("beforeend", renderCompactCard(article, fontClass));
    });
  }

  // Update source filter display names if we have them in the map
  updateSourceFilterDisplayNames();
}

function updateSourceFilterDisplayNames() {
  // TODO: Remove this function once API provides source names directly
  // Currently needed because /api/v1/search/sources/by-language only returns IDs
  const sourceFilters = document.getElementById("source-filters");
  if (!sourceFilters) return;

  const pills = sourceFilters.querySelectorAll(".source-pill[data-source-id]");
  pills.forEach(pill => {
    const sourceId = pill.dataset.sourceId;
    if (sourceId && sourceId !== "all") {
      const displayName = homeState.sourceMap.get(sourceId);
      if (displayName) {
        pill.textContent = displayName;
      }
    }
  });
}

function renderCompactCard(article, fontClass) {
  const imgSrc = article.image_url || getFallbackImage();
  return `<div class="card card-compact article-item animate-fade-in-up" onclick="navigateToArticle('${article.id}')">
    <img src="${imgSrc}" class="card-image" alt="${article.title}" onerror="this.src='${getFallbackImage()}'; this.onerror=null;" />
    <div class="flex-1">
      <h3 class="card-title ${fontClass}">${article.title}</h3>
      <span class="text-xs text-muted mt-1">${article.source_name} • ${formatDate(article.published_at)}</span>
    </div>
  </div>`;
}

function loadSources() {
  getSourcesByLanguage(homeState.currentLanguage)
    .then(sourcesByLang => {
      const sourceFilters = document.getElementById("source-filters");
      sourcesByLang.sources.forEach(source => {
        const pill = document.createElement("div");
        pill.className = "source-pill";

        // TODO: Remove compatibility code once new API is live
        // Handle both old format (string) and new format (object with id/name)
        let sourceId, sourceName;
        if (typeof source === 'string') {
          // Old API format: source is just the ID
          sourceId = source;
          sourceName = homeState.sourceMap.get(sourceId) || formatSourceId(sourceId);
        } else {
          // New API format: source is an object with id and name
          sourceId = source.id;
          sourceName = source.name;
          homeState.sourceMap.set(sourceId, sourceName);
        }

        pill.dataset.sourceId = sourceId;
        pill.textContent = sourceName;
        sourceFilters.appendChild(pill);
      });
      applySavedSourceSelection();
    })
    .catch(error => console.error("Failed to load sources:", error));
}

function formatSourceId(sourceId) {
  // TODO: Remove this fallback function once API provides source names directly
  // or when all source IDs have proper display names in the sourceMap
  return sourceId
    .split("-")
    .map(word => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ");
}

function applySavedSourceSelection() {
  const sourceFilters = document.getElementById("source-filters");
  const allPill = sourceFilters.querySelector('[data-source-id="all"]');

  if (homeState.currentSourceIds.length === 0) {
    allPill.classList.add("active");
  } else {
    allPill.classList.remove("active");
    homeState.currentSourceIds.forEach(sourceId => {
      const pill = sourceFilters.querySelector(`[data-source-id="${sourceId}"]`);
      if (pill) pill.classList.add("active");
    });
  }
}

function handleSourceClick(e) {
  if (!e.target.classList.contains("source-pill")) return;
  
  const sourceFilters = document.getElementById("source-filters");
  const clickedSourceId = e.target.dataset.sourceId;
  const allPill = sourceFilters.querySelector('[data-source-id="all"]');

  if (clickedSourceId === "all") {
    const pills = sourceFilters.querySelectorAll(".source-pill");
    pills.forEach(p => p.classList.remove("active"));
    e.target.classList.add("active");
    homeState.currentSourceIds = [];
  } else {
    allPill.classList.remove("active");

    if (e.target.classList.contains("active")) {
      e.target.classList.remove("active");
      homeState.currentSourceIds = homeState.currentSourceIds.filter(id => id !== clickedSourceId);
      if (homeState.currentSourceIds.length === 0) allPill.classList.add("active");
    } else {
      e.target.classList.add("active");
      if (!homeState.currentSourceIds.includes(clickedSourceId)) {
        homeState.currentSourceIds.push(clickedSourceId);
      }
    }
  }

  setSelectedSourceIds(homeState.currentSourceIds);
  loadArticles(false);
}

function handleScroll() {
  if (homeState.isLoading || !homeState.hasMore) return;

  const scrollPosition = window.innerHeight + window.scrollY;
  const threshold = document.documentElement.scrollHeight - 300;

  if (scrollPosition >= threshold) loadArticles(true);
}

function handleHeaderScroll() {
  const header = document.querySelector(".glass-header");
  if (window.scrollY > 10) {
    header.classList.add("scrolled");
  } else {
    header.classList.remove("scrolled");
  }
}

function restoreHomeState() {
  const savedState = getSavedHomeState();
  if (!savedState) return false;

  // Restore state variables
  homeState.articlesData = savedState.articles || [];
  homeState.currentPage = savedState.currentPage || 0;
  homeState.hasMore = savedState.hasMore !== undefined ? savedState.hasMore : true;
  // TODO: Remove backward compatibility for old currentSources format once all users are migrated
  homeState.currentSourceIds = savedState.currentSourceIds || savedState.currentSources || [];

  // Hide skeleton and show content
  const skeleton = document.getElementById("loading-skeleton");
  const content = document.getElementById("news-content");
  if (skeleton) skeleton.classList.add("hidden");
  if (content) {
    content.classList.remove("hidden");
  }

  // Render articles from cache
  if (homeState.articlesData.length > 0) {
    renderArticles(homeState.articlesData, false);
  }

  // Restore scroll position after a brief delay to ensure DOM is ready
  requestAnimationFrame(() => {
    window.scrollTo(0, savedState.scrollPosition || 0);
    // Clear state after successful restoration
    clearHomeState();
  });

  return true;
}

document.addEventListener("DOMContentLoaded", () => {
  const sourceScroll = document.querySelector(".source-scroll");
  if (sourceScroll) initTouchScroll(sourceScroll);

  // Try to restore saved state first
  const wasRestored = restoreHomeState();

  if (!wasRestored) {
    // No saved state, load fresh articles
    loadArticles(false);
  }

  loadSources();

  document.getElementById("source-filters").addEventListener("click", handleSourceClick);
  window.addEventListener("scroll", handleScroll);
  window.addEventListener("scroll", handleHeaderScroll);
});
