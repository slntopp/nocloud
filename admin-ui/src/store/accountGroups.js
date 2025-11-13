import api from "@/api.js";

export default {
  namespaced: true,
  state: {
    accountGroups: [],
    one: null,
    loading: false,
  },
  mutations: {
    setAccountGroups(state, accountGroups) {
      state.accountGroups = accountGroups;
    },
    pushAccountGroup(state, accountGroup) {
      const index = state.accountGroups.findIndex(
        (ag) => ag.uuid === accountGroup.uuid
      );

      if (index !== -1) {
        state.accountGroups[index] = accountGroup;
      } else {
        state.accountGroups.push(accountGroup);
      }
    },
    replaceAccountGroup(state, accountGroup) {
      const index = state.accountGroups.findIndex(
        (ag) => ag.uuid === accountGroup.uuid
      );

      if (index !== -1) {
        state.accountGroups.splice(index, 1, accountGroup);
      }
    },
    removeAccountGroup(state, uuid) {
      state.accountGroups = state.accountGroups.filter(
        (ag) => ag.uuid !== uuid
      );
    },
    setOne(state, accountGroup) {
      state.one = accountGroup;
    },
    setLoading(state, data) {
      state.loading = data;
    },
  },
  actions: {
    async fetch({ commit, state }) {
      if (state.accountGroups.length > 0 || state.loading) {
        return;
      }

      commit("setAccountGroups", []);
      commit("setLoading", true);
      try {
        const response = await api.get("account_groups");
        commit("setAccountGroups", response.pool);
      } finally {
        commit("setLoading", false);
      }
    },

    async fetchById({ commit }, id) {
      commit("setLoading", true);
      try {
        const accountGroup = await api.get(`account_groups/${id}`);
        commit("setOne", accountGroup);
      } finally {
        commit("setLoading", false);
      }
    },
    async create({ commit }, accountGroup) {
      commit("setLoading", true);
      try {
        const response = await api.post("account_groups", accountGroup);
        commit("pushAccountGroup", response);
      } finally {
        commit("setLoading", false);
      }
    },
    async update({ commit }, accountGroup) {
      commit("setLoading", true);

      try {
        const response = await api.patch(
          `account_groups/${accountGroup.uuid}`,
          accountGroup
        );

        commit("replaceAccountGroup", response);
        return response;
      } finally {
        commit("setLoading", false);
      }
    },
    async delete({ commit }, uuid) {
      commit("setLoading", true);
      try {
        await api.delete(`account_groups/${uuid}`);
        commit("removeAccountGroup", uuid);
      } finally {
        commit("setLoading", false);
      }
    },
  },
  getters: {
    all(state) {
      return state.accountGroups;
    },
    one(state) {
      return state.one;
    },
    isLoading(state) {
      return state.loading;
    },
  },
};
