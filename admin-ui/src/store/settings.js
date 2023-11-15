import api from "@/api.js";

export default {
  namespaced: true,
  state: {
    keys: [],
    rawKeys: [],
    values: {},
    loading: false,
  },
  mutations: {
    setValues(state, values) {
      state.values = { ...state.values, ...values };
    },
    setKeys(state, keys) {
      state.keys = keys;
    },
    setRawKeys(state, rawKeys) {
      state.rawKeys = rawKeys;
    },
    setLoading(state, data) {
      state.loading = data;
    },
  },
  actions: {
    fetchKeys({ commit }) {
      commit("setLoading", true);
      return new Promise((resolve, reject) => {
        api.settings
          .list()
          .then((response) => {
            commit(
              "setRawKeys",
              response.pool.map((el) => el.key)
            );
            commit("setKeys", response.pool);
            resolve(response.pool);
          })
          .catch((error) => {
            reject(error);
          })
          .finally(() => {
            commit("setLoading", false);
          });
      });
    },
    fetchValues({ commit, state }, values = []) {
      commit("setLoading", true);
      const interestedKeys = values.length == 0 ? state.rawKeys : values;
      return new Promise((resolve, reject) => {
        if (interestedKeys.length !== 0) {
          api.settings
            .get(interestedKeys)
            .then((response) => {
              commit("setValues", response);
              resolve(response);
            })
            .catch((error) => {
              reject(error);
            })
            .finally(() => {
              commit("setLoading", false);
            });
        } else {
          commit("setLoading", false);
          resolve([]);
        }
      });
    },
    async fetch({ dispatch }) {
      await dispatch("fetchKeys");
      return new Promise((resolve, reject) => {
        dispatch("fetchValues").then(resolve).catch(reject);
      });
    },
  },
  getters: {
    all(state) {
      return state.keys.map((key) => {
        return { ...key, value: state.values[key.key] };
      });
    },
    values(state) {
      return state.values;
    },
    keys(state) {
      return state.keys;
    },
    isLoading(state) {
      return state.loading;
    },
    whmcsApi(state) {
      return JSON.parse(
        state.values["whmcs"] || "{}"
      ).api;
    },
  },
};
