import api from "@/api";

const sendVmAction = {
  data: () => ({
    isActionLoading: false,
  }),
  methods: {
    sendVmAction(action, { uuid, type }, data) {
      if (action === "vnc") {
        return this.openVnc(uuid, type);
      }
      if (action === "dns") {
        return this.openDns(uuid);
      }
      if (action === "open_ipmi") {
        return this.openIPMI(uuid);
      }

      this.isActionLoading = true;
      return api.instances
        .action({ uuid, action, params: data })
        .then((data) => {
          this.$store.commit("snackbar/showSnackbar", { message: "Done!" });
          return data;
        })
        .catch((err) => {
          const opts = {
            message: `Error: ${err?.response?.data?.message ?? "Unknown"}.`,
          };
          this.$store.commit("snackbar/showSnackbar", opts);
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
        window.open(data.meta.url, "_blanc");
      }
    },
    async openIPMI(uuid) {
      const { result, meta } = await this.sendVmAction("ipmi", { uuid });
      if (result) {
        window.open(meta.url, "_blanc");
      } else {
        this.$store.commit("snackbar/showSnackbarSuccess", {message:meta.message});
      }
    },
    openDns(uuid) {
      this.$router.push({
        name: "InstanceDns",
        params: { instanceId: uuid },
      });
    },
  },
};

export default sendVmAction;
