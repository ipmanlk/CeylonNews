const DB_CONFIG = {
  NAME: "ceylon_news_db",
  VERSION: 2,
  STORES: {
    SAVED: "saved_articles",
    HISTORY: "read_history",
  },
  LIMITS: {
    HISTORY: 30,
  },
};

let dbInstance = null;

function getDB() {
  if (dbInstance) return Promise.resolve(dbInstance);

  return new Promise((resolve, reject) => {
    const request = indexedDB.open(DB_CONFIG.NAME, DB_CONFIG.VERSION);

    request.onerror = () => reject(request.error);
    request.onsuccess = () => {
      dbInstance = request.result;
      resolve(dbInstance);
    };

    request.onupgradeneeded = (event) => {
      const db = event.target.result;

      if (!db.objectStoreNames.contains(DB_CONFIG.STORES.SAVED)) {
        const store = db.createObjectStore(DB_CONFIG.STORES.SAVED, {
          keyPath: "id",
        });
        store.createIndex("saved_at", "saved_at", { unique: false });
      }

      if (!db.objectStoreNames.contains(DB_CONFIG.STORES.HISTORY)) {
        const store = db.createObjectStore(DB_CONFIG.STORES.HISTORY, {
          keyPath: "id",
        });
        store.createIndex("read_at", "read_at", { unique: false });
      }
    };
  });
}

async function queryDB(storeName, mode, callback) {
  const db = await getDB();
  return new Promise((resolve, reject) => {
    const tx = db.transaction(storeName, mode);
    const store = tx.objectStore(storeName);

    // Execute the IDB operation (get, put, delete, getAll)
    const request = callback(store);

    request.onsuccess = () => resolve(request.result);
    request.onerror = () => reject(request.error);
  });
}

const savedArticles = {
  getAll() {
    return queryDB(DB_CONFIG.STORES.SAVED, "readonly", (store) =>
      store.getAll(),
    );
  },

  get(id) {
    return queryDB(DB_CONFIG.STORES.SAVED, "readonly", (store) =>
      store.get(id),
    );
  },

  save(article) {
    const data = { ...article, saved_at: new Date().toISOString() };
    return queryDB(DB_CONFIG.STORES.SAVED, "readwrite", (store) =>
      store.put(data),
    );
  },

  remove(id) {
    return queryDB(DB_CONFIG.STORES.SAVED, "readwrite", (store) =>
      store.delete(id),
    );
  },

  async has(id) {
    const result = await this.get(id);
    return !!result;
  },

  async toggle(article) {
    const exists = await this.has(article.id);
    if (exists) {
      await this.remove(article.id);
      return false;
    }
    await this.save(article);
    return true;
  },
};

const readHistory = {
  async getAll() {
    const results = await queryDB(
      DB_CONFIG.STORES.HISTORY,
      "readonly",
      (store) => {
        return store.index("read_at").getAll();
      },
    );

    // Sort mostly recent first
    return results.sort((a, b) => new Date(b.read_at) - new Date(a.read_at));
  },

  async add(article) {
    const entry = {
      id: article.id,
      title: article.title,
      source_name: article.source_name,
      image_url: article.image_url,
      published_at: article.published_at,
      read_at: new Date().toISOString(),
    };

    await queryDB(DB_CONFIG.STORES.HISTORY, "readwrite", (store) =>
      store.put(entry),
    );

    return this._trim();
  },

  async _trim() {
    const all = await this.getAll();
    if (all.length <= DB_CONFIG.LIMITS.HISTORY) return;

    const toDelete = all.slice(DB_CONFIG.LIMITS.HISTORY);

    const deletePromises = toDelete.map((item) =>
      queryDB(DB_CONFIG.STORES.HISTORY, "readwrite", (store) =>
        store.delete(item.id),
      ),
    );

    return Promise.all(deletePromises);
  },

  clear() {
    return queryDB(DB_CONFIG.STORES.HISTORY, "readwrite", (store) =>
      store.clear(),
    );
  },
};
