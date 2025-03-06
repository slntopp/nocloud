<template>
  <div class="pa-4">
    <h1 class="page__title">Create transaction</h1>
    <v-form v-model="isValid" ref="form">
      <v-row>
        <v-col lg="6" cols="12">
          <v-row align="center">
            <v-col cols="3">
              <v-subheader>Type</v-subheader>
            </v-col>
            <v-col cols="9">
              <v-select
                :disabled="isEdit"
                item-value="id"
                item-text="title"
                label="Type"
                v-model="typeId"
                :items="types"
                :loading="isFetchLoading"
              >
                <template v-slot:item="{ item }">
                  <span>{{ item.title }} - {{ item.amount.title }}</span>
                </template>
                <template v-slot:selection="{ item }">
                  <span>{{ item.title }} - {{ item.amount.title }}</span>
                </template>
              </v-select>
            </v-col>
          </v-row>
        </v-col>
      </v-row>
      <v-row>
        <v-col lg="6" cols="12">
          <v-row align="center">
            <v-col cols="3">
              <v-subheader>Account</v-subheader>
            </v-col>
            <v-col cols="9">
              <div class="d-flex align-center">
                <accounts-autocomplete
                  fetch-value
                  :rules="generalRule"
                  :disabled="isEdit"
                  :loading="isFetchLoading"
                  label="Account"
                  return-object
                  v-model="transaction.account"
                />
                <v-btn
                  @click="openAccountWindow"
                  icon
                  v-if="isEdit && !isFetchLoading"
                >
                  <v-icon>mdi-login</v-icon>
                </v-btn>
              </div>
            </v-col>
          </v-row>

          <v-row align="center">
            <v-col cols="3">
              <v-subheader>Service</v-subheader>
            </v-col>
            <v-col cols="9">
              <v-autocomplete
                :filter="defaultFilterObject"
                label="Service"
                item-value="uuid"
                item-text="title"
                clearable
                return-object
                v-model="transaction.service"
                :items="services"
                :loading="isServicesLoading"
              />
            </v-col>
          </v-row>

          <v-row v-if="transaction.service && instances?.length" align="center">
            <v-col cols="3">
              <v-subheader>Instances</v-subheader>
            </v-col>
            <v-col cols="9">
              <v-autocomplete
                :filter="defaultFilterObject"
                multiple
                label="Instances"
                item-value="uuid"
                item-text="title"
                v-model="transaction.meta.instances"
                :items="instances"
              />
            </v-col>
          </v-row>

          <v-row align="center">
            <v-col cols="3">
              <v-subheader>Amount</v-subheader>
            </v-col>
            <v-col cols="9">
              <v-text-field
                type="number"
                label="Amount"
                :suffix="accountCurrency?.code"
                v-model.number="transaction.total"
                :rules="amountRule"
              />
            </v-col>
          </v-row>

          <v-row align="center" v-if="isTransaction">
            <v-col cols="3">
              <v-subheader>Date</v-subheader>
            </v-col>
            <v-col cols="4" v-for="type of [date, time]" :key="type.title">
              <v-menu
                offset-y
                min-width="auto"
                transition="scale-transition"
                v-model="type.visible"
                :ref="`menu${type.title}`"
                :close-on-content-click="false"
                :return-value.sync="type.value"
              >
                <template v-slot:activator="{ on, attrs }">
                  <v-text-field
                    readonly
                    v-model="type.value"
                    v-bind="attrs"
                    v-on="on"
                    :label="type.title"
                  />
                </template>
                <v-date-picker
                  no-title
                  scrollable
                  v-model="type.value"
                  v-if="date.visible"
                >
                  <v-spacer />
                  <v-btn text color="primary" @click="type.visible = false">
                    Cancel
                  </v-btn>
                  <v-btn
                    text
                    color="primary"
                    @click="$refs.menuDate[0].save(type.value)"
                  >
                    OK
                  </v-btn>
                </v-date-picker>
                <v-time-picker
                  use-seconds
                  format="24hr"
                  v-if="time.visible"
                  v-model="type.value"
                  @click:second="$refs.menuTime[0].save(type.value)"
                />
              </v-menu>
            </v-col>
          </v-row>

          <v-row v-if="!isAdminNoteHide" class="mx-5">
            <v-textarea
              no-resize
              label="Admin note"
              v-model="transaction.meta.note"
            ></v-textarea>
          </v-row>

          <v-row class="mx-5">
            <v-textarea
              no-resize
              label="Items descriptions"
              v-model="transaction.meta.description"
            ></v-textarea>
          </v-row>
          <v-expansion-panels v-if="history.length" class="mt-4">
            <v-expansion-panel>
              <v-expansion-panel-header color="background-light">
                <span class="text-h6">History</span>
                <template v-slot:actions>
                  <v-icon x-large> $expand </v-icon>
                </template>
              </v-expansion-panel-header>
              <v-expansion-panel-content color="background-light">
                <invoice-items-table
                  sort-by="date"
                  :account="transaction.account"
                  :items="historyItems"
                  readonly
                  show-date
                />
              </v-expansion-panel-content>
            </v-expansion-panel>
          </v-expansion-panels>
          <v-expansion-panels class="mt-4">
            <v-expansion-panel>
              <v-expansion-panel-header color="background-light">
                <span class="text-h6">Meta</span>
                <template v-slot:actions>
                  <v-icon x-large> $expand </v-icon>
                </template>
              </v-expansion-panel-header>
              <v-expansion-panel-content color="background-light">
                <json-editor
                  :json="transaction.meta"
                  @changeValue="(data) => (transaction.meta = data)"
                />
              </v-expansion-panel-content>
            </v-expansion-panel>
          </v-expansion-panels>
        </v-col>
      </v-row>

      <v-row justify="start" class="mb-4">
        <v-btn
          class="mx-3"
          color="background-light"
          :loading="isLoading"
          @click="tryToSend(false)"
        >
          Publish
        </v-btn>
      </v-row>
    </v-form>
  </div>
