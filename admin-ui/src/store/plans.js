import api from "@/api.js";

export default {
  namespaced: true,
  state: {
    plans: [],
    plan: {},
    loading: false,
    instanceCountLoading: false,
    instanceCountMap: {},
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
    isInstanceCountLoading(state) {
      return state.instanceCountLoading;
    },
    instanceCountMap(state) {
      return state.instanceCountMap;
    },
  },
  mutations: {
    setPlans(state, plans) {
      state.plans = plans;
    },
    setPlan(state, plan) {
      state.plan = plan;
    },
    setIsInstanceCountLoading(state, val) {
      state.instanceCountLoading = val;
    },
    setIsInstanceCountMap(state, map) {
      state.instanceCountMap = map;
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
    fetch(
      { commit, dispatch },
      options = { params: { anonymously: false }, withCount: false }
    ) {
      console.log(options)
      if (!options?.silent) {
        commit("setPlans", []);
        commit("setLoading", true);
      }

      if (options.withCount) {
        dispatch("fetchCount", options);
      }

      return new Promise((resolve, reject) => {
        api.plans
          .list(options?.params)
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
    fetchCount({ commit }, options) {
      commit("setIsInstanceCountLoading", true);
      return api.plans
        .instancesCountMap(options?.params)
        .then((response) => {
          commit("setIsInstanceCountMap", response.plans);
        })
        .finally(() => {
          commit("setIsInstanceCountLoading", false);
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
