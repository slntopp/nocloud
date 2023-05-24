<template>
  <div class="pa-4">
    <v-btn
      class="mr-2"
      color="background-light"
      :to="{ name: 'Transactions create' }"
    >
      Create
    </v-btn>

    <v-select
      label="Account"
      item-text="title"
      item-value="uuid"
      class="d-inline-block mr-2"
      v-model="accountId"
      :items="accounts"
    />
    <v-select
      label="Service"
      item-text="title"
      item-value="uuid"
      class="d-inline-block"
      v-model="serviceId"
      :items="servicesByAccount"
    />

    <v-progress-linear indeterminate class="pt-1" v-if="chartLoading" />
    <template v-else-if="series.length < 1">
      <v-subheader v-if="balance.values.length > 1"> Balance: </v-subheader>
      <v-sparkline
        color="primary"
        height="25vh"
        line-width="1"
        label-size="4"
        :labels="balance.labels"
        :value="balance.values"
      />
    </template>
    <apexcharts
      v-else
      type="line"
      height="250"
      :options="chartOptions"
      :series="series"
    />

    <transactions-table
      :page="page"
      @update:options="updateOptions"
      :count="count"
      :transactions="filtredTransactions"
      :selectTransaction="selectTransaction"
    />
  </div>
</template>

<script>
import snackbar from "@/mixins/snackbar.js";
import search from "@/mixins/search.js";
import apexcharts from "vue-apexcharts";
import transactionsTable from "@/components/transactions_table.vue";
import { filterArrayIncludes, filterArrayBy } from "@/functions";
import { mapGetters } from "vuex";
export default {
  name: "transactions-view",
  components: { apexcharts, transactionsTable },
  mixins: [snackbar, search],
  data: () => ({
    accountId: null,
    serviceId: null,
    series: [],
    chartLoading: false,
    chartOptions: {
      chart: { height: 250, type: "line" },
      dataLabels: { enabled: false },
      stroke: { curve: "smooth" },
      xaxis: { type: "datetime" },
      tooltip: { x: { format: "dd.MM.yy HH:mm" } },
      theme: { palette: "palette10", mode: "dark" },
      legend: { showForSingleSeries: true },
    },
  }),
  methods: {
    getTransactions() {
      const accounts = [];
      this.accounts.forEach((acc) => {
        if (acc.uuid) accounts.push(acc.uuid);
      });

      this.fetchTransactions();
    },
    setTransactions(dates, labels, values) {
      const min = Math.min(...dates);
      let counter = 1;
      for (let i = 1; i < dates.length; i++) {
        const curr = Math.round(dates[i] / min);
        const spaces = curr < 10 ? curr : 9;
        if (spaces < 2) continue;
        const newValues = values.splice(i + counter);
        const newLabels = labels.splice(i + counter);
        const diff = (newValues[0] - values.at(-1)) / spaces;
        for (let j = 0; j < spaces - 1; j++) {
          const prev = values[i + j + counter - 1];
          values[i + j + counter] = prev + diff;
          labels[i + j + counter] = " ";
        }
        counter += spaces - 1;
        values = values.concat(newValues);
        labels = labels.concat(newLabels);
      }
      return [labels, values];
    },
    selectTransaction(value) {
      this.series = [];
      this.chartLoading = true;
      value.forEach(({ total, service, proc }) => {
        const name = service.slice(0, 8);
        const data = { data: [{ x: proc * 1000, y: total }], name, service };
        const i = this.series.findIndex((item) => item.name === name);
        if (i !== -1) {
          this.series[i].data.push({ x: proc * 1000, y: total });
        } else {
          this.series.push(data);
        }
      });
      setTimeout(() => {
        this.chartLoading = false;
      }, 300);
      if (this.series.length < 1) {
        this.showSnackbar({
          message: "Records not found",
          buttonColor: "white",
          color: "blue darken-3",
        });
      }
    },
    setListenerToLegend() {
      const legend = document.querySelectorAll(".apexcharts-legend-text");
      legend.forEach((el) => {
        el.addEventListener("click", (e) => {
          const { service } = this.series.find(
            (item) => item.name === e.target.innerText
          );
          this.$router.push({
            name: "Service",
            params: { serviceId: service },
          });
        });
      });
    },
    updateOptions(options) {
      options.itemsPerPage =
        options.itemsPerPage === -1 ? 0 : options.itemsPerPage;
      this.$store
        .dispatch("transactions/changeFiltres", {
          options,
          data: this.transactionData,
        })
        .then(() => {
          this.fetchError = "";
        })
        .catch((err) => {
          console.error(err);
          this.fetchError = "Can't reach the server";
          if (err.response) {
            this.fetchError += `: [ERROR]: ${err.response.data.message}`;
          } else {
            this.fetchError += `: [ERROR]: ${err.toJSON().message}`;
          }
        });
    },
    initTransactions() {
      this.$store.dispatch("transactions/init", this.transactionData);
    },
    fetchTransactions() {
      this.$store.dispatch("transactions/fetch", this.transactionData);
    },
    changeTransactionData() {
      this.$store.commit("transactions/setPage", 1);
      this.$store.commit("transactions/setFilter", { field: "", sort: "" });
      this.initTransactions();
      this.fetchTransactions();
    },
  },
  mounted() {
    if (this.$route.query.account) {
      this.accountId = this.$route.query.account;
    } else {
      this.accountId = this.user.uuid || null;
    }

    const accounts = [];
    if (this.accounts.length < 2) {
      this.$store.dispatch("accounts/fetch");
    }
    if (this.services.length < 2) {
      this.$store.dispatch("services/fetch");
    }
    if (this.namespaces.length < 2) {
      this.$store.dispatch("namespaces/fetch");
    }

    this.accounts.forEach((acc) => {
      if (acc.uuid) accounts.push(acc.uuid);
    });

    this.$store.commit("reloadBtn/setCallback", {
      type: "transactions/init",
      params: this.transactionData,
    });
  },
  computed: {
    ...mapGetters("transactions", ["count", "page", "isLoading", "all"]),
    transactions() {
      return this.all;
    },
    filtredTransactions() {
      if (this.searchParam) {
        const byUuid = filterArrayIncludes(this.transactions, {
          keys: ["uuid", "service"],
          value: this.searchParam,
        });
        const byTotal = filterArrayBy(this.transactions, {
          key: "total",
          value: +this.searchParam,
        });
        return [...new Set([...byUuid, ...byTotal])];
      }
      return this.transactions;
    },
    user() {
      return this.$store.getters["auth/userdata"];
    },
    namespaces() {
      return this.$store.getters["namespaces/all"];
    },
    accounts() {
      const accounts = this.$store.getters["accounts/all"];
      return [{ title: "all", uuid: null }].concat(accounts);
    },
    services() {
      return this.$store.getters["services/all"];
    },
    servicesByAccount() {
      let filtredServices = null;
      if (this.accountId) {
        const namespaceId = this.namespaces.find(
          (n) => this.accountId === n.access.namespace
        )?.uuid;
        filtredServices = this.services.filter((s) => {
          return s.access.namespace === namespaceId;
        });
      } else {
        filtredServices = this.services;
      }

      return [{ title: "all", uuid: null }].concat(
        filtredServices.map((el) => ({
          title: `${el.title} (${el.uuid.slice(0, 8)})`,
          uuid: el.uuid,
        }))
      );
    },
    balance() {
      const dates = [];
      let labels = [`0 ${this.defaultCurrency}`];
      let values = [0];
      let balance = 0;
      if (!this.accountId) {
        return { labels, values };
      }
      this.transactions?.forEach((el, i, arr) => {
        values.push((balance -= el.total));
        labels.push(`${balance.toFixed(2)} ${this.defaultCurrency}`);
        dates.push(
          el.proc - arr[i - 1]?.proc || arr[i + 1]?.proc - el.proc || el.proc
        );
      });
      [labels, values] = this.setTransactions(dates, labels, values);
      const amount = values.length - 12;
      return {
        labels: amount > 0 ? labels.slice(amount) : labels,
        values: amount > 0 ? values.slice(amount) : values,
      };
    },
    defaultCurrency() {
      return this.$store.getters["currencies/default"];
    },
    searchParam() {
      return this.$store.getters["appSearch/param"];
    },
    transactionData() {
      const data = {};
      if (this.accountId) {
        data.account = this.accountId;
      }
      if (this.accountId) {
        data.service = this.serviceId;
      }
      return data;
    },
  },
  watch: {
    chartLoading() {
      setTimeout(this.setListenerToLegend);
    },
    user() {
      this.accountId = this.user.uuid;
    },
    accountId() {
      if (this.serviceId === null) {
        this.serviceId = null;
        this.changeTransactionData();
      } else {
        this.serviceId = null;
      }
    },
    serviceId() {
      this.changeTransactionData();
    },
    accounts() {
      if (!this.all.length && !this.isLoading) {
        this.getTransactions();
      }
    },
  },
};
</script>
<style>
.apexcharts-svg {
  background: none !important;
}
</style>
