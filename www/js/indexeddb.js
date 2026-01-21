const DB_NAME = 'ceylon_news_db';
const DB_VERSION = 2;
const SAVED_ARTICLES_STORE = 'saved_articles';
const READ_HISTORY_STORE = 'read_history';
const READ_HISTORY_LIMIT = 30;

let db = null;

function openDB() {
  return new Promise((resolve, reject) => {
    if (db) {
      resolve(db);
      return;
    }

    const request = indexedDB.open(DB_NAME, DB_VERSION);

    request.onerror = () => {
      reject(request.error);
    };

    request.onsuccess = () => {
      db = request.result;
      resolve(db);
    };

    request.onupgradeneeded = (event) => {
      const database = event.target.result;
      
      if (!database.objectStoreNames.contains(SAVED_ARTICLES_STORE)) {
        const store = database.createObjectStore(SAVED_ARTICLES_STORE, { keyPath: 'id' });
        store.createIndex('saved_at', 'saved_at', { unique: false });
      }
      
      if (!database.objectStoreNames.contains(READ_HISTORY_STORE)) {
        const store = database.createObjectStore(READ_HISTORY_STORE, { keyPath: 'id' });
        store.createIndex('read_at', 'read_at', { unique: false });
      }
    };
  });
}

const savedArticles = {
  async getAll() {
    const database = await openDB();
    return new Promise((resolve, reject) => {
      const transaction = database.transaction(SAVED_ARTICLES_STORE, 'readonly');
      const store = transaction.objectStore(SAVED_ARTICLES_STORE);
      const request = store.getAll();

      request.onsuccess = () => resolve(request.result);
      request.onerror = () => reject(request.error);
    });
  },

  async get(id) {
    const database = await openDB();
    return new Promise((resolve, reject) => {
      const transaction = database.transaction(SAVED_ARTICLES_STORE, 'readonly');
      const store = transaction.objectStore(SAVED_ARTICLES_STORE);
      const request = store.get(id);

      request.onsuccess = () => resolve(request.result);
      request.onerror = () => reject(request.error);
    });
  },

  async save(article) {
    const database = await openDB();
    return new Promise((resolve, reject) => {
      const transaction = database.transaction(SAVED_ARTICLES_STORE, 'readwrite');
      const store = transaction.objectStore(SAVED_ARTICLES_STORE);
      article.saved_at = new Date().toISOString();
      const request = store.put(article);

      request.onsuccess = () => resolve(true);
      request.onerror = () => reject(request.error);
    });
  },

  async remove(id) {
    const database = await openDB();
    return new Promise((resolve, reject) => {
      const transaction = database.transaction(SAVED_ARTICLES_STORE, 'readwrite');
      const store = transaction.objectStore(SAVED_ARTICLES_STORE);
      const request = store.delete(id);

      request.onsuccess = () => resolve(true);
      request.onerror = () => reject(request.error);
    });
  },

  async has(id) {
    const article = await this.get(id);
    return !!article;
  },

  async toggle(article) {
    const saved = await this.has(article.id);
    if (saved) {
      await this.remove(article.id);
      return false;
    } else {
      await this.save(article);
      return true;
    }
  }
};

const readHistory = {
  async getAll() {
    const database = await openDB();
    return new Promise((resolve, reject) => {
      const transaction = database.transaction(READ_HISTORY_STORE, 'readonly');
      const store = transaction.objectStore(READ_HISTORY_STORE);
      const index = store.index('read_at');
      const request = index.getAll();

      request.onsuccess = () => {
        const results = request.result.sort((a, b) => 
          new Date(b.read_at) - new Date(a.read_at)
        );
        resolve(results);
      };
      request.onerror = () => reject(request.error);
    });
  },

  async add(article) {
    const database = await openDB();
    
    const articleToSave = {
      id: article.id,
      title: article.title,
      source_name: article.source_name,
      image_url: article.image_url,
      published_at: article.published_at,
      read_at: new Date().toISOString()
    };

    return new Promise((resolve, reject) => {
      const transaction = database.transaction(READ_HISTORY_STORE, 'readwrite');
      const store = transaction.objectStore(READ_HISTORY_STORE);
      
      store.put(articleToSave);

      transaction.oncomplete = async () => {
        await this._trimToLimit();
        resolve(true);
      };
      transaction.onerror = () => reject(transaction.error);
    });
  },

  async _trimToLimit() {
    const database = await openDB();
    const all = await this.getAll();
    
    if (all.length <= READ_HISTORY_LIMIT) {
      return;
    }

    const toRemove = all.slice(READ_HISTORY_LIMIT);
    
    return new Promise((resolve, reject) => {
      const transaction = database.transaction(READ_HISTORY_STORE, 'readwrite');
      const store = transaction.objectStore(READ_HISTORY_STORE);
      
      toRemove.forEach(article => {
        store.delete(article.id);
      });

      transaction.oncomplete = () => resolve(true);
      transaction.onerror = () => reject(transaction.error);
    });
  },

  async clear() {
    const database = await openDB();
    return new Promise((resolve, reject) => {
      const transaction = database.transaction(READ_HISTORY_STORE, 'readwrite');
      const store = transaction.objectStore(READ_HISTORY_STORE);
      const request = store.clear();

      request.onsuccess = () => resolve(true);
      request.onerror = () => reject(request.error);
    });
  }
};
