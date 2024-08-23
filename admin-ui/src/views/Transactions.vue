<template>
  <div class="pa-4">
    <v-row align="start" class="ml-2 mb-4">
      <v-btn
        class="ml-2"
        color="background-light"
        :to="{ name: 'Transactions create' }"
      >
        Create
      </v-btn>
      <v-btn
        class="ml-2"
        color="background-light"
        @click="downloadTransactionsReport"
        :loading="isReportLoading"
        :disabled="isPlansLoading"
        >Report</v-btn
      >
    </v-row>

    <v-progress-linear indeterminate class="pt-1" v-if="chartLoading" />

    <reports-table
      table-name="transaction-table"
      :filters="filters"
      :duration="duration"
      @input:unique="setUniques"
    />
  </div>
</template>

<script>
import snackbar from "@/mixins/snackbar.js";
import search from "@/mixins/search.js";
import XlsxService from "@/services/XlsxService";
import api from "@/api";
import { mapGetters } from "vuex";
import reportsTable from "@/components/reports_table.vue";
import AccountsAutocomplete from "@/components/ui/accountsAutocomplete.vue";
import { getTodayFullDate } from "../functions";

export default {
  name: "transactions-view",
  components: { reportsTable },
  mixins: [snackbar, search({ name: "transactions" })],
  data: () => ({
    types: [],
    resources: [],
    products: [],
    series: [],
    chartLoading: false,
    duration: { to: null, from: null },

    isReportLoading: false,
  }),
  mounted() {
    this.$store.dispatch("plans/fetch");
  },
  methods: {
    setUniques({ resources, products, types }) {
      this.resources = resources;
      this.types = types;
      this.products = products;
    },
    async downloadTransactionsReport() {
      try {
        this.isReportLoading = true;

        const transactions = await api.reports.list({ filters: this.filters });

        const resultData = {};

        transactions.records.forEach((transaction) => {
          const productOrResource = transaction.product || transaction.resource;
          if (!productOrResource) {
            return;
          }

          let data = {};
          if (resultData[transaction.account]) {
            data = resultData[transaction.account];
          }

          if (!data[productOrResource]) {
            data[productOrResource] = 0;
          }
          data[productOrResource] += Math.abs(+transaction.total);

          resultData[transaction.account] = data;
        });

        const accounts = await Promise.allSettled(
          Object.keys(resultData).map((key) => api.accounts.get(key))
        );

        return XlsxService.downloadXlsx(
          "transactions_report_" + getTodayFullDate(),
          Object.entries(resultData).map(([key, value]) => {
            const account = accounts.find(
              (account) => account.value?.uuid == key
            ).value;

            Object.keys(value).forEach((key) => {
              value[key] = `${value[key].toFixed(2)} ${
                account.currency || this.defaultCurrency
              }`;
            });

            return {
              name: `${account?.title || key} (${key})`,
              headers: Object.keys(value).map((key) => ({
                key,
                title: key.replaceAll("_", " "),
              })),
              items: [value],
            };
          })
        );
      } finally {
        this.isReportLoading = false;
      }
    },
  },
  computed: {
    ...mapGetters("transactions", ["count", "page", "isLoading"]),
    ...mapGetters("plans", { plans: "all", isPlansLoading: "isLoading" }),
    ...mapGetters("appSearch", ["filter"]),
    filters() {
      const total = {};
      if (this.filter.total?.to) {
        total.to = +this.filter.total.to;
      }
      if (this.filter.total?.from) {
        total.from = +this.filter.total.from;
      }

      const dates = {};
      const dateKeys = ["exec", "start", "end", "payment_date"];
      dateKeys.forEach((key) => {
        if (!this.filter[key]) {
          return;
        }
        dates[key] = {};

        if (this.filter[key][0]) {
          dates[key].from = new Date(this.filter[key][0]).getTime() / 1000;
        }
        if (this.filter[key][1]) {
          dates[key].to = new Date(this.filter[key][1]).getTime() / 1000;
        }
      });

      const resource = [];
      const product = [];

      resource.push(...(this.filter.resource || []));
      product.push(...(this.filter.product || []));

      if (this.filter.plans?.length) {
        this.filter.plans.forEach((uuid) => {
          const plan = this.plans.find((p) => p.uuid === uuid);

          if (!plan) {
            return;
          }

          product.push(...Object.keys(plan.products || {}));
          resource.push(...(plan.resources.map((r) => r.key) || []));
        });
      }

      return {
        ...dates,
        account: this.filter.account?.length ? this.filter.account : undefined,
        instance: this.filter.instance?.length
          ? this.filter.instance
          : undefined,
        transactionType: this.filter.type?.length
          ? this.filter.type
          : undefined,
        total: Object.keys(total).length ? total : undefined,
        resource,
        product,
      };
    },
    defaultCurrency() {
      return this.$store.getters["currencies/default"];
    },
    transactionTypes() {
      return this.$store.getters["transactions/types"];
    },
    searchFields() {
      return [
        {
          key: "type",
          type: "select",
          items: this.transactionTypes,
          item: { value: "key", title: "title" },
          title: "Type",
        },
        {
          key: "account",
          type: "select",
          custom: true,
          component: AccountsAutocomplete,
          label: "Accounts",
          multiple: true,
          clearable: true,
          fetchValue: true,
        },
        {
          key: "plans",
          type: "select",
          items: this.plans.map(({ title, uuid }) => ({ title, uuid })),
          item: { value: "uuid", title: "title" },
          title: "Plan",
        },
        {
          key: "product",
          type: "select",
          items: this.products,
          title: "Product",
        },
        {
          key: "resource",
          type: "select",
          items: this.resources,
          title: "Resource",
        },
        { key: "exec", type: "date", title: "Executed date" },
        { key: "start", type: "date", title: "Start date" },
        { key: "end", type: "date", title: "End date" },
        { key: "payment_date", type: "date", title: "Payment date" },
        { key: "total", type: "number-range", title: "Total" },
      ];
    },
  },
  watch: {
    searchFields() {
      this.$store.commit("appSearch/setFields", this.searchFields);
    },
  },
};
</script>
<style>
.apexcharts-svg {
  background: none !important;
}
</style>
