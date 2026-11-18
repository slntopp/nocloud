import { ref } from "vue";
import api from "@/api.js";
import store from "@/store";
import { createListCache } from "@/store/cache.js";
import {
  GetInvoicesRequest,
  ListRequest,
} from "nocloud-proto/proto/es/billing/billing_pb";
import { ListInstancesRequest } from "nocloud-proto/proto/es/instances/instances_pb";

// Full SP list, shared by all components; dropped by the reload button
// and by any change to an SP (see store/cache.js)
const spListCache = createListCache({ invalidateOn: /^\/?sp(\/|\?|$)/ });

// Max uuids per list request
const CHUNK_SIZE = 100;

// Each loader takes a list of uuids and returns the matching entities
// (entities that don't exist or aren't accessible are simply missing).
const loaders = {
  async accounts(uuids) {
    const { pool } = await api.post("accounts", { filters: { uuid: uuids } });
    return pool ?? [];
  },
  async namespaces(uuids) {
    const { pool } = await api.post("namespaces", { filters: { uuid: uuids } });
    return pool ?? [];
  },
  async services(uuids) {
    const { pool } = await api.post("services", {
      filters: { uuid: uuids },
      showDeleted: true,
    });
    return pool ?? [];
  },
  // SP list has no uuid filter, but there are few SPs: take them from one
  // cached list request instead of a request per SP
  async servicesProviders(uuids) {
    const params = { anonymously: false, showDeleted: true };
    const { pool = [] } = await spListCache.run(params, () =>
      api.servicesProviders.list(params),
    );
    return pool.filter((sp) => uuids.includes(sp.uuid));
  },
  async plans(uuids) {
    const response = await store.getters["plans/plansClient"].listPlans(
      ListRequest.fromJson({ filters: { uuid: uuids }, showDeleted: true }),
    );
    return response.toJson().pool ?? [];
  },
  async instances(uuids) {
    const response = await store.getters["instances/instancesClient"].list(
      ListInstancesRequest.fromJson({ filters: { uuid: uuids } }),
    );
    return response.pool.map((i) => ({
      ...i,
      ...i.instance.toJson(),
      instance: undefined,
    }));
  },
  async invoices(uuids) {
    const response = await store.getters["invoices/invoicesClient"].getInvoices(
      GetInvoicesRequest.fromJson({ filters: { uuid: uuids } }),
    );
    return response.toJson().pool ?? [];
  },
};

// Requests made within the same tick (also from different components)
// are merged into a single list request per entity type.
const queues = {};

const flush = async (type) => {
  const queue = queues[type];
  delete queues[type];

  const uuids = [...queue.keys()];
  const chunks = [];
  for (let i = 0; i < uuids.length; i += CHUNK_SIZE) {
    chunks.push(uuids.slice(i, i + CHUNK_SIZE));
  }

  await Promise.all(
    chunks.map(async (chunk) => {
      try {
        const entities = await loaders[type](chunk);
        const byUuid = new Map(entities.map((e) => [e.uuid, e]));
        chunk.forEach((uuid) => queue.get(uuid).resolve(byUuid.get(uuid)));
      } catch (e) {
        chunk.forEach((uuid) => queue.get(uuid).reject(e));
      }
    }),
  );
};

const enqueue = (type, uuid) => {
  if (!queues[type]) {
    queues[type] = new Map();
    setTimeout(() => flush(type), 0);
  }

  const queue = queues[type];
  if (!queue.has(uuid)) {
    let resolve, reject;
    const promise = new Promise((res, rej) => {
      resolve = res;
      reject = rej;
    });
    queue.set(uuid, { promise, resolve, reject });
  }

  return queue.get(uuid).promise;
};

/** Load one entity by uuid, batched with other calls made in the same tick. */
export const loadEntity = (type, uuid) => {
  if (!loaders[type]) throw new Error(`Unknown entity type: ${type}`);
  return enqueue(type, uuid);
};

/** Load many entities by uuid in batched list requests. Returns { [uuid]: entity }. */
export const loadEntities = async (type, uuids) => {
  const unique = [...new Set(uuids.filter(Boolean))];
  const entities = await Promise.all(unique.map((u) => loadEntity(type, u)));
  return Object.fromEntries(unique.map((u, i) => [u, entities[i]]));
};

/**
 * Component-level uuid -> entity map with batched loading.
 * `map[uuid]` is undefined until loaded (or if the entity wasn't found).
 * The map lives as long as the component: no cross-page caching.
 */
const useEntityMap = (type) => {
  const map = ref({});
  const isLoading = ref(false);
  const requested = new Set();
  let pending = 0;

  const load = async (uuids) => {
    const missing = [...new Set(uuids)].filter(
      (uuid) => uuid && !requested.has(uuid),
    );
    if (!missing.length) return map.value;

    missing.forEach((uuid) => requested.add(uuid));
    pending++;
    isLoading.value = true;

    try {
      const loaded = await loadEntities(type, missing);
      Object.assign(map.value, loaded);
    } catch (e) {
      // allow a retry on the next load
      missing.forEach((uuid) => requested.delete(uuid));
      console.error(`Failed to load ${type}`, e);
    } finally {
      pending--;
      isLoading.value = pending > 0;
    }

    return map.value;
  };

  // Forget loaded entities so the next load() fetches them again
  const reset = () => {
    requested.clear();
    map.value = {};
  };

  return { map, isLoading, load, reset };
};

export default useEntityMap;
