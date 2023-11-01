<template>
  <div class="pa-4">
    <v-row align="start">
      <v-col cols="1">
        <v-btn
          class="mr-2"
          color="background-light"
          :to="{ name: 'Transactions create' }"
        >
          Create
        </v-btn>
      </v-col>
      <v-col>
        <date-picker dense label="from" v-model="duration.from" />
      </v-col>
      <v-col>
        <date-picker dense label="to" v-model="duration.to" />
      </v-col>
      <v-col>
        <v-autocomplete
          :filter="defaultFilterObject"
          label="Types"
          dense
          v-model="selectedTypes"
          multiple
          :items="types"
        />
      </v-col>
      <v-col>
        <v-autocomplete
          dense
          :filter="defaultFilterObject"
          label="Accounts"
          item-text="title"
          item-value="uuid"
          v-model="selectedAccounts"
          multiple
          :items="accounts"
        />
      </v-col>
      <v-col>
        <v-autocomplete
          :filter="defaultFilterObject"
          dense
          label="Instances"
          item-text="title"
          item-value="uuid"
          v-model="selectedInstances"
          multiple
          :items="instances"
        />
      </v-col>
    </v-row>

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

    <reports-table
      v-if="!isInitLoading"
      table-name="transaction-table"
      :filters="filters"
      :duration="duration"
      @input:unique="types = $event.transactionType"
      :select-record="selectTransaction"
    />
  </div>
</template>

<script>
import snackbar from "@/mixins/snackbar.js";
import search from "@/mixins/search.js";
import apexcharts from "vue-apexcharts";

import { defaultFilterObject } from "@/functions";
import { mapGetters } from "vuex";
import reportsTable from "@/components/reports_table.vue";
import DatePicker from "@/components/ui/datePicker.vue";
export default {
  name: "transactions-view",
  components: { DatePicker, reportsTable, apexcharts },
  mixins: [snackbar, search],
  data: () => ({
    selectedAccounts: [],
    selectedInstances: [],
    selectedTypes: [],
    types: [],
    series: [],
    chartLoading: false,
    isInitLoading: true,
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
    defaultFilterObject,
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
    fetchData() {
      this.$store.dispatch("accounts/fetch");
      this.$store.dispatch("services/fetch");
      this.$store.dispatch("namespaces/fetch");
    },
  },
  created() {
    if (this.$route.query.account) {
      this.selectedAccounts = [this.$route.query.account];
    } else {
      this.selectedAccounts = [];
    }
    this.isInitLoading = false;
  },
  mounted() {
    this.fetchData();
    this.$store.commit("reloadBtn/setCallback", {
      event: async () => {
        this.isInitLoading = true;
        this.fetchData();
        setTimeout(() => (this.isInitLoading = false), 0);
      },
    });
  },
  computed: {
    ...mapGetters("transactions", ["count", "page", "isLoading", "all"]),
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
      return {
        account: this.selectedAccounts.length
          ? this.selectedAccounts
          : undefined,
        instance: this.selectedInstances.length
          ? this.selectedInstances
          : undefined,
        transactionType: this.selectedTypes.length
          ? this.selectedTypes
          : undefined,
      };
    },
    servicesByAccount() {
      const namespaces = this.namespaces
        .filter((n) => this.selectedAccounts.includes(n.access.namespace))
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
      if (!this.selectedAccounts) {
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
  },
  watch: {
    chartLoading() {
      setTimeout(this.setListenerToLegend);
    },
    selectedAccounts: {
      handler() {
        this.selectedInstances = this.selectedInstances.filter((si) =>
          this.instances.find((i) => i.uuid === si)
        );
      },
      deep: true,
    },
  },
};
</script>
<style>
.apexcharts-svg {
  background: none !important;
}
</style>
