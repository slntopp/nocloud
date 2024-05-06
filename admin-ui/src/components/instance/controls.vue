<template>
  <div class="controls">
    <template v-for="btn in vmControlBtns">
      <instance-control-btn
        v-if="!btn.component"
        :hint="btn.title || btn.action"
        :key="btn.action + btn.title"
      >
        <v-btn
          class="ma-1"
          :loading="runningActionName === btn.action"
          @click="btn.type === 'method' ? btn.method() : sendAction(btn)"
        >
          <v-icon>
            {{ btn.icon }}
          </v-icon>
        </v-btn>
      </instance-control-btn>
      <component
        v-else
        :is="btn.component"
        :key="btn.action + btn.title"
        :disabled="
          btn.disabled ||
          (!!runningActionName && runningActionName !== btn.action) ||
          isDeleted
        "
        :loading="runningActionName === btn.action"
        :template="template"
        @click="
          btn.type === 'method' ? btn.method($event) : sendAction(btn, $event)
        "
      />
    </template>

    <instance-control-btn hint="Playbook">
      <v-dialog style="height: 100%" v-if="isAnsibleActive && !isDeleted">
        <template v-slot:activator="{ on, attrs }">
          <v-btn class="ma-1" v-bind="attrs" v-on="on">
            <v-icon>mdi-book</v-icon>
          </v-btn>
        </template>
        <plugin-iframe
          style="height: 80vh"
          :params="{ instances: [template] }"
          :url="ansiblePlaybookUrl"
        />
      </v-dialog>
    </instance-control-btn>

    <instance-control-btn
      :hint="isPending ? 'CREATE INVOICE' : 'Renewal invoice'"
    >
      <confirm-dialog @confirm="sendInvoice">
        <v-btn class="ma-1" :loading="isInvoiceLoading">
          <v-icon>mdi-invoice-text-outline</v-icon>
        </v-btn>
      </confirm-dialog>
    </instance-control-btn>

    <instance-control-btn
      :hint="template.data.lock ? 'User unlock' : 'User lock'"
    >
      <confirm-dialog @confirm="lockInstance" :disabled="isDeleted">
        <v-btn :loading="isLockLoading" class="ma-1" :disabled="isDeleted">
          <v-icon>
            {{ template.data.lock ? "mdi-lock-off" : "mdi-lock" }}
          </v-icon>
        </v-btn>
      </confirm-dialog>
    </instance-control-btn>

    <instance-control-btn hint="Terminate">
      <confirm-dialog :disabled="isDeleted" @confirm="deleteInstance">
        <v-btn class="ma-1" :disabled="isDeleted" :loading="isDeleteLoading">
          <v-icon> mdi-delete </v-icon>
        </v-btn>
      </confirm-dialog>
    </instance-control-btn>

    <div class="save_button">
      <instance-control-btn top hint="Save">
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
              <v-icon> mdi-content-save </v-icon>
            </v-btn>
          </template>
          <v-card color="background-light">
            <v-card-title
              >Do you really want to change your current price
              model?</v-card-title
            >
            <v-card-subtitle class="mt-1"
              >You can also create a new price model based on the current
              one.</v-card-subtitle
            >
            <v-card-actions class="d-flex justify-end">
              <v-btn
                class="mr-2"
                :loading="isDeleteLoading"
                @click="isBillingDialog = false"
              >
                Close
              </v-btn>
              <v-btn
                class="mr-2"
                :loading="isDeleteLoading"
                @click="save(true)"
              >
                Create
              </v-btn>
              <v-btn :loading="isDeleteLoading" @click="save(false)">
                Edit
              </v-btn>
            </v-card-actions>
          </v-card>
        </v-dialog>
      </instance-control-btn>
    </div>
  </div>
</template>
<script>
import api from "@/api";
import snackbar from "@/mixins/snackbar.js";
import ConfirmDialog from "@/components/confirmDialog.vue";
import { getTodayFullDate } from "@/functions";
import { mapActions, mapGetters } from "vuex";
import PluginIframe from "@/components/plugin/iframe.vue";
import InstanceControlBtn from "@/components/ui/hintBtn.vue";

