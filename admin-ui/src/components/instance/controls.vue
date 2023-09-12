<template>
  <div>
    <v-btn
      class="mr-2"
      v-for="btn in vmControlBtns"
      :key="btn.action + btn.title"
      :disabled="btn.disabled"
      :loading="isSendActionLoading"
      @click="
        sendVmAction({
          action: btn.action,
          template: { ...template, type: type },
          params: btn.data,
        })
      "
    >
      {{ btn.title || btn.action }}
    </v-btn>
    <confirm-dialog @confirm="deleteInstance">
      <v-btn class="mr-2" :loading="isLoading"> Delete </v-btn>
    </confirm-dialog>

    <confirm-dialog
      v-if="isBillingChange"
      text="Billing plan has changed, a new plan will be created"
      @confirm="save"
    >
      <v-btn
        class="mr-2"
        :loading="isSaveLoading"
        :color="isChanged ? 'primary' : ''"
      >
        Save
      </v-btn>
    </confirm-dialog>
    <v-btn
      v-else
      @click="save"
      class="mr-2"
      :loading="isSaveLoading"
      :color="isChanged ? 'primary' : ''"
    >
      Save
    </v-btn>
  </div>
</template>
<script>
import api from "@/api";
import snackbar from "@/mixins/snackbar.js";
import ConfirmDialog from "@/components/confirmDialog.vue";
import { getTodayFullDate } from "@/functions";
import { mapActions, mapGetters } from "vuex";

export default {
  name: "instance-actions",
  components: { ConfirmDialog },
  mixins: [snackbar],
  props: {
    template: { type: Object, required: true },
    copyTemplate: { type: Object },
    sp: { type: Object },
  },
  data: () => ({ isLoading: false, isSaveLoading: false }),
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
        setTimeout(() => {
          this.$router.push({ name: "Instances" });
        }, 100);
      } catch (err) {
        this.showSnackbarError({
          message: `Error: ${err?.response?.data?.message ?? "Unknown"}.`,
        });
      } finally {
        this.isLoading = false;
      }
    },
    async save() {
      const tempService = JSON.parse(JSON.stringify(this.service));
      const instance = JSON.parse(JSON.stringify(this.copyTemplate));
      const igIndex = tempService.instancesGroups.findIndex((ig) =>
        ig.instances.find((i) => i.uuid === this.template.uuid)
      );
      const instanceIndex = tempService.instancesGroups[
        igIndex
      ].instances.findIndex((i) => i.uuid === this.template.uuid);

      tempService.instancesGroups[igIndex].instances[instanceIndex] = instance;
      if (this.isBillingChange) {
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
  },
  computed: {
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
        virtual: [
          {
            action: "change_state",
            data: { state: 3 },
            title: "start",
            disabled: this.virtualActions.start,
          },
          {
            action: "change_state",
            data: { state: 2 },
            title: "stop",
            disabled: this.virtualActions.stop,
          },
          {
            action: "change_state",
            data: { state: 6 },
            title: "suspend",
            disabled: this.virtualActions.suspend,
          },
        ],
        opensrs: [{ action: "dns" }],
        cpanel: [{ action: "session" }],
      };

      return types[this.type];
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
    virtualActions() {
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
    getPlanTitle() {
      const type = this.template.type.includes("ovh")
        ? "ovh"
        : this.template.type;

      switch (type) {
        case "virtual":
        case "openai":
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
    product() {
      switch (this.template.type) {
        case "ovh": {
          return (
            this.template.config.duration + " " + this.template.config.planCode
          );
        }
        case "ione":
        case "virtual": {
          return this.template.product;
        }
      }

      return null;
    },
  },
};
</script>
