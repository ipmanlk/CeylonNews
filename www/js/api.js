const API_BASE_URL = "https://cnapi.navinda.me";

async function apiRequest(endpoint, options = {}) {
  const url = `${API_BASE_URL}${endpoint}`;
  const defaultOptions = {
    headers: {
      "Content-Type": "application/json",
    },
  };

  const response = await fetch(url, { ...defaultOptions, ...options });

  if (!response.ok) {
    const error = await response
      .json()
      .catch(() => ({ error: "Request failed" }));
    throw new Error(error.error || "Request failed");
  }

  return response.json();
}

async function getArticles(params = {}) {
  const queryParams = new URLSearchParams();

  if (params.limit) queryParams.append("limit", params.limit);
  if (params.offset) queryParams.append("offset", params.offset);
  if (params.languages) {
    queryParams.append(
      "languages",
      Array.isArray(params.languages)
        ? params.languages.join(",")
        : params.languages,
    );
  }
  if (params.sourceNames && Array.isArray(params.sourceNames)) {
    params.sourceNames.forEach((source) =>
      queryParams.append("source_names", source),
    );
  }
  if (params.startDate) queryParams.append("start_date", params.startDate);
  if (params.endDate) queryParams.append("end_date", params.endDate);

  const queryString = queryParams.toString();
  const endpoint = queryString
    ? `/api/v1/articles?${queryString}`
    : "/api/v1/articles";

  return apiRequest(endpoint);
}

async function getArticleById(id) {
  return apiRequest(`/api/v1/articles/${id}`);
}

async function searchArticles(params = {}) {
  const queryParams = new URLSearchParams();

  if (params.query) queryParams.append("query", params.query);
  if (params.languages) {
    queryParams.append(
      "languages",
      Array.isArray(params.languages)
        ? params.languages.join(",")
        : params.languages,
    );
  }
  if (params.sourceNames) {
    params.sourceNames.forEach((source) =>
      queryParams.append("source_names", source),
    );
  }
  if (params.startDate) queryParams.append("start_date", params.startDate);
  if (params.endDate) queryParams.append("end_date", params.endDate);
  if (params.limit) queryParams.append("limit", params.limit);
  if (params.offset) queryParams.append("offset", params.offset);

  return apiRequest(`/api/v1/search?${queryParams.toString()}`);
}

async function getAvailableSources() {
  return apiRequest("/api/v1/search/sources");
}

async function getAvailableLanguages() {
  return apiRequest("/api/v1/search/languages");
}

async function getSourcesByLanguage(language) {
  return apiRequest(`/api/v1/search/sources/by-language?language=${language}`);
}

async function getRecentArticles(params = {}) {
  const queryParams = new URLSearchParams();

  if (params.languages) {
    queryParams.append(
      "languages",
      Array.isArray(params.languages)
        ? params.languages.join(",")
        : params.languages,
    );
  }
  if (params.sourceNames) {
    params.sourceNames.forEach((source) =>
      queryParams.append("source_names", source),
    );
  }
  if (params.limit) queryParams.append("limit", params.limit);

  const queryString = queryParams.toString();
  const endpoint = queryString
    ? `/api/v1/search/recent?${queryString}`
    : "/api/v1/search/recent";

  return apiRequest(endpoint);
}

async function healthCheck() {
  return apiRequest("/health");
}

function formatDate(dateString) {
  const date = new Date(dateString);
  const now = new Date();
  const diffMs = now - date;
  const diffMins = Math.floor(diffMs / 60000);
  const diffHours = Math.floor(diffMs / 3600000);
  const diffDays = Math.floor(diffMs / 86400000);

  if (diffMins < 60) return `${diffMins} min ago`;
  if (diffHours < 24) return `${diffHours} hr${diffHours > 1 ? "s" : ""} ago`;
  if (diffDays < 7) return `${diffDays} day${diffDays > 1 ? "s" : ""} ago`;

  return date.toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

function truncateText(text, maxLength) {
  if (text.length <= maxLength) return text;
  return text.substring(0, maxLength).trim() + "...";
}

function getFallbackImage() {
  const images = [
    "https://images.unsplash.com/photo-1503694978374-8a2fa686963a?q=80&w=800&auto=format&fit=crop",
    "https://images.unsplash.com/photo-1573812195421-50a396d17893?q=80&w=400&auto=format&fit=crop",
    "https://images.unsplash.com/photo-1503694978374-8a2fa686963a?q=80&w=400&auto=format&fit=crop",
    "https://images.unsplash.com/photo-1503428593586-e225b39bddfe?q=80&w=400&auto=format&fit=crop",
    "https://images.unsplash.com/photo-1529243856184-fd5465488984?q=80&w=400&auto=format&fit=crop",
  ];
  return images[Math.floor(Math.random() * images.length)];
}

function setupImageErrorHandlers() {
  document.addEventListener(
    "error",
    function (e) {
      if (e.target.tagName === "IMG" && !e.target.dataset.fallbackLoaded) {
        e.target.src = getFallbackImage();
        e.target.dataset.fallbackLoaded = "true";
      }
    },
    true,
  );
}

document.addEventListener("DOMContentLoaded", setupImageErrorHandlers);
