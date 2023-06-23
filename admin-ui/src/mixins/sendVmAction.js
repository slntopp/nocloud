import api from "@/api";
import { useStore } from "@/store";

const store = useStore();

const sendVmAction = {
  data: () => ({
    isActionLoading: false,
  }),
  methods: {
    sendVmAction(action, { uuid, type }, params) {
      if (action === "vnc") {
        return this.openVnc(uuid, type);
      }
      if (action === "dns") {
        return this.openDns(uuid);
      }

      this.isActionLoading = true;
      return api.instances
        .action({ uuid, action, params })
        .then((data) => {
          store.commit('snackbar/showSnackbarSuccess', { message: "Done!" });
          return data;
        })
        .catch((err) => {
          const opts = {
            message: `Error: ${err?.response?.data?.message ?? "Unknown"}.`,
          };
          store.commit('snackbar/showSnackbarError', opts);
        })
        .finally(() => {
          this.isActionLoading = false;
        });
    },
    async openVnc(uuid, type) {
      if (type === "ione") {
        this.$router.push({
          name: "Vnc",
          params: { instanceId: uuid },
        });
      } else {
        const data = await this.sendVmAction("start_vnc", { uuid });

        window.open(data.meta.url, "_blank");
      }
    },
    openDns(uuid) {
      this.$router.push({
        name: "InstanceDns",
        params: { instanceId: uuid },
      });
    },
  }
};

export default sendVmAction;
