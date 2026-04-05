const CACHE_VERSION = "v4";

const CONFIG = {
  caches: {
    api: `ceylon-news-api-${CACHE_VERSION}`,
    static: `ceylon-news-static-${CACHE_VERSION}`,
    images: `ceylon-news-images-${CACHE_VERSION}`,
  },
  durations: {
    ARTICLES: 600,
    ARTICLE_DETAIL: 3600,
    SOURCES: 86400,
    STATIC: 604800,
    IMAGES: 86400,
    SEARCH: 300,
  },
  limits: {
    api: 100,
    images: 500,
  },
  apiBase: "https://cnapi.navinda.me",
};

self.addEventListener("install", (event) => {
  event.waitUntil(self.skipWaiting());
});

self.addEventListener("activate", (event) => {
  event.waitUntil(
    caches
      .keys()
      .then((keys) => {
        const activeCaches = Object.values(CONFIG.caches);
        return Promise.all(
          keys.map((key) => {
            if (!activeCaches.includes(key)) {
              return caches.delete(key);
            }
          }),
        );
      })
      .then(() => self.clients.claim()),
  );
});

self.addEventListener("fetch", (event) => {
  const url = new URL(event.request.url);

  if (shouldIgnore(url)) {
    return;
  }

  if (url.origin === CONFIG.apiBase) {
    event.respondWith(routeApiRequest(event.request, url));
    return;
  }

  // check for image extensions to catch external CDN URLs
  if (url.pathname.match(/\.(jpg|jpeg|png|gif|webp|svg)$/i)) {
    event.respondWith(
      cacheFirst(
        event.request,
        CONFIG.caches.images,
        CONFIG.durations.IMAGES,
        CONFIG.limits.images,
      ),
    );
    return;
  }
});

function shouldIgnore(url) {
  // Localhost/Loopback (Dev environment)
  if (["localhost", "127.0.0.1"].includes(url.hostname)) return true;

  // Cordova / File Protocol
  if (url.protocol === "file:" || url.pathname.includes("cordova")) return true;

  // Bundled Assets
  const isLocalOrigin =
    url.origin === self.location.origin || url.protocol === "file:";

  if (isLocalOrigin) {
    // root or HTML files
    if (url.pathname === "/" || url.pathname.endsWith(".html")) return true;

    // local assets
    if (/^\/(js|css|fonts|img|res)\//i.test(url.pathname)) return true;
  }

  return false;
}

function routeApiRequest(request, url) {
  const { caches, durations, limits } = CONFIG;

  if (
    url.pathname.match(/\/articles\/\d+$/) ||
    url.pathname.includes("/sources")
  ) {
    return cacheFirst(
      request,
      caches.api,
      durations.ARTICLE_DETAIL,
      limits.api,
    );
  }

  if (url.pathname.includes("/search")) {
    return staleWhileRevalidate(
      request,
      caches.api,
      durations.SEARCH,
      limits.api,
    );
  }

  return staleWhileRevalidate(
    request,
    caches.api,
    durations.ARTICLES,
    limits.api,
  );
}

// Strategies

function staleWhileRevalidate(request, cacheName, maxAge, maxEntries) {
  return caches.open(cacheName).then((cache) => {
    return cache.match(request).then((cachedResp) => {
      const isValid = cachedResp && isCacheValid(cachedResp, maxAge);
      const fetchPromise = fetchAndCache(request, cacheName, maxEntries);

      if (isValid) {
        fetchPromise.catch(() => {});
        return cachedResp;
      }

      return fetchPromise;
    });
  });
}

function cacheFirst(request, cacheName, maxAge, maxEntries) {
  return caches.open(cacheName).then((cache) => {
    return cache.match(request).then((cachedResp) => {
      if (cachedResp && isCacheValid(cachedResp, maxAge)) {
        return cachedResp;
      }
      return fetchAndCache(request, cacheName, maxEntries);
    });
  });
}

// Helpers

function fetchAndCache(request, cacheName, maxEntries) {
  return fetch(request).then((networkResp) => {
    // Allow Opaque responses for cross-origin images
    // Opaque responses have status 0 and ok: false, but contain valid image data
    const isOpaque = networkResp.type === "opaque";

    if (!networkResp.ok && !isOpaque) {
      throw new Error("Network response invalid");
    }

    // We can only clone and modify headers for NON-opaque responses.
    // Opaque responses must be stored "as is".
    let respToCache;

    if (isOpaque) {
      respToCache = networkResp.clone();
    } else {
      const headers = new Headers(networkResp.headers);
      headers.append("sw-fetched-on", Date.now().toString());

      respToCache = new Response(networkResp.clone().body, {
        status: networkResp.status,
        statusText: networkResp.statusText,
        headers: headers,
      });
    }

    caches.open(cacheName).then((cache) => {
      cache.put(request, respToCache);
      if (maxEntries) trimCache(cacheName, maxEntries);
    });

    return networkResp;
  });
}

function isCacheValid(response, maxAgeSeconds) {
  // If response is Opaque (CORS image), we cannot read headers.
  // We assume it's valid and let the LRU (Cache Limit) handle cleanup.
  if (response.type === "opaque") return true;

  const timestamp = response.headers.get("sw-fetched-on");
  if (!timestamp) return false;

  return (Date.now() - parseInt(timestamp, 10)) / 1000 < maxAgeSeconds;
}

function trimCache(cacheName, maxEntries) {
  caches.open(cacheName).then((cache) => {
    cache.keys().then((keys) => {
      if (keys.length > maxEntries) {
        const itemsToDelete = keys.slice(0, keys.length - maxEntries);
        itemsToDelete.forEach((req) => cache.delete(req));
      }
    });
  });
}
