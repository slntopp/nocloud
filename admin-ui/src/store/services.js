import api from "@/api.js";

export default {
  namespaced: true,
  state: {
    services: [],
    service: [],
    instances: [],
    loading: false,
    loadingItem: false,
  },
  getters: {
    all(state) {
      return state.services.map((s) => ({
        ...s,
        title: s.title.startsWith("SRV_") ? s.title : `SRV_${s.title}`,
      }));
    },
    one(state) {
      return state.service;
    },
    getInstances(state) {
      return state.instances;
    },
    isLoading(state) {
      return state.loading;
    },
    isLoadingItem(state) {
      return state.loadingItem;
    },
  },
  mutations: {
    setServices(state, services) {
      state.services = services;
    },
    setLoading(state, data) {
      state.loading = data;
    },
    setLoadingItem(state, data) {
      state.loadingItem = data;
    },
    setService(state, service) {
      if (state.service.length) {
        let isProductExists = false;
        state.service.find((item) => {
          if (item.uuid === service.uuid) {
            isProductExists = true;
          }
        });
        if (!isProductExists) {
          state.service.push(service);
        }
      } else {
        state.service.push(service);
      }
    },
    setInstances(state, services) {
      state.instances = [];
      services.forEach(({ instancesGroups, uuid, access }) => {
        instancesGroups.forEach(({ instances, sp, type }) => {
          instances.forEach((inst) => {
            state.instances.push({ ...inst, service: uuid, sp, type, access });
          });
        });
      });
    },
    updateService(state, service) {
      if (!state.services.length) state.services.push(service);
      state.services = state.services.map((serv) =>
        serv.uuid === service.uuid ? service : serv
      );
      this.commit("services/setInstances", state.services);
    },
    updateInstance(state, { value, uuid, key = "state" }) {
      const i = state.services.findIndex((el) => uuid === el.uuid);
      const service = state.services[i];

      service.instancesGroups.forEach((el, i, groups) => {
        el.instances.forEach(({ uuid }, j) => {
          if (uuid === value.uuid) {
            groups[i].instances[j][key] = value[key];
          }
        });
      });
      this.commit("services/updateService", service);
    },
    fetchByIdElem(state, data) {
      state.service = data;
    },
  },
  actions: {
    fetch({ commit }, params) {
      commit("setLoading", true);
      commit("setInstances", []);
      commit("setServices", []);
      return new Promise((resolve, reject) => {
        api.services
          .list(params)
          .then((response) => {
            commit("setServices", response.pool);
            commit("setInstances", response.pool);
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
        api.services
          .get(id)
          .then((response) => {
            commit("updateService", response);
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
    fetchByIdItem({ commit }, id) {
      commit("setLoadingItem", true);
      return new Promise((resolve, reject) => {
        api.services
          .get(id)
          .then((response) => {
            commit("setService", response);
            resolve(response);
          })
          .catch((error) => {
            reject(error);
          })
          .finally(() => {
            commit("setLoadingItem", false);
          });
      });
    },
  },
};
