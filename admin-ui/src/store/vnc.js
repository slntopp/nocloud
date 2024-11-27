import api from "@/api.js";

export default {
  namespaced: true,
  state: {
    services: [],
    instances: [],
    servicesFull: [],
    loading: false,
    loadingInvoke: false,
  },
  mutations: {
    setServices(state, services) {
      state.services = services;
    },
    setInstances(state, data) {
      state.instances = state.instances.filter(
        ({ uuidService }) => uuidService !== data.uuid
      );
      data.instancesGroups.forEach((group) => {
        group.instances.forEach((inst) => {
          state.instances.push({
            ...inst,
            uuidService: data.uuid,
            uuidInstancesGroups: group.uuid,
            type: group.type,
            sp: group.sp,
          });
        });
      });
    },
    setServicesFull(state, data) {
      if (state.servicesFull.length) {
        let servicesFull = false;
        state.servicesFull.forEach((item) => {
          if (item.uuid === data.uuid) {
            servicesFull = true;
          }
        });
        if (!servicesFull) {
          state.servicesFull.push(data);
        }
      } else {
        state.servicesFull.push(data);
      }
    },
    setLoading(state, data) {
      state.loading = data;
    },
    setLoadingInvoke(state, data) {
      state.loadingInvoke = data;
    },
  },
  actions: {
    actionVMInvoke({ commit }, data) {
      return new Promise((resolve, reject) => {
        api.instances
          .action(data)
          .then((response) => {
            resolve(response);
          })
          .catch((err) => {
            reject(err);
          })
          .finally(() => {
            commit("setLoadingInvoke", false);
          });
      });
    },
  },
  getters: {
    all(state) {
      return state.services;
    },
    isLoading: (state) => state.loading,

    getInstances(state) {
      return state.instances;
    },
  },
};
