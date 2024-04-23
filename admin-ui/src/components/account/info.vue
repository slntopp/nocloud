<template>
  <v-card elevation="0" color="background-light" class="pa-4">
    <div style="z-index: 0; position: relative; top: -25px; right: 40px">
      <div class="d-flex justify-end mt-1 align-center flex-wrap">
        <hint-btn hint="Create instance">
          <v-btn
            class="ma-1"
            :disabled="isLocked"
            :to="{
              name: 'Instance create',
              params: {
                accountId: account.uuid,
              },
            }"
          >
            <v-icon>mdi-server</v-icon>
          </v-btn>
        </hint-btn>
        <hint-btn hint="Invoice based">
          <v-dialog v-model="isChangeRegularPaymentOpen" max-width="500">
            <template v-slot:activator="{ on, attrs }">
              <v-btn
                :disabled="isChangeRegularPaymentLoading"
                :loading="isChangeRegularPaymentLoading"
                class="ma-1"
                v-bind="attrs"
                v-on="on"
              >
                <v-icon>mdi-invoice-check-outline</v-icon>
              </v-btn>
            </template>
            <v-card color="background-light pa-5">
              <v-card-actions class="d-flex justify-center">
                <v-btn class="mr-2" @click="changeRegularPayment(false)">
                  Disable to all
                </v-btn>
                <v-btn class="mr-2" @click="changeRegularPayment(true)">
                  Enable to all</v-btn
                >
              </v-card-actions>
            </v-card>
          </v-dialog>
        </hint-btn>

        <hint-btn
          v-for="button in stateButtons"
          :key="button.title"
          :hint="button.hint"
        >
          <confirm-dialog
            @confirm="
              button.method
                ? button.method()
                : changeStatus(button.newStatusValue)
            "
          >
            <v-btn
              :loading="button.newStatusValue === statusChangeValue"
              class="mr-2"
            >
              <v-icon>{{ button.icon }}</v-icon>
            </v-btn>
          </confirm-dialog>
        </hint-btn>
        <hint-btn hint="Create transaction/invoice">
          <v-chip @click="openTransaction" class="ma-1" color="primary" outlined
            >Balance: {{ account.balance?.toFixed(2) || 0 }}
            {{ account.currency }}</v-chip
          >
        </hint-btn>
      </div>
    </div>

    <v-row>
      <v-col cols="2">
        <v-text-field v-model="uuid" readonly label="UUID" />
      </v-col>
      <v-col cols="2">
        <v-text-field v-model="title" label="name" style="width: 330px">
          <template v-slot:append>
            <login-in-account-icon :uuid="account.uuid" />
          </template>
        </v-text-field>
      </v-col>
      <v-col cols="2">
        <v-select
          :readonly="isCurrencyReadonly"
          :items="currencies"
          v-model="currency"
          label="currency"
          style="width: 330px"
        />
      </v-col>
    </v-row>

    <nocloud-expansion-panels
      class="account-additional"
      title="Additional info"
    >
      <v-row>
        <v-col lg="3" md="4" sm="6">
          <v-text-field readonly :value="account.data?.email" label="Email" />
        </v-col>

        <v-col lg="3" md="4" sm="6">
          <v-text-field
            readonly
            :value="account.data?.company"
            label="Company"
          />
        </v-col>

        <v-col lg="3" md="4" sm="6">
          <v-text-field readonly :value="account.data?.phone" label="Phone" />
        </v-col>

        <v-col lg="3" md="4" sm="6">
          <v-text-field
            readonly
            :value="formatSecondsToDate(account.data?.date_create || 0)"
            label="Date of create"
          />
        </v-col>

        <v-col lg="3" md="4" sm="6">
          <v-text-field
            readonly
            :value="account.data?.country"
            label="Country"
          />
        </v-col>

        <v-col lg="3" md="4" sm="6">
          <v-text-field readonly :value="account.data?.city" label="City" />
        </v-col>

        <v-col lg="3" md="4" sm="6">
          <v-text-field
            readonly
            :value="account.data?.address"
            label="Address"
          />
        </v-col>

        <v-col lg="1" md="2" sm="4">
          <v-text-field
            readonly
            :value="account.data?.whmcs_id"
            label="WHMCS id"
          >
            <template v-slot:append>
              <whmcs-btn :account="account" />
            </template>
          </v-text-field>
        </v-col>
      </v-row>
    </nocloud-expansion-panels>

    <div class="d-flex align-center">
      <v-card-title class="px-0 instances-panel">Instances:</v-card-title>
      <v-switch
        class="ml-3 mt-5"
        dense
        label="Show deleted"
        v-model="showDeletedInstances"
      />
    </div>
    <instances-table
      :items="filteredInstances"
      no-search
      :show-select="false"
    />

    <v-card-title class="px-0">SSH keys:</v-card-title>

    <div class="pt-4">
      <v-menu
        bottom
        offset-y
        transition="slide-y-transition"
        v-model="isVisible"
        :close-on-content-click="false"
      >
        <template v-slot:activator="{ on, attrs }">
          <v-btn class="mr-2" v-bind="attrs" v-on="on"> Create </v-btn>
        </template>
        <v-card class="pa-4">
          <v-row>
            <v-col>
              <v-text-field
                dense
                label="title"
                v-model="newKey.title"
                :rules="generalRule"
              />
            </v-col>
          </v-row>
          <v-row>
            <v-col>
              <v-textarea
                dense
                label="key"
                v-model="newKey.value"
                :rules="generalRule"
              />
            </v-col>
          </v-row>
          <v-row>
            <v-col>
              <v-btn @click="addKey"> Send </v-btn>
            </v-col>
          </v-row>
        </v-card>
      </v-menu>

      <v-btn class="mr-8" :disabled="selected.length < 1" @click="deleteKeys">
        Delete
      </v-btn>
    </div>

    <nocloud-table
      table-name="account-info"
      item-key="value"
      v-model="selected"
      :items="keys"
      :headers="headers"
    />

    <v-btn class="mt-4 mr-2" :loading="isEditLoading" @click="editAccount">
      Save
    </v-btn>
  </v-card>