export default {
  name: "instance-actions",
  components: { InstanceControlBtn, PluginIframe, ConfirmDialog },
  mixins: [snackbar],
  props: {
    template: { type: Object, required: true },
    account: { type: Object, required: true },
    copyTemplate: { type: Object },
    sp: { type: Object },
  },
  data: () => ({
    isDeleteLoading: false,
    isSaveLoading: false,
    isInvoiceLoading: false,
    isLockLoading: false,
    runningActionName: "",
    isBillingDialog: false,
  }),
  methods: {
    ...mapActions("actions", ["sendVmAction"]),
    async deleteInstance() {
      this.isDeleteLoading = true;
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
        this.isDeleteLoading = false;
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

          this.$emit("refresh");
        })
        .catch((err) => {
          this.showSnackbarError({ message: err });
        })
        .finally(() => {
          this.isSaveLoading = false;
        });
    },
    async freezeInstance(date) {
      this.sendAction(this.freezeBtn, {date})
    },
    onSaveClick() {
      if (!this.isChanged) {
        return;
      }

      if (!this.isBillingChange && !this.isDeleted) {
        this.save();
      }
    },
    async save(createNewPlan = false, instance) {
      if (!instance) {
        instance = JSON.parse(JSON.stringify(this.copyTemplate));
      }

      const tempService = JSON.parse(JSON.stringify(this.service));
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
          meta: {
            ...(this.copyTemplate.billingPlan.meta || {}),
            isIndividual: true,
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
          resources:
            this.type === "opensrs"
              ? { ...this.copyTemplate.billingPlan.resources }
              : [
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

          this.$emit("refresh");
        })
        .catch((err) => {
          this.showSnackbarError({ message: err });
        })
        .finally(() => {
          this.isSaveLoading = false;
        });
    },
    async startInstance(instance) {
      if (!instance) {
        instance = JSON.parse(JSON.stringify(this.template));
      }

      try {
        await this.save(
          false,
          JSON.parse(
            JSON.stringify({
              ...instance,
              config: { ...instance.config, auto_start: true },
            })
          )
        );

        await api.services.up(this.template.service);
      } catch (e) {
        this.$store.commit("snackbar/showSnackbarError", {
          message: e.response?.data?.message || "Error during start instance",
        });
      }
    },
    async rebuildVps(os) {
      await this.sendAction(
        this.vmControlBtns.find((a) => a.action === "rebuild"),
        { imageId: os.id }
      );

      const tempService = JSON.parse(JSON.stringify(this.service));
      const instance = JSON.parse(JSON.stringify(this.template));
      const igIndex = tempService.instancesGroups.findIndex((ig) =>
        ig.instances.find((i) => i.uuid === this.template.uuid)
      );
      const instanceIndex = tempService.instancesGroups[
        igIndex
      ].instances.findIndex((i) => i.uuid === this.template.uuid);

      instance.config.configuration = {
        ...instance.config.configuration,
        vps_os: os.name,
      };
      tempService.instancesGroups[igIndex].instances[instanceIndex] = instance;
      await api.services._update(tempService);
      this.$emit("refresh");
    },
    async sendAction(btn, data) {
      this.runningActionName = btn.action;
      try {
        await this.sendVmAction({
          action: btn.action,
          template: { ...this.template, type: this.type },
          params: btn.data || data,
        });
      } finally {
        this.runningActionName = "";
      }
    },
    async sendInvoice() {
      this.isInvoiceLoading = true;
      try {
        await fetch(
          /https:\/\/(.+?\.?\/)/.exec(this.whmcsApi)[0] +
            `modules/addons/nocloud/api/index.php?run=create_renewal_invoice&account=${this.account.uuid}&instance=${this.template.uuid}`
        );
      } catch (e) {
        this.$store.commit("snackbar/showSnackbarError", {
          message:
            e.response?.data?.message || "Error during create renewal invoice",
        });
      } finally {
        this.isInvoiceLoading = false;
      }
    },
  },
  computed: {
    ...mapGetters("actions", ["isSendActionLoading"]),
    type() {
      return this.template.billingPlan.type;
    },
    baseVmControls() {
      return [
        this.isDetached
          ? {
              action: "attach",
              title: "Show in user app",
              type: "method",
              icon: "mdi-paperclip",
              method: this.attachInstance,
            }
          : {
              action: "detach",
              title: "Hide in user app",
              icon: "mdi-paperclip-off",
              type: "method",
              method: () => this.attachInstance(true),
            },
      ];
    },
    freezeBtn() {
      return {
        action: "freeze",
        title: "freeze",
        type: "method",
        method: this.freezeInstance,
        component: () => import("@/components/dialogs/freezeInstance.vue"),
      };
    },
    vmControlBtns() {
      const types = {
        ione: [
          {
            action: "start",
            type: "method",
            component: () => import("@/components/dialogs/startInstance.vue"),
            method: this.startInstance,
            disabled: this.ioneActions?.start,
          },
          {
            action: "poweroff",
            disabled: this.ioneActions?.poweroff,
            icon: "mdi-stop",
          },
          {
            action: "resume",
            disabled: this.ioneActions?.resume,
            icon: "mdi-play",
          },
          this.freezeBtn,
          {
            action: "suspend",
            disabled: this.ioneActions?.suspend,
            icon: "mdi-power-sleep",
          },
          {
            action: "reboot",
            disabled: this.ioneActions?.reboot,
            icon: "mdi-restart",
          },
          {
            action: "vnc",
            title: "Console", //not reqired, use 'action' for a name if not found
            disabled: this.ioneActions?.vnc,
            icon: "mdi-console",
          },
          ...this.baseVmControls,
        ],
        "ovh dedicated": [
          {
            action: "start",
            type: "method",
            component: () => import("@/components/dialogs/startInstance.vue"),
            method: this.startInstance,
            disabled: this.ovhActions?.start,
          },
          { action: "poweroff", disabled: true, icon: "mdi-stop" },
          { action: "resume", disabled: true, icon: "mdi-play" },
          this.freezeBtn,
          { action: "suspend", disabled: true, icon: "mdi-power-sleep" },
          { action: "reboot", disabled: true, icon: "mdi-restart" },
          {
            action: "open_ipmi",
            title: "console",
            disabled: this.ovhActions?.reboot,
            icon: "mdi-console",
          },
          ...this.baseVmControls,
        ],
        "ovh cloud": [
          {
            action: "stop_vm",
            title: "poweroff",
            icon: "mdi-stop",
            disabled: this.ovhActions?.poweroff,
          },
          {
            action: "resume_vm",
            title: "resume",
            icon: "mdi-play",
            disabled: this.ovhActions?.resume,
          },
          this.freezeBtn,
          {
            action: "suspend_vm",
            title: "suspend",
            disabled: this.ovhActions?.suspend,
            icon: "mdi-power-sleep",
          },
          {
            action: "reboot_vm",
            title: "reboot",
            component: () => import("@/components/dialogs/rebootInstance.vue"),
            disabled: this.ovhActions?.reboot,
          },
          {
            action: "start_vm",
            title: "poweron",
            icon: "mdi-power",
            disabled: this.ovhActions?.resume,
          },
          {
            action: "start_vnc_vm",
            title: "Console",
            disabled: this.ovhActions?.reboot,
            icon: "mdi-console",
          },
          ...this.baseVmControls,
        ],
        "ovh vps": [
          {
            action: "start",
            type: "method",
            method: this.startInstance,
            component: () => import("@/components/dialogs/startInstance.vue"),
            disabled: this.ovhActions?.start,
          },
          {
            action: "poweroff",
            disabled: this.ovhActions?.poweroff,
            icon: "mdi-stop",
          },
          {
            action: "resume",
            disabled: this.ovhActions?.resume,
            icon: "mdi-play",
          },
          this.freezeBtn,
          {
            action: "suspend",
            disabled: this.ovhActions?.suspend,
            icon: "mdi-power-sleep",
          },
          {
            action: "reboot",
            disabled: this.ovhActions?.reboot,
            icon: "mdi-restart",
          },
          {
            action: "rebuild",
            type: "method",
            method: this.rebuildVps,
            component: () => import("@/components/dialogs/rebuildVps.vue"),
            disabled: this.ovhActions?.rebuild,
            icon: "mdi-account-convert",
          },
          {
            action: "vnc",
            title: "Console",
            disabled: this.ovhActions?.reboot,
            icon: "mdi-console",
          },
          ...this.baseVmControls,
        ],
        empty: [
          {
            action: "change_state",
            data: { state: 3 },
            title: "start",
            component: () => import("@/components/dialogs/startInstance.vue"),
            disabled: this.emptyActions?.start,
          },
          {
            action: "change_state",
            data: { state: 2 },
            title: "stop",
            icon: "mdi-stop",
            disabled: this.emptyActions?.stop,
          },
          this.freezeBtn,
          {
            action: "change_state",
            data: { state: 6 },
            title: "suspend",
            icon: "mdi-power-sleep",
            disabled: this.emptyActions?.suspend,
          },
          ...this.baseVmControls,
        ],
        keyweb: [
          {
            action: "auto_start",
            type: "method",
            title: "start",
            component: () => import("@/components/dialogs/startInstance.vue"),
            method: this.startInstance,
            disabled: !this.keywebActions?.auto_start,
          },
          {
            action: "start",
            title: "resume",
            icon: "mdi-play",
            disabled: !this.keywebActions?.start,
          },
          {
            action: "stop",
            title: "stop",
            icon: "mdi-stop",
            disabled: !this.keywebActions?.stop,
          },
          {
            action: "reboot",
            title: "reboot",
            icon: "mdi-restart",
            disabled: !this.keywebActions?.reboot,
          },
          this.freezeBtn,
          {
            action: "suspend",
            title: "suspend",
            disabled: !this.keywebActions?.suspend,
            icon: "mdi-power-sleep",
          },
          {
            action: "unsuspend",
            title: "unsuspend",
            disabled: !this.keywebActions?.unsuspend,
            icon: "mdi-weather-sunny",
          },
          {
            action: "vnc",
            title: "Console",
            icon: "mdi-console",
            disabled: !this.keywebActions?.vnc,
          },
          ...this.baseVmControls,
        ],
        opensrs: [{ action: "dns", icon: "mdi-dns", ...this.baseVmControls }],
        cpanel: [
          {
            action: "start",
            type: "method",
            component: () => import("@/components/dialogs/startInstance.vue"),
            method: this.startInstance,
            disabled: this.template.config.auto_start,
          },
          { action: "session", icon: "mdi-console" },
          ...this.baseVmControls,
        ],
      };

      return (
        types[this.type]?.map((b) => ({ ...b, type: b.type || "action" })) || []
      );
    },
    ioneActions() {
      if (
        !this.template?.state ||
        !this.template.config.auto_start ||
        this.template.state?.meta?.state === 1 ||
        this.isDetached
      ) {
        return {
          start: this.template.config.auto_start,
          resume: true,
          poweroff: true,
          reboot: true,
          suspend: true,
          vnc: true,
        };
      }
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
        start: true,
      };
    },
    ovhActions() {
      if (!this.template?.state) return;
      if (this.template.state.state === "PENDING")
        return {
          start: this.template.config.auto_start,
          poweroff: true,
          reboot: true,
          resume: true,
          suspend: true,
          vnc: true,
          rebuild: true,
        };
      const isRebootDisabled =
        this.template.state.state === "SUSPENDED" ||
        this.template.state.meta.state === "BUILD" ||
        this.template.state.state === "STOPPED";

      return {
        poweroff:
          this.template.state.state === "SUSPENDED" ||
          (this.template.state.state !== "RUNNING" &&
            this.template.state.state === "STOPPED"),
        reboot: isRebootDisabled,
        rebuild: isRebootDisabled,
        resume:
          this.template.state.state === "RUNNING" &&
          this.template.state.state !== "STOPPED",
        suspend: this.template.state.state === "SUSPENDED",
        vnc: this.template.state.state !== "RUNNING",
        start: this.template.config.auto_start,
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
      if (!this.template?.state) return;

      const state = this.template?.state.state;

      switch (state) {
        case "RUNNING": {
          return {
            stop: true,
            reboot: true,
            suspend: true,
            vnc: true,
            auto_start: !this.template.config.auto_start,
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
          return {
            auto_start: !this.template.config.auto_start,
          };
        }
        case "SUSPENDED": {
          return {
            auto_start: !this.template.config.auto_start,
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
        case "cpanel":
        case "keyweb":
        case "opensrs":
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
        case "cpanel":
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
    isPending() {
      return this.template.state.state === "PENDING";
    },
    namespace() {
      return this.$store.getters["namespaces/all"].find(
        (n) => n.uuid === this.template.access.namespace
      );
    },
    whmcsApi() {
      return this.$store.getters["settings/whmcsApi"];
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

<style>
.save_button {
  position: fixed;
  top: 140px;
  right: 40px;
}
</style>
