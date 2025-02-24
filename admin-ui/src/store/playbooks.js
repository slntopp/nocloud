import api from "@/api";

export default {
  namespaced: true,
  state: {
    playbooks: [],
  },
  mutations: {
    setPlaybooks(state, data) {
      state.playbooks = data;
    },
  },
  actions: {
    async fetch({ commit, state }) {
      if (state.playbooks.length) {
        return;
      }

      try {
        const data = await api.get("/ansible/ansible/playbooks");

        commit("setPlaybooks", data.playbooks);
      } catch (e) {
        console.log(e);
      }
    },
  },
  getters: {
    all(state) {
      return state.playbooks;
    },
  },
};
