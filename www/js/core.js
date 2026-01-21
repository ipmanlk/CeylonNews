(function() {
  const savedTheme = localStorage.getItem("app_theme") || "light";
  document.documentElement.setAttribute("data-theme", savedTheme);
})();

function toggleTheme() {
  const current = document.documentElement.getAttribute("data-theme");
  const next = current === "light" ? "dark" : "light";
  document.documentElement.setAttribute("data-theme", next);
  localStorage.setItem("app_theme", next);
}

function getFontClass() {
  const lang = localStorage.getItem("ceylon_news_lang") || "en";
  const useCustomFont = localStorage.getItem("use_custom_font") === "true";
  return useCustomFont ? lang + "-font" : "";
}

function applyCustomFont(elements) {
  const fontClass = getFontClass();
  const classes = ["en-font", "si-font", "ta-font"];
  
  elements.forEach(function(el) {
    classes.forEach(function(c) { el.classList.remove(c); });
    if (fontClass) el.classList.add(fontClass);
  });
}

function initTouchScroll(container) {
  let isDown = false;
  let startX;
  let scrollLeft;

  container.addEventListener("mousedown", function(e) {
    isDown = true;
    container.style.cursor = "grabbing";
    startX = e.pageX - container.offsetLeft;
    scrollLeft = container.scrollLeft;
  });

  container.addEventListener("mouseleave", function() {
    isDown = false;
    container.style.cursor = "grab";
  });

  container.addEventListener("mouseup", function() {
    isDown = false;
    container.style.cursor = "grab";
  });

  container.addEventListener("mousemove", function(e) {
    if (!isDown) return;
    e.preventDefault();
    const x = e.pageX - container.offsetLeft;
    const walk = (x - startX) * 2;
    container.scrollLeft = scrollLeft - walk;
  });

  container.addEventListener("touchstart", function(e) {
    isDown = true;
    startX = e.touches[0].pageX - container.offsetLeft;
    scrollLeft = container.scrollLeft;
  });

  container.addEventListener("touchend", function() {
    isDown = false;
  });

  container.addEventListener("touchmove", function(e) {
    if (!isDown) return;
    const x = e.touches[0].pageX - container.offsetLeft;
    const walk = (x - startX) * 2;
    container.scrollLeft = scrollLeft - walk;
  });
}

function formatDate(dateString) {
  const date = new Date(dateString);
  const now = new Date();
  const diffMs = now - date;
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMs / 3600000);
  const diffDays = Math.floor(diffMs / 86400000);

  if (diffMins < 60) return diffMins + " min ago";
  if (diffHours < 24) return diffHours + " hr" + (diffHours > 1 ? "s" : "") + " ago";
  if (diffDays < 7) return diffDays + " day" + (diffDays > 1 ? "s" : "") + " ago";

  return date.toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric"
  });
}

function truncateText(text, maxLength) {
  if (text.length <= maxLength) return text;
  return text.substring(0, maxLength).trim() + "...";
}

const FALLBACK_IMAGES = [
  "https://images.unsplash.com/photo-1503694978374-8a2fa686963a?q=80&w=800&auto=format&fit=crop",
  "https://images.unsplash.com/photo-1573812195421-50a396d17893?q=80&w=400&auto=format&fit=crop",
  "https://images.unsplash.com/photo-1503428593586-e225b39bddfe?q=80&w=400&auto=format&fit=crop",
  "https://images.unsplash.com/photo-1529243856184-fd5465488984?q=80&w=400&auto=format&fit=crop"
];

function getFallbackImage() {
  return FALLBACK_IMAGES[Math.floor(Math.random() * FALLBACK_IMAGES.length)];
}

function setupImageErrorHandlers() {
  document.addEventListener("error", function(e) {
    if (e.target.tagName === "IMG" && !e.target.dataset.fallbackLoaded) {
      e.target.src = getFallbackImage();
      e.target.dataset.fallbackLoaded = "true";
    }
  }, true);
}

document.addEventListener("DOMContentLoaded", setupImageErrorHandlers);

document.addEventListener("deviceready", function() {
  console.log("Cordova ready: " + cordova.platformId + "@" + cordova.version);
}, false);
