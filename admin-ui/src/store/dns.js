import api from "@/api.js";

export default {
  namespaced: true,
  state: {
    zones: [],
    hosts: {},
    loading: false,
  },
  mutations: {
    setZones(state, zones) {
      state.zones = zones;
    },
    setLoading(state, data) {
      state.loading = data;
    },
    setHosts(state, data) {
      state.hosts = data;
    },
  },
  actions: {
    fetch({ commit }) {
      commit("setZones", []);
      commit("setLoading", true);
      return new Promise((resolve, reject) => {
        api.dns
          .list()
          .then((response) => {
            commit("setZones", response.zones);
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
    fetchHosts({ commit }, zone) {
      commit("setLoading", true);
      return new Promise((resolve, reject) => {
        api.dns
          .get(zone)
          .then((response) => {
            commit("setHosts", { [zone]: response.locations });
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
  getters: {
    all(state) {
      return state.zones;
    },
    isLoading(state) {
      return state.loading;
    },
    hosts(state) {
      return state.hosts;
    },
    getHost(state, getters) {
      return (dnsname) => {
        const hosts = getters["hosts"][dnsname];
        return hosts;
      };
    },
  },
};
