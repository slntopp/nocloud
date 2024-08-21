<template>
  <div class="pa-4">
    <v-row align="start">
      <v-col cols="1">
        <v-btn
          class="ma-2"
          color="background-light"
          :to="{ name: 'Transactions create' }"
        >
          Create
        </v-btn>
      </v-col>
      <v-col cols="1">
        <v-btn
          class="ma-2"
          color="background-light"
          @click="downloadTransactionsReport"
          :loading="isReportLoading"
          >Report</v-btn
        >
      </v-col>
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
            return {
              name:
                accounts.find((account) => account.value?.uuid == key).value
                  ?.title || key,
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
        resource: this.filter.resource,
        product: this.filter.product,
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
        { key: "exec", type: "date", title: "Exec" },
        { key: "start", type: "date", title: "Start" },
        { key: "end", type: "date", title: "End" },
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
