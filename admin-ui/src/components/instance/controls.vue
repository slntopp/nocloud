<template>
  <div class="controls">
    <v-btn
      class="ma-1"
      v-for="btn in vmControlBtns"
      :key="btn.action + btn.title"
      :disabled="
        btn.disabled ||
        (!!runningActionName && runningActionName !== btn.action) ||
        isDeleted
      "
      :loading="runningActionName === btn.action"
      @click="btn.type === 'method' ? btn.method() : sendAction(btn)"
    >
      {{ btn.title || btn.action }}
    </v-btn>
    <v-dialog style="height: 100%" v-if="isAnsibleActive && !isDeleted">
      <template v-slot:activator="{ on, attrs }">
        <v-btn class="ma-1" v-bind="attrs" v-on="on"> Playbook </v-btn>
      </template>
      <plugin-iframe
        style="height: 80vh"
        :params="{ instances: [template] }"
        :url="ansiblePlaybookUrl"
      />
    </v-dialog>
    <confirm-dialog @confirm="lockInstance" :disabled="isDeleted">
      <v-btn :loading="isLockLoading" class="ma-1" :disabled="isDeleted">
        {{ template.data.lock ? "User unlock" : "User lock" }}
      </v-btn>
    </confirm-dialog>
    <confirm-dialog :disabled="isDeleted" @confirm="deleteInstance">
      <v-btn class="ma-1" :disabled="isDeleted" :loading="isLoading">
        Terminate
      </v-btn>
    </confirm-dialog>

    <v-dialog persistent v-model="isBillingDialog" max-width="600px">
      <template v-slot:activator="{ on, attrs }">
        <v-btn
          v-bind="isBillingChange && !isDeleted ? attrs : undefined"
          v-on="isBillingChange && !isDeleted ? on : undefined"
          :disabled="isDeleted"
          @click="onSaveClick"
          class="ma-1"
          :loading="isSaveLoading"
          :color="isChanged ? 'primary' : ''"
        >
          Save
        </v-btn>
      </template>
      <v-card color="background-light">
        <v-card-title
          >Do you really want to change your current price model?</v-card-title
        >
        <v-card-subtitle class="mt-1"
          >You can also create a new price model based on the current
          one.</v-card-subtitle
        >
        <v-card-actions class="d-flex justify-end">
          <v-btn
            class="mr-2"
            :loading="isLoading"
            @click="isBillingDialog = false"
          >
            Close
          </v-btn>
          <v-btn class="mr-2" :loading="isLoading" @click="save(true)">
            Create
          </v-btn>
          <v-btn :loading="isLoading" @click="save(false)"> Edit </v-btn>
        </v-card-actions>
      </v-card>
    </v-dialog>
  </div>
</template>
<script>
import api from "@/api";
import snackbar from "@/mixins/snackbar.js";
import ConfirmDialog from "@/components/confirmDialog.vue";
import { getTodayFullDate } from "@/functions";
import { mapActions, mapGetters } from "vuex";
import PluginIframe from "@/components/plugin/iframe.vue";
import { is } from "date-fns/locale";

