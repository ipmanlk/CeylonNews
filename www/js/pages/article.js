let currentArticle = null;

function sanitizeContentHTML(html) {
  const parser = new DOMParser();
  const doc = parser.parseFromString(html, "text/html");
  doc.body.querySelectorAll("*").forEach(el => {
    el.removeAttribute("style");
    el.removeAttribute("width");
    el.removeAttribute("height");
    el.removeAttribute("class");
  });
  return doc.body.innerHTML;
}

function applyArticleFont() {
  const titleEl = document.getElementById("article-title");
  const bodyEl = document.getElementById("article-body");
  applyCustomFont([titleEl, bodyEl]);
}

async function loadArticle(id) {
  const loadingEl = document.getElementById("article-loading");
  const contentEl = document.getElementById("article-content");
  const heroLoading = document.getElementById("hero-loading");
  const articleHero = document.getElementById("article-hero");

  try {
    const article = await getArticleById(id);
    currentArticle = article;

    heroLoading.classList.add("hidden");
    articleHero.classList.remove("hidden");
    loadingEl.classList.add("hidden");
    contentEl.classList.remove("hidden");

    document.getElementById("source-name").textContent = article.source_name;
    document.getElementById("article-title").textContent = article.title;
    document.getElementById("article-date").textContent = new Date(article.published_at)
      .toLocaleDateString("en-US", { month: "long", day: "numeric", year: "numeric" });

    const heroImage = document.getElementById("article-hero-img");
    heroImage.src = article.image_url || getFallbackImage();
    heroImage.alt = article.title;
    heroImage.onerror = function() {
      this.src = getFallbackImage();
      this.onerror = null;
    };

    document.getElementById("article-body").innerHTML = sanitizeContentHTML(article.content_html);

    const readOriginalBtn = document.getElementById("read-original-btn");
    readOriginalBtn.onclick = () => window.open(article.url, "_system");
    readOriginalBtn.innerHTML = `Read Original on ${article.source_name}<i class="ph ph-arrow-square-out icon-md ml-2"></i>`;

    await updateBookmarkButton();
    addToReadHistory(article);
    loadRelatedArticles(article.source_name, article.language);
    applyArticleFont();
  } catch (error) {
    console.error("Failed to load article:", error);
    heroLoading.classList.add("hidden");
    loadingEl.innerHTML = `<div class="empty-state-container">
      <i class="ph ph-warning-circle empty-state-icon"></i>
      <p class="empty-state-text">Failed to load article</p>
      <button onclick="history.back()" class="btn btn-outline mt-4 btn-auto-width">Go Back</button>
    </div>`;
  }
}

async function updateBookmarkButton() {
  const bookmarkBtn = document.getElementById("bookmark-btn");
  const bookmarkIcon = document.getElementById("bookmark-icon");

  if (currentArticle) {
    bookmarkBtn.classList.remove("hidden");
    const saved = await isArticleSaved(currentArticle.id);
    if (saved) {
      bookmarkBtn.classList.add("bookmarked");
      bookmarkIcon.className = "ph-fill ph-bookmark-simple icon-lg";
    } else {
      bookmarkBtn.classList.remove("bookmarked");
      bookmarkIcon.className = "ph ph-bookmark-simple icon-lg";
    }
  }
}

async function handleBookmarkClick() {
  if (currentArticle) {
    await toggleSavedArticle(currentArticle);
    await updateBookmarkButton();
  }
}

function loadRelatedArticles(source, language) {
  getRecentArticlesFromApi({
    sourceNames: [source],
    languages: [language],
    limit: 3
  })
    .then(response => {
      const relatedContainer = document.getElementById("related-articles");
      if (response.articles && response.articles.length > 0) {
        response.articles.slice(0, 2).forEach(article => {
          const card = document.createElement("div");
          card.className = "card card-compact mb-3";
          const imgSrc = article.image_url || getFallbackImage();
          card.innerHTML = `<img src="${imgSrc}" class="card-image" alt="${article.title}" onerror="this.src='${getFallbackImage()}'; this.onerror=null;" />
            <div class="flex-1">
              <h3 class="card-title">${article.title}</h3>
              <span class="text-xs text-muted mt-1">${article.source_name} • ${formatDate(article.published_at)}</span>
            </div>`;
          card.onclick = () => window.location.href = `article.html?id=${article.id}`;
          relatedContainer.appendChild(card);
        });
      }
    })
    .catch(error => console.error("Failed to load related articles:", error));
}

document.addEventListener("DOMContentLoaded", () => {
  const urlParams = new URLSearchParams(window.location.search);
  const articleId = urlParams.get("id");

  document.getElementById("bookmark-btn").addEventListener("click", handleBookmarkClick);

  if (articleId) {
    loadArticle(articleId);
  } else {
    const heroLoading = document.getElementById("hero-loading");
    const loadingEl = document.getElementById("article-loading");
    heroLoading.classList.add("hidden");
    loadingEl.innerHTML = `<div class="empty-state-container">
      <i class="ph ph-warning-circle empty-state-icon"></i>
      <p class="empty-state-text">No article specified</p>
      <button onclick="history.back()" class="btn btn-outline mt-4 btn-auto-width">Go Back</button>
    </div>`;
  }
});
