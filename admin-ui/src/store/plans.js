import api from "@/api.js";

export default {
  namespaced: true,
  state: {
    plans: [],
    plan: {},
    loading: false,
  },
  getters: {
    all(state) {
      return state.plans;
    },
    one(state) {
      return state.plan;
    },
    isLoading(state) {
      return state.loading;
    },
  },
  mutations: {
    setPlans(state, plans) {
      state.plans = plans;
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
    fetch({ commit }, params) {
      commit("setLoading", true);

      return new Promise((resolve, reject) => {
        api.plans
          .list(params)
          .then((response) => {
            commit("setPlans", response.pool);
            resolve(response);
          })
          .catch((error) => {
            reject(error);
          })
          .finally(() => {
            commit("setLoading", false);
          });
      });
    },
    fetchById({ commit }, id) {
      commit("setLoading", true);

      return new Promise((resolve, reject) => {
        api.plans
          .get(id)
          .then((response) => {
            commit("updatePlan", response);
            resolve(response);
          })
          .catch((error) => {
            reject(error);
          })
          .finally(() => {
            commit("setLoading", false);
          });
      });
    },
    fetchItem({ commit }, id) {
      commit("setLoading", true);

      return new Promise((resolve, reject) => {
        api.plans
          .get(id)
          .then((response) => {
            commit("setPlan", response);
            resolve(response);
          })
          .catch((error) => {
            reject(error);
          })
          .finally(() => {
            commit("setLoading", false);
          });
      });
    },
  },
};
