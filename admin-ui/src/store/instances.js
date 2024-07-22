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
    async fetch({ commit, getters }, params) {
      commit("setInstances", []);
      commit("setLoading", true);
      try {
        const response = await getters["instancesClient"].list(
          ListInstancesRequest.fromJson(params)
        );
        commit(
          "setInstances",
          response.pool.map((i) => ({ ...i, instance: i.instance.toJson() }))
        );
        commit("setTotal", Number(response.count));

        return response.pool;
      } finally {
        commit("setLoading", false);
      }
    },
    async get({ commit, getters }, uuid) {
      commit("setLoading", true);
      try {
        const response = await getters["instancesClient"].get({ uuid });
        commit("setOne", response.toJson());
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

      const response = await state.cached.get(uuid);

      commit("setToCached", { instance: response.toJson(), uuid });

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
        rootGetters["app/transport"]
      );
    },
  },
};
