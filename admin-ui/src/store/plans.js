import { ListRequest } from "nocloud-proto/proto/es/billing/billing_pb";
import { BillingService } from "nocloud-proto/proto/es/billing/billing_connect";
import { createPromiseClient } from "@connectrpc/connect";

export default {
  namespaced: true,
  state: {
    plans: [],
    total: 0,
    plan: {},
    loading: false,
  },
  getters: {
    all(state) {
      return state.plans;
    },
    total(state) {
      return state.total;
    },
    one(state) {
      return state.plan;
    },
    loading(state) {
      return state.loading;
    },
    plansClient(state, getters, rootState, rootGetters) {
      return createPromiseClient(BillingService, rootGetters["app/transport"]);
    },
  },
  mutations: {
    setPlans(state, plans) {
      state.plans = plans;
    },
    setTotal(state, val) {
      state.total = +val;
    },
    setPlan(state, plan) {
      state.plan = plan;
    },
    setLoading(state, data) {
      state.loading = data;
    },
    updatePlan(state, newPlan) {
      state.plan = state.plan.map((plan) =>
        newPlan.uuid === plan.uuid ? newPlan : plan
      );
    },
  },
  actions: {
    async fetch({ commit, getters }, options) {
      commit("setLoading", true);
      commit("setPlans", []);
      try {
        const response = await getters.plansClient.listPlans(
          ListRequest.fromJson(options)
        );

        const data = response.toJson();
        commit("setPlans", data.pool);
        commit("setTotal", data.total);
        return data.pool;
      } finally {
        commit("setLoading", false);
      }
    },
    async fetchById({ commit, getters }, id) {
      commit("setLoading", true);

      try {
        const response = await getters.plansClient.getPlan({ uuid: id });
        const data = response.toJson();
        commit("updatePlan", data);

        return data;
      } finally {
        commit("setLoading", false);
      }
    },
    async fetchItem({ commit, getters }, id) {
      commit("setLoading", true);

      try {
        const response = await getters.plansClient.getPlan({ uuid: id });
        const data = response.toJson();
        commit("setPlan", data);

        return data;
      } finally {
        commit("setLoading", false);
      }
    },
  },
};
