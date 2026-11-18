import api from "@/api.js";

// How long reference data (SPs, showcases, account groups, currencies) is
// reused between pages before it's requested again
export const CACHE_TTL = 5 * 60 * 1000;

const caches = new Set();

/**
 * Remembers which params the module's list state was loaded with and when.
 * The state itself stays in the store, so local mutations (push/remove/replace)
 * survive cache hits.
 *
 * `invalidateOn` - a REST url pattern: any successful non-GET request to it
 * (create, update, delete, bind_plan, ...) drops the cache.
 */
export const createListCache = ({ invalidateOn } = {}) => {
  let slot = null; // { key, at, promise, pending }

  const cache = {
    invalidateOn,

    /** Run `load` unless the state already holds fresh data for `params`. */
    run(params, load, { force = false } = {}) {
      const key = JSON.stringify(params ?? {});
      const isFresh =
        slot?.key === key &&
        (slot.pending || Date.now() - slot.at < CACHE_TTL);
      if (!force && isFresh) return slot.promise;

      const entry = { key, at: Date.now(), pending: true };
      entry.promise = Promise.resolve()
        .then(load)
        .then(
          (response) => {
            entry.pending = false;
            entry.at = Date.now();
            return response;
          },
          (error) => {
            if (slot === entry) slot = null;
            throw error;
          },
        );
      slot = entry;
      return entry.promise;
    },

    invalidate() {
      slot = null;
    },
  };

  caches.add(cache);
  return cache;
};

/** Drop every reference cache: the next ensure() goes to the server. */
export const invalidateAllCaches = () => {
  caches.forEach((cache) => cache.invalidate());
};

api.axios.interceptors.response.use((response) => {
  const { method = "get", url = "" } = response.config ?? {};
  if (method.toLowerCase() !== "get") {
    caches.forEach((cache) => {
      if (cache.invalidateOn?.test(url)) cache.invalidate();
    });
  }
  return response;
});
