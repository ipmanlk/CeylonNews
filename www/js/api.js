const API_BASE_URL = "https://cnapi.navinda.me";

function apiRequest(endpoint, options = {}) {
  const url = API_BASE_URL + endpoint;
  const defaultOptions = {
    headers: { "Content-Type": "application/json" }
  };

  return fetch(url, { ...defaultOptions, ...options })
    .then(response => {
      if (!response.ok) {
        return response.json()
          .catch(() => ({ error: "Request failed" }))
          .then(err => { throw new Error(err.error || "Request failed"); });
      }
      return response.json();
    });
}

function getArticles(params = {}) {
  const queryParams = new URLSearchParams();

  if (params.limit) queryParams.append("limit", params.limit);
  if (params.offset) queryParams.append("offset", params.offset);
  if (params.languages) {
    queryParams.append("languages", 
      Array.isArray(params.languages) ? params.languages.join(",") : params.languages
    );
  }
  if (params.sourceNames && Array.isArray(params.sourceNames)) {
    params.sourceNames.forEach(source => queryParams.append("source_names", source));
  }
  if (params.startDate) queryParams.append("start_date", params.startDate);
  if (params.endDate) queryParams.append("end_date", params.endDate);

  const queryString = queryParams.toString();
  const endpoint = queryString ? "/api/v1/articles?" + queryString : "/api/v1/articles";
  return apiRequest(endpoint);
}

function getArticleById(id) {
  return apiRequest("/api/v1/articles/" + id);
}

function searchArticles(params = {}) {
  const queryParams = new URLSearchParams();

  if (params.query) queryParams.append("query", params.query);
  if (params.languages) {
    queryParams.append("languages",
      Array.isArray(params.languages) ? params.languages.join(",") : params.languages
    );
  }
  if (params.sourceNames) {
    params.sourceNames.forEach(source => queryParams.append("source_names", source));
  }
  if (params.startDate) queryParams.append("start_date", params.startDate);
  if (params.endDate) queryParams.append("end_date", params.endDate);
  if (params.limit) queryParams.append("limit", params.limit);
  if (params.offset) queryParams.append("offset", params.offset);

  return apiRequest("/api/v1/search?" + queryParams.toString());
}

function getSourcesByLanguage(language) {
  return apiRequest("/api/v1/search/sources/by-language?language=" + language);
}

function getRecentArticlesFromApi(params = {}) {
  const queryParams = new URLSearchParams();

  if (params.languages) {
    queryParams.append("languages",
      Array.isArray(params.languages) ? params.languages.join(",") : params.languages
    );
  }
  if (params.sourceNames) {
    params.sourceNames.forEach(source => queryParams.append("source_names", source));
  }
  if (params.limit) queryParams.append("limit", params.limit);

  const queryString = queryParams.toString();
  const endpoint = queryString ? "/api/v1/search/recent?" + queryString : "/api/v1/search/recent";
  return apiRequest(endpoint);
}

function healthCheck() {
  return apiRequest("/health");
}
