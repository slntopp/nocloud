import { createPromiseClient } from "@connectrpc/connect";
import { AddonsService } from "nocloud-proto/proto/es/billing/billing_connect";
import {
  CountAddonsRequest,
  ListAddonsRequest,
} from "nocloud-proto/proto/es/billing/addons/addons_pb";

export default {
  namespaced: true,
  state: {
    addons: [],
    one: null,
    loading: false,
  },
  mutations: {
    setAddons(state, addons) {
      state.addons = addons;
    },
    setOne(state, addon) {
      state.one = addon;
    },
    setLoading(state, data) {
      state.loading = data;
    },
  },
  actions: {
    async fetch({ commit, getters }, options) {
      commit("setAddons", []);
      commit("setLoading", true);
      try {
        const response = await getters.addonsClient.list(
          ListAddonsRequest.fromJson(options)
        );
        commit("setAddons", response.addons);
      } finally {
        commit("setLoading", false);
      }
    },
    async count({ getters }, options) {
      return getters.addonsClient.count(CountAddonsRequest.fromJson(options));
    },
    async fetchById({ commit, getters }, id) {
      commit("setLoading", true);
      try {
        const response = await getters.addonsClient.get({ uuid: id });
        commit("setOne", response.toJson());
      } finally {
        commit("setLoading", false);
      }
    },
  },
  getters: {
    addonsClient(state, getters, rootState, rootGetters) {
      return createPromiseClient(AddonsService, rootGetters["app/transport"]);
    },
    all(state) {
      return state.addons;
    },
    one(state) {
      return state.one;
    },
    isLoading(state) {
      return state.loading;
    },
  },
};
