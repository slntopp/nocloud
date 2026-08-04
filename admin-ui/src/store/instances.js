import { createPromiseClient } from "@connectrpc/connect";
import { InstancesService } from "nocloud-proto/proto/es/instances/instances_connect";
import { ListInstancesRequest } from "nocloud-proto/proto/es/instances/instances_pb";

export default {
  namespaced: true,
  state: {
    instances: [],
    cached: new Map(),
    one: null,
    loading: false,
    total: 0,

    fetchSeq: 0,
    activeFetches: 0,
  },
  mutations: {
    setInstances(state, instances) {
      state.instances = instances;
    },
    setOne(state, instance) {
      state.one = instance;
    },
    setTotal(state, total) {
      state.total = +total;
    },
    setLoading(state, data) {
      if (state.loading !== data) {
        console.log("[instances/store] setLoading", state.loading, "→", data);
      }
      state.loading = data;
    },
    setCached(state, data) {
      state.cached = data;
    },
    setToCached(state, { instance, uuid }) {
      state.cached.set(uuid, instance);
    },
  },
  actions: {
    async fetch({ commit, state, getters }, params) {
      const requestId = ++state.fetchSeq;
      state.activeFetches++;
      console.log("[instances/fetch] START", {
        requestId,
        activeFetches: state.activeFetches,
        loading: state.loading,
        params: JSON.parse(JSON.stringify(params)),
      });
      console.trace("[instances/fetch] caller");
      commit("setLoading", true);
      try {
        const response = await getters["instancesClient"].list(
          ListInstancesRequest.fromJson(params),
        );

        if (requestId !== state.fetchSeq) {
          console.warn("[instances/fetch] STALE drop", {
            requestId,
            currentSeq: state.fetchSeq,
            activeFetches: state.activeFetches,
            count: Number(response.count),
            pool: response.pool?.length,
          });
          return;
        }

        const instances = response.pool.map((i) => ({
          ...i,
          ...i.instance.toJson(),
          instance: undefined,
        }));

        console.log("[instances/fetch] APPLY", {
          requestId,
          count: Number(response.count),
          pool: instances.length,
          states: instances.map((i) => i.state?.state),
        });
        commit("setInstances", instances);
        commit("setTotal", Number(response.count));
        return instances;
      } finally {
        state.activeFetches = Math.max(0, state.activeFetches - 1);
        if (state.activeFetches === 0) {
          commit("setLoading", false);
        }
        console.log("[instances/fetch] END", {
          requestId,
          activeFetches: state.activeFetches,
          loading: state.loading,
          storeTotal: state.total,
          storeLen: state.instances.length,
        });
      }
    },
    async get({ commit, getters }, uuid) {
      commit("setLoading", true);
      try {
        const response = await getters["instancesClient"].get({ uuid });
        const data = response.toJson();
        commit("setOne", {
          ...data,
          instance: {
            ...data.instance,
            config: !data.instance.config ? {} : data.instance.config,
          },
        });

        return data;
      } finally {
        commit("setLoading", false);
      }
    },
    async fetchToCached({ state, commit, getters }, uuid) {
      if (state.cached.has(uuid)) {
        return state.cached.get(uuid);
      }

      commit("setToCached", {
        instance: getters["instancesClient"].get({ uuid }),
        uuid,
      });

      const response = (await state.cached.get(uuid)).toJson();

      commit("setToCached", {
        instance: {
          ...response.instance,
          ...response,
          instance: undefined,
        },
        uuid,
      });

      return response;
    },
  },
  getters: {
    all(state) {
      return state.instances;
    },
    cached(state) {
      return state.cached;
    },
    one(state) {
      return state.one;
    },
    isLoading(state) {
      return state.loading;
    },
    total(state) {
      return state.total;
    },
    instancesClient(state, getters, rootState, rootGetters) {
      return createPromiseClient(
        InstancesService,
        rootGetters["app/transport"],
      );
    },
  },
};
