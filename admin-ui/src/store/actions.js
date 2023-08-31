import api from "@/api";
import { getClientIP } from "@/functions";
import router from "@/router";

export default {
  namespaced: true,
  state: {
    isSendActionLoading: false,
  },
  mutations: {
    sendIsSendActionLoading(state, val) {
      state.isSendActionLoading = val;
    },
  },
  actions: {
    async sendVmAction({ commit, dispatch }, { action, template, params }) {
      const { uuid, type } = template;
      if (action === "vnc") {
        return dispatch("openVnc", { uuid, type });
      }
      if (action === "dns") {
        return dispatch("openDns", { uuid });
      }
      if (action === "open_ipmi") {
        return dispatch("openIPMI", { uuid });
      }

      commit("sendIsSendActionLoading", true);

      try {
        const data = await api.instances.action({ uuid, action, params });
        commit(
          "snackbar/showSnackbarSuccess",
          { message: "Done!" },
          { root: true }
        );
        return data;
      } catch (err) {
        const opts = {
          message: `Error: ${err?.response?.data?.message ?? "Unknown"}.`,
        };
        commit("snackbar/showSnackbarError", opts, { root: true });
      } finally {
        commit("sendIsSendActionLoading", false);
      }
    },
    async openVnc({ dispatch }, { uuid, type }) {
      let action = "start_vnc";
      if (type === "ione") {
        router.push({
          name: "Vnc",
          params: { instanceId: uuid },
        });
      } else {
        if (type === "ovh cloud") {
          action = "start_vnc_vm";
        }
        const data = await dispatch("sendVmAction", {
          action,
          template: { uuid, type },
        });
        window.open(data.meta.url, "_blanc");
      }
    },
    async openIPMI({ dispatch, commit }, { uuid }) {
      const { result, meta } = await dispatch("sendVmAction", {
        action: "ipmi",
        template: { uuid },
        params: { ip: await getClientIP() },
      });
      if (result) {
        window.open(meta.url, "_blanc");
      } else {
        commit(
          "snackbar/showSnackbarSuccess",
          {
            message: meta.message,
          },
          { root: true }
        );
      }
    },
    openDns(uuid) {
      router.push({
        name: "InstanceDns",
        params: { instanceId: uuid },
      });
    },
  },
  getters: {
    isSendActionLoading(state) {
      return state.sendIsSendActionLoading;
    },
  },
};
