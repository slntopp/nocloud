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
    </v-row>

    <v-progress-linear indeterminate class="pt-1" v-if="chartLoading" />
<!--    <template v-else-if="series.length < 1">-->
<!--      <v-subheader v-if="balance.values.length > 1"> Balance: </v-subheader>-->
<!--      <v-sparkline-->
<!--        color="primary"-->
<!--        height="25vh"-->
<!--        line-width="1"-->
<!--        label-size="4"-->
<!--        :labels="balance.labels"-->
<!--        :value="balance.values"-->
<!--      />-->
<!--    </template>-->
<!--    <apexcharts-->
<!--      v-else-->
<!--      type="line"-->
<!--      height="250"-->
<!--      :options="chartOptions"-->
<!--      :series="series"-->
<!--    />-->

    <reports-table
      table-name="transaction-table"
      :filters="filters"
      :duration="duration"
      @input:unique="setUniques"
      :select-record="selectTransaction"
    />
  </div>
</template>

<script>
import snackbar from "@/mixins/snackbar.js";
import search from "@/mixins/search.js";

import { mapGetters } from "vuex";
import reportsTable from "@/components/reports_table.vue";
export default {
  name: "transactions-view",
  components: { reportsTable },
  mixins: [snackbar, search("transactions")],
  data: () => ({
    types: [],
    resources: [],
    products: [],
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
    duration: { to: null, from: null },
  }),
  methods: {
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
      // this.chartLoading = true;
      value.forEach(({ total, item, exec }) => {
        const name = item.slice(0, 8);
        const data = { data: [{ x: exec * 1000, y: total }], name, item };
        const i = this.series.findIndex((item) => item.name === name);
        if (i !== -1) {
          this.series[i].data.push({ x: exec * 1000, y: total });
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
    setUniques({ resources, products, types }) {
      this.resources = resources;
      this.types = types;
      this.products = products;
    },
    fetchData() {
      this.$store.dispatch("accounts/fetch");
      this.$store.dispatch("services/fetch", { showDeleted: true });
      this.$store.dispatch("namespaces/fetch");
    },
  },
  mounted() {
    this.fetchData();
    this.$store.commit("reloadBtn/setCallback", {
      event: async () => {
        this.fetchData();
      },
    });
  },
  computed: {
    ...mapGetters("transactions", ["count", "page", "isLoading", "all"]),
    ...mapGetters("appSearch", ["filter"]),
    transactions() {
      return this.all;
    },
    user() {
      return this.$store.getters["auth/userdata"];
    },
    namespaces() {
      return this.$store.getters["namespaces/all"];
    },
    accounts() {
      return this.$store.getters["accounts/all"];
    },
    services() {
      return this.$store.getters["services/all"];
    },
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
    servicesByAccount() {
      const namespaces = this.namespaces
        .filter((n) => this.filter.account?.includes(n.access.namespace))
        .map((n) => n.uuid);

      return this.services.filter((s) => {
        return namespaces.includes(s.access.namespace);
      });
    },
    instances() {
      const instances = [];
      this.servicesByAccount.forEach((s) => {
        s.instancesGroups.forEach((ig) => {
          ig.instances.forEach((i) =>
            instances.push({ title: i.title, uuid: i.uuid })
          );
        });
      });

      return instances;
    },
    balance() {
      const dates = [];
      let labels = [`0 ${this.defaultCurrency}`];
      let values = [0];
      let balance = 0;
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
    searchFields() {
      return [
        { key: "type", type: "select", items: this.types, title: "Type" },
        {
          key: "instance",
          type: "select",
          item: { value: "uuid", title: "title" },
          items: this.instances,
          title: "Instances",
        },
        {
          key: "account",
          type: "select",
          item: { value: "uuid", title: "title" },
          items: this.accounts,
          title: "Accounts",
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
    chartLoading() {
      setTimeout(this.setListenerToLegend);
    },
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
