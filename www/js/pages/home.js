const homeState = {
  isLoading: false,
  currentPage: 0,
  perPage: 10,
  hasMore: true,
  currentSources: getSelectedSources(),
  currentLanguage: getLanguage(),
  moreStoriesAdded: false
};

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

  if (homeState.currentSources.length > 0) {
    params.sourceNames = homeState.currentSources;
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

function renderArticles(articles, append) {
  const newsContent = document.getElementById("news-content");
  const fontClass = getFontClass();

  if (!append) {
    let heroHTML = "";
    let standardHTML = "";
    let compactHTML = "";

    articles.forEach((article, index) => {
      const imgSrc = article.image_url || getFallbackImage();
      
      if (index === 0) {
        heroHTML = `<div class="card card-hero animate-fade-in-up stagger-1" onclick="window.location.href='article.html?id=${article.id}'">
          <img src="${imgSrc}" class="card-image" alt="${article.title}" onerror="this.src='${getFallbackImage()}'; this.onerror=null;" />
          <span class="card-source">${article.source_name}</span>
          <div class="card-overlay">
            <h2 class="card-title ${fontClass}">${article.title}</h2>
            <div class="card-meta"><span>${formatDate(article.published_at)}</span></div>
          </div>
        </div>`;
      } else if (index >= 1 && index <= 3) {
        standardHTML += `<div class="card card-standard animate-fade-in-up stagger-${index + 1}" onclick="window.location.href='article.html?id=${article.id}'">
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
    const articlesList = document.getElementById("articles-list");
    articles.forEach(article => {
      articlesList.insertAdjacentHTML("beforeend", renderCompactCard(article, fontClass));
    });
  }
}

function renderCompactCard(article, fontClass) {
  const imgSrc = article.image_url || getFallbackImage();
  return `<div class="card card-compact article-item animate-fade-in-up" onclick="window.location.href='article.html?id=${article.id}'">
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
        pill.dataset.source = source;
        pill.textContent = source;
        sourceFilters.appendChild(pill);
      });
      applySavedSourceSelection();
    })
    .catch(error => console.error("Failed to load sources:", error));
}

function applySavedSourceSelection() {
  const sourceFilters = document.getElementById("source-filters");
  const allPill = sourceFilters.querySelector('[data-source="all"]');

  if (homeState.currentSources.length === 0) {
    allPill.classList.add("active");
  } else {
    allPill.classList.remove("active");
    homeState.currentSources.forEach(source => {
      const pill = sourceFilters.querySelector(`[data-source="${source}"]`);
      if (pill) pill.classList.add("active");
    });
  }
}

function handleSourceClick(e) {
  if (!e.target.classList.contains("source-pill")) return;
  
  const sourceFilters = document.getElementById("source-filters");
  const clickedSource = e.target.dataset.source;
  const allPill = sourceFilters.querySelector('[data-source="all"]');

  if (clickedSource === "all") {
    const pills = sourceFilters.querySelectorAll(".source-pill");
    pills.forEach(p => p.classList.remove("active"));
    e.target.classList.add("active");
    homeState.currentSources = [];
  } else {
    allPill.classList.remove("active");

    if (e.target.classList.contains("active")) {
      e.target.classList.remove("active");
      homeState.currentSources = homeState.currentSources.filter(s => s !== clickedSource);
      if (homeState.currentSources.length === 0) allPill.classList.add("active");
    } else {
      e.target.classList.add("active");
      if (!homeState.currentSources.includes(clickedSource)) {
        homeState.currentSources.push(clickedSource);
      }
    }
  }

  setSelectedSources(homeState.currentSources);
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

document.addEventListener("DOMContentLoaded", () => {
  const sourceScroll = document.querySelector(".source-scroll");
  if (sourceScroll) initTouchScroll(sourceScroll);

  loadArticles(false);
  loadSources();

  document.getElementById("source-filters").addEventListener("click", handleSourceClick);
  window.addEventListener("scroll", handleScroll);
  window.addEventListener("scroll", handleHeaderScroll);
});
