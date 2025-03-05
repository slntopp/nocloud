import api from "@/api";

const getCacheKey = (params) => JSON.stringify(params);

export default {
  namespaced: true,
  state: {
    cached: {},
    loading: false,
  },
  mutations: {
    setToCached(state, { key, value }) {
      state.cached[key] = value;
    },
    setLoading(state, val) {
      state.loading = !!val;
    },
  },
  actions: {
    async fetch({ commit, state }, params) {
      const cacheKey = getCacheKey(params);

      if (state.cached[cacheKey]) {
        return state.cached[cacheKey];
      }

      try {
        commit("setLoading", true);

        commit("setToCached", {
          key: cacheKey,
          value: api.get(`/statistic/${params.entity}`, {
            params: params.params,
          }),
        });

        const response = await state.cached[cacheKey];

        commit("setToCached", { key: cacheKey, value: response });

        return response;
      } catch {
        commit("setToCached", { key: cacheKey, value: null });
      } finally {
        commit(
          "setLoading",
          Object.keys(state.cached).some(
            (key) => state.cached[key] instanceof Promise
          )
        );
      }
    },
    get(_, params) {
      return api.get(`/statistic/${params.entity}`, {
        params: params.params,
      });
    },
  },
  getters: {
    loading(state) {
      return state.loading;
    },
    cached(state) {
      return (params) => {
        const cacheKey = getCacheKey(params);

        return state.cached[cacheKey];
      };
    },
  },
};
