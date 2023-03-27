<template>
  <div>
    <v-btn
      class="mr-2"
      v-for="btn in vmControlBtns"
      :key="btn.action"
      :disabled="btn.disabled"
      :loading="isLoading"
      @click="sendVmAction(btn.action)"
    >
      {{ btn.title || btn.action }}
    </v-btn>
    <confirm-dialog @confirm="deleteInstance">
      <v-btn :loading="isLoading"> Delete </v-btn>
    </confirm-dialog>

    <v-snackbar
      v-model="snackbar.visibility"
      :timeout="snackbar.timeout"
      :color="snackbar.color"
    >
      {{ snackbar.message }}
      <template v-if="snackbar.route && Object.keys(snackbar.route).length > 0">
        <router-link :to="snackbar.route"> Look up. </router-link>
      </template>

      <template v-slot:action="{ attrs }">
        <v-btn
          :color="snackbar.buttonColor"
          text
          v-bind="attrs"
          @click="snackbar.visibility = false"
        >
          Close
        </v-btn>
      </template>
    </v-snackbar>
  </div>
</template>
<script>
import api from "@/api";
import snackbar from "@/mixins/snackbar.js";
import ConfirmDialog from "@/components/confirmDialog.vue";

export default {
  name: "instance-actions",
  components: { ConfirmDialog },
  mixins: [snackbar],
  props: { template: { type: Object, required: true } },
  data: () => ({ isLoading: false }),
  methods: {
    sendVmAction(action) {
      if (action === "vnc") {
        this.openVnc();
        return;
      }
      if (action === "dns") {
        this.openDns();
        return;
      }

      this.isLoading = true;
      api.instances
        .action({ uuid: this.template.uuid, action })
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
          this.isLoading = false;
        });
    },
    openVnc() {
      this.$router.push({
        name: "Vnc",
        params: { instanceId: this.template.uuid },
      });
    },
    openDns() {
      this.$router.push({
        name: "InstanceDns",
        params: { instanceId: this.template.uuid },
      });
    },
    deleteInstance() {
      this.isLoading = true;
      api
        .delete(`/instances/${this.template.uuid}`)
        .then(() => {
          this.showSnackbarSuccess({ message: "Done!" });

          setTimeout(() => {
            this.$router.push({ name: "Instances" });
          }, 100);
        })
        .catch((err) => {
          this.showSnackbarError({
            message: `Error: ${err?.response?.data?.message ?? "Unknown"}.`,
          });
        })
        .finally(() => {
          this.isLoading = false;
        });
    },
  },
  computed: {
    vmControlBtns() {
      const types = {
        ione: [
          { action: "poweroff", disabled: this.ioneActions?.poweroff },
          { action: "resume", disabled: this.ioneActions?.resume },
          { action: "suspend", disabled: this.ioneActions?.suspend },
          { action: "reboot", disabled: this.ioneActions?.reboot },
          {
            action: "vnc",
            title: "Console", //not reqired, use 'action' for a name if not found
            disabled: this.ioneActions?.vnc,
          },
        ],
        ovh: [
          { action: "poweroff", disabled: this.ovhActions?.poweroff },
          { action: "resume", disabled: this.ovhActions?.resume },
          { action: "suspend", disabled: this.ovhActions?.suspend },
          { action: "reboot", disabled: this.ovhActions?.reboot },
        ],
        opensrs: [{ action: "dns" }],
        cpanel: [{ action: "session" }],
      };

      return types[this.template.billingPlan?.type];
    },
    ioneActions() {
      if (!this.template?.state) return;
      if (this.template.state.meta.state === 1)
        return {
          resume: true,
          poweroff: true,
          reboot: true,
          suspend: true,
        };
      return {
        poweroff:
          this.template.state.meta.state === 5 ||
          (this.template.state.meta.state !== 3 &&
            [0, 18, 20].includes(this.template.state.meta.lcm_state)),
        reboot:
          this.template.state.meta.state === 5 ||
          (this.template.state.meta.state !== 3 &&
            (this.template.state.meta.lcm_state === 18 ||
              this.template.state.meta.lcm_state === 20)) ||
          (this.template.state.meta.lcm_state === 0 &&
            this.template.state.meta.state === 8),
        resume:
          this.template.state.meta.state === 5 ||
          (this.template.state.meta.state === 3 &&
            ![18, 20].includes(this.template.state.meta.lcm_state)),
        suspend: this.template.state.meta.state === 5,
        vnc: this.template.state.meta.state === 5,
      };
    },
    ovhActions() {
      if (!this.template?.state) return;
      if (this.template.state.state === "PENDING")
        return {
          poweroff: true,
          reboot: true,
          resume: true,
          suspend: true,
        };
      return {
        poweroff:
          this.template.state.state === "SUSPENDED" ||
          (this.template.state.state !== "RUNNING" &&
            this.template.state.state === "STOPPED"),
        reboot:
          this.template.state.state === "SUSPENDED" ||
          this.template.state.meta.state === "BUILD" ||
          this.template.state.state === "STOPPED",
        resume:
          this.template.state.state === "SUSPENDED" ||
          (this.template.state.state === "RUNNING" &&
            this.template.state.state !== "STOPPED"),
        suspend: this.template.state.state === "SUSPENDED",
      };
    },
  },
};
</script>
