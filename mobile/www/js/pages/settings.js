function updateLanguageDisplay() {
  const lang = getLanguage();
  const langNames = { en: "English", si: "සිංහල", ta: "தமிழ்" };
  const langElement = document.getElementById("current-language");
  if (langElement) {
    langElement.textContent = langNames[lang];
  }
}

function toggleCustomFont() {
  const isEnabled = document.getElementById("custom-font-toggle").checked;
  setCustomFontEnabled(isEnabled);
}

function showContactModal() {
  document.getElementById("contact-modal").classList.add("active");
  document.body.style.overflow = "hidden";
}

function hideContactModal(e) {
  if (e && e.target !== e.currentTarget) return;
  document.getElementById("contact-modal").classList.remove("active");
  document.body.style.overflow = "";
}

function showLanguageModal() {
  document.getElementById("language-modal").classList.add("active");
  document.body.style.overflow = "hidden";
}

function hideLanguageModal(e) {
  if (e && e.target !== e.currentTarget) return;
  document.getElementById("language-modal").classList.remove("active");
  document.body.style.overflow = "";
}

function changeLanguage(lang) {
  setLanguage(lang);
  updateLanguageDisplay();
  hideLanguageModal();
  window.location.reload();
}

function updateCacheSize() {
  if (window.serviceWorkerSupported === false) {
    const cacheSizeElement = document.getElementById("cache-size");
    if (cacheSizeElement) {
      cacheSizeElement.textContent = "Not supported";
    }
    const clearButton = document.querySelector('[onclick="clearCacheData()"]');
    if (clearButton) {
      clearButton.disabled = true;
      clearButton.style.opacity = "0.5";
      clearButton.style.cursor = "not-allowed";
    }
    return;
  }
  
  getCacheSize()
    .then(function (bytes) {
      const cacheSizeElement = document.getElementById("cache-size");
      if (cacheSizeElement) {
        cacheSizeElement.textContent = formatBytes(bytes);
      }
    })
    .catch(function (error) {
      console.error("Error getting cache size:", error);
      const cacheSizeElement = document.getElementById("cache-size");
      if (cacheSizeElement) {
        cacheSizeElement.textContent = "Not available";
      }
    });
}

function clearCacheData() {
  if (window.serviceWorkerSupported === false) {
    alert("Cache functionality is not available in this environment.");
    return;
  }
  
  if (
    confirm(
      "Are you sure you want to clear all cached data? This may slow down the app initially.",
    )
  ) {
    clearCache()
      .then(function () {
        alert("Cache cleared successfully!");
        updateCacheSize();
      })
      .catch(function (error) {
        console.error("Error clearing cache:", error);
        alert("Failed to clear cache: " + error.message);
      });
  }
}

function showFontSizeModal() {
  document.getElementById("font-size-modal").classList.add("active");
  document.body.style.overflow = "hidden";
}

function hideFontSizeModal(e) {
  if (e && e.target !== e.currentTarget) return;
  document.getElementById("font-size-modal").classList.remove("active");
  document.body.style.overflow = "";
}

function setArticleSize(size) {
  setArticleFontSize(size);
  updateFontSizeDisplay();
  hideFontSizeModal();
}

function updateFontSizeDisplay() {
  const size = getArticleFontSize();
  const sizeNames = { 
    small: "Small", 
    medium: "Medium", 
    large: "Large", 
    "extra-large": "Extra Large" 
  };
  const fontSizeElement = document.getElementById("current-font-size");
  if (fontSizeElement) {
    fontSizeElement.textContent = sizeNames[size];
  }
}

document.addEventListener("DOMContentLoaded", () => {
  if (localStorage.getItem("app_theme") === "dark") {
    document.getElementById("theme-toggle").checked = true;
  }

  if (isCustomFontEnabled()) {
    document.getElementById("custom-font-toggle").checked = true;
  }

  updateFontSizeDisplay();
  updateLanguageDisplay();

  if (typeof window.serviceWorkerSupported !== 'undefined') {
    updateCacheSize();
  } else {
    setTimeout(updateCacheSize, 500);
  }
});
