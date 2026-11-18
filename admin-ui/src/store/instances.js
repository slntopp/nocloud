import { createPromiseClient } from "@connectrpc/connect";
import { InstancesService } from "nocloud-proto/proto/es/instances/instances_connect";
import { ListInstancesRequest } from "nocloud-proto/proto/es/instances/instances_pb";

export default {
  namespaced: true,
  state: {
    instances: [],
    one: null,
    loading: false,
    total: 0,

    currentRequestId: null,
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
  },
  actions: {
    async fetch({ commit, state, getters }, params) {
      const requestId = Date.now();
      state.currentRequestId = requestId;

      commit("setLoading", true);
      try {
        const response = await getters["instancesClient"].list(
          ListInstancesRequest.fromJson(params),
        );

        if (requestId !== state.currentRequestId) return;

        const instances = response.pool.map((i) => ({
          ...i,
          ...i.instance.toJson(),
          instance: undefined,
        }));

        commit("setInstances", instances);
        commit("setTotal", Number(response.count));
        return instances;
      } finally {
        if (requestId === state.currentRequestId) {
          commit("setLoading", false);
        }
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
  },
  getters: {
    all(state) {
      return state.instances;
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