</template>

<script>
import api from "@/api.js";
import snackbar from "@/mixins/snackbar.js";
import JsonEditor from "@/components/JsonEditor.vue";
import { defaultFilterObject } from "@/functions";
import InvoiceItemsTable from "@/components/invoiceItemsTable.vue";
import AccountsAutocomplete from "@/components/ui/accountsAutocomplete.vue";

export default {
  components: { AccountsAutocomplete, InvoiceItemsTable, JsonEditor },
  name: "transactionsCreate-view",
  mixins: [snackbar],
  data: () => ({
    transaction: {
      priority: 1,
      account: {},
      service: "",
      total: "",
      exec: 0,
      meta: { instances: [], description: "", transactionType: "", items: [] },
    },
    isFetchLoading: false,
    namespace: {},
    date: {
      title: "Date",
      value: "",
      visible: false,
    },
    time: {
      title: "Time",
      value: "",
      visible: false,
    },
    generalRule: [(v) => !!v || "This field is required!"],
    isValid: false,
    isLoading: false,

    types: [
      {
        id: 3,
        value: "transaction",
        title: "Transaction",
        amount: { title: "Top-up", value: false },
      },
      {
        id: 4,
        value: "transaction",
        title: "Transaction",
        amount: { title: "Debit", value: true },
      },
      {
        id: 5,
        value: "transaction",
        title: "Transaction",
        amount: { title: "Set account balance", value: null },
      },
    ],
    typeId: 4,
    isEdit: false,
    history: [],
    services: [],
    isServicesLoading: false,
  }),
  methods: {
    defaultFilterObject,
    setTransactionType() {
      let amount = "";

      if (this.fullType.amount.value === null) {
        amount = "account-balance";
      } else {
        amount = this.fullType.amount.value ? "payment" : "top-up";
      }
      this.transaction.meta.transactionType = [
        this.fullType.value,
        amount,
      ].join(" ");
    },
    async tryToSend(withEmail = false) {
      if (!this.isValid) {
        this.$refs.form.validate();

        this.showSnackbarError({
          message: "Validation failed!",
        });
        return;
      }
      this.isLoading = true;
      this.refreshData();

      try {
        if (this.isEdit) {
          await this.editTransaction(withEmail);
        } else {
          await this.createTransaction(withEmail);
        }

        if (
          this.$route.query.account &&
          this.transaction.account.uuid === this.$route.query.account
        ) {
          this.$router.push({
            name: "Account",
            params: { accountId: this.$route.query.account },
          });
        } else {
          this.$router.push({ name: "Transactions" });
        }
      } catch (err) {
        this.showSnackbarError({
          message: err,
        });
      } finally {
        this.isLoading = false;
      }
    },
    async editTransaction() {
      this.showSnackbarSuccess({
        message: "Transaction edited successfully",
      });
    },
    async createTransaction() {
      let total = this.transaction.total;
      const amountType = this.fullType.amount.value;
      if (amountType === null) {
        const balance = this.transaction.account.balance || 0;
        const difference = Math.abs(total - balance);
        total = (balance > total ? +difference : -difference).toFixed(2);
      } else {
        total = Math.abs(total);
        total = amountType ? total : -total;
      }

      await api.transactions.create({
        ...this.transaction,
        account: this.transaction.account.uuid,
        total,
        currency: this.accountCurrency,
      });

      this.showSnackbarSuccess({
        message: "Transaction created successfully",
      });
    },
    refreshData() {
      this.transaction.service = this.transaction.service?.uuid;

      this.transaction.exec = this.exec;
      this.transaction.total *= 1;
    },
    resetDate() {
      this.date.value = null;
      this.time.value = null;
    },
    initDate() {
      const date = new Date();
      const day = date.getDate();
      const month = date.getMonth() + 1;
      const year = date.getFullYear();
      const time = date.toString().split(" ")[4];

      this.date.value = `${year}-${
        month.toString().length < 2 ? "0" + month : month
      }-${day.toString().length < 2 ? "0" + day : day}`;
      this.time.value = `${time}`;
    },
    openAccountWindow() {
      return window.open(
        "/admin/accounts/" + this.transaction.account.uuid,
        "_blanc"
      );
    },
  },
  async created() {
    if (this.$route.query.account) {
      this.transaction.account.uuid = this.$route.query.account;
    }

    this.initDate();
    this.setTransactionType();

    if (this.$route.params.uuid) {
      this.isFetchLoading = true;

      try {
        const { pool } = await api.transactions.get(this.$route.params.uuid);
        this.transaction = pool[0];
        this.isEdit = true;
        this.typeId =
          this.types.find(
            (t) =>
              this.transaction.meta.transactionType.startsWith(t.value) &&
              t.amount.value === !!this.transaction.total
          )?.id || 2;

        const { records = [] } = await api.reports.list({
          filters: { base: [this.$route.params.uuid] },
        });
        this.history = records;
      } catch (err) {
        this.$router.back();
      } finally {
        this.isFetchLoading = false;
      }
    }
  },
  computed: {
    whmcsApi() {
      return this.$store.getters["settings/whmcsApi"];
    },
    defaultCurrency() {
      return this.$store.getters["currencies/default"];
    },
    accountCurrency() {
      return this.transaction.account?.currency || this.defaultCurrency;
    },
    instances() {
      if (!this.transaction.service) {
        return;
      }

      const instances = [];

      this.transaction.service?.instancesGroups.forEach((ig) => {
        ig.instances.forEach((i) =>
          instances.push({ uuid: i.uuid, title: i.title })
        );
      });

      return instances;
    },
    exec() {
      return new Date(`${this.date.value}T${this.time.value}`).getTime() / 1000;
    },
    isTransaction() {
      return this.fullType.value === "transaction";
    },
    amountRule() {
      return [
        (v) =>
          (this.fullType.amount.value === null ? v === 0 || !!v : !!v) ||
          "This field is required!",
      ];
    },
    fullType() {
      return this.types.find((t) => t.id === this.typeId);
    },
    isAdminNoteHide() {
      return this.isTransaction;
    },
    historyItems() {
      const items = [];
      this.history.forEach((historyItem) => {
        historyItem.meta.items?.forEach((i, index) =>
          items.push({
            ...i,
            title: "Item " + (index + 1),
            date: new Date(historyItem.exec * 1000).toLocaleString(),
          })
        );
      });

      return items;
    },
  },
  watch: {
    "transaction.service"() {
      this.transaction.meta.instances = [];
    },
    typeId() {
      this.setTransactionType();
      if (this.isTransaction) {
        this.transaction.meta.items = undefined;
        this.transaction.meta.note = undefined;
        this.initDate();
      }
    },
    async "transaction.account"() {
      if (!this.transaction.account?.uuid) {
        this.services = [];
        this.namespace = null;
        return;
      }

      try {
        const { pool: namespaces } = await this.$store.dispatch(
          "namespaces/fetch",
          {
            filters: { account: this.transaction.account.uuid },
          }
        );
        this.namespace = namespaces[0];

        this.isServicesLoading = true;
        const { pool: services } = await this.$store.dispatch(
          "services/fetch",
          {
            filters: { account: this.transaction.account.uuid },
          }
        );
        this.services = services;
      } catch (err) {
        this.showSnackbarError({
          message: err,
        });
      } finally {
        this.isServicesLoading = false;
      }
    },
  },
};
</script>

<style scoped>
.page__title {
  color: var(--v-primary-base);
  font-weight: 400;
  font-size: 32px;
  font-family: "Quicksand", sans-serif;
  line-height: 1em;
  margin-bottom: 10px;
}
</style>