</template>

<script>
import config from "@/config.js";
import api from "@/api.js";
import snackbar from "@/mixins/snackbar.js";
import nocloudTable from "@/components/table.vue";
import InstancesTable from "@/components/instances_table.vue";
import ConfirmDialog from "@/components/confirmDialog.vue";
import LoginInAccountIcon from "@/components/ui/loginInAccountIcon.vue";
import NocloudExpansionPanels from "@/components/ui/nocloudExpansionPanels.vue";
import hintBtn from "@/components/ui/hintBtn.vue";
import { formatSecondsToDate } from "@/functions";
import whmcsBtn from "@/components/ui/whmcsBtn.vue";

export default {
  name: "account-info",
  components: {
    NocloudExpansionPanels,
    LoginInAccountIcon,
    ConfirmDialog,
    InstancesTable,
    nocloudTable,
    hintBtn,
    whmcsBtn,
  },
  mixins: [snackbar],
  props: ["account"],
  data: () => ({
    newKey: { title: "", value: "" },
    headers: [
      { text: "Title", value: "title" },
      { text: "Key", value: "value" },
    ],
    generalRule: [(v) => !!v || "Required field"],
    navTitles: config.navTitles ?? {},
    uuid: "",
    title: "",
    currency: "",
    keys: [],
    selected: [],
    isVisible: false,
    isEditLoading: false,
    statusChangeValue: "",
    isChangeRegularPaymentLoading: false,
    isChangeRegularPaymentOpen: false,
    showDeletedInstances: false,
  }),
  methods: {
    formatSecondsToDate,
    navTitle(title) {
      if (title && this.navTitles[title]) {
        return this.navTitles[title];
      }

      return title;
    },
    addKey() {
      this.keys.push(this.newKey);
      this.isVisible = false;
      this.newKey = { title: "", value: "" };
    },
    deleteKeys() {
      if (this.selected.length < 1) return;
      const arr = this.selected.map((el) => el.value);

      this.keys = this.keys.filter((el) => !arr.includes(el.value));
      this.selected = [];
    },
    updateAccount(newAccount) {
      return api.accounts.update(this.account.uuid, newAccount).catch((err) => {
        this.showSnackbarError({ message: err });
      });
    },
    async editAccount() {
      const newAccount = {
        ...this.account,
        title: this.title,
        currency: this.currency,
      };
      if (!newAccount.data) {
        newAccount.data = {};
      }
      newAccount.data.ssh_keys = this.keys;

      this.isEditLoading = true;
      try {
        await this.updateAccount(newAccount);
        this.showSnackbarSuccess({
          message: "Account edited successfully",
        });

        this.$router.push({ name: "Accounts" });
      } finally {
        this.isEditLoading = false;
      }
    },
    async changeStatus(newStatus) {
      this.statusChangeValue = newStatus;
      try {
        await fetch(
          /https:\/\/(.+?\.?\/)/.exec(this.whmcsApi)[0] +
            `modules/addons/nocloud/api/index.php?run=status_user&account=${
              this.account.uuid
            }&status=${newStatus === "ACTIVE" ? "open" : "close"}`
        );
        await this.updateAccount({ ...this.account, status: newStatus });
        this.$set(this.account, "status", newStatus);
        this.showSnackbarSuccess({
          message: "Status change successfully",
        });
      } finally {
        this.statusChangeValue = "";
      }
    },
    async permanentLock() {
      const newStatus = "PERMANENT_LOCK";
      this.statusChangeValue = newStatus;
      try {
        const accountServices = this.services.filter(
          (s) => s.access.namespace === this.accountNamespace?.uuid
        );

        const servicesForDown = accountServices.filter(
          (s) => s.status !== "INIT"
        );
        await Promise.all(
          servicesForDown.map((s) => api.services.down(s.uuid))
        );
        await Promise.all(
          accountServices.map((s) => api.services.delete(s.uuid))
        );
        await this.changeStatus(newStatus);
      } catch {
        this.showSnackbarError({
          message: "Error while change status",
        });
      } finally {
        this.statusChangeValue = "";
      }
    },
    async changeRegularPayment(value) {
      this.isChangeRegularPaymentLoading = true;
      this.isChangeRegularPaymentOpen = false;
      try {
        const services = [];

        this.accountsByInstance.forEach((instance) => {
          const tempService =
            services.find((s) => s.uuid === instance.service) ||
            JSON.parse(
              JSON.stringify(
                this.services.find((s) => s.uuid === instance.service)
              )
            );
          const igIndex = tempService.instancesGroups.findIndex((ig) =>
            ig.instances.find((i) => i.uuid === instance.uuid)
          );
          const instanceIndex = tempService.instancesGroups[
            igIndex
          ].instances.findIndex((i) => i.uuid === instance.uuid);

          instance.config.regular_payment = value;

          tempService.instancesGroups[igIndex].instances[instanceIndex] =
            instance;

          const sIndex = services.findIndex((s) => s.uuid === instance.service);
          if (sIndex !== -1) {
            services[sIndex] = tempService;
          } else {
            services.push(tempService);
          }
        });
        await Promise.all(services.map((s) => api.services._update(s)));
      } catch {
        this.showSnackbarError({
          message: "Error while change invoice based",
        });
      } finally {
        this.isChangeRegularPaymentLoading = false;
      }
    },
    openTransaction() {
      this.$router.push({
        name: "Transactions create",
        params: { account: this.account.uuid },
      });
    },
  },
  mounted() {
    this.title = this.account.title;
    this.currency = this.account.currency;
    this.uuid = this.account.uuid;
    this.keys = this.account.data?.ssh_keys || [];
    this.$store.dispatch("namespaces/fetch");
    this.$store.dispatch("services/fetch", { showDeleted: true });
    this.$store.dispatch("servicesProviders/fetch", { anonymously: true });
  },
  computed: {
    namespaces() {
      return this.$store.getters["namespaces/all"];
    },
    services() {
      return this.$store.getters["services/all"];
    },
    currencies() {
      return this.$store.getters["currencies/all"].filter((c) => c !== "NCU");
    },
    servicesProviders() {
      return this.$store.getters["servicesProviders/all"];
    },
    instances() {
      return this.$store.getters["services/getInstances"];
    },
    accountNamespace() {
      return this.namespaces.find(
        (n) => n.access.namespace === this.account?.uuid
      );
    },
    accountsByInstance() {
      return this.instances.filter(
        (i) => i.access.namespace === this.accountNamespace?.uuid
      );
    },
    filteredInstances() {
      if (this.showDeletedInstances) {
        return this.accountsByInstance;
      }

      return this.accountsByInstance.filter(
        (inst) => inst.state?.state !== "DELETED"
      );
    },
    isCurrencyReadonly() {
      return this.account.currency && this.account.currency !== "NCU";
    },
    isLocked() {
      return this.account.status !== "ACTIVE";
    },
    stateButtons() {
      const status = this.account.status?.toLowerCase();
      const permanentLock = {
        hint: "Delete user",
        icon: "mdi-delete",
        newStatusValue: "PERMANENT_LOCK",
        method: this.permanentLock,
      };

      switch (status) {
        case "lock": {
          return [
            {
              hint: "Unlock access",
              newStatusValue: "ACTIVE",
              icon: "md-lock-off",
            },
            permanentLock,
          ];
        }
        case "active": {
          return [
            { hint: "Block access", icon: "mdi-lock", newStatusValue: "LOCK" },
            permanentLock,
          ];
        }
        default: {
          return [];
        }
      }
    },
    whmcsApi() {
      return this.$store.getters["settings/whmcsApi"];
    },
  },
};
</script>

<style scoped lang="scss">
.page__title {
  color: var(--v-primary-base);
  font-weight: 400;
  font-size: 32px;
  font-family: "Quicksand", sans-serif;
  line-height: 1em;
  margin-bottom: 10px;
}

.instances-panel {
  @media (max-width: 1300px) {
    margin-top: 25px;
  }
}

.account-additional {
  @media (max-width: 1600px) {
    margin-top: 50px;
  }
  @media (max-width: 1250px) {
    margin-top: 100px;
  }
}
</style>