export default {
  name: "instance-actions",
  components: { PluginIframe, ConfirmDialog },
  mixins: [snackbar],
  props: {
    template: { type: Object, required: true },
    copyTemplate: { type: Object },
    sp: { type: Object },
  },
  data: () => ({
    isLoading: false,
    isSaveLoading: false,
    isLockLoading: false,
    runningActionName: "",
    isBillingDialog: false,
  }),
  methods: {
    ...mapActions("actions", ["sendVmAction"]),
    async deleteInstance() {
      this.isLoading = true;
      try {
        await api.delete(`/instances/${this.template.uuid}`);
        if (this.template.type === "ione") {
          const tempService = JSON.parse(JSON.stringify(this.service));
          const instance = JSON.parse(JSON.stringify(this.template));
          const igIndex = tempService.instancesGroups.findIndex((ig) =>
            ig.instances.find((i) => i.uuid === this.template.uuid)
          );
          Object.keys(tempService.instancesGroups[igIndex].resources).forEach(
            (key) => {
              if (instance.resources[key]) {
                tempService.instancesGroups[igIndex].resources[key] -=
                  instance.resources[key];
              }
            }
          );

          await api.services._update(tempService);
        }

        this.showSnackbarSuccess({ message: "Done!" });
        this.$router.push({ name: "Instances" });
      } catch (err) {
        this.showSnackbarError({
          message: `Error: ${err?.response?.data?.message ?? "Unknown"}.`,
        });
      } finally {
        this.isLoading = false;
      }
    },
    async attachInstance(detach = false) {
      const action = detach ? "detach" : "attach";
      this.runningActionName = action;

      try {
        await api.delete(`/instances/${action}/${this.template.uuid}`);
        this.showSnackbarSuccess({ message: "Done!" });
        this.$router.push({ name: "Instances" });
      } catch (err) {
        this.showSnackbarError({
          message: `Error: ${err?.response?.data?.message ?? "Unknown"}.`,
        });
      } finally {
        this.runningActionName = "";
      }
    },
    lockInstance() {
      const lock = !this.template.data.lock;

      const tempService = JSON.parse(JSON.stringify(this.service));
      const igIndex = tempService.instancesGroups.findIndex((ig) =>
        ig.instances.find((i) => i.uuid === this.template.uuid)
      );
      const instanceIndex = tempService.instancesGroups[
        igIndex
      ].instances.findIndex((i) => i.uuid === this.template.uuid);

      tempService.instancesGroups[igIndex].instances[instanceIndex] = {
        ...this.template,
        data: { ...this.template.data, lock },
      };

      this.isLockLoading = true;
      api.services
        ._update(tempService)
        .then(() => {
          this.showSnackbarSuccess({
            message: `Instance ${lock ? "lock" : "unlock"} successfully`,
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
    onSaveClick() {
      if (!this.isChanged) {
        return;
      }

      if (!this.isBillingChange && !this.isDeleted) {
        this.save();
      }
    },
    async save(createNewPlan = false) {
      const tempService = JSON.parse(JSON.stringify(this.service));
      const instance = JSON.parse(JSON.stringify(this.copyTemplate));
      const igIndex = tempService.instancesGroups.findIndex((ig) =>
        ig.instances.find((i) => i.uuid === this.template.uuid)
      );
      const instanceIndex = tempService.instancesGroups[
        igIndex
      ].instances.findIndex((i) => i.uuid === this.template.uuid);

      tempService.instancesGroups[igIndex].instances[instanceIndex] = instance;

      if (this.isBillingChange && createNewPlan) {
        const title = this.getPlanTitle(this.template);
        const billingPlan = {
          ...this.copyTemplate.billingPlan,
          title,
          products: {
            [this.product]: this.product
              ? this.copyTemplate.billingPlan.products[this.product]
              : undefined,
          },
          public: false,
        };
        delete billingPlan.uuid;
        this.isSaveLoading = true;

        try {
          const data = await api.plans.create(billingPlan);
          await api.servicesProviders.bindPlan(this.sp.uuid, [data.uuid]);
          tempService.instancesGroups[igIndex].instances[
            instanceIndex
          ].billingPlan = data;
        } catch (e) {
          this.$store.commit("snackbar/showSnackbarError", {
            message:
              e.response?.data?.message ||
              "Error during create individual plan",
          });
        }
      } else if (this.isBillingChange) {
        const title = this.getPlanTitle(this.template);
        const ogPlan = this.$store.getters["plans/all"].find(
          (p) => p.uuid === this.copyTemplate.billingPlan.uuid
        );
        const updatedPlan = {
          ...ogPlan,
          ...this.copyTemplate.billingPlan,
          products: {
            ...ogPlan.products,
            ...this.copyTemplate.billingPlan.products,
          },
          resources: [
            ...ogPlan.resources,
            ...this.copyTemplate.billingPlan.resources,
          ],
          title,
        };

        this.isSaveLoading = true;
        try {
          const data = await api.plans.update(updatedPlan.uuid, updatedPlan);
          tempService.instancesGroups[igIndex].instances[
            instanceIndex
          ].billingPlan = data;
        } catch (e) {
          this.$store.commit("snackbar/showSnackbarError", {
            message: e.response?.data?.message || "Error during update plan",
          });
        }
      }

      this.isSaveLoading = true;
      api.services
        ._update(tempService)
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
    async sendAction(btn) {
      this.runningActionName = btn.action;
      try {
        await this.sendVmAction({
          action: btn.action,
          template: { ...this.template, type: this.type },
          params: btn.data,
        });
      } finally {
        this.runningActionName = "";
      }
    },
  },
  computed: {
    is() {
      return is;
    },
    ...mapGetters("actions", ["isSendActionLoading"]),
    type() {
      return this.template.billingPlan.type;
    },
    ovhButtons() {
      return [
        { action: "poweroff", disabled: this.ovhActions?.poweroff },
        { action: "resume", disabled: this.ovhActions?.resume },
        { action: "suspend", disabled: this.ovhActions?.suspend },
        { action: "reboot", disabled: this.ovhActions?.reboot },
      ];
    },
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
          this.isDetached
            ? { action: "attach", type: "method", method: this.attachInstance }
            : {
                action: "detach",
                type: "method",
                method: () => this.attachInstance(true),
              },
        ],
        "ovh dedicated": [
          { action: "poweroff", disabled: true },
          { action: "resume", disabled: true },
          { action: "suspend", disabled: true },
          { action: "reboot", disabled: true },
          {
            action: "open_ipmi",
            title: "console",
            disabled: this.ovhActions?.reboot,
          },
        ],
        "ovh cloud": [
          ...this.ovhButtons,
          {
            action: "vnc",
            title: "Console",
            disabled: this.ovhActions?.reboot,
          },
        ],
        "ovh vps": [
          ...this.ovhButtons,
          {
            action: "vnc",
            title: "Console",
            disabled: this.ovhActions?.reboot,
          },
        ],
        empty: [
          {
            action: "change_state",
            data: { state: 3 },
            title: "start",
            disabled: this.emptyActions.start,
          },
          {
            action: "change_state",
            data: { state: 2 },
            title: "stop",
            disabled: this.emptyActions.stop,
          },
          {
            action: "change_state",
            data: { state: 6 },
            title: "suspend",
            disabled: this.emptyActions.suspend,
          },
        ],
        keyweb: [
          {
            action: "start",
            title: "start",
            disabled: !this.keywebActions.start,
          },
          {
            action: "stop",
            title: "stop",
            disabled: !this.keywebActions.stop,
          },
          {
            action: "reboot",
            title: "reboot",
            disabled: !this.keywebActions.reboot,
          },
          {
            action: "suspend",
            title: "suspend",
            disabled: !this.keywebActions.suspend,
          },
          {
            action: "unsuspend",
            title: "unsuspend",
            disabled: !this.keywebActions.unsuspend,
          },
        ],
        opensrs: [{ action: "dns" }],
        cpanel: [{ action: "session" }],
      };

      return (
        types[this.type]?.map((b) => ({ ...b, type: b.type || "action" })) || []
      );
    },
    ioneActions() {
      if (!this.template?.state) return;
      if (this.template.state.meta.state === 1 || this.isDetached)
        return {
          resume: true,
          poweroff: true,
          reboot: true,
          suspend: true,
          vnc: true,
        };
      return {
        poweroff:
          this.template.state.meta.state === 5 ||
          (this.template.state.meta.state !== 3 &&
            [0, 18, 20].includes(this.template.state.meta.lcm_state)),
        reboot:
          this.template.state.meta.lcm_state === 6 ||
          this.template.state.meta.lcm_state === 21 ||
          this.template.state.meta.state === 5 ||
          (this.template.state.meta.state !== 3 &&
            (this.template.state.meta.lcm_state === 18 ||
              this.template.state.meta.lcm_state === 20)) ||
          (this.template.state.meta.lcm_state === 0 &&
            this.template.state.meta.state === 8),
        resume:
          this.template.state.meta.lcm_state === 21 ||
          this.template.state.meta.lcm_state === 6 ||
          (this.template.state.meta.state === 3 &&
            ![18, 20].includes(this.template.state.meta.lcm_state)),
        suspend:
          this.template.state.meta.state === 5 ||
          this.template.state.meta.lcm_state === 21 ||
          this.template.state.meta.lcm_state === 6,
        vnc:
          this.template.state.meta.state === 5 ||
          this.template.state.meta.lcm_state === 21 ||
          this.template.state.meta.lcm_state === 6,
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
    emptyActions() {
      if (!this.template?.state || this.template.state.state === "PENDING")
        return {
          stop: true,
          suspend: true,
        };
      return {
        stop: this.template.state.state === "INIT",
        suspend: this.template.state.state === "SUSPENDED",
        start: this.template.state.state === "RUNNING",
      };
    },
    keywebActions() {
      const state = this.template?.state.state;

      switch (state) {
        case "RUNNING": {
          return {
            stop: true,
            reboot: true,
            suspend: true,
          };
        }
        case "STOPPED": {
          return {
            start: true,
            reboot: true,
            suspend: true,
          };
        }
        case "DELETED":
        case "PENDING":
        case "OPERATION": {
          return {};
        }
        case "SUSPENDED": {
          return {
            unsuspend: true,
          };
        }
        default: {
          return {
            stop: true,
            reboot: true,
            suspend: true,
          };
        }
      }
    },
    getPlanTitle() {
      const type = this.template.type.includes("ovh")
        ? "ovh"
        : this.template.type;

      switch (type) {
        case "empty":
        case "openai":
        case "keyweb":
        case "ione": {
          return (item) => {
            let planTitle = `IND_${this.sp.title}_${
              item.billingPlan.title
            }_${getTodayFullDate()}`;
            if (item.billingPlan.title.startsWith("IND_")) {
              const titleKeys = item.billingPlan.title.split("_");
              titleKeys[3] = getTodayFullDate();
              planTitle = titleKeys.join("_");
            }
            return planTitle;
          };
        }
        case "ovh": {
          return (item) => {
            let planTitle = `IND_${item.title}_${getTodayFullDate()}`;

            if (item.billingPlan.title.startsWith("IND_")) {
              const titleKeys = item.billingPlan.title.split("_");
              titleKeys[2] = getTodayFullDate();
              planTitle = titleKeys.join("_");
            }
            return planTitle;
          };
        }
        default: {
          return null;
        }
      }
    },
    service() {
      return this.$store.getters["services/all"]?.find(
        (s) => s.uuid == this.template.service
      );
    },
    isChanged() {
      return (
        JSON.stringify(this.template) !== JSON.stringify(this.copyTemplate)
      );
    },
    isBillingChange() {
      return (
        JSON.stringify(this.copyTemplate.billingPlan) !==
        JSON.stringify(this.template.billingPlan)
      );
    },
    isDetached() {
      return this.template?.status?.toLowerCase() === "detached";
    },
    product() {
      switch (this.template.type) {
        case "ovh": {
          return (
            this.template.config.duration + " " + this.template.config.planCode
          );
        }
        case "ione":
        case "openai":
        case "keyweb":
        case "empty": {
          return this.template.product;
        }
      }

      return null;
    },
    plugins() {
      return this.$store.getters["plugins/all"];
    },
    ansiblePlugin() {
      return this.plugins.find((p) =>
        p.title.toLowerCase().includes("ansible")
      );
    },
    isAnsibleActive() {
      const allowedTypes = ["ione"];
      return allowedTypes.includes(this.template.type) && !!this.ansiblePlugin;
    },
    ansiblePlaybookUrl() {
      if (!this.isAnsibleActive) {
        return;
      }

      return `${this.ansiblePlugin.url}playbooks-preview`;
    },
    isDeleted() {
      return this.template.state?.state === "DELETED";
    },
  },
};
</script>

<style scoped lang="scss">
.controls {
  max-width: calc(100% - 450px);
  min-width: 60%;
}
</style>
