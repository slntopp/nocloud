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
      class="d-inline-block ml-2"
      v-model="accountTitle"
      :items="accountsTitles"
    />

    <v-progress-linear indeterminate class="pt-1" v-if="chartLoading" />
    <template v-else-if="series.length < 1">
      <v-subheader v-if="balance.values.length > 1">
        Balance:
      </v-subheader>
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

    <nocloud-table
      class="mt-4"
      sort-by="proc"
      :items="transactions"
      :headers="headers"
      :loading="isLoading"
      :sort-desc="true"
      :footer-error="fetchError"
      @input="selectTransaction"
    >
      <template v-slot:[`item.account`]="{ item }">
        {{ account(item.account) }}
      </template>

      <template v-slot:[`item.service`]="{ item, index }">
        <template v-if="item.service">
          <router-link
            :to="{ name: 'Service', params: { serviceId: item.service } }"
          >
            {{ service(item.service) }}
          </router-link>

          <v-icon
            class="ml-2"
            v-if="!visibleItems.includes(index)"
            @click="visibleItems.push(index)"
          >
            mdi-eye-outline
          </v-icon>
          <template v-else>
            ({{ hashTrim(item.service) }})
            <v-btn icon @click="addToClipboard(item.service, index)">
              <v-icon v-if="copyed === index"> mdi-check </v-icon>
              <v-icon v-else> mdi-content-copy </v-icon>
            </v-btn>
          </template>
        </template>
        <template v-else>-</template>
      </template>

      <template v-slot:[`item.total`]="{ item }">
        <balance :value="-item.total" />
      </template>
      <template v-slot:[`item.proc`]="{ item }">
        {{ date(item.proc) }}
      </template>
    </nocloud-table>

    <v-snackbar
      v-model="snackbar.visibility"
      :timeout="snackbar.timeout"
      :color="snackbar.color"
    >
      {{ snackbar.message }}
      <template v-if="snackbar.route && Object.keys(snackbar.route).length > 0">
        <router-link :to="snackbar.route"> Look up. </router-link>
      </template>

      <template v-slot:action="{ attrs }">
        <v-btn
          :color="snackbar.buttonColor"
          text
          v-bind="attrs"
          @click="snackbar.visibility = false"
        >
          Close
        </v-btn>
      </template>
    </v-snackbar>
  </div>
</template>

<script>
import api from '@/api.js';
import snackbar from '@/mixins/snackbar.js';
import nocloudTable from '@/components/table.vue';
import balance from '@/components/balance.vue';
import apexcharts from 'vue-apexcharts';

