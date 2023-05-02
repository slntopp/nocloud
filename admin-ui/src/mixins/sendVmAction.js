import api from "@/api";
import snackbar from "@/mixins/snackbar";

const sendVmAction = {
  data: () => ({
    isActionLoading: false,
  }),
  methods: {
    sendVmAction(action, uuid) {
      if (action === "vnc") {
        this.openVnc(uuid);
        return;
      }
      if (action === "dns") {
        this.openDns(uuid);
        return;
      }

      this.isActionLoading = true;
      api.instances
        .action({ uuid, action })
        .then(() => {
          this.showSnackbarSuccess({ message: "Done!" });
        })
        .catch((err) => {
          const opts = {
            message: `Error: ${err?.response?.data?.message ?? "Unknown"}.`,
          };
          this.showSnackbarError(opts);
        })
        .finally(() => {
          this.isActionLoading = false;
        });
    },
    openVnc(uuid) {
      this.$router.push({
        name: "Vnc",
        params: { instanceId: uuid },
      });
    },
    openDns(uuid) {
      this.$router.push({
        name: "InstanceDns",
        params: { instanceId: uuid },
      });
    },
  },
  mixins: [snackbar],
};

export default sendVmAction;
