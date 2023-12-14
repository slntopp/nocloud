import api from "@/api.js";

export default {
  namespaced: true,
  state: {
    showcases: [],
    loading: false,
  },
  mutations: {
    setShowcases(state, showcases) {
      state.showcases = showcases;
    },
    pushShowcase(state, showcase) {
      const index = state.showcases.findIndex((a) => a.uuid === showcase.uuid);

      if (index !== -1) {
        state.showcases[index] = showcase;
      } else {
        state.showcases.push(showcase);
      }
    },
    removeShowcase(state, uuid) {
      state.showcases = state.showcases.filter((s) => s.uuid !== uuid);
    },
    replaceShowcase(state, value) {
      state.showcases = state.showcases.map((s) =>
        s.uuid === value.uuid ? value : s
      );
    },
    setLoading(state, data) {
      state.loading = data;
    },
  },
  actions: {
    async fetch({ commit }, params) {
      commit("setShowcases", []);
      commit("setLoading", true);

      try {
        const response = await api.showcases.list(params);
        commit("setShowcases", response.showcases);
      } catch (e) {
        throw new Error(e);
      } finally {
        commit("setLoading", false);
      }
    },
    async fetchById({ commit }, id) {
      commit("setLoading", true);

      try {
        const response = await api.showcases.get(id);
        commit("pushShowcase", response);
      } catch (e) {
        throw new Error(e);
      } finally {
        commit("setLoading", false);
      }
    },
    async delete({ commit }, uuid) {
      await api.showcases.delete(uuid);
      commit("removeShowcase", uuid);
    },
  },
  getters: {
    all(state) {
      return state.showcases;
    },
    isLoading(state) {
      return state.loading;
    },
  },
};
