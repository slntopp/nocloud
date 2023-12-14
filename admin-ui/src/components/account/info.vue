<template>
  <v-card elevation="0" color="background-light" class="pa-4">
    <div style="position: absolute; top: 0; right: 25px; max-width: 45%">
      <div>
        <v-chip class="ma-1" color="primary" outlined
          >Balance: {{ account.balance?.toFixed(2) || 0 }}
          {{ account.currency }}</v-chip
        >
        <v-btn
          class="ma-1"
          :disabled="isLocked"
          :to="{
            name: 'Transactions create',
            params: { account: account.uuid },
          }"
          >Create transaction/invoice</v-btn
        >
        <v-btn
          :disabled="isLocked"
          class="ma-1"
          :to="{
            name: 'Instance create',
            params: {
              accountId: account.uuid,
            },
          }"
        >
          Create instance
        </v-btn>
      </div>
      <div class="d-flex justify-end mt-3 align-center">
        <v-switch
          :loading="isChangeRegularPaymentLoading"
          :input-value="
            account.data?.regular_payment === undefined
              ? true
              : account.data?.regular_payment
          "
          @change="changeRegularPayment"
          label="Invoice based"
          class="mr-4"
        />
        <confirm-dialog
          v-for="button in stateButtons"
          :key="button.title"
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
            {{ button.title }}
          </v-btn>
        </confirm-dialog>
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
    <v-card-title class="px-0">Instances:</v-card-title>

    <instances-table :items="accountInstances" :show-select="false" />

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
      class="mt-4"
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

export default {
  name: "account-info",
  components: {
    LoginInAccountIcon,
    ConfirmDialog,
    InstancesTable,
    nocloudTable,
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
  }),
  methods: {
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
      try {
        await this.updateAccount({
          ...this.account,
          data: {
            ...this.account.data,
            regular_payment: value,
          },
        });
        this.$set(this.account.data, "regular_payment", value);
      } catch {
        this.showSnackbarError({
          message: "Error while change invoice based",
        });
      } finally {
        this.isChangeRegularPaymentLoading = false;
      }
    },
  },
  mounted() {
    this.title = this.account.title;
    this.currency = this.account.currency;
    this.uuid = this.account.uuid;
    this.keys = this.account.data?.ssh_keys || [];
    this.$store.dispatch("namespaces/fetch");
    this.$store.dispatch("services/fetch", { showDeleted: true });
    this.$store.dispatch("servicesProviders/fetch",{anonymously:true});
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
    accountInstances() {
      return this.instances.filter(
        (i) => i.access.namespace === this.accountNamespace?.uuid
      );
    },
    isCurrencyReadonly() {
      return this.account.currency && this.account.currency !== "NCU";
    },
    isLocked() {
      return this.account.status !== "ACTIVE";
    },
    stateButtons() {
      const status = this.account.status.toLowerCase();
      const permanentLock = {
        title: "Permanent lock",
        newStatusValue: "PERMANENT_LOCK",
        method: this.permanentLock,
      };

      switch (status) {
        case "lock": {
          return [{ title: "Unlock", newStatusValue: "ACTIVE" }, permanentLock];
        }
        case "active": {
          return [{ title: "Lock", newStatusValue: "LOCK" }, permanentLock];
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
</style>