export default {
  name: 'transactions-view',
  components: { nocloudTable, balance, apexcharts },
  mixins: [snackbar],
  data: () => ({
    headers: [
      { text: 'Account ', value: 'account' },
      { text: 'Service ', value: 'service' },
      { text: 'Amount ', value: 'total' },
      { text: 'Date ', value: 'proc' }
    ],
    accountTitle: '',
    visibleItems: [],
    selected: [],
    copyed: -1,
    fetchError: '',

    series: [],
    chartLoading: false,
    chartOptions: {
      chart: { height: 250, type: 'line' },
      dataLabels: { enabled: false },
      stroke: { curve: 'smooth' },
      xaxis: { type: 'datetime', categories: [] },
      tooltip: { x: { format: 'dd.MM.yy HH:mm' } },
      theme: { palette: 'palette10', mode: 'dark' },
      legend: { showForSingleSeries: true }
    }
  }),
  methods: {
    account(uuid) {
      return this.accounts.find((acc) =>
        acc.uuid === uuid
      )?.title;
    },
    service(uuid) {
      const service = this.$store.getters['services/all']
        .find((serv) => serv.uuid === uuid);

      return service?.title;
    },
    date(timestamp, bool) {
      const date = new Date(timestamp * 1000);
      const time = date.toUTCString().split(' ')[4];
      
      const day = date.getUTCDate();
      const month = date.getUTCMonth() + 1;
      const year = date.toUTCString().split(' ')[3];

      if (bool) return `${day}-${month}-${year}T${time}Z`;
      return `${day}.${month}.${year} ${time}`;
    },
    hashTrim(hash) {
      if (hash) return ` ${hash.slice(0, 12)}... `;
      else return ' XXXXXXXX... ';
    },
    addToClipboard(text, index) {
      if (navigator?.clipboard) {
        navigator.clipboard
          .writeText(text)
          .then(() => {
            this.copyed = index;
          })
          .catch((err) => {
            this.showSnackbarError({
              message: err
            });
          });
      } else {
        this.showSnackbarError({
          message: 'Clipboard is not supported!'
        });
      }
    },
    getTransactions() {
      const { title } = this.$store.getters['auth/userdata'];
      const accounts = this.accounts.map((acc) => acc.uuid);

      this.accountTitle = title;
      this.$store.dispatch('services/fetch')
      this.$store.dispatch('transactions/fetch', accounts)
        .then(() => {
          this.fetchError = '';
        })
        .catch((err) => {
          console.error(err);

          this.fetchError = 'Can\'t reach the server';
          if (err.response) {
            this.fetchError += `: [ERROR]: ${err.response.data.message}`;
          } else {
            this.fetchError += `: [ERROR]: ${err.toJSON().message}`;
          }
        });
    },
    setTransactions(dates, labels, values) {
      const min = Math.min(...dates);
      let counter = 1;

      for (let i = 1; i < dates.length; i++) {
        const curr = Math.round(dates[i] / min);
        const spaces = (curr < 10) ? curr : 9;

        if (spaces < 2) continue;
        const newValues = values.splice(i + counter);
        const newLabels = labels.splice(i + counter);
        const diff = (newValues[0] - values.at(-1)) / spaces;

        for (let j = 0; j < spaces - 1; j++) {
          const prev = values[i + j + counter - 1];

          values[i + j + counter] = prev + diff;
          labels[i + j + counter] = ' ';
        }

        counter += spaces - 1;
        values = values.concat(newValues);
        labels = labels.concat(newLabels);
      }

      return [labels, values];
    },
    selectTransaction(value) {
      if (value.length < 1) return;

      this.series = [];
      this.chartLoading = true;
      this.chartOptions.xaxis.categories = [];

      value.forEach(({ uuid, service }) => {
        api.transactions.get(uuid)
          .then(({ pool }) => {
            pool.forEach((el) => {
              const name = el.instance.slice(0, 8);
              const data = { data: [el.total], name, service };
              const i = this.series.findIndex(
                (item) => item.name === name
              );

              if (i !== -1) {
                this.series[i].data.push(el.total);
              } else {
                this.series.push(data);
              }

              this.chartOptions.xaxis.categories
                .push(this.date(el.exec, true));
            })
          })
          .catch(() => {
            this.showSnackbar({
              message: 'Records not found',
              buttonColor: 'white',
              color: 'blue darken-3'
            });
          })
          .finally(() => {
            this.series = [...this.series];
            this.selected = value;
            this.chartLoading = false;
          });
      });
    },
    setListenerToLegend() {
      const legend = document.querySelectorAll('.apexcharts-legend-text');

      legend.forEach((el) => {
        el.addEventListener('click', (e) => {
          const { service } = this.series.find((item) =>
            item.name === e.target.innerText
          );

          this.$router.push({
            name: 'Service',
            params: { serviceId: service }
          });
        });
      });
    }
  },
  created() {
    this.getTransactions();
  },
  mounted() {
    this.$store.commit("reloadBtn/setCallback", {
      type: "transactions/fetch",
      params: this.accounts.map((acc) => acc.uuid)
    });
  },
  computed: {
    transactions() {
      const transactions = this.$store.getters['transactions/all'];
      const account = this.accounts.find((acc) =>
        acc.title === this.accountTitle
      );

      if (this.accountTitle === 'all') {
        return transactions;
      }

      return transactions.filter((item) =>
        item.account === account?.uuid
      );
    },
    isLoading() {
      return this.$store.getters['transactions/isLoading'];
    },
    accounts() {
      return this.$store.getters['accounts/all'];
    },
    accountsTitles() {
      return [...this.accounts.map((acc) => acc.title), 'all'];
    },
    balance() {
      const dates = [];
      let labels = ['0 NCU'];
      let values = [0];
      let balance = 0;

      if (this.accountTitle === 'all') {
        return { labels, values };
      }

      this.transactions?.forEach((el, i, arr) => {
        values.push(balance -= el.total);
        labels.push(`${balance} NCU`);
        dates.push(el.proc - arr[i - 1]?.proc ||
          arr[i + 1].proc - el.proc);
      });

      [labels, values] = this.setTransactions(dates, labels, values);

      const amount = values.length - 12;
      return {
        labels: (amount > 0) ? labels.slice(amount) : labels,
        values: (amount > 0) ? values.slice(amount) : values
      };
    }
  },
  watch: {
    transactions() {
      this.fetchError = '';
    },
    accounts() {
      this.getTransactions();
    },
    chartLoading() {
      setTimeout(this.setListenerToLegend);
    }
  }
}
</script>

<style>
.apexcharts-svg {
  background: none !important;
}
</style>
