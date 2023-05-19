<template>
  <div>
    <v-btn
      class="mr-2"
      v-for="btn in vmControlBtns"
      :key="btn.action"
      :disabled="btn.disabled"
      :loading="isActionLoading"
      @click="sendVmAction(btn.action, template)"
    >
      {{ btn.title || btn.action }}
    </v-btn>
    <confirm-dialog @confirm="deleteInstance">
      <v-btn class="mr-2" :loading="isLoading"> Delete </v-btn>
    </confirm-dialog>
    <v-btn class="mr-2" :loading="isSaveLoading" @click="save"> Save </v-btn>

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
import sendVmAction from "@/mixins/sendVmAction";

export default {
  name: "instance-actions",
  components: { ConfirmDialog },
  mixins: [snackbar, sendVmAction],
  props: { template: { type: Object, required: true } },
  data: () => ({ isLoading: false, isSaveLoading: false }),
  methods: {
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
    save() {
      const instance = this.template;
      const service = JSON.parse(JSON.stringify(this.service));

      const igIndex = service.instancesGroups.findIndex((ig) =>
        ig.instances.find((i) => i.uuid === this.template.uuid)
      );
      const instanceIndex = service.instancesGroups[
        igIndex
      ].instances.findIndex((i) => i.uuid === this.template.uuid);

      service.instancesGroups[igIndex].instances[instanceIndex] = instance;

      this.isSaveLoading = true;
      api.services
        ._update(service)
        .then(() => {
          this.showSnackbarSuccess({
            message: "Instance saved successfully",
          });

          this.$store.dispatch("services/fetch", this.template.uuid);
          this.$store.dispatch("servicesProviders/fetch");
        })
        .catch((err) => {
          this.showSnackbarError({ message: err });
        })
        .finally(() => {
          this.isSaveLoading = false;
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
          {
            action: "vnc",
            title: "Console",
            disabled: this.ovhActions?.reboot,
          },
        ],
        opensrs: [{ action: "dns" }],
        cpanel: [{ action: "session" }],
      };

      const type = this.template.billingPlan?.type.includes("ovh")
        ? "ovh"
        : this.template.billingPlan?.type;

      return types[type];
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
          this.template.state.meta.state === 3 &&
          ![18, 20].includes(this.template.state.meta.lcm_state),
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
          vnc: true,
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
          this.template.state.state === "RUNNING" &&
          this.template.state.state !== "STOPPED",
        suspend: this.template.state.state === "SUSPENDED",
        vnc: this.template.state.state !== "RUNNING",
      };
    },
    service() {
      return this.$store.getters["services/all"]?.find(
        (s) => s.uuid == this.template.service
      );
    },
  },
};
</script>
