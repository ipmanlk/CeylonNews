async function createSavedArticleCard(article) {
  const card = document.createElement("div");
  card.className = "card card-compact mb-3";
  const imgSrc = article.image_url || getFallbackImage();
  const fontClass = getFontClass();
  
  card.innerHTML = `<img src="${imgSrc}" class="card-image" alt="${article.title}" onerror="this.src='${getFallbackImage()}'; this.onerror=null;" />
    <div class="flex-1 flex-center gap-2">
      <div class="flex-1">
        <h3 class="card-title ${fontClass}">${article.title}</h3>
        <span class="text-xs text-muted mt-1">${article.source_name} • ${formatDate(article.published_at)}</span>
      </div>
      <button class="unsave-btn btn-clear clickable text-muted" data-id="${article.id}">
        <i class="ph-fill ph-bookmark-simple icon-md text-accent"></i>
      </button>
    </div>`;

  card.querySelector(".card-title").addEventListener("click", () => {
    window.location.href = `article.html?id=${article.id}`;
  });

  card.querySelector(".card-image").addEventListener("click", () => {
    window.location.href = `article.html?id=${article.id}`;
  });

  card.querySelector(".unsave-btn").addEventListener("click", async e => {
    e.stopPropagation();
    await removeArticle(article.id);
    card.style.opacity = "0";
    card.style.transform = "translateX(-20px)";
    setTimeout(() => {
      card.remove();
      loadSavedArticlesUI();
    }, 200);
  });

  return card;
}

async function loadSavedArticlesUI() {
  const saved = await getSavedArticles();
  const container = document.getElementById("articles-container");
  const emptyState = document.getElementById("empty-state");

  container.innerHTML = "";

  if (saved.length === 0) {
    emptyState.classList.remove("hidden");
    return;
  }

  emptyState.classList.add("hidden");
  saved.reverse().forEach(async article => {
    const card = await createSavedArticleCard(article);
    container.appendChild(card);
  });
}

document.addEventListener("DOMContentLoaded", () => {
  loadSavedArticlesUI();
});
