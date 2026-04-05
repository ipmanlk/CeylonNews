function createHistoryArticleCard(article) {
  const card = document.createElement("div");
  card.className = "card card-compact mb-3 clickable";
  const imgSrc = article.image_url || getFallbackImage();
  const fontClass = getFontClass();
  
  card.innerHTML = `<img src="${imgSrc}" class="card-image" alt="${article.title}" onerror="this.src='${getFallbackImage()}'; this.onerror=null;" />
    <div class="flex-1">
      <h3 class="card-title ${fontClass}">${article.title}</h3>
      <span class="text-xs text-muted mt-1">${article.source_name} • ${formatDate(article.published_at)}</span>
    </div>`;

  card.addEventListener("click", () => {
    window.location.href = `article.html?id=${article.id}`;
  });

  return card;
}

async function loadReadHistoryUI() {
  const history = await getReadHistory();
  const container = document.getElementById("articles-container");
  const emptyState = document.getElementById("empty-state");
  const clearBtn = document.getElementById("clear-history-btn");

  container.innerHTML = "";

  if (history.length === 0) {
    emptyState.classList.remove("hidden");
    clearBtn.classList.add("hidden");
    return;
  }

  emptyState.classList.add("hidden");
  clearBtn.classList.remove("hidden");
  
  history.forEach(article => {
    const card = createHistoryArticleCard(article);
    container.appendChild(card);
  });
}

async function confirmClearHistory() {
  if (confirm("Clear all read history?")) {
    await clearReadHistory();
    loadReadHistoryUI();
  }
}

document.addEventListener("DOMContentLoaded", () => {
  loadReadHistoryUI();
});
