import { createPromiseClient } from "@connectrpc/connect";
import { InstancesService } from "nocloud-proto/proto/es/instances/instances_connect";
import { ListInstancesRequest } from "nocloud-proto/proto/es/instances/instances_pb";
import api from "@/api.js";

export default {
  namespaced: true,
  state: {
    instances: [],
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
  },
  actions: {
    async fetch({ commit }, params) {
      commit("setInstances", []);
      commit("setLoading", true);
      try {
        console.log(ListInstancesRequest.fromJson(params));
        // const response = await getters["instancesClient"].list(
        //     ListInstancesRequest.fromJson(params)
        // );

        const response = await api.post("/instances", params);

        commit("setInstances", response.pool);
        commit("setTotal", response.count);
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
        rootGetters["app/transport"]
      );
    },
  },
};
