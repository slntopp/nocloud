import api from "@/api.js";

export default {
  namespaced: true,
  state: {
    services: [],
    service: [],
    instances: [],
    loading: false,
    loadingItem: false,
    total: 0,
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
    total(state) {
      return state.total;
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
    setTotal(state, value) {
      state.total = value;
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
      const service = state.services.find((el) => uuid === el.uuid);
      const igIndex = service.instancesGroups.findIndex(
        (ig) => !!ig.instances.find((inst) => value.uuid === inst.uuid)
      );
      const instIndex = service.instancesGroups[igIndex].instances.findIndex(
        (inst) => value.uuid === inst.uuid
      );
      const oldData =
        service.instancesGroups[igIndex].instances[instIndex][key];
      const newData = { ...oldData, ...value[key] };

      if (JSON.stringify(oldData) !== JSON.stringify(newData)) {
        service.instancesGroups[igIndex].instances[instIndex][key] = newData;
        this.commit("services/updateService", service);
      }
    },
    fetchByIdElem(state, data) {
      state.service = data;
    },
  },
  actions: {
    async fetch({ commit }, options) {
      commit("setInstances", []);
      commit("setServices", []);
      commit("setLoading", true);
      try {
        const response = await api.post("services", options);
        commit("setServices", response.pool);
        commit("setTotal", response.count);
        commit("setInstances", response.pool);
        return response;
      } finally {
        commit("setLoading", false);
      }
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
