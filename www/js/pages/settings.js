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

document.addEventListener("DOMContentLoaded", () => {
  if (localStorage.getItem("app_theme") === "dark") {
    document.getElementById("theme-toggle").checked = true;
  }

  if (isCustomFontEnabled()) {
    document.getElementById("custom-font-toggle").checked = true;
  }

  updateLanguageDisplay();
});
